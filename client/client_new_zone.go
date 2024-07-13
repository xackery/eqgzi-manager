package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/xackery/eqgzi-manager/gui"
)

func (c *Client) newZoneInit() {
	newZone, err := gui.ShowNewZone()
	if err != nil {
		c.logf("Failed new zone: %s", err)
		return
	}

	if newZone.Name == "" {
		gui.MessageBox("Error", "Zone name cannot be empty", true)
		return
	}

	name := strings.ToLower(strings.TrimSpace(newZone.Name))

	//  TODO: fix name to sanitize for file creation
	_, err = os.Stat(fmt.Sprintf("zones/%s", name))
	if err == os.ErrNotExist {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to stat folder", name), true)
		return
	}

	err = os.Mkdir(fmt.Sprintf("zones/%s/", newZone), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create folder: %s", name, err), true)
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/convert.bat", newZone), convertText.Content(), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create convert.bat: %s", name, err), true)
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_eq.bat", newZone), copyEQText.Content(), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create copy_eq.bat: %s", name, err), true)
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/copy_server.bat", newZone), copyServerText.Content(), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create copy_server.bat: %s", name, err), true)
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/%s.blend", newZone, newZone), baseBlend.Content(), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create %s.blend: %s", name, name, err), true)
		return
	}

	err = os.WriteFile(fmt.Sprintf("zones/%s/white.png", newZone), whitePng.Content(), os.ModePerm)
	if err != nil {
		gui.MessageBox("Error", fmt.Sprintf("Zone %s failed to create white.png: %s", name, err), true)
		return
	}

	c.onZoneRefresh()
	gui.SetCurrentZone(newZone.Name)
	c.logf("Created zones/%s", newZone)
}
