package client

import (
	"context"
	"fmt"
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
	mu                     sync.RWMutex
	currentPath            string
	canvas                 fyne.CanvasObject
	mainCanvas             fyne.CanvasObject
	downloadCanvas         fyne.CanvasObject
	zoneCombo              *widget.Select
	blenderPathInput       *widget.Entry
	newZonePopup           *widget.PopUp
	newZoneName            *widget.Entry
	newZoneSaveButton      *widget.Button
	newZoneCancelButton    *widget.Button
	popupStatus            *widget.Label
	setServerPathButton    *widget.Button
	setServerPopup         *widget.PopUp
	setServerName          *widget.Entry
	setServerSaveButton    *widget.Button
	setServerCancelButton  *widget.Button
	setEverQuestPathButton *widget.Button
	progressBar            *widget.ProgressBar
	window                 fyne.Window
	cfg                    *config.Config
	statusLabel            *widget.Label
	blenderOpenButton      *widget.Button
	folderOpenButton       *widget.Button
	eqgziOpenButton        *widget.Button
	convertButton          *widget.Button
	exportEQGCheck         *widget.Check
	exportServerCheck      *widget.Check
	downloadEQGZIButton    *widget.Button
	blenderDetectButton    *widget.Button
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
	c.setEverQuestPathButton = widget.NewButtonWithIcon("Set EQ Path", theme.FolderNewIcon(), func() {
		c.setServerPopup.Show()
		c.window.Canvas().Focus(c.setServerName)
	})

	c.exportEQGCheck = widget.NewCheck("Copy .eqg to EverQuest", c.onExportEQGCheck)
	c.exportEQGCheck.Checked = c.cfg.IsEQCopy
	if c.cfg.IsEQCopy {
		c.setEverQuestPathButton.Show()
	} else {
		c.setEverQuestPathButton.Hide()
	}
	c.setServerPathButton = widget.NewButtonWithIcon("Set Server Path", theme.FolderNewIcon(), func() {
		c.setServerPopup.Show()
		c.window.Canvas().Focus(c.setServerName)
	})

	c.exportServerCheck = widget.NewCheck("Copy nav meshes to Server", c.onExportServerCheck)
	c.exportServerCheck.Checked = c.cfg.IsServerCopy
	if c.cfg.IsServerCopy {
		c.setServerPathButton.Show()
	} else {
		c.setServerPathButton.Hide()
	}

	zones := c.zoneRefresh()

	c.zoneCombo = widget.NewSelect(zones, c.onZoneCombo)

	zoneRefreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), c.onZoneRefresh)

	newZoneButton := widget.NewButtonWithIcon("Create New Zone", theme.FolderNewIcon(), func() {
		c.newZonePopup.Show()
		c.window.Canvas().Focus(c.newZoneName)
	})

	c.popupStatus = widget.NewLabel("")
	c.popupStatus.Wrapping = fyne.TextWrapBreak
	c.popupStatus.Alignment = fyne.TextAlignCenter

	c.newZoneSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onNewZoneSaveButton)
	c.newZoneCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onNewZoneCancelButton)

	c.newZoneName = widget.NewEntry()
	c.newZoneName.OnSubmitted = func(string) { c.onNewZoneSaveButton() }
	c.newZonePopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Create a new zone"),
			c.newZoneName,
			container.NewHBox(
				c.newZoneSaveButton,
				c.newZoneCancelButton,
			),
			c.popupStatus,
		),
		c.window.Canvas(),
	)

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
		newZoneButton,
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
				c.setEverQuestPathButton,
			),
			container.NewHBox(
				c.exportServerCheck,
				c.setServerPathButton,
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

func (c *Client) onConvertButton() {
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	isEQCopy := c.cfg.IsEQCopy
	eqPath := c.cfg.EQPath
	isServerCopy := c.cfg.IsServerCopy
	serverPath := c.cfg.ServerPath
	blenderPath := c.cfg.BlenderPath
	c.mu.RUnlock()
	c.logf("Converting %s", zone)

	env := []string{
		fmt.Sprintf(`PATH=%s;%s\tools;%s`, blenderPath, currentPath, `C:\src\eqgzi\out`),
		fmt.Sprintf(`EQPATH=%s`, eqPath),
		fmt.Sprintf(`EQGZI=%s\tools\`, currentPath),
		fmt.Sprintf(`ZONE=%s`, zone),
		fmt.Sprintf(`EQSERVERPATH=%s`, serverPath),
		fmt.Sprintf(`BLENDERPATH=%s`, blenderPath),
	}

	cmd := exec.Command(fmt.Sprintf("%s/zones/%s/convert.bat", currentPath, zone))
	cmd.Dir = fmt.Sprintf("%s/zones/%s/", currentPath, zone)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		c.logf("Failed to run convert.bat: %s", err)
		return
	}

	if isEQCopy {
		cmd = exec.Command(fmt.Sprintf("%s/%s/copy_eq.bat", currentPath, zone))
		cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, zone)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env
		err = cmd.Run()
		if err != nil {
			c.logf("Failed to run copy_eq.bat: %s", err)
			return
		}
	}
	if isServerCopy {
		cmd = exec.Command(fmt.Sprintf("%s/%s/copy_server.bat", currentPath, zone))
		cmd.Dir = fmt.Sprintf("%s/%s/", currentPath, zone)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env
		err = cmd.Run()
		if err != nil {
			c.logf("Failed to run copy_server.bat: %s", err)
			return
		}
	}

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
		c.setEverQuestPathButton.Show()
	} else {
		c.setEverQuestPathButton.Hide()
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
		c.setServerPathButton.Show()
	} else {
		c.setServerPathButton.Hide()
	}
	err := c.cfg.Save()
	if err != nil {
		c.logf("Failed saving after servercheck: %s", err)
		return
	}
}

func (c *Client) onNewZoneSaveButton() {
	if c.newZoneName.Text == "" {
		c.popupStatus.SetText("Failed: zone cannot be empty")
		return
	}

	newZone := strings.ToLower(strings.TrimSpace(c.newZoneName.Text))
	_, err := os.Stat(fmt.Sprintf("zones/%s", newZone))
	if err == os.ErrNotExist {
		c.popupStatus.SetText(fmt.Sprintf("Failed: zone %s already exists", newZone))
		return
	}

	err = os.Mkdir(fmt.Sprintf("zones/%s/", newZone), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating folder: %s", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/convert.bat", newZone), convertText.Content(), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating convert.bat: %s", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_eq.bat", newZone), copyEQText.Content(), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating copy_eq.bat: %s", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_server.bat", newZone), copyServerText.Content(), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating copy_server.bat: %s", err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/%s.blend", newZone, newZone), baseBlend.Content(), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating %s.blend: %s", newZone, err))
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/white.png", newZone), whitePng.Content(), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed creating white.png: %s", newZone, err))
		return
	}

	c.onZoneRefresh()
	c.zoneCombo.SetSelected(newZone)
	c.logf("Created zones/%s", newZone)
	c.newZonePopup.Hide()
}

func (c *Client) onNewZoneCancelButton() {
	c.logf("Cancelled new zone")
	c.newZonePopup.Hide()
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
