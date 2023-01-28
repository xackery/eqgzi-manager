package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	Version string
)

var (
	mu               sync.RWMutex
	currentPath      string
	currentZone      = "chess"
	zones            = []string{}
	zoneCombo        *widget.Select
	currentZoneLabel *widget.Label
	newZonePopup     *widget.PopUp
)

func main() {
	if Version == "" {
		Version = "EXPERIMENTAL"
	}
	log.Println("initialized", Version)
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("eqgzi-manager v%s", Version))

	var err error

	currentPath, err = os.Getwd()
	if err != nil {
		log.Fatalf("wd invalid: %s", err)
		return
	}

	currentPath = `C:\src\eqp\client\zones`

	convertButton := widget.NewButtonWithIcon("convert", theme.ConfirmIcon(), convert)
	openButton := widget.NewButtonWithIcon("open in blender", theme.FolderOpenIcon(), open)

	refreshZones()

	mu.Lock()
	zoneCombo = widget.NewSelect(zones, func(value string) {
		mu.Lock()
		currentZone = value
		currentZoneLabel.SetText(fmt.Sprintf("current zone: %s", currentZone))
		mu.Unlock()
		log.Println("Select set to", value)
	})
	mu.Unlock()

	zoneRefreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), refreshZones)

	newZoneButton := widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
		newZonePopup.Show()
	})

	newZonePopup = widget.NewModalPopUp(container.NewVBox(widget.NewButtonWithIcon("", theme.CancelIcon(), func() { newZonePopup.Hide() }), widget.NewLabel("Test")), w.Canvas())

	currentZoneLabel = widget.NewLabel(fmt.Sprintf("current zone: %s", currentZone))
	w.SetContent(container.NewVBox(container.NewHBox(newZoneButton, widget.NewLabel("zone: "), zoneCombo, zoneRefreshButton), currentZoneLabel, convertButton, openButton))
	w.ShowAndRun()
}

func refreshZones() {
	mu.Lock()
	defer mu.Unlock()
	zones = []string{}

	entries, err := os.ReadDir(`C:\src\eqp\client\zones`)
	if err != nil {
		log.Println("failed to read dir:", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		zones = append(zones, filepath.Base(entry.Name()))
	}

	if zoneCombo != nil {
		zoneCombo.Options = zones
	}
	log.Println("refreshed zone list")
}

func convert() {
	fmt.Println("converting")

	cmd := exec.Command(fmt.Sprintf("%s/%s/convert.bat", currentPath, currentZone))
	cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, currentZone)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s;%s;%s;%s", os.Getenv("PATH"), `d:\games\eq\tools`, `C:\Program Files\Blender Foundation\Blender 2.93`, `C:\src\eqgzi\out`),
		`EQPATH=c:\games\eq\everquestparty\`,
	}
	fmt.Println(cmd.Env)
	err := cmd.Run()
	if err != nil {
		fmt.Println("run failed:", err)
	}
}

func open() {

}
