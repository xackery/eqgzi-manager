package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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
	progress              float64
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
	navMeshEditButton     *widget.Button
	downloadButton        *widget.Button
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

	c.downloadButton = widget.NewButtonWithIcon("Download Update", theme.DownloadIcon(), c.onDownloadButton)

	c.convertButton = widget.NewButtonWithIcon("Create zone.eqg", theme.NewThemedResource(eqIcon), c.onConvertButton)
	c.blenderOpenButton = widget.NewButtonWithIcon("Open zone in blender", theme.NewThemedResource(blenderIcon), c.onBlenderOpen)
	c.folderOpenButton = widget.NewButtonWithIcon("Open zone folder", theme.FolderOpenIcon(), c.onFolderOpen)
	c.eqgziOpenButton = widget.NewButtonWithIcon("Debug zone in eqgzi-gui", theme.QuestionIcon(), c.onEqgziOpenButton)
	c.downloadEQGZIButton = widget.NewButtonWithIcon("Download EQGZI & Lantern", theme.DownloadIcon(), c.onDownloadEQGZIButton)
	c.navMeshEditButton = widget.NewButtonWithIcon("Edit Navmesh", theme.GridIcon(), c.onNavMeshEditButton)
	c.blenderPathInput = widget.NewEntry()
	if c.cfg.BlenderPath != "" {
		c.blenderPathInput.SetText(c.cfg.BlenderPath)
	}

	c.blenderDetectButton = widget.NewButtonWithIcon("Detect Blender Path", theme.SearchIcon(), c.onBlenderDetectButton)
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
		c.downloadButton,
		widget.NewLabel(""),
		container.NewVBox(
			container.New(
				layout.NewFormLayout(),
				widget.NewLabel("Blender Path:"),
				container.NewMax(c.blenderPathInput),
			),
			c.blenderDetectButton,
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
			c.navMeshEditButton,
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
		_, err = os.Stat("tools/LanternExtractor.exe")
		if err != nil {
			c.canvas = c.downloadCanvas
		} else {
			c.canvas = c.mainCanvas
			c.window.Resize(fyne.NewSize(600, 600))
		}
	}

	go c.loop()

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
	cmd := c.createCommand(false, blenderPath+"blender.exe", fmt.Sprintf("%s/zones/%s/%s.blend", currentPath, zone, zone))

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

	cmd := c.createCommand(false, exePath, fmt.Sprintf(`%s\zones\%s`, currentPath, zone))
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

	path := fmt.Sprintf("%s/tools/gui/settings.lua", currentPath)
	settings, err := os.ReadFile(path)
	if err != nil {
		c.logf("Failed to read settings.lua: %s", err)
		return
	}

	w, err := os.Create(path)
	if err != nil {
		c.logf("Failed to create settings.lua: %s", err)
		return
	}
	defer w.Close()

	buf := bufio.NewReader(bytes.NewBuffer(settings))
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			c.logf("Failed to read settings.lua: %s", err)
			return
		}

		if strings.Contains(line, `folder = "`) {
			line = fmt.Sprintf("	folder = \"%s\",\n", filepath.ToSlash(fmt.Sprintf("%s/zones/%s/out/%s.eqg", currentPath, zone, zone)))
		}

		_, err = w.WriteString(line)
		if err != nil {
			c.logf("Failed to write to settings.lua: %s", err)
			return
		}
	}

	cmd := c.createCommand(false, fmt.Sprintf("%s/tools/eqgzi-gui.exe", currentPath), fmt.Sprintf("%s/zones/%s/out/%s.eqg", currentPath, zone, zone))
	cmd.Dir = fmt.Sprintf("%s/tools/", currentPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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

func (c *Client) onNavMeshEditButton() {
	c.mu.RLock()
	currentPath := c.currentPath
	zone := c.cfg.LastZone
	c.mu.RUnlock()

	cmd := c.createCommand(false, fmt.Sprintf("%s/tools/map_edit/map_edit.exe", currentPath), zone)
	c.logf("running command: map_edit %s", zone)
	cmd.Dir = fmt.Sprintf("%s/tools/map_edit/", currentPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		c.logf("Failed map-edit: %s", err)
		return
	}
}

func (c *Client) addProgress(amount float64) float64 {
	c.progress += amount

	if c.progress > 1 {
		fmt.Printf("progress > 1: %0.2f\n", c.progress)
		c.progress = 1
	}
	return c.progress
}

func (c *Client) loop() {
	err := c.updateCheck()
	if err != nil {
		fmt.Println("Failed loop updateCheck:", err)
	}
	for {
		time.Sleep(24 * time.Hour)
		err := c.updateCheck()
		if err != nil {
			fmt.Println("Failed loop updateCheck:", err)
		}
	}
}

func (c *Client) updateCheck() error {
	err := c.updateCheckLantern()
	if err != nil {
		return fmt.Errorf("updateCheckLantern: %w", err)
	}
	return nil
}

func (c *Client) updateCheckLantern() error {
	c.mu.Lock()
	lanternVersion := c.cfg.LanternVersion
	c.mu.Unlock()

	gitReply := &gitReply{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/LanternEQ/LanternExtractor/releases/latest", nil)
	if err != nil {
		return fmt.Errorf("git request: %w", err)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do git request: %w", err)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(gitReply)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("decode git request: %w", err)
	}
	assetURL := ""

	zipName := fmt.Sprintf("LanternExtractor-%s.zip", gitReply.TagName)
	for _, asset := range gitReply.Assets {
		if asset.Name != zipName {
			continue
		}
		assetURL = asset.BrowserDownloadURL
	}
	if assetURL == "" {
		return fmt.Errorf("download eqgzi zip not found")
	}

	if gitReply.TagName == lanternVersion {
		return nil
	}
	return nil
}
