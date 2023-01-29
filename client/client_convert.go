package client

import (
	"fmt"
	"os"
	"os/exec"
)

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
		fmt.Sprintf(`PATH=%s;%s\tools`, blenderPath, currentPath),
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
