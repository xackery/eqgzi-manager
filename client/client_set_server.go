package client

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (c *Client) newSetServerInit() {
	c.setServerSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onSetServerSaveButton)
	c.setServerCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onSetServerCancelButton)

	c.setServerName = widget.NewEntry()
	c.setServerName.OnSubmitted = func(string) { c.onSetServerSaveButton() }
	c.setServerButton = widget.NewButtonWithIcon("Set Server Path", theme.FolderNewIcon(), func() {
		c.setServerPopup.Show()
		c.window.Canvas().Focus(c.setServerName)
	})

	c.setServerPopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Set Server Path"),
			c.setServerName,
			container.NewHBox(
				c.setServerSaveButton,
				c.setServerCancelButton,
			),
			c.popupStatus,
		),
		c.window.Canvas(),
	)
}

func (c *Client) onSetServerSaveButton() {
	if c.setServerName.Text == "" {
		c.popupStatus.SetText("Failed: path cannot be empty")
		return
	}
	setServer := strings.TrimSpace(c.setServerName.Text)
	if !strings.HasSuffix(setServer, "/") && !strings.HasSuffix(setServer, `\`) {
		setServer += "/"
	}
	setServer = strings.ReplaceAll(setServer, `\`, "/")
	err := c.validateSetServerPath(setServer)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed %s", err))
		return
	}
	c.cfg.ServerPath = setServer
	c.cfg.Save()
	c.labelServer.SetText(setServer)
	c.logf("Updated Server Path")

	err = os.WriteFile(fmt.Sprintf("%s/tools/map_edit/config.json", strings.ReplaceAll(c.currentPath, `\`, "/")), []byte(fmt.Sprintf(`{
	"paths": {
		"base": "%s/base/",
		"project": "project/",
		"nav": "%s/nav/",
		"water": "%s/water/",
		"volume": "%s/volume/"
	}
}`, setServer, setServer, setServer, setServer)), os.ModePerm)
	if err != nil {
		c.popupStatus.SetText(fmt.Sprintf("Failed setting config.json: %s", err))
		return
	}

	c.setServerPopup.Hide()
}

func (c *Client) validateSetServerPath(path string) error {

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path %s: %w", path, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path %s is not a directory", path)
	}

	subdirs := []string{
		"base",
		"nav",
		"volume",
		"water",
	}
	for _, subdir := range subdirs {
		fullPath := fmt.Sprintf("%s%s", path, subdir)
		fi, err = os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path %s does not exist", fullPath)
			}
			return fmt.Errorf("path %s: %w", fullPath, err)
		}
		if !fi.IsDir() {
			return fmt.Errorf("path %s is not a directory", fullPath)
		}
	}
	return nil
}

func (c *Client) onSetServerCancelButton() {
	c.logf("Cancelled set server")
	c.setServerPopup.Hide()
}
