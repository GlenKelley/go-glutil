package render

import (
	"errors"
	glfw "github.com/go-gl/glfw3"
	"time"
)

type GameTime struct {
	Now     time.Time
	Elapsed time.Duration
	Delta   time.Duration
}

type DoSimulation func()

type WindowDelegate interface {
	Init(window *glfw.Window)
	Draw(window *glfw.Window)
	Reshape(window *glfw.Window, width, height int)
	MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey)
	MouseMove(window *glfw.Window, xpos float64, ypos float64)
	KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey)
	Scroll(window *glfw.Window, xoff float64, yoff float64)
	Simulate(time GameTime)
	OnClose(window *glfw.Window)
	IsIdle() bool
	NeedsRender() bool
}

type WindowDelegator struct {
	Delegate WindowDelegate
}

func (wd *WindowDelegator) Init(window *glfw.Window) {
	wd.Delegate.Init(window)
}
func (wd *WindowDelegator) Draw(window *glfw.Window) {
	wd.Delegate.Draw(window)
}
func (wd *WindowDelegator) Reshape(window *glfw.Window, width, height int) {
	wd.Delegate.Reshape(window, width, height)
}
func (wd *WindowDelegator) MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	wd.Delegate.MouseClick(window, button, action, mod)
}
func (wd *WindowDelegator) MouseMove(window *glfw.Window, xpos float64, ypos float64) {
	wd.Delegate.MouseMove(window, xpos, ypos)
}
func (wd *WindowDelegator) KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
	wd.Delegate.KeyPress(window, k, s, action, mods)
}
func (wd *WindowDelegator) Scroll(window *glfw.Window, xoff float64, yoff float64) {
	wd.Delegate.Scroll(window, xoff, yoff)
}
func (wd *WindowDelegator) Simulate(time GameTime) {
	wd.Delegate.Simulate(time)
}
func (wd *WindowDelegator) OnClose(window *glfw.Window) {
	wd.Delegate.OnClose(window)
}
func (wd *WindowDelegator) IsIdle() bool {
	return wd.Delegate.IsIdle()
}
func (wd *WindowDelegator) NeedsRender() bool {
	return wd.Delegate.NeedsRender()
}

type IdleSimulatorWindowDelegator struct {
	WindowDelegator
	DoSimulate DoSimulation
}

func (wd *IdleSimulatorWindowDelegator) MouseClick(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if wd.WindowDelegator.IsIdle() {
		wd.DoSimulate()
	}
	wd.WindowDelegator.MouseClick(window, button, action, mod)
}
func (wd *IdleSimulatorWindowDelegator) MouseMove(window *glfw.Window, xpos float64, ypos float64) {
	if wd.WindowDelegator.IsIdle() {
		wd.DoSimulate()
	}
	wd.WindowDelegator.MouseMove(window, xpos, ypos)
}
func (wd *IdleSimulatorWindowDelegator) KeyPress(window *glfw.Window, k glfw.Key, s int, action glfw.Action, mods glfw.ModifierKey) {
	if wd.WindowDelegator.IsIdle() {
		wd.DoSimulate()
	}
	wd.WindowDelegator.KeyPress(window, k, s, action, mods)
}
func (wd *IdleSimulatorWindowDelegator) Scroll(window *glfw.Window, xoff float64, yoff float64) {
	if wd.WindowDelegator.IsIdle() {
		wd.DoSimulate()
	}
	wd.WindowDelegator.Scroll(window, xoff, yoff)
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

		vidModes, _ := monitor.GetVideoModes()
		maxResolution := vidModes[len(vidModes)-1]
		width = maxResolution.Width
		height = maxResolution.Height
	}
	window, err := glfw.CreateWindow(width, height, name, monitor, nil)
	if err != nil {
		return err
	}

	if fullscreen {
		window.SetInputMode(glfw.Cursor, glfw.CursorDisabled)
	}

	start := time.Now()
	last := start
	doSimulation := func() {
		now := time.Now()
		gameTime := GameTime{
			now,
			now.Sub(start),
			now.Sub(last),
		}
		delegate.Simulate(gameTime)
		last = now
	}

	bindEvents(window, &IdleSimulatorWindowDelegator{
		WindowDelegator{delegate},
		doSimulation,
	})
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	delegate.Init(window)
	frameWidth, frameHeight := window.GetFramebufferSize()
	delegate.Reshape(window, frameWidth, frameHeight)
	for !window.ShouldClose() {
		doSimulation()
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
