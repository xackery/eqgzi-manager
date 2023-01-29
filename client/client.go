package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/xackery/eqgzi-manager/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Client struct {
	mu                    sync.RWMutex
	currentPath           string
	canvas                fyne.CanvasObject
	mainCanvas            fyne.CanvasObject
	downloadCanvas        fyne.CanvasObject
	zoneCombo             *widget.Select
	blenderPathInput      *widget.Entry
	newZoneButton         *widget.Button
	newZonePopup          *widget.PopUp
	newZoneName           *widget.Entry
	newZoneSaveButton     *widget.Button
	newZoneCancelButton   *widget.Button
	popupStatus           *widget.Label
	exportServerCheck     *widget.Check
	setServerButton       *widget.Button
	setServerPopup        *widget.PopUp
	setServerName         *widget.Entry
	setServerSaveButton   *widget.Button
	setServerCancelButton *widget.Button
	labelServer           *widget.Label
	exportEQGCheck        *widget.Check
	setEQButton           *widget.Button
	setEQPopup            *widget.PopUp
	setEQName             *widget.Entry
	setEQSaveButton       *widget.Button
	setEQCancelButton     *widget.Button
	labelEQ               *widget.Label
	progressBar           *widget.ProgressBar
	window                fyne.Window
	cfg                   *config.Config
	statusLabel           *widget.Label
	blenderOpenButton     *widget.Button
	folderOpenButton      *widget.Button
	eqgziOpenButton       *widget.Button
	convertButton         *widget.Button
	downloadEQGZIButton   *widget.Button
	blenderDetectButton   *widget.Button
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
	c.statusLabel.Wrapping = fyne.TextWrapBreak
	c.statusLabel.Alignment = fyne.TextAlignCenter

	c.popupStatus = widget.NewLabel("")
	c.popupStatus.Wrapping = fyne.TextWrapBreak
	c.popupStatus.Alignment = fyne.TextAlignCenter
	c.newZoneInit()
	c.newSetServerInit()
	c.newSetEQInit()

	c.convertButton = widget.NewButtonWithIcon("Create zone.eqg", theme.NewThemedResource(eqIcon), c.onConvertButton)
	c.blenderOpenButton = widget.NewButtonWithIcon("Open zone in blender", theme.NewThemedResource(blenderIcon), c.onBlenderOpen)
	c.folderOpenButton = widget.NewButtonWithIcon("Open zone folder", theme.FolderOpenIcon(), c.onFolderOpen)
	c.eqgziOpenButton = widget.NewButtonWithIcon("Debug zone in eqgzi-gui", theme.QuestionIcon(), c.onEqgziOpenButton)
	c.downloadEQGZIButton = widget.NewButtonWithIcon("Download EQGZI", theme.DownloadIcon(), c.onDownloadEQGZIButton)
	c.blenderPathInput = widget.NewEntry()
	if c.cfg.BlenderPath != "" {
		c.blenderPathInput.SetText(c.cfg.BlenderPath)
	}

	c.blenderDetectButton = widget.NewButtonWithIcon("Detect", theme.SearchIcon(), c.onBlenderDetectButton)
	if c.cfg.BlenderPath == "" {
		c.onBlenderDetectButton()
	}

	c.exportEQGCheck = widget.NewCheck("Copy .eqg to EverQuest", c.onExportEQGCheck)
	c.exportEQGCheck.Checked = c.cfg.IsEQCopy
	if c.cfg.IsEQCopy {
		c.setEQButton.Show()
	} else {
		c.setEQButton.Hide()
	}
	c.labelEQ = widget.NewLabel(c.cfg.EQPath)

	c.exportServerCheck = widget.NewCheck("Copy nav meshes to Server", c.onExportServerCheck)
	c.exportServerCheck.Checked = c.cfg.IsServerCopy
	if c.cfg.IsServerCopy {
		c.setServerButton.Show()
	} else {
		c.setServerButton.Hide()
	}
	c.labelServer = widget.NewLabel(c.cfg.ServerPath)

	zones := c.zoneRefresh()

	c.zoneCombo = widget.NewSelect(zones, c.onZoneCombo)

	zoneRefreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), c.onZoneRefresh)

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
		} else {
			c.disableActions()
		}
	}

	c.zoneCombo.SetSelected(c.cfg.LastZone)

	c.mainCanvas = container.NewVBox(
		container.New(
			layout.NewFormLayout(),
			widget.NewLabel("Blender Path:"),
			container.NewHBox(
				c.blenderPathInput,
				c.blenderDetectButton,
			),
			/*widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
				dia := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {

				}, c.window)
				dia.SetFilter(storage.NewExtensionFileFilter(extensions []string) .FileFilter)
				storage.FileFilter
				//dia.SetFilter(filter storage.FileFilter)
				dia.Show()
			}),*/
		),
		widget.NewLabel(""),
		c.newZoneButton,
		container.NewCenter(
			container.NewHBox(
				widget.NewLabel("Zone: "),
				c.zoneCombo,
				zoneRefreshButton,
			),
		),
		container.NewVBox(
			c.folderOpenButton,
			c.blenderOpenButton,
			container.NewHBox(
				c.exportEQGCheck,
				c.setEQButton,
				c.labelEQ,
			),
			container.NewHBox(
				c.exportServerCheck,
				c.setServerButton,
				c.labelServer,
			),
			c.convertButton,
			c.eqgziOpenButton,
		),
		c.progressBar,
		c.statusLabel,
	)

	c.downloadCanvas = container.NewVBox(
		c.downloadEQGZIButton,
		c.progressBar,
		c.statusLabel,
	)

	_, err = os.Stat("tools/eqgzi.exe")
	if err != nil {
		c.canvas = c.downloadCanvas
	} else {
		c.canvas = c.mainCanvas
		c.window.Resize(fyne.NewSize(600, 600))
	}

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

	_, err := os.Stat("zones")
	if os.IsNotExist(err) {
		err = os.Mkdir("zones", os.ModePerm)
		if err != nil {
			c.logf("Failed to mkdir zone: %s", err)
			return zones
		}
	}

	entries, err := os.ReadDir(fmt.Sprintf("%s/zones/", currentPath))
	if err != nil {
		c.logf("Failed to read dir: %s", err)
		return zones
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		zones = append(zones, filepath.Base(entry.Name()))
	}

	return zones
}

func (c *Client) onBlenderOpen() {
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	blenderPath := c.cfg.BlenderPath
	c.mu.RUnlock()

	c.logf("Opening %s in Blender", zone)
	cmd := exec.Command(blenderPath+"blender.exe", fmt.Sprintf("%s/zones/%s/%s.blend", currentPath, zone, zone))
	//cmd.Dir = fmt.Sprintf("%s/", blenderPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{}
	err := cmd.Start()
	if err != nil {
		c.logf("Failed to run blender: %s", err)
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

	cmd := exec.Command(exePath, fmt.Sprintf(`%s\zones\%s`, currentPath, zone))
	//cmd.Dir = fmt.Sprintf("%s/", blenderPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{}
	err := cmd.Start()
	if err != nil {
		c.logf("Failed to open: %s", err)
		return
	}
	c.logf("Opened %s folder", zone)
}

func (c *Client) onZoneCombo(value string) {
	c.mu.Lock()
	c.cfg.LastZone = value
	err := c.cfg.Save()
	if err != nil {
		c.logf("Failed saving after zone select: %s", err)
		return
	}
	c.blenderOpenButton.SetText(fmt.Sprintf("Open %s in Blender", c.cfg.LastZone))
	c.convertButton.SetText(fmt.Sprintf("Create %s.eqg", c.cfg.LastZone))
	c.folderOpenButton.SetText(fmt.Sprintf("Open %s folder", c.cfg.LastZone))
	c.eqgziOpenButton.SetText(fmt.Sprintf("Debug %s in eqgzi-gui", c.cfg.LastZone))
	c.enableActions()
	c.mu.Unlock()
	c.logf("Focused on %s", value)
}

func (c *Client) onExportEQGCheck(value bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg.IsEQCopy = value
	if c.cfg.IsEQCopy {
		c.setEQButton.Show()
	} else {
		c.setEQButton.Hide()
	}
	err := c.cfg.Save()
	if err != nil {
		c.logf("Failed saving after eqgcheck: %s", err)
		return
	}
}

func (c *Client) onExportServerCheck(value bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg.IsServerCopy = value
	if c.cfg.IsServerCopy {
		c.setServerButton.Show()
	} else {
		c.setServerButton.Hide()
	}
	err := c.cfg.Save()
	if err != nil {
		c.logf("Failed saving after servercheck: %s", err)
		return
	}
}

func (c *Client) onEqgziOpenButton() {
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	c.mu.RUnlock()

	cmd := exec.Command(fmt.Sprintf("%s/tools/eqgzi-gui.exe", currentPath), fmt.Sprintf("%s/zones/%s/out/%s.eqg", currentPath, zone, zone))
	cmd.Dir = fmt.Sprintf("%s/tools/", currentPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		c.logf("Failed eqgzi-gui: %s", err)
		return
	}
}

func (c *Client) logf(format string, a ...interface{}) {
	text := fmt.Sprintf(format, a...)
	fmt.Println(text)
	c.statusLabel.SetText(text)
}

func (c *Client) disableActions() {
	c.blenderOpenButton.Disable()
	c.folderOpenButton.Disable()
	c.eqgziOpenButton.Disable()
	c.convertButton.Disable()
	c.exportEQGCheck.Disable()
	c.exportServerCheck.Disable()
}

func (c *Client) enableActions() {
	c.blenderOpenButton.Enable()
	c.folderOpenButton.Enable()
	c.eqgziOpenButton.Enable()
	c.convertButton.Enable()
	c.exportEQGCheck.Enable()
	c.exportServerCheck.Enable()
}
