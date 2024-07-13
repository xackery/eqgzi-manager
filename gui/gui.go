//go:build windows
// +build windows

package gui

import (
	"context"
	"fmt"
	"strings"

	"github.com/xackery/eqgzi-manager/config"
	"github.com/xackery/wlk/cpl"
	"github.com/xackery/wlk/walk"
)

type Gui struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	mw                    *walk.MainWindow
	progress              *walk.ProgressBar
	log                   *walk.TextEdit
	newHandler            []func()
	openHandler           []func(path string, file string) error
	savePFSHandler        []func(path string) error
	saveContentHandler    []func(path string, file string) error
	saveAllContentHandler []func(path string) error
	downloadEQGZIHandler  []func()
	refreshHandler        []func()
	statusBar             *walk.StatusBarItem
	blenderText           *walk.TextEdit
	downloadEQGZI         *walk.PushButton
	zones                 []string
	zonesCombo            *walk.ComboBox
	eqPathCheck           *walk.CheckBox
	eqPathChange          *walk.PushButton
	eqPathLabel           *walk.Label
	eqPath                *walk.LineEdit
	serverPathCheck       *walk.CheckBox
	serverPathChange      *walk.PushButton
	serverPathLabel       *walk.Label
	blenderOpen           *walk.PushButton
	folderOpen            *walk.PushButton
	eqgziOpen             *walk.PushButton
	convert               *walk.PushButton
	exportEQGCheck        *walk.CheckBox
	exportServerCheck     *walk.CheckBox
}

var (
	gui *Gui
)

// NewMainWindow creates a new main window
func NewMainWindow(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, version string) error {
	walk.SetDarkModeAllowed(true)
	gui = &Gui{
		ctx:    ctx,
		cancel: cancel,
	}

	var err error

	cmw := cpl.MainWindow{
		Title:    "quail-gui v" + version,
		MinSize:  cpl.Size{Width: 300, Height: 300},
		Layout:   cpl.VBox{},
		Visible:  false,
		Name:     "quail-gui",
		AssignTo: &gui.mw,
		Children: []cpl.Widget{
			cpl.HSplitter{Children: []cpl.Widget{
				cpl.Composite{Children: []cpl.Widget{
					cpl.Label{Text: "Blender Path:"},
					cpl.TextEdit{
						AssignTo: &gui.blenderText,
						Text:     cfg.BlenderPath,
					},
					cpl.PushButton{
						Text: "Detect",
						OnClicked: func() {
							fmt.Println("detect clicked")
						},
					},
				}},
				cpl.Composite{Children: []cpl.Widget{
					cpl.CheckBox{
						AssignTo: &gui.eqPathCheck,
						Text:     "EQ Path:",
						Checked:  cfg.EQPath != "",
					},
					cpl.PushButton{
						Text:     "Change",
						AssignTo: &gui.eqPathChange,
					},
					cpl.Label{
						AssignTo: &gui.eqPathLabel,
						Text:     cfg.EQPath,
						Enabled:  cfg.EQPath != "",
					},
				}},
				cpl.PushButton{
					Text:     "Download EQGZI",
					AssignTo: &gui.downloadEQGZI,
				},
				cpl.LineEdit{
					Text:     "EQ Path:",
					AssignTo: &gui.eqPath,
				},
				cpl.ProgressBar{
					AssignTo: &gui.progress,
					Visible:  false,
					MaxValue: 100,
					MinValue: 0,
				},
			}},
		},
		StatusBarItems: []cpl.StatusBarItem{
			{
				AssignTo: &gui.statusBar,
				Text:     "Ready",
				OnClicked: func() {
					fmt.Println("status bar clicked")
				},
			},
		},
	}
	err = cmw.Create()
	if err != nil {
		return fmt.Errorf("create main window: %w", err)
	}
	return nil
}

func Run() int {
	if gui == nil {
		return 1
	}

	gui.mw.SetVisible(true)
	return gui.mw.Run()
}

func SubscribeClose(fn func(cancelled *bool, reason byte)) {
	if gui == nil {
		return
	}
	gui.mw.Closing().Attach(fn)
}

// Logf logs a message to the gui
func Logf(format string, a ...interface{}) {
	if gui == nil {
		return
	}

	line := fmt.Sprintf(format, a...)
	if strings.Contains(line, "\n") {
		line = line[0:strings.Index(line, "\n")]
	}
	gui.statusBar.SetText(line)

	//convert \n to \r\n
	format = strings.ReplaceAll(format, "\n", "\r\n")
	gui.log.AppendText(fmt.Sprintf(format, a...))

}

func LogClear() {
	if gui == nil {
		return
	}
	gui.log.SetText("")
}

func SetMaxProgress(value int) {
	if gui == nil {
		return
	}
	gui.progress.SetRange(0, value)
}

func SetProgress(value int) {
	if gui == nil {
		return
	}
	gui.progress.SetValue(value)
	gui.progress.SetVisible(value > 0)
}

func AddProgress(value int) {
	if gui == nil {
		return
	}
	gui.progress.SetValue(gui.progress.Value() + value)
	gui.progress.SetVisible(gui.progress.Value() > 0)
}

func MessageBox(title string, message string, isError bool) {
	if gui == nil {
		return
	}
	// convert style to msgboxstyle
	icon := walk.MsgBoxIconInformation
	if isError {
		icon = walk.MsgBoxIconError
	}
	walk.MsgBox(gui.mw, title, message, icon)
}

func MessageBoxYesNo(title string, message string) bool {
	if gui == nil {
		return false
	}
	// convert style to msgboxstyle
	icon := walk.MsgBoxIconInformation
	result := walk.MsgBox(gui.mw, title, message, icon|walk.MsgBoxYesNo)
	return result == walk.DlgCmdYes
}

func MessageBoxf(title string, format string, a ...interface{}) {
	if gui == nil {
		return
	}
	// convert style to msgboxstyle
	icon := walk.MsgBoxIconInformation
	walk.MsgBox(gui.mw, title, fmt.Sprintf(format, a...), icon)
}

func SetTitle(title string) {
	if gui == nil {
		return
	}
	gui.mw.SetTitle(title)
}

func Close() {
	if gui == nil {
		return
	}
	gui.mw.Close()
}

func SubscribeNew(fn func()) {
	if gui == nil {
		return
	}
	gui.newHandler = append(gui.newHandler, fn)
}

func SubscribeOpen(fn func(path string, file string) error) {
	if gui == nil {
		return
	}
	gui.openHandler = append(gui.openHandler, fn)
}

func SubscribeRefresh(fn func()) {
	if gui == nil {
		return
	}
	gui.refreshHandler = append(gui.refreshHandler, fn)
}

func ShowOpen(title string, filter string, initialDirPath string) (string, error) {
	if gui == nil {
		return "", fmt.Errorf("gui not initialized")
	}
	dialog := walk.FileDialog{
		Title:          title,
		Filter:         filter,
		InitialDirPath: initialDirPath,
	}
	ok, err := dialog.ShowOpen(gui.mw)
	if err != nil {
		return "", fmt.Errorf("show open: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("show open: cancelled")
	}
	return dialog.FilePath, nil
}

func SubscribeSavePFS(fn func(path string) error) {
	if gui == nil {
		return
	}
	gui.savePFSHandler = append(gui.savePFSHandler, fn)
}

func SubscribeSaveContent(fn func(path string, file string) error) {
	if gui == nil {
		return
	}
	gui.saveContentHandler = append(gui.saveContentHandler, fn)
}

func ShowSave(title string, fileName string, initialDirPath string) (string, error) {
	if gui == nil {
		return "", fmt.Errorf("gui not initialized")
	}
	dialog := walk.FileDialog{
		Title:          title,
		FilePath:       fileName,
		InitialDirPath: initialDirPath,
	}
	ok, err := dialog.ShowSave(gui.mw)
	if err != nil {
		return "", fmt.Errorf("show save: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("show save: cancelled")
	}
	return dialog.FilePath, nil
}

func SubscribeSaveAllContent(fn func(path string) error) {
	if gui == nil {
		return
	}
	gui.saveAllContentHandler = append(gui.saveAllContentHandler, fn)
}

func SubscribeDownloadEQGZI(fn func()) {
	if gui == nil {
		return
	}
	gui.downloadEQGZIHandler = append(gui.downloadEQGZIHandler, fn)
}

func ShowDirSave(title string, filter string, initialDirPath string) (string, error) {
	if gui == nil {
		return "", fmt.Errorf("gui not initialized")
	}
	dialog := walk.FileDialog{
		Title:          title,
		Filter:         filter,
		InitialDirPath: initialDirPath,
	}
	ok, err := dialog.ShowBrowseFolder(gui.mw)
	if err != nil {
		return "", fmt.Errorf("show save: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("show save: cancelled")
	}
	return dialog.FilePath, nil
}

func Progress() int {
	if gui == nil {
		return 0
	}
	return gui.progress.Value()
}

func SetBlenderText(value string) {
	if gui == nil {
		return
	}
	gui.blenderText.SetText(value)
}

func SetDownloadEQGZIEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.downloadEQGZI.SetEnabled(value)
}

func SetCurrentZone(name string) {
	if gui == nil {
		return
	}

	zoneIndex := -1
	for idx, zone := range gui.zones {
		if zone != name {
			continue
		}
		zoneIndex = idx
		break
	}
	if zoneIndex == -1 {
		return
	}
	gui.zonesCombo.SetCurrentIndex(zoneIndex)
}

func SetZones(zones []string) {
	if gui == nil {
		return
	}
	gui.zones = zones
	gui.zonesCombo.SetModel(zones)
}

func SetEQPathVisible(value bool) {
	if gui == nil {
		return
	}
	gui.eqPathChange.SetVisible(value)
	gui.eqPathLabel.SetVisible(value)
}

func SetServerVisible(value bool) {
	if gui == nil {
		return
	}
	gui.serverPathChange.SetVisible(value)
	gui.serverPathLabel.SetVisible(value)
}

func SetBlenderOpenEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.blenderOpen.SetEnabled(value)
}

func SetFolderOpenEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.folderOpen.SetEnabled(value)
}

func SetEQGZIOpenEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.eqgziOpen.SetEnabled(value)
}

func SetConvertEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.convert.SetEnabled(value)
}

func SetExportEQGEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.exportEQGCheck.SetEnabled(value)
}

func SetExportServerEnabled(value bool) {
	if gui == nil {
		return
	}
	gui.exportServerCheck.SetEnabled(value)
}
