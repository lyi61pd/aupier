package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"aupier/internal/audio"
	"aupier/internal/config"
)

type ClipInfo struct {
	Filename string
	Time     time.Time
	Duration float64
}

type MainWindow struct {
	app      fyne.App
	window   fyne.Window
	config   *config.Config
	recorder *audio.Recorder
	player   *audio.Player

	statusLabel   *widget.Label
	clipsList     *widget.List
	clips         []ClipInfo
	selectedIndex int
	loopCheck     *widget.Check
	volumeSlider  *widget.Slider

	onRecordToggle func()
	onPlayLast     func()
}

func NewMainWindow(cfg *config.Config) (*MainWindow, error) {
	a := app.New()
	w := a.NewWindow("Aupier - 看电影学英语音频截取工具")

	recorder, err := audio.NewRecorder(cfg.SampleRate, cfg.Channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create recorder: %w", err)
	}

	player, err := audio.NewPlayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	mw := &MainWindow{
		app:           a,
		window:        w,
		config:        cfg,
		recorder:      recorder,
		player:        player,
		clips:         make([]ClipInfo, 0),
		selectedIndex: -1,
	}

	mw.buildUI()
	mw.loadExistingClips()

	return mw, nil
}

func (mw *MainWindow) buildUI() {
	mw.statusLabel = widget.NewLabel("状态: 空闲")

	mw.clipsList = widget.NewList(
		func() int {
			return len(mw.clips)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			clip := mw.clips[id]
			label.SetText(fmt.Sprintf("%s (%.1fs)",
				filepath.Base(clip.Filename),
				clip.Duration))
		},
	)

	mw.clipsList.OnSelected = func(id widget.ListItemID) {
		mw.selectedIndex = id
	}

	playBtn := widget.NewButton("▶ 播放", func() {
		if mw.selectedIndex >= 0 && mw.selectedIndex < len(mw.clips) {
			mw.playClip(mw.clips[mw.selectedIndex].Filename)
		}
	})

	stopBtn := widget.NewButton("■ 停止", func() {
		mw.player.Stop()
	})

	mw.loopCheck = widget.NewCheck("循环播放", func(checked bool) {
		mw.player.SetLoop(checked)
	})

	mw.volumeSlider = widget.NewSlider(0, 2)
	mw.volumeSlider.Value = 1.0
	mw.volumeSlider.OnChanged = func(value float64) {
		mw.player.SetVolume(float32(value))
	}

	deleteBtn := widget.NewButton("删除所选", func() {
		if mw.selectedIndex >= 0 && mw.selectedIndex < len(mw.clips) {
			mw.deleteClip(mw.selectedIndex)
		}
	})

	openFolderBtn := widget.NewButton("打开文件夹", func() {
		absPath, _ := mw.config.GetAbsOutputDir()
		mw.openFolder(absPath)
	})

	statusBox := container.NewVBox(
		mw.statusLabel,
		widget.NewLabel(fmt.Sprintf("录音源: 默认麦克风 (%dHz, %d声道)",
			mw.config.SampleRate, mw.config.Channels)),
	)

	controlBox := container.NewVBox(
		container.NewHBox(playBtn, stopBtn),
		mw.loopCheck,
		container.NewBorder(nil, nil, widget.NewLabel("音量:"), nil, mw.volumeSlider),
		container.NewHBox(deleteBtn, openFolderBtn),
	)

	content := container.NewBorder(
		statusBox,
		controlBox,
		nil,
		nil,
		mw.clipsList,
	)

	mw.window.SetContent(content)
	mw.window.Resize(fyne.NewSize(600, 400))
}

func (mw *MainWindow) loadExistingClips() {
	files, err := os.ReadDir(mw.config.OutputDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".wav" {
			fullPath := filepath.Join(mw.config.OutputDir, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			mw.clips = append(mw.clips, ClipInfo{
				Filename: fullPath,
				Time:     info.ModTime(),
				Duration: 0,
			})
		}
	}

	sort.Slice(mw.clips, func(i, j int) bool {
		return mw.clips[i].Time.After(mw.clips[j].Time)
	})

	mw.clipsList.Refresh()
}

func (mw *MainWindow) ToggleRecording() {
	if mw.recorder.IsRecording() {
		filename := audio.GenerateFilename(mw.config.OutputDir)
		err := mw.recorder.Stop(filename)
		if err != nil {
			dialog.ShowError(err, mw.window)
			mw.statusLabel.SetText("状态: 录音失败")
			return
		}

		mw.statusLabel.SetText("状态: 录音已保存")

		clip := ClipInfo{
			Filename: filename,
			Time:     time.Now(),
			Duration: 0,
		}
		mw.clips = append([]ClipInfo{clip}, mw.clips...)
		mw.clipsList.Refresh()

		mw.selectedIndex = 0
		mw.clipsList.Select(0)
		mw.playClip(filename)

	} else {
		err := mw.recorder.Start()
		if err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		mw.statusLabel.SetText("状态: 🔴 正在录音...")
	}
}

func (mw *MainWindow) PlayLastClip() {
	if len(mw.clips) > 0 {
		mw.selectedIndex = 0
		mw.clipsList.Select(0)
		mw.playClip(mw.clips[0].Filename)
	}
}

func (mw *MainWindow) playClip(filename string) {
	mw.player.Stop()

	err := mw.player.LoadWAV(filename)
	if err != nil {
		dialog.ShowError(fmt.Errorf("无法加载音频: %w", err), mw.window)
		return
	}

	err = mw.player.Play()
	if err != nil {
		dialog.ShowError(fmt.Errorf("无法播放音频: %w", err), mw.window)
		return
	}
}

func (mw *MainWindow) deleteClip(index int) {
	if index < 0 || index >= len(mw.clips) {
		return
	}

	filename := mw.clips[index].Filename

	confirm := dialog.NewConfirm("确认删除",
		fmt.Sprintf("确定要删除 %s 吗？", filepath.Base(filename)),
		func(yes bool) {
			if yes {
				os.Remove(filename)
				mw.clips = append(mw.clips[:index], mw.clips[index+1:]...)
				mw.selectedIndex = -1
				mw.clipsList.UnselectAll()
				mw.clipsList.Refresh()
			}
		}, mw.window)

	confirm.Show()
}

func (mw *MainWindow) openFolder(path string) {
	cmd := exec.Command("explorer", path)
	cmd.Start()
}

func (mw *MainWindow) SetRecordToggleCallback(callback func()) {
	mw.onRecordToggle = callback
}

func (mw *MainWindow) SetPlayLastCallback(callback func()) {
	mw.onPlayLast = callback
}

func (mw *MainWindow) Show() {
	mw.window.ShowAndRun()
}

func (mw *MainWindow) Close() {
	mw.recorder.Close()
	mw.player.Close()
}
