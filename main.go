package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xackery/eqgzi-manager/client"
	"github.com/xackery/eqgzi-manager/config"
	"github.com/xackery/eqgzi-manager/gui"
	"github.com/xackery/eqgzi-manager/slog"
	"github.com/xackery/wlk/walk"
)

var (
	// Version is the version of the build
	Version string
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		os.Exit(1)
	}

}

func run() error {
	if Version == "" {
		Version = string(client.VersionText.Content())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exeName, err := os.Executable()
	if err != nil {
		gui.MessageBox("Error", "Failed to get executable name", true)
		os.Exit(1)
	}
	baseName := filepath.Base(exeName)
	if strings.Contains(baseName, ".") {
		baseName = baseName[0:strings.Index(baseName, ".")]
	}
	cfg, err := config.New(context.Background())
	if err != nil {
		slog.Printf("Failed to load config: %s", err.Error())
		os.Exit(1)
	}

	err = gui.NewMainWindow(ctx, cancel, cfg, Version)
	if err != nil {
		return fmt.Errorf("new main window: %w", err)
	}

	_, err = client.New()
	if err != nil {
		fmt.Println("client new:", err)
		os.Exit(1)
	}
	defer slog.Dump(baseName + ".txt")

	gui.SubscribeClose(func(canceled *bool, reason byte) {
		if ctx.Err() != nil {
			fmt.Println("Accepting exit")
			return
		}
		*canceled = true
		fmt.Println("Got close message")
		gui.SetTitle("Closing...")
		cancel()
	})

	go func() {
		<-ctx.Done()
		fmt.Println("Doing clean up process...")
		gui.Close()
		walk.App().Exit(0)
		fmt.Println("Done, exiting")
		slog.Dump(baseName + ".txt")
		os.Exit(0)
	}()

	errCode := gui.Run()
	if errCode != 0 {
		fmt.Println("Failed to run:", errCode)
		os.Exit(1)
	}
	return nil
}
