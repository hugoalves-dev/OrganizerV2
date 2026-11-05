//go:build windows

package notifier

import (
    "strings"
    "syscall"
    "unsafe"
)

// SHQueryUserNotificationState is Microsoft's recommended check before
// presenting notification UI. The foreground-window fallback also detects
// fullscreen browsers and borderless games not reported by the shell.
var (
    notificationShell32 = syscall.NewLazyDLL("shell32.dll")
    notificationUser32  = syscall.NewLazyDLL("user32.dll")
    shQueryUserNotificationState = notificationShell32.NewProc("SHQueryUserNotificationState")
    getForegroundWindow = notificationUser32.NewProc("GetForegroundWindow")
    getWindowRect = notificationUser32.NewProc("GetWindowRect")
    getClassName = notificationUser32.NewProc("GetClassNameW")
    isIconic = notificationUser32.NewProc("IsIconic")
    monitorFromWindow = notificationUser32.NewProc("MonitorFromWindow")
    getMonitorInfo = notificationUser32.NewProc("GetMonitorInfoW")
)

type fullscreenRect struct {
    Left, Top, Right, Bottom int32
}

type fullscreenMonitorInfo struct {
    Size uint32
    Monitor fullscreenRect
    Work fullscreenRect
    Flags uint32
}

func windowsFullscreenActive() bool {
    const (
        qunsBusy = 2
        qunsExclusiveD3D = 3
        qunsPresentation = 4
    )
    var state uint32
    hr, _, _ := shQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&state)))
    if hr == 0 && (state == qunsBusy || state == qunsExclusiveD3D || state == qunsPresentation) {
        return true
    }

    hwnd, _, _ := getForegroundWindow.Call()
    if hwnd == 0 {
        return false
    }
    minimized, _, _ := isIconic.Call(hwnd)
    if minimized != 0 {
        return false
    }

    // The desktop fills a monitor, but obviously isn't a fullscreen app.
    var name [256]uint16
    length, _, _ := getClassName.Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
    if length > 0 {
        switch strings.ToLower(syscall.UTF16ToString(name[:int(length)])) {
        case "progman", "workerw", "shell_traywnd", "shell_secondarytraywnd":
            return false
        }
    }

    var window fullscreenRect
    ok, _, _ := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&window)))
    if ok == 0 {
        return false
    }
    monitor, _, _ := monitorFromWindow.Call(hwnd, 2) // MONITOR_DEFAULTTONEAREST
    if monitor == 0 {
        return false
    }
    info := fullscreenMonitorInfo{Size: uint32(unsafe.Sizeof(fullscreenMonitorInfo{}))}
    ok, _, _ = getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info)))
    return ok != 0 && fullscreenCoversMonitor(window, info.Monitor)
}

// Compare monitor bounds, not WorkArea: maximized windows normally leave the
// taskbar exposed, while true fullscreen windows cover it. A 2px tolerance
// accommodates rounding at non-integer display scales.
func fullscreenCoversMonitor(window, monitor fullscreenRect) bool {
    const tolerance int32 = 2
    return window.Right > window.Left && window.Bottom > window.Top &&
        monitor.Right > monitor.Left && monitor.Bottom > monitor.Top &&
        window.Left <= monitor.Left+tolerance &&
        window.Top <= monitor.Top+tolerance &&
        window.Right >= monitor.Right-tolerance &&
        window.Bottom >= monitor.Bottom-tolerance
}
