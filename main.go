package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/xackery/eqgzi-manager/client"
)

var (
	Version string
)

func main() {
	if Version == "" {
		Version = string(client.VersionText.Content())
	}
	log.Println("initializing", Version)

	a := app.New()

	w := a.NewWindow(fmt.Sprintf("eqgzi-manager v%s", Version))
	c, err := client.New(w)
	if err != nil {
		fmt.Println("client new:", err)
		os.Exit(1)
	}

	w.SetContent(c.GetContent())
	w.CenterOnScreen()
	w.ShowAndRun()
}
