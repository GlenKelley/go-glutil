package render

import (
   "os"
   "encoding/json"
	// "fmt"
	glm "github.com/Jragonmiris/mathgl"
	glfw "github.com/go-gl/glfw3"
	"reflect"
	"regexp"
	"strconv"
)

type keyEvent struct {
	Key    glfw.Key
	Action glfw.Action
}

type mouseButtonEvent struct {
	Button glfw.MouseButton
	Action glfw.Action
}

type Action func()
type MouseMoveAction func(position, delta glm.Vec2d)
type ScrollAction func(xoff, yoff float64)

type ControlBindings struct {
	keyBindings          map[keyEvent]Action
	mouseButtonBindings  map[mouseButtonEvent]Action
	mouseMovementBinding MouseMoveAction
	scrollBinding  ScrollAction
   
   lastMousePosition glm.Vec2d
   hasLastMousePosition bool
}

func (c *ControlBindings) ResetBindings() {
	c.keyBindings = map[keyEvent]Action{}
	c.mouseButtonBindings = map[mouseButtonEvent]Action{}
	c.mouseMovementBinding = nil
   c.scrollBinding = nil
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

func (c *ControlBindings) BindScroll(action ScrollAction) {
	c.scrollBinding = action
}

func (c *ControlBindings) UnbindScroll() {
	c.scrollBinding = nil
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

func (c *ControlBindings) DoKeyAction(key glfw.Key, keyAction glfw.Action) {
	boundAction, ok := c.FindKeyAction(key, keyAction)
	if ok {
		boundAction()
	}
}

func (c *ControlBindings) DoMouseButtonAction(button glfw.MouseButton, mouseAction glfw.Action) {
	boundAction, ok := c.FindClickAction(button, mouseAction)
	if ok {
		boundAction()
	}
}

func MouseCoord(window *glfw.Window, xpos, ypos float64) glm.Vec2d {
   width, height := window.GetSize()
   return glm.Vec2d{xpos / float64(width), 1 - ypos / float64(height)}
}

func (c *ControlBindings) DoMouseMoveAction(window *glfw.Window, xpos, ypos float64) {
   pos := MouseCoord(window, xpos, ypos)
   boundAction, ok := c.FindMouseMovementAction()
   if ok {
      delta := pos.Sub(c.lastMousePosition)
      if !c.hasLastMousePosition {
         c.hasLastMousePosition = true
         delta = glm.Vec2d{}
      }
      boundAction(pos, delta)
   }
   c.lastMousePosition = pos
}

func (c *ControlBindings) DoScrollAction(xoff, yoff float64) {
   boundAction, ok := c.FindScrollAction()
   if ok {
      boundAction(xoff, yoff)
   }
}

func (c *ControlBindings) FindKeyAction(key glfw.Key, keyAction glfw.Action) (Action, bool) {
	action, ok := c.keyBindings[keyEvent{key, keyAction}]
	return action, ok
}

func (c *ControlBindings) FindClickAction(button glfw.MouseButton, buttonAction glfw.Action) (Action, bool) {
	action, ok := c.mouseButtonBindings[mouseButtonEvent{button, buttonAction}]
	return action, ok

}

func (c *ControlBindings) FindMouseMovementAction() (MouseMoveAction, bool) {
	return c.mouseMovementBinding, c.mouseMovementBinding != nil
}

func (c *ControlBindings) FindScrollAction() (ScrollAction, bool) {
	return c.scrollBinding, c.scrollBinding != nil
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
		return MouseMoveAction(m.Interface().(func(glm.Vec2d, glm.Vec2d)))
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
		stopName := "Stop" + name
		if len(k) == 1 {
			key := glfw.Key(k[0])
			if name == "" {
				c.UnbindKeyPress(key)
			} else {
				startAction := FindActionMethod(receiverValue, name)
				stopAction := FindActionMethod(receiverValue, stopName)
				c.BindKeyPress(key, startAction, stopAction)
			}
		} else if mb.MatchString(k) {
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
		} else if k == "mousemove" {
			if name == "" {
				c.UnbindMouseMovement()
			} else {
				moveAction := FindMouseMoveActionMethod(receiverValue, name)
				c.BindMouseMovement(moveAction)
			}
		}
	}
}

func LoadConfiguration(confFile string, constants interface{}, bindings *ControlBindings, receiver interface{}) error {
   file, err := os.Open(confFile)
   if err == nil {
      defer file.Close()
      decoder := json.NewDecoder(file)
      root := map[string]interface{}{}
      err = decoder.Decode(&root)
      if err != nil { return err }
      if constants, ok := root["constants"]; ok {
         bytes, err := json.Marshal(constants)
         if err != nil { return err }
         err = json.Unmarshal(bytes, &constants)
         if err != nil { return err }
      }
      if controls, ok := root["controls"]; ok {
         sc := make(map[string]string)
         for k, v := range controls.(map[string]interface{}) {
            sc[k] = v.(string)
         }
         bindings.Apply(receiver, sc)
      }
   }
   return err
}

