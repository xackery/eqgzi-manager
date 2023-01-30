package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (c *Client) newSetEQInit() {
	c.setEQSaveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), c.onSetEQSaveButton)
	c.setEQCancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), c.onSetEQCancelButton)

	c.setEQName = widget.NewEntry()
	c.setEQName.OnSubmitted = func(string) { c.onSetEQSaveButton() }
	c.setEQButton = widget.NewButtonWithIcon("Set EQ Path", theme.FolderNewIcon(), func() {
		c.setEQPopup.Show()
		c.window.Canvas().Focus(c.setEQName)
	})

	c.setEQPopup = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Set EQ Path"),
			c.setEQName,
			container.NewHBox(
				c.setEQSaveButton,
				c.setEQCancelButton,
			),
			c.popupStatus,
		),
		c.window.Canvas(),
	)
}

func (c *Client) onSetEQSaveButton() {
	if c.setEQName.Text == "" {
		c.popupStatus.SetText("Failed: path cannot be empty")
		return
	}

	setEQ := strings.TrimSpace(c.setEQName.Text)
	setEQ = strings.ReplaceAll(setEQ, `\`, "/")
	_, err := os.Stat(setEQ)
	if err == os.ErrNotExist {
		c.popupStatus.SetText(fmt.Sprintf("Failed: zone %s already exists", setEQ))
		return
	}

	path := fmt.Sprintf("%s/tools/settings.txt", c.currentPath)
	settings, err := os.ReadFile(path)
	if err != nil {
		c.logf("Failed to read settings.txt: %s", err)
		return
	}

	w, err := os.Create(path)
	if err != nil {
		c.logf("Failed to create settings.txt: %s", err)
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

		if strings.Contains(line, `EverQuestDirectory =`) {
			line = fmt.Sprintf("EverQuestDirectory = %s\r\n", setEQ)
		}

		_, err = w.WriteString(line)
		if err != nil {
			c.logf("Failed to write to settings.lua: %s", err)
			return
		}
	}

	c.cfg.EQPath = setEQ
	c.cfg.Save()
	c.labelEQ.SetText(setEQ)
	c.logf("Updated EQ Path to %s", setEQ)
	c.setEQPopup.Hide()
}

func (c *Client) onSetEQCancelButton() {
	c.logf("Cancelled set EQ Path")
	c.setEQPopup.Hide()
}
