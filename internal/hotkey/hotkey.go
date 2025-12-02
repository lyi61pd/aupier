package hotkey

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WM_HOTKEY   = 0x0312
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
)

type Manager struct {
	callbacks map[int]func()
	nextID    int
	mu        sync.Mutex
	quit      chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		callbacks: make(map[int]func()),
		nextID:    1,
		quit:      make(chan struct{}),
	}
}

func parseHotkey(hotkeyStr string) (uint32, uint32, error) {
	var modifiers uint32 = 0
	var key uint32 = 0

	switch hotkeyStr {
	case "Ctrl+Shift+R":
		modifiers = MOD_CONTROL | MOD_SHIFT
		key = 'R'
	case "Ctrl+Shift+P":
		modifiers = MOD_CONTROL | MOD_SHIFT
		key = 'P'
	default:
		return 0, 0, fmt.Errorf("unsupported hotkey format: %s", hotkeyStr)
	}

	return modifiers, key, nil
}

func (m *Manager) Register(hotkeyStr string, callback func()) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	modifiers, key, err := parseHotkey(hotkeyStr)
	if err != nil {
		return 0, err
	}

	id := m.nextID
	m.nextID++

	ret, _, err := procRegisterHotKey.Call(
		0,
		uintptr(id),
		uintptr(modifiers),
		uintptr(key),
	)

	if ret == 0 {
		return 0, fmt.Errorf("failed to register hotkey: %v", err)
	}

	m.callbacks[id] = callback
	return id, nil
}

func (m *Manager) Unregister(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ret, _, err := procUnregisterHotKey.Call(0, uintptr(id))
	if ret == 0 {
		return fmt.Errorf("failed to unregister hotkey: %v", err)
	}

	delete(m.callbacks, id)
	return nil
}

func (m *Manager) Listen() {
	type MSG struct {
		HWND   uintptr
		UINT   uint32
		WPARAM int32
		LPARAM int64
		DWORD  uint32
		POINT  struct{ X, Y int32 }
	}

	msg := &MSG{}

	for {
		select {
		case <-m.quit:
			return
		default:
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(msg)),
				0,
				0,
				0,
			)

			if ret == 0 {
				return
			}

			if msg.UINT == WM_HOTKEY {
				m.mu.Lock()
				if callback, ok := m.callbacks[int(msg.WPARAM)]; ok {
					go callback()
				}
				m.mu.Unlock()
			}

			procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
			procDispatchMessage.Call(uintptr(unsafe.Pointer(msg)))
		}
	}
}

func (m *Manager) Stop() {
	close(m.quit)
}
