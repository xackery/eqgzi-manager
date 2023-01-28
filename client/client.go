package client

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Client struct {
	mu           sync.RWMutex
	currentPath  string
	canvas       fyne.CanvasObject
	zone         string
	zoneCombo    *widget.Select
	zoneLabel    *widget.Label
	newZonePopup *widget.PopUp
	window       fyne.Window
}

func New(window fyne.Window) (*Client, error) {
	var err error
	c := &Client{
		zone:   "chess",
		window: window,
	}

	c.currentPath, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("wd invalid: %w", err)
	}

	c.currentPath = `C:\src\eqp\client\zones`

	convertButton := widget.NewButtonWithIcon("convert", theme.ConfirmIcon(), c.onConvert)
	openButton := widget.NewButtonWithIcon("open in blender", theme.FolderOpenIcon(), c.onOpen)

	zones := c.zoneRefresh()

	c.zoneCombo = widget.NewSelect(zones, c.onZoneSelect)

	zoneRefreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), c.onZoneRefresh)

	newZoneButton := widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
		c.newZonePopup.Show()
	})

	c.newZonePopup = widget.NewModalPopUp(container.NewVBox(widget.NewButtonWithIcon("", theme.CancelIcon(), func() { c.newZonePopup.Hide() }), widget.NewLabel("Test")), c.window.Canvas())

	c.zoneLabel = widget.NewLabel(fmt.Sprintf("current zone: %s", c.zone))

	c.canvas = container.NewVBox(container.NewHBox(newZoneButton, widget.NewLabel("zone: "), c.zoneCombo, zoneRefreshButton), c.zoneLabel, convertButton, openButton)
	return c, nil
}

func (c *Client) GetContent() fyne.CanvasObject {
	return c.canvas
}

func (c *Client) onZoneRefresh() {
	zones := c.zoneRefresh()
	c.mu.Lock()
	if c.zoneCombo != nil {
		c.zoneCombo.Options = zones
	}
	c.mu.Unlock()
}

func (c *Client) zoneRefresh() []string {
	zones := []string{}
	c.mu.RLock()
	currentPath := c.currentPath
	c.mu.RUnlock()

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		log.Println("failed to read dir:", err)
		return zones
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		zones = append(zones, filepath.Base(entry.Name()))
	}

	log.Println("refreshed zone list")
	return zones
}

func (c *Client) onConvert() {
	fmt.Println("converting")
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.zone
	c.mu.RUnlock()

	cmd := exec.Command(fmt.Sprintf("%s/%s/convert.bat", currentPath, zone))
	cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, zone)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s;%s;%s;%s", os.Getenv("PATH"), `d:\games\eq\tools`, `C:\Program Files\Blender Foundation\Blender 2.93`, `C:\src\eqgzi\out`),
		`EQPATH=c:\games\eq\everquestparty\`,
	}
	fmt.Println(cmd.Env)
	err := cmd.Run()
	if err != nil {
		log.Println("run failed:", err)
	}
}

func (c *Client) onOpen() {

}

func (c *Client) onZoneSelect(value string) {
	c.mu.Lock()
	c.zone = value
	c.zoneLabel.SetText(fmt.Sprintf("current zone: %s", c.zone))
	c.mu.Unlock()
	log.Println("zone changed to", value)
}
