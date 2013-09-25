package render

import (
    "time"
    "errors"
    glfw "github.com/go-gl/glfw3"
)

type WindowDelegate interface {
    Init(window *glfw.Window)
    Draw(window *glfw.Window)
    Reshape(window *glfw.Window, width, height int)
    MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey)
    MouseMove(window *glfw.Window, xpos float64, ypos float64)
    KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey)
    Scroll(window *glfw.Window, xoff float64, yoff float64)
    Simulate(now time.Time, elapsed time.Duration, duration time.Duration)
    OnClose(window *glfw.Window)
    IsIdle() bool
    NeedsRender() bool
}

func WindowAspectRatio(window *glfw.Window) float64 {
    frameWidth, frameHeight := window.GetFramebufferSize()
    return float64(frameWidth) / float64(frameHeight)
}

func bindEvents(window *glfw.Window, delegate WindowDelegate) {
    window.SetFramebufferSizeCallback(delegate.Reshape)
    window.SetMouseButtonCallback(delegate.MouseClick)
    window.SetCursorPositionCallback(delegate.MouseMove)
    window.SetKeyCallback(delegate.KeyPress)
    window.SetScrollCallback(delegate.Scroll)
    window.SetCloseCallback(delegate.OnClose)
}

func CreateWindow(width, height int, name string, fullscreen bool, delegate WindowDelegate) error {
    if !glfw.Init() {
        return errors.New("Failed to initialize GLFW")
    }
    defer glfw.Terminate()
        
    glfw.WindowHint(glfw.DepthBits, 16)
    glfw.WindowHint(glfw.ContextVersionMajor, 3)
    glfw.WindowHint(glfw.ContextVersionMinor, 2)
    glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
    glfw.WindowHint(glfw.OpenglForwardCompatible, 1)

    var monitor *glfw.Monitor = nil
    var err error = nil
    if fullscreen {
        monitor, err = glfw.GetPrimaryMonitor()
        if err != nil {
            return err
        }
    }
    window, err := glfw.CreateWindow(width, height, name, monitor, nil)
    if err != nil {
        return err
    }
    
    if fullscreen {
        window.SetInputMode(glfw.Cursor, glfw.CursorDisabled)        
    }

    bindEvents(window, delegate)
    
    window.MakeContextCurrent()
    glfw.SwapInterval(1)

    delegate.Init(window)
    frameWidth, frameHeight := window.GetFramebufferSize()
    delegate.Reshape(window, frameWidth, frameHeight)
    start := time.Now()
    last := start
    for !window.ShouldClose() {
        now := time.Now()
        duration := now.Sub(last)
        elapsed := now.Sub(start)
        delegate.Simulate(now, elapsed, duration)
        if delegate.NeedsRender() {
            delegate.Draw(window)
        }
        window.SwapBuffers()
        if delegate.IsIdle() {
            glfw.WaitEvents()            
        } else {
            glfw.PollEvents()
        }
    }
    return nil
}
