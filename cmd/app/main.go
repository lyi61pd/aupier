package main

import (
"fmt"
"log"
"os"
"runtime"

"aupier/internal/config"
"aupier/internal/hotkey"
"aupier/internal/ui"
)

const configFile = "config.json"

func main() {
	if runtime.GOOS != "windows" {
		log.Fatal("This application is designed for Windows only")
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	mainWindow, err := ui.NewMainWindow(cfg)
	if err != nil {
		log.Fatalf("Failed to create main window: %v", err)
	}
	defer mainWindow.Close()

	hotkeyManager := hotkey.NewManager()

	_, err = hotkeyManager.Register(cfg.RecordHotkey, func() {
		mainWindow.ToggleRecording()
	})
	if err != nil {
		log.Printf("Warning: Failed to register record hotkey: %v", err)
	}

	_, err = hotkeyManager.Register(cfg.PlayLastClipHotkey, func() {
		mainWindow.PlayLastClip()
	})
	if err != nil {
		log.Printf("Warning: Failed to register play last clip hotkey: %v", err)
	}

	go hotkeyManager.Listen()

	fmt.Printf("Application started\n")
	fmt.Printf("Record hotkey: %s\n", cfg.RecordHotkey)
	fmt.Printf("Play last clip hotkey: %s\n", cfg.PlayLastClipHotkey)
	fmt.Printf("Output directory: %s\n", cfg.OutputDir)

	mainWindow.Show()

	hotkeyManager.Stop()
}
