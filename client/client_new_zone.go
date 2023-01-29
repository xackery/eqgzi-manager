package client

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (c *Client) newZoneInit() {
	c.newZoneSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onNewZoneSaveButton)
	c.newZoneCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onNewZoneCancelButton)

	c.newZoneName = widget.NewEntry()
	c.newZoneName.OnSubmitted = func(string) { c.onNewZoneSaveButton() }
	c.newZoneButton = widget.NewButtonWithIcon("Create New Zone", theme.FolderNewIcon(), func() {
		c.newZonePopup.Show()
		c.window.Canvas().Focus(c.newZoneName)
	})

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

	if strings.Contains(newZone, ".") {
		c.popupStatus.SetText(fmt.Sprintf("Failed: zone %s shouldn't have a period", newZone))
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
		c.popupStatus.SetText(fmt.Sprintf("Failed creating white.png: %s", err))
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
