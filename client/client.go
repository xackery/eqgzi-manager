package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/xackery/eqgzi-manager/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Client struct {
	mu                  sync.RWMutex
	currentPath         string
	canvas              fyne.CanvasObject
	zoneCombo           *widget.Select
	blenderPathInput    *widget.Entry
	everQuestPathInput  *widget.Entry
	newZonePopup        *widget.PopUp
	newZoneName         *widget.Entry
	newZoneSaveButton   *widget.Button
	newZoneCancelButton *widget.Button
	progressBar         *widget.ProgressBar
	window              fyne.Window
	cfg                 *config.Config
	statusLabel         *widget.Label
	blenderOpenButton   *widget.Button
	folderOpenButton    *widget.Button
	convertButton       *widget.Button
	exportEQGCheck      *widget.Check
	downloadEQGZIButton *widget.Button
}

func New(window fyne.Window) (*Client, error) {
	var err error
	c := &Client{
		window: window,
	}

	c.cfg, err = config.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("config.new: %w", err)
	}

	c.currentPath, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("wd invalid: %w", err)
	}

	//c.currentPath = `C:\src\eqp\client\zones`

	c.statusLabel = widget.NewLabel("")
	c.statusLabel.Alignment = fyne.TextAlignCenter
	c.convertButton = widget.NewButtonWithIcon("TODO convert", theme.ConfirmIcon(), c.onConvertButton)
	c.blenderOpenButton = widget.NewButtonWithIcon("TODO in blender", theme.NewThemedResource(blenderIcon), c.onBlenderOpen)
	c.folderOpenButton = widget.NewButtonWithIcon("TODO in folder", theme.FolderOpenIcon(), c.onFolderOpen)

	c.downloadEQGZIButton = widget.NewButtonWithIcon("Download EQGZI", theme.FolderOpenIcon(), c.onDownloadEQGZIButton)
	c.exportEQGCheck = widget.NewCheck("Copy .eqg to EQ Path", c.onExportEQGCheck)

	zones := c.zoneRefresh()

	c.zoneCombo = widget.NewSelect(zones, c.onZoneSelect)

	zoneRefreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), c.onZoneRefresh)

	newZoneButton := widget.NewButtonWithIcon("New Zone", theme.FolderNewIcon(), func() {
		c.newZonePopup.Show()
	})

	c.newZoneSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onNewZoneSaveButton)
	c.newZoneCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onNewZoneCancelButton)

	c.newZoneName = widget.NewEntry()
	c.newZonePopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Create a new zone"),
			c.newZoneName,
			container.NewHBox(
				c.newZoneSaveButton,
				c.newZoneCancelButton,
			),
		),
		c.window.Canvas(),
	)

	c.blenderPathInput = widget.NewEntry()
	if c.cfg.BlenderPath != "" {
		c.blenderPathInput.SetText(c.cfg.BlenderPath)
	}

	c.everQuestPathInput = widget.NewEntry()
	if c.cfg.EQPath != "" {
		c.everQuestPathInput.SetText(c.cfg.EQPath)
	}

	c.progressBar = widget.NewProgressBar()
	c.progressBar.Hide()

	if len(zones) > 0 && c.cfg.LastZone == "" {
		c.cfg.LastZone = zones[0]
	}
	isValidZone := false
	for _, zone := range zones {
		if zone == c.cfg.LastZone {
			isValidZone = true
			break
		}
	}
	if !isValidZone {
		if len(zones) > 0 {
			c.cfg.LastZone = zones[0]
		}
	}

	c.zoneCombo.SetSelected(c.cfg.LastZone)

	c.canvas = container.NewVBox(
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("Blender Path:"),
			c.blenderPathInput,
		),
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("EQ Path:"),
			c.everQuestPathInput,
		),
		widget.NewLabel(""),
		newZoneButton,
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("zone: "),
			container.NewHBox(c.zoneCombo, zoneRefreshButton),
		),
		container.NewVBox(
			c.folderOpenButton,
			c.blenderOpenButton,
			c.convertButton,
		),
		c.exportEQGCheck,
		c.progressBar,
		c.statusLabel,
	)
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

	entries, err := os.ReadDir(fmt.Sprintf("%s/zones/", currentPath))
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

	c.statusLabel.SetText("Refreshed zones")
	return zones
}

func (c *Client) onConvertButton() {
	fmt.Println("converting")
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	isEQCopy := c.cfg.IsEQCopy
	eqPath := c.cfg.EQPath
	c.mu.RUnlock()

	cmd := exec.Command(fmt.Sprintf("%s/%s/convert.bat", currentPath, zone))
	cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, zone)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		fmt.Sprintf(`PATH=%s;%s;%s\tools;%s`, os.Getenv("PATH"), `C:\Program Files\Blender Foundation\Blender 2.93`, currentPath, `C:\src\eqgzi\out`),
		fmt.Sprintf(`EQPATH=%s`, eqPath),
		fmt.Sprintf(`EQGZI=%s\tools\`, currentPath),
		fmt.Sprintf(`ZONE=%s`, zone),
	}
	err := cmd.Run()
	if err != nil {
		log.Println("run convert.bat failed:", err)
		return
	}

	if !isEQCopy {
		return
	}

	cmd = exec.Command(fmt.Sprintf("%s/%s/copy.bat", currentPath, zone))
	cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, zone)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s;%s;%s;%s", os.Getenv("PATH"), `d:\games\eq\tools`, `C:\Program Files\Blender Foundation\Blender 2.93`, `C:\src\eqgzi\out`),
		`EQPATH=c:\games\eq\everquestparty\`,
	}
	err = cmd.Run()
	if err != nil {
		log.Println("run copy.bat failed:", err)
		return
	}

}

func (c *Client) onBlenderOpen() {
	fmt.Println("opening")
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	blenderPath := c.cfg.BlenderPath
	c.mu.RUnlock()

	cmd := exec.Command(blenderPath, fmt.Sprintf("%s/zones/%s/%s.blend", currentPath, zone, zone))
	//cmd.Dir = fmt.Sprintf("%s/", blenderPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{}
	err := cmd.Run()
	if err != nil {
		log.Println("run failed:", err)
	}
}

func (c *Client) onFolderOpen() {
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	c.mu.RUnlock()

	exePath := "explorer.exe"
	if runtime.GOOS == "darwin" {
		exePath = "open"
	}

	cmd := exec.Command(exePath, fmt.Sprintf("%s/zones/%s", currentPath, zone))
	//cmd.Dir = fmt.Sprintf("%s/", blenderPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{}
	err := cmd.Run()
	if err != nil {
		log.Println("run failed:", err)
	}
}

func (c *Client) onZoneSelect(value string) {
	c.mu.Lock()
	c.cfg.LastZone = value
	err := c.cfg.Save()
	if err != nil {
		log.Println("save failed:", err)
	}
	c.blenderOpenButton.SetText(fmt.Sprintf("Open %s in Blender", c.cfg.LastZone))
	c.convertButton.SetText(fmt.Sprintf("Convert %s to .eqg", c.cfg.LastZone))
	c.folderOpenButton.SetText(fmt.Sprintf("Open %s folder", c.cfg.LastZone))
	c.mu.Unlock()
	log.Println("zone changed to", value)
}

func (c *Client) onExportEQGCheck(value bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg.IsEQCopy = value
	err := c.cfg.Save()
	if err != nil {
		c.statusLabel.SetText(fmt.Sprintf("failed save: %s", err))
		return
	}
}

func (c *Client) onDownloadEQGZIButton() {
	c.statusLabel.Hide()
	c.progressBar.Show()
	defer func() {
		c.statusLabel.Show()
		c.progressBar.Hide()
	}()
}

func (c *Client) onNewZoneSaveButton() {
	if c.newZoneName.Text == "" {
		c.newZoneName.SetValidationError(fmt.Errorf("cannot be empty"))
		return
	}
	newZone := strings.ToLower(strings.TrimSpace(c.newZoneName.Text))
	_, err := os.Stat(fmt.Sprintf("zones/%s", newZone))
	if err == os.ErrNotExist {
		c.newZoneName.SetValidationError(fmt.Errorf("already exists"))
		return
	}

	err = os.Mkdir(fmt.Sprintf("zones/%s/", newZone), os.ModePerm)
	if err != nil {
		c.newZoneName.SetValidationError(fmt.Errorf("mkdir: %w", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/convert.bat", newZone), convertText.Content(), os.ModePerm)
	if err != nil {
		c.newZoneName.SetValidationError(fmt.Errorf("create convert.bat: %w", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_eq.bat", newZone), copyEQText.Content(), os.ModePerm)
	if err != nil {
		c.newZoneName.SetValidationError(fmt.Errorf("create copy_eq.bat: %w", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_server.bat", newZone), copyServerText.Content(), os.ModePerm)
	if err != nil {
		c.newZoneName.SetValidationError(fmt.Errorf("create copy_server.bat: %w", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/%s.blend", newZone, newZone), baseBlend.Content(), os.ModePerm)
	if err != nil {
		c.newZoneName.SetValidationError(fmt.Errorf("create %s.blend: %w", newZone, err))
		return
	}

	c.onZoneRefresh()
	c.zoneCombo.SetSelected(newZone)
	c.newZonePopup.Hide()
}

func (c *Client) onNewZoneCancelButton() {
	c.newZonePopup.Hide()
}
