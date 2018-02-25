package dropper

import (
	"github.com/MJKWoolnough/gopherjs/files"
	"honnef.co/go/js/dom"
)

func Initialise() (enter, leave chan struct{}, drop chan []files.File) {
	var window = dom.GetWindow()
	var document = window.Document().(dom.HTMLDocument)
	enter = make(chan struct{})
	leave = make(chan struct{})
	drop = make(chan []files.File)
	document.AddEventListener("drop", true, func(ev dom.Event) {
		ev.PreventDefault()
		dtf := ev.Underlying().Get("dataTransfer").Get("files")
		var out []files.File
		for i := 0; i < dtf.Length(); i++ {
			file := files.NewFile(&dom.File{Object: dtf.Index(i)})
			out = append(out, file)
		}
		select {
		case drop <- out:
			// great!
		default:
			// nothing was listening.
		}
	})
	document.AddEventListener("dragover", true, func(ev dom.Event) {
		ev.PreventDefault()
	})
	document.AddEventListener("dragenter", true, func(ev dom.Event) {
		ev.PreventDefault()
		select {
		case enter <- struct{}{}:
			// great!
		default:
			// nothing was listening.
		}
	})
	document.AddEventListener("dragleave", true, func(ev dom.Event) {
		ev.PreventDefault()
		select {
		case leave <- struct{}{}:
			// great!
		default:
			// nothing was listening.
		}
	})
	return
}
