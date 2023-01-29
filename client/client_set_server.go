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
	c.setServerSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onSetZoneSaveButton)
	c.setServerCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onSetZoneCancelButton)

	c.setServerName = widget.NewEntry()
	c.setServerName.OnSubmitted = func(string) { c.onSetZoneSaveButton() }
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

func (c *Client) onSetZoneSaveButton() {
	if c.setServerName.Text == "" {
		c.popupStatus.SetText("Failed: path cannot be empty")
		return
	}

	setServer := strings.TrimSpace(c.setServerName.Text)
	_, err := os.Stat(setServer)
	if err == os.ErrNotExist {
		c.popupStatus.SetText(fmt.Sprintf("Failed: path %s already exists", setServer))
		return
	}

	c.cfg.ServerPath = setServer
	c.cfg.Save()
	c.labelServer.SetText(setServer)
	c.logf("Updated Server Path")
	c.setServerPopup.Hide()
}

func (c *Client) onSetZoneCancelButton() {
	c.logf("Cancelled set server")
	c.setServerPopup.Hide()
}
