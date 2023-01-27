package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	Version string
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello World")

	log.Println("initialized", Version)
	progress := widget.NewProgressBar()
	gatherButton := widget.NewButton("Gather "+Version, func() {
		log.Println("gather Enabled")
	})

	go func() {
		i := 0.0
		isPositive := true
		progress.TextFormatter = func() string {
			return ""
		}
		for {
			time.Sleep(time.Millisecond * 50)
			progress.SetValue(i)

			if i >= 1.0 {
				isPositive = false
			}
			if i <= 0 {
				isPositive = true
			}
			if isPositive {
				i += 0.01
			}
			if !isPositive {
				i -= 0.01
			}
		}
	}()
	w.SetContent(container.NewVBox(widget.NewLabel("Hello World!"), progress, gatherButton))
	w.ShowAndRun()
}
