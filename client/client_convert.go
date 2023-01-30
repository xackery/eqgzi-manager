package client

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

	c.progressBar.Show()
	c.progressBar.Value = 0
	c.statusLabel.Hide()
	defer func() {
		c.progressBar.Hide()
		c.statusLabel.Show()
	}()

	env := []string{
		fmt.Sprintf(`PATH=%s;%s\tools`, blenderPath, currentPath),
		fmt.Sprintf(`EQPATH=%s`, eqPath),
		fmt.Sprintf(`EQGZI=%s\tools\`, currentPath),
		fmt.Sprintf(`ZONE=%s`, zone),
		fmt.Sprintf(`EQSERVERPATH=%s`, strings.ReplaceAll(serverPath, "/", `\`)),
		fmt.Sprintf(`BLENDERPATH=%s`, blenderPath),
	}

	cmd := exec.Command(fmt.Sprintf("%s/zones/%s/convert.bat", currentPath, zone))
	cmd.Dir = fmt.Sprintf("%s/zones/%s/", currentPath, zone)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.logf("Failed to start convert.bat: stdoutpipe: %s", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.logf("Failed to start convert.bat: stderrpipe: %s", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		c.logf("Failed to run convert.bat: %s", err)
		return
	}

	reader := io.MultiReader(stdout, stderr)

	err = c.processOutput(reader, currentPath, zone, "convert.log")
	if err != nil {
		c.logf("Failed stdout: %s", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		c.logf("Failed convert.bat: %s", err)
		return
	}

	if isEQCopy {
		cmd = exec.Command(fmt.Sprintf("%s/zones/%s/copy_eq.bat", currentPath, zone))
		cmd.Dir = fmt.Sprintf("%s/zones/%s/", currentPath, zone)
		cmd.Env = env
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			c.logf("Failed to start copy_eq.bat: stdoutpipe: %s", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			c.logf("Failed to start copy_eq.bat: stderrpipe: %s", err)
			return
		}
		reader = io.MultiReader(stdout, stderr)
		err = cmd.Start()
		if err != nil {
			c.logf("Failed to run copy_eq.bat: %s", err)
			return
		}
		err = c.processOutput(reader, currentPath, zone, "copy_eq.log")
		if err != nil {
			c.logf("Failed copy_eq stdout: %s", err)
			return
		}
		err = cmd.Wait()
		if err != nil {
			c.logf("Failed copy_eq.bat: %s", err)
			return
		}

	}

	if isServerCopy {
		cmd = exec.Command(fmt.Sprintf("%s/zones/%s/copy_server.bat", currentPath, zone))
		cmd.Dir = fmt.Sprintf("%s/zones/%s/", currentPath, zone)
		cmd.Env = env
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			c.logf("Failed to start copy_server.bat: stdoutpipe: %s", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			c.logf("Failed to start copy_server.bat: stderrpipe: %s", err)
			return
		}
		err = cmd.Start()
		if err != nil {
			c.logf("Failed to run copy_server.bat: %s", err)
			return
		}
		reader = io.MultiReader(stdout, stderr)
		err = c.processOutput(reader, currentPath, zone, "copy_server.log")
		if err != nil {
			c.logf("Failed copy_server stdout: %s", err)
			return
		}
		err = cmd.Wait()
		if err != nil {
			c.logf("Failed copy_server.bat: %s", err)
			return
		}

	}
	c.logf("Created %s.eqg", zone)
}

func (c *Client) processOutput(in io.Reader, currentPath string, zone string, logName string) error {
	buf := bufio.NewReader(in)
	lineNumber := 0
	outLog, err := os.Create(fmt.Sprintf("%s/zones/%s/%s", currentPath, zone, logName))
	if err != nil {
		return fmt.Errorf("create %s: %s", logName, err)
	}
	defer outLog.Close()
	_, err = outLog.WriteString(fmt.Sprintf("Initialized from eqgzi-manager v%s", VersionText.Content()))
	if err != nil {
		return fmt.Errorf("write to %s: %s", logName, err)
	}
	failedMessage := ""
	isMainCommandError := false

	step := 0

	for {
		lineNumber++
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read %s: %s", logName, err)
		}
		fmt.Printf("%s:%d %s", logName, lineNumber, line)

		if logName == "convert.log" && strings.HasPrefix(line, "Step") && len(line) > 7 {
			stepNum, err := strconv.Atoi(line[5:6])
			if err != nil { //can be ignored
				c.logf("minor note %s:%d %s %s", logName, lineNumber, line, err)
			} else {
				step = stepNum
			}
		}

		//Keyerrors are great source errors
		if strings.HasPrefix(line, "KeyError:") {
			context := ""
			if strings.Contains(line, "not found") && strings.Contains(line, "bpy_prop_collection") {
				context = " (an image texture is not properly exported)"
			}
			failedMessage = fmt.Sprintf("%s:%d %s%s", logName, lineNumber, line, context)
		}

		if failedMessage == "" && strings.Contains(line, "GPUTexture: Blender Texture Not Loaded!") {
			failedMessage = fmt.Sprintf("%s:%d %s (a reference to a texture in blender is broken)", logName, lineNumber, line)
		}

		if failedMessage == "" && isMainCommandError {
			failedMessage = fmt.Sprintf("%s:%d %s", logName, lineNumber, line)
			isMainCommandError = false
		}
		if failedMessage == "" && strings.Contains(line, "missing") && strings.Contains(line, "not copying") {
			failedMessage = fmt.Sprintf("%s:%d %s", logName, lineNumber, line)
		}
		if failedMessage == "" && strings.Contains(line, "error") && !strings.Contains(line, "main_cmd error") {
			failedMessage = fmt.Sprintf("%s:%d %s", logName, lineNumber, line)
		}
		if step >= 7 && failedMessage == "" && strings.Contains(line, "main_cmd error:") {
			isMainCommandError = true
		}
		if failedMessage == "" && strings.Contains(line, "PermissionError: [Errno 13] Permission denied: '.'") {
			failedMessage = fmt.Sprintf("%s:%d %s (This is usually caused by an embedded image)", logName, lineNumber, line)
		}
		_, err = outLog.WriteString(line)
		if err != nil {
			return fmt.Errorf("write string to %s: %s", logName, err)
		}
	}

	if failedMessage != "" {
		return fmt.Errorf(failedMessage)
	}
	if logName == "convert.log" && step < 7 {
		return fmt.Errorf("convert.bat failed at step %d", step)
	}
	return nil
}
