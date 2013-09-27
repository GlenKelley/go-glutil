package render

import (
    // "fmt"
    "regexp"
    "reflect"
    "strconv"
    glm "github.com/Jragonmiris/mathgl"
    glfw "github.com/go-gl/glfw3"
)
type keyEvent struct {
    Key glfw.Key
    Action glfw.Action
}

type mouseButtonEvent struct {
    Button glfw.MouseButton
    Action glfw.Action
}

type Action func()
type MouseMoveAction func(position, delta glm.Vec2f)

type ControlBindings struct {
    keyBindings map[keyEvent]Action
    mouseButtonBindings map[mouseButtonEvent]Action
    mouseMovementBinding MouseMoveAction
}

func (c *ControlBindings) ResetBindings() {
    c.keyBindings = map[keyEvent]Action{}
    c.mouseButtonBindings = map[mouseButtonEvent]Action{}
    c.mouseMovementBinding = nil
}

func (c *ControlBindings) BindKeyPress(key glfw.Key, press Action, release Action) {
    if press != nil {
        c.keyBindings[keyEvent{key, glfw.Press}] = press
    }
    if release != nil {
        c.keyBindings[keyEvent{key, glfw.Release}] = release
    }
}

func (c *ControlBindings) BindMouseClick(button glfw.MouseButton, press Action, release Action) {
    if press != nil {
        c.mouseButtonBindings[mouseButtonEvent{button, glfw.Press}] = press
    }
    if release != nil {
        c.mouseButtonBindings[mouseButtonEvent{button, glfw.Release}] = release
    }
}

func (c *ControlBindings) UnbindKeyPress(key glfw.Key) {
    delete(c.keyBindings, keyEvent{key, glfw.Press})
    delete(c.keyBindings, keyEvent{key, glfw.Release})
}

func (c *ControlBindings) UnbindMouseClick(button glfw.MouseButton) {
    delete(c.mouseButtonBindings, mouseButtonEvent{button, glfw.Press})
    delete(c.mouseButtonBindings, mouseButtonEvent{button, glfw.Release})
}

func (c *ControlBindings) BindMouseMovement(action MouseMoveAction) {
    c.mouseMovementBinding = action
}

func (c *ControlBindings) UnbindMouseMovement() {
    c.mouseMovementBinding = nil
}

func (c *ControlBindings) FindKeyAction(key glfw.Key, keyAction glfw.Action) (Action, bool) {
    action, ok := c.keyBindings[keyEvent{key,keyAction}]
    return action, ok
}

func (c *ControlBindings) FindClickAction(button glfw.MouseButton, buttonAction glfw.Action) (Action, bool) {
    action, ok := c.mouseButtonBindings[mouseButtonEvent{button,buttonAction}]
    return action, ok

}

func (c *ControlBindings) FindMouseMovementAction() (MouseMoveAction, bool) {
    return c.mouseMovementBinding, c.mouseMovementBinding != nil
}

func FindActionMethod(v reflect.Value, name string) Action {
    m := v.MethodByName(name)
    if m.IsValid() {
        return Action(m.Interface().(func()))
    } else {
        return nil
    }
}

func FindMouseMoveActionMethod(v reflect.Value, name string) MouseMoveAction {
    m := v.MethodByName(name)
    if m.IsValid() {
        return MouseMoveAction(m.Interface().(func(glm.Vec2f,glm.Vec2f)))
    } else {
        return nil
    }
}


func (c *ControlBindings) Apply(receiver interface{}, bindings map[string]string) {
    mb, err := regexp.Compile("^mouse([1-9])$")
    if err != nil {
        panic(err)
    }
    receiverValue := reflect.ValueOf(receiver)
    for k, name := range bindings {
        stopName := "Stop"+name
        if len(k) == 1 {
            key := glfw.Key(k[0])
            if name == "" {
                c.UnbindKeyPress(key)
            } else {
                startAction := FindActionMethod(receiverValue, name)
                stopAction := FindActionMethod(receiverValue, stopName)
                c.BindKeyPress(key, startAction, stopAction)
            }
        } else if (mb.MatchString(k)) {
            i, err := strconv.Atoi(mb.FindStringSubmatch(k)[1])
            if err == nil {
                button := glfw.MouseButton1 + glfw.MouseButton(i-1)
                if name == "" {
                    c.UnbindMouseClick(button)
                } else {
                    // fmt.Println("bind", i, button, glfw.MouseButton1)
                    startAction := FindActionMethod(receiverValue, name)
                    stopAction := FindActionMethod(receiverValue, stopName)
                    c.BindMouseClick(button, startAction, stopAction)
                }                
            }
        } else if (k == "mousemove") {
            if name == "" {
                c.UnbindMouseMovement()
            } else {
                moveAction := FindMouseMoveActionMethod(receiverValue, name)
                c.BindMouseMovement(moveAction)
            }
        }
    }
}