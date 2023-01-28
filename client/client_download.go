package client

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (c *Client) onDownloadEQGZIButton() {
	c.statusLabel.Hide()
	c.progressBar.Show()
	defer func() {
		c.statusLabel.Show()
		c.progressBar.Hide()
	}()

	err := c.downloadEQGZI()
	if err != nil {
		fmt.Println("failed", err)
		c.statusLabel.SetText("failed " + err.Error())
		return
	}

	c.window.SetContent(c.mainCanvas)
}

func (c *Client) downloadEQGZI() error {

	type Reply struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			Size               int    `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	gitReply := &Reply{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/xackery/eqgzi/releases/latest", nil)
	if err != nil {
		return fmt.Errorf("new git request: %w", err)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do git request: %w", err)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(gitReply)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("decode git request: %w", err)
	}
	assetURL := ""
	zipName := fmt.Sprintf("eqgzi-%s.zip", gitReply.TagName)
	for _, asset := range gitReply.Assets {
		if asset.Name != zipName {
			continue
		}
		assetURL = asset.BrowserDownloadURL
	}
	if assetURL == "" {
		return fmt.Errorf("download eqgzi zip not found")
	}
	fmt.Println("downloading", assetURL)

	err = os.Mkdir("cache", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	_, err = os.Stat(fmt.Sprintf("cache/%s", zipName))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat cache/%s: %w", zipName, err)
		}
		req, err = http.NewRequest("GET", assetURL, nil)
		if err != nil {
			return fmt.Errorf("asset req: %w", err)
		}

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("asset do: %w", err)
		}

		w, err := os.Create(fmt.Sprintf("cache/%s", zipName))
		if err != nil {
			return fmt.Errorf("create zip: %w", err)
		}
		defer w.Close()

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return fmt.Errorf("copy zip: %w", err)
		}
	}

	fmt.Println("extracting", zipName)
	err = os.Mkdir("tools", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir tools: %w", err)
	}
	zr, err := zip.OpenReader(fmt.Sprintf("cache/%s", zipName))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	for _, zf := range zr.File {
		filePath := fmt.Sprintf("tools/%s", zf.Name)

		if zf.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("mkdir %s: %w", filePath, err)
			}
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return fmt.Errorf("open fs %s: %w", filePath, err)
		}

		fileInArchive, err := zf.Open()
		if err != nil {
			return fmt.Errorf("open zip %s: %w", zf.Name, err)
		}

		_, err = io.Copy(dstFile, fileInArchive)
		if err != nil {
			return fmt.Errorf("copy %s: %w", zf.Name, err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
