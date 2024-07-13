package client

func (c *Client) onSetEQSaveButton(path string) {
	/*path := fmt.Sprintf("%s/tools/settings.txt", c.currentPath)
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
			c.logf("Failed to read settings.txt: %s", err)
			return
		}

		if strings.Contains(line, `EverQuestDirectory =`) {
			line = fmt.Sprintf("EverQuestDirectory = %s\r\n", setEQ)
		}

		_, err = w.WriteString(line)
		if err != nil {
			c.logf("Failed to write to settings.txt: %s", err)
			return
		}
	}

	c.cfg.EQPath = setEQ
	c.cfg.Save()
	c.labelEQ.SetText(setEQ)
	c.logf("Updated EQ Path to %s", setEQ)
	c.setEQPopup.Hide()
	*/
}
