//go build windows
//+windows

package client

import (
	"strings"

	"github.com/xackery/eqgzi-manager/gui"
	"golang.org/x/sys/windows/registry"
)

func (c *Client) onBlenderDetectButton() {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\blendfile\DefaultIcon`, registry.QUERY_VALUE)
	if err != nil {
		c.logf("Failed registry: %s", err)
		return
	}
	defer k.Close()

	s, _, err := k.GetStringValue("")
	if err != nil {
		c.logf("Failed registry get: %s", err)
		return
	}
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.TrimSuffix(s, ", 1")
	s = strings.TrimSuffix(s, "blender-launcher.exe")

	gui.SetBlenderText(s)
	c.cfg.BlenderPath = s
	err = c.cfg.Save()
	if err != nil {
		c.logf("Failed save: %s", err)
		return
	}
	c.logf("Updated blender path")
}
