package dropper

import (
	"path/filepath"

	"sync"

	"github.com/MJKWoolnough/gopherjs/files"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

type File struct {
	files.File
	dir string
}

func (f File) Path() string {
	return filepath.Join(f.dir, f.Name())
}

func (f File) Reader() *files.FileReader {
	return files.NewFileReader(f.File)
}

func Initialise(target dom.EventTarget) (enter, leave chan struct{}, drop chan []File) {
	if target == nil {
		target = dom.GetWindow().Document()
	}
	enter = make(chan struct{})
	leave = make(chan struct{})
	drop = make(chan []File)
	var over bool
	target.AddEventListener("drop", true, func(ev dom.Event) {
		ev.PreventDefault()
		over = false
		items := ev.Underlying().Get("dataTransfer").Get("items")
		var out []File
		var wait sync.WaitGroup
		var processEntry func(string, *js.Object)
		processEntry = func(dir string, entry *js.Object) {
			if entry.Get("isFile").Bool() {
				wait.Add(1)
				entry.Call("file", func(f *js.Object) {
					file := File{
						File: files.NewFile(&dom.File{Object: f}),
						dir:  dir,
					}
					out = append(out, file)
					wait.Done()
				})
			} else {
				sub := filepath.Join(dir, entry.Get("name").String())
				wait.Add(1)
				entry.Call("createReader").Call("readEntries", func(entries []*js.Object) {
					for _, child := range entries {
						processEntry(sub, child)
					}
					wait.Done()
				})
			}
		}
		for i := 0; i < items.Length(); i++ {
			item := items.Index(i)
			var entry *js.Object
			if item.Get("getAsEntry").Bool() {
				entry = item.Call("getAsEntry")
			} else if item.Get("webkitGetAsEntry").Bool() {
				entry = item.Call("webkitGetAsEntry")
			}
			processEntry("/", entry)
		}
		go func() {
			wait.Wait()
			select {
			case drop <- out:
				// great!
			default:
				// nothing was listening.
			}
		}()
	})
	target.AddEventListener("dragover", true, func(ev dom.Event) {
		ev.PreventDefault()
		if !over {
			over = true
			select {
			case enter <- struct{}{}:
				// great!
			default:
				// nothing was listening.
			}
		}
	})
	target.AddEventListener("dragenter", true, func(ev dom.Event) {
		ev.PreventDefault()
		if !over {
			over = true
			select {
			case enter <- struct{}{}:
				// great!
			default:
				// nothing was listening.
			}
		}
	})
	target.AddEventListener("dragleave", true, func(ev dom.Event) {
		ev.PreventDefault()
		if over {
			over = false
			select {
			case leave <- struct{}{}:
				// great!
			default:
				// nothing was listening.
			}
		}
	})
	return
}
