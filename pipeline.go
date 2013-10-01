package render

import (
	"os"
	"fmt"
   "math"
	"image"
	"errors"
	"unsafe"
	"reflect"
	"io/ioutil"
	_ "image/png"
	_ "image/jpeg"
	"runtime/debug"
	glm "github.com/Jragonmiris/mathgl"
	gl "github.com/GlenKelley/go-gl/gl32"
	collada "github.com/GlenKelley/go-collada"
)

type Color [4]gl.Float
type ColorPallet []Color

var (
	Red   = Color{1, 0, 0, 1}
	Green = Color{0, 1, 0, 1}
	Blue  = Color{0, 0, 1, 1}

	Black = Color{0, 0, 0, 1}
	White = Color{1, 1, 1, 1}

	Orange  = Color{1, 0.5, 0, 1}
	Pink    = Color{1, 0, 0.5, 1}
	Lime    = Color{0.5, 1, 0, 1}
	Aqua    = Color{0, 1, 0.5, 1}
	Purple  = Color{0.5, 0, 1, 1}
	SkyBlue = Color{0, 0.5, 1, 1}

	Yellow  = Color{1, 1, 0, 1}
	Cyan    = Color{0, 1, 1, 1}
	Magenta = Color{1, 0, 1, 1}

	SoftRed   = Color{0.8, 0.2, 0.2, 1}
	SoftGreen = Color{0.2, 0.8, 0.2, 1}
	SoftBlue  = Color{0.2, 0.2, 0.8, 1}

	SoftWhite = Color{0.8, 0.8, 0.8, 1}
	SoftBlack = Color{0.2, 0.2, 0.2, 1}

	DebugPallet = ColorPallet{
		SoftRed, SoftGreen, SoftBlue, Orange, Pink, Lime, Aqua, Purple, Yellow, Cyan, Magenta,
	}
   
)

func ToVec3D(v glm.Vec4d) glm.Vec3d {
   return glm.Vec3d{v[0], v[1], v[2]}
}

func ToHomogVec4D(v glm.Vec3d) glm.Vec4d {
   return glm.Vec4d{v[0], v[1], v[2], 0.0}
}

func (pallet ColorPallet) Pick(n int) *Color {
	return &pallet[n%len(pallet)]
}

type ShaderLibrary struct {
	FragmentShaders map[string]gl.FragmentShader
	VertexShaders   map[string]gl.VertexShader
	Programs        map[string]gl.Program
}

func NewShaderLibrary() ShaderLibrary {
	return ShaderLibrary{
		make(map[string]gl.FragmentShader),
		make(map[string]gl.VertexShader),
		make(map[string]gl.Program),
	}
}

func (lib *ShaderLibrary) LoadFragmentShader(tag, filename string) {
	_, ok := lib.FragmentShaders[tag]
	if !ok {
		shader := gl.FragmentShader(gl.CreateShader(gl.FRAGMENT_SHADER))
		err := LoadFragmentShaderSource(shader, filename)
		if err != nil {
			panic(err)
		}
		lib.FragmentShaders[tag] = shader
	}
}

func (lib *ShaderLibrary) LoadVertexShader(tag, filename string) {
	_, ok := lib.VertexShaders[tag]
	if !ok {
		shader := gl.VertexShader(gl.CreateShader(gl.VERTEX_SHADER))
		err := LoadVertexShaderSource(shader, filename)
		if err != nil {
			panic(err)
		}
		lib.VertexShaders[tag] = shader
	}
}

func (lib *ShaderLibrary) LoadProgram(tag, vsfilename, fsfilename string) {
	vtag := tag + "_vs"
	ftag := tag + "_fs"
	lib.LoadVertexShader(vtag, vsfilename)
	lib.LoadFragmentShader(ftag, fsfilename)
	program := gl.CreateProgram()
	err := LoadProgram(program, lib.VertexShaders[vtag], lib.FragmentShaders[ftag])
	if err != nil {
		panic(err)
	}
	_, ok := lib.Programs[tag]
	if !ok {
		lib.Programs[tag] = program
	} else {
		panic("program: '" + tag + "' already defined")
	}
}

func (lib *ShaderLibrary) BindProgramLocations(tag string, obj interface{}) {
	program, ok := lib.GetProgram(tag)
	if ok {
		BindProgramLocations(program, obj)
	}
}

func (lib *ShaderLibrary) UseProgram(tag string) {
	program, ok := lib.GetProgram(tag)
	if ok {
		gl.UseProgram(program)
	}
}

func (lib *ShaderLibrary) GetProgram(tag string) (gl.Program, bool) {
	program, ok := lib.Programs[tag]
	return program, ok
}

func MatArray(d glm.Mat4d) *gl.Float {
   n := len(d)
   f := make([]gl.Float, n)
   for i := 0; i < n; i++ {
      f[i] = gl.Float(d[i])
   }
   return &f[0]
}

func ArrayPtr(data interface{}) (gl.Pointer, gl.Sizeiptr) {
	var size gl.Sizeiptr
	var ptr gl.Pointer
	switch data := data.(type) {
	case []float64:
		duplicate := make([]gl.Float, len(data))
		for i, v := range data {
			duplicate[i] = gl.Float(v)
		}
		ptr, size = ArrayPtr(duplicate)
	case []float32:
		if len(data) == 0 {
			size = 0
			ptr = gl.Pointer(nil)
		} else {
			var v float32
			size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
			ptr = gl.Pointer(&data[0])
		}
	case []gl.Float:
		if len(data) == 0 {
			size = 0
			ptr = gl.Pointer(nil)
		} else {
			var v gl.Float
			size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
			ptr = gl.Pointer(&data[0])
		}
	case []int16:
		if len(data) == 0 {
			size = 0
			ptr = gl.Pointer(nil)
		} else {
			var v int16
			size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
			ptr = gl.Pointer(&data[0])
		}
	case []gl.Ushort:
		if len(data) == 0 {
			size = 0
			ptr = gl.Pointer(nil)
		} else {
			var v gl.Ushort
			size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
			ptr = gl.Pointer(&data[0])
		}
	case []int:
		duplicate := make([]gl.Ushort, len(data))
		for i, v := range data {
			duplicate[i] = gl.Ushort(v)
		}
		ptr, size = ArrayPtr(duplicate)
	case []gl.Int:
		if len(data) == 0 {
			size = 0
			ptr = gl.Pointer(nil)
		} else {
			var v gl.Int
			size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
			ptr = gl.Pointer(&data[0])
		}
	default:
		panic(fmt.Sprintln("unknown data type:", reflect.TypeOf(data)))
	}
	return ptr, size
}

func BindArrayData(buffer gl.Buffer, data interface{}) {
	ptr, size := ArrayPtr(data)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, size, ptr, gl.STATIC_DRAW)
}

func ImageData(img image.Image) (gl.Sizei, gl.Sizei, gl.Enum, gl.Enum, gl.Pointer) {
	switch img := img.(type) {
	case *image.NRGBA:
		return gl.Sizei(img.Rect.Dx()), gl.Sizei(img.Rect.Dy()), gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(&img.Pix[0])
	case *image.RGBA:
		return gl.Sizei(img.Rect.Dx()), gl.Sizei(img.Rect.Dy()), gl.RGBA, gl.UNSIGNED_BYTE, gl.Pointer(&img.Pix[0])
	default:
		panic(reflect.TypeOf(img))
	}
	return 0, 0, gl.RGB, gl.UNSIGNED_BYTE, nil
}

func LoadTexture(texture gl.Texture, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	width, height, format, channelType, pixels := ImageData(img)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, /* target, level of detail */
		gl.RGB8,          /* internal format */
		width, height, 0, /* width, height, border */
		format, channelType, /* external format, type */
		pixels, /* pixels */
	)
	return nil
}

func LoadVertexShaderSource(shader gl.VertexShader, filename string) error {
	return LoadShader(gl.Uint(shader), filename)
}

func LoadFragmentShaderSource(shader gl.FragmentShader, filename string) error {
	return LoadShader(gl.Uint(shader), filename)
}

func LoadShader(shader gl.Uint, filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	source := string(bytes)
	gl.ShaderSource(shader, []string{source})
	gl.CompileShader(shader)
	var ok gl.Int
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &ok)
	if ok == 0 {
		fmt.Fprintln(os.Stderr, gl.GetShaderInfoLog(shader))
		gl.DeleteShader(shader)
		return errors.New("Failed to compile " + filename + "\n")
	}
	return nil
}

func LoadProgram(program gl.Program, vertexShader gl.VertexShader, fragmentShader gl.FragmentShader) error {
	gl.AttachShader(program, gl.Uint(vertexShader))
	gl.AttachShader(program, gl.Uint(fragmentShader))
	gl.LinkProgram(program)
	var ok gl.Int
	gl.GetProgramiv(program, gl.LINK_STATUS, &ok)
	if ok == 0 {
		fmt.Fprintln(os.Stderr, gl.GetProgramInfoLog(program))
		gl.DeleteProgram(program)
		return errors.New("Failed to link shader program")
	}
	return nil
}

var bufferType = reflect.TypeOf(gl.Buffer(0))
var textureType = reflect.TypeOf(gl.Texture(0))
var vaoType = reflect.TypeOf(gl.VertexArrayObject(0))
var vertexShaderType = reflect.TypeOf(gl.VertexShader(0))
var fragmentShaderType = reflect.TypeOf(gl.FragmentShader(0))
var programType = reflect.TypeOf(gl.Program(0))
var uniformLocationType = reflect.TypeOf(gl.UniformLocation(0))
var attributeLocationType = reflect.TypeOf(gl.AttributeLocation(0))

func Bind(bindings interface{}) {
	value := reflect.ValueOf(bindings).Elem()
	n := value.NumField()

	for i := 0; i < n; i++ {
		field := value.Field(i)
		switch field.Type() {
		case bufferType:
			buffer := gl.GenBuffer()
			field.Set(reflect.ValueOf(buffer))
		case textureType:
			texture := gl.GenTexture()
			field.Set(reflect.ValueOf(texture))
		case vaoType:
			vao := gl.GenVertexArray()
			field.Set(reflect.ValueOf(vao))
		case vertexShaderType:
			shader := gl.VertexShader(gl.CreateShader(gl.VERTEX_SHADER))
			field.Set(reflect.ValueOf(shader))
		case fragmentShaderType:
			shader := gl.FragmentShader(gl.CreateShader(gl.FRAGMENT_SHADER))
			field.Set(reflect.ValueOf(shader))
		case programType:
			program := gl.CreateProgram()
			field.Set(reflect.ValueOf(program))
		}
	}
}

func BindProgramLocations(program gl.Program, bindings interface{}) {
	value := reflect.ValueOf(bindings).Elem()
	n := value.NumField()
	for i := 0; i < n; i++ {
		field := value.Field(i)
		name := value.Type().Field(i).Tag.Get("gl")
		if name != "" {
			switch field.Type() {
			case uniformLocationType:
				location := gl.GetUniformLocation(program, name)
				field.Set(reflect.ValueOf(location))
				PanicOnError()
			case attributeLocationType:
				location := gl.GetAttribLocation(program, name)
				field.Set(reflect.ValueOf(location))
				PanicOnError()
			}
		}
	}
}

func AttachTexture(location gl.UniformLocation, textureEnum gl.Enum, target gl.Enum, texture gl.Texture) {
	gl.ActiveTexture(textureEnum)
	gl.BindTexture(target, texture)
	gl.Uniform1i(location, gl.Int(textureEnum-gl.TEXTURE0))
}

func PanicOnError() {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		switch err {
		case gl.INVALID_ENUM:
			fmt.Println("INVALID_ENUM")
		case gl.INVALID_VALUE:
			fmt.Println("INVALID_VALUE")
		case gl.INVALID_OPERATION:
			fmt.Println("GL_INVALID_OPERATION")
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			fmt.Println("GL_INVALID_FRAMEBUFFER_OPERATION")
		case gl.OUT_OF_MEMORY:
			fmt.Println("GL_OUT_OF_MEMORY")
		default:
			fmt.Println("other error", err)
		}
		debug.PrintStack()
		panic(err)
	}
}

type StencilOp struct {
}

var Stencil StencilOp

func (s *StencilOp) Enable() *StencilOp {
	gl.Enable(gl.STENCIL_TEST)
	s.Draw()
	s.Mask(0)
	s.Keep()
	s.Depth()
	s.DepthMask()
	s.DepthLT()
	return s
}

func (s *StencilOp) Disable() *StencilOp {
	gl.Disable(gl.STENCIL_TEST)
	s.Draw()
	s.Depth()
	s.DepthMask()
	s.DepthLT()
	return s
}

func (s *StencilOp) Draw() *StencilOp {
	gl.ColorMask(gl.TRUE, gl.TRUE, gl.TRUE, gl.TRUE)
	return s
}

func (s *StencilOp) DepthLE() *StencilOp {
	gl.DepthFunc(gl.LEQUAL)
	return s
}

func (s *StencilOp) DepthLT() *StencilOp {
	gl.DepthFunc(gl.LESS)
	return s
}

func (s *StencilOp) DepthAlways() *StencilOp {
	gl.DepthFunc(gl.ALWAYS)
	return s
}

func (s *StencilOp) NoDraw() *StencilOp {
	gl.ColorMask(gl.FALSE, gl.FALSE, gl.FALSE, gl.FALSE)
	return s
}

func (s *StencilOp) Mask(level int) *StencilOp {
	gl.StencilFunc(gl.EQUAL, gl.Int(level), ^gl.Uint(0))
	return s
}

func (s *StencilOp) Unmask(level int) *StencilOp {
	gl.StencilFunc(gl.ALWAYS, gl.Int(level), ^gl.Uint(0))
	return s
}

func (s *StencilOp) Depth() *StencilOp {
	gl.Enable(gl.DEPTH_TEST)
	return s
}

func (s *StencilOp) NoDepth() *StencilOp {
	gl.Disable(gl.DEPTH_TEST)
	return s
}

func (s *StencilOp) DepthMask() *StencilOp {
	gl.DepthMask(gl.TRUE)
	return s
}

func (s *StencilOp) NoDepthMask() *StencilOp {
	gl.DepthMask(gl.FALSE)
	return s
}

func (s *StencilOp) Keep() *StencilOp {
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	return s
}

func (s *StencilOp) Replace() *StencilOp {
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.REPLACE)
	return s
}

func (s *StencilOp) Increment() *StencilOp {
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.INCR)
	return s
}

func (s *StencilOp) Decrement() *StencilOp {
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.DECR)
	return s
}

type Model struct {
	Transform glm.Mat4d
	Geometry  []*Geometry
	Children  []*Model
}

type Geometry struct {
	VertexBuffer gl.Buffer
	NormalBuffer gl.Buffer
	Elements     []*DrawElements
}

type DrawElements struct {
	Buffer   gl.Buffer
	DrawType gl.Enum
	Count    int
}

func NewSingleModel(verticies, normals []float64, elements []int16, drawType gl.Enum, transform glm.Mat4d) *Model {
	drawElements := []*DrawElements{NewDrawElements(elements, drawType)}
	geometries := []*Geometry{NewGeometry(verticies, normals, drawElements)}
	model := NewModel([]*Model{}, geometries, transform)
	return model
}

func NewModel(children []*Model, geometry []*Geometry, transform glm.Mat4d) *Model {
	return &Model{
		transform,
		geometry,
		children,
	}
}

func EmptyModel() *Model {
	return &Model{
		glm.Ident4d(),
		[]*Geometry{},
		[]*Model{},
	}
}

func (model *Model) AddChild(child *Model) {
	model.Children = append(model.Children, child)
}

func (model *Model) AddGeometry(geometry ...*Geometry) {
	model.Geometry = append(model.Geometry, geometry...)
}

func NewGeometry(verticies, normals []float64, elements []*DrawElements) *Geometry {
	var vertexBuffer gl.Buffer = 0
	var normalBuffer gl.Buffer = 0
	if len(verticies) > 0 {
		vertexBuffer = gl.GenBuffer()
		BindArrayData(vertexBuffer, verticies)
	}
	if len(normals) > 0 {
		normalBuffer = gl.GenBuffer()
		BindArrayData(normalBuffer, normals)
	}
	return &Geometry{
		vertexBuffer,
		normalBuffer,
		elements,
	}
}

func (geometry *Geometry) AddDrawElements(drawElements *DrawElements) {
	geometry.Elements = append(geometry.Elements, drawElements)
}

func NewDrawElements(elements []int16, drawType gl.Enum) *DrawElements {
	count := len(elements)
	if count > 0 {
		buffer := gl.GenBuffer()
		BindArrayData(buffer, elements)
		return &DrawElements{
			buffer,
			drawType,
			count,
		}
	}
	return nil
}

func MakeElements(elementMap map[gl.Enum][]int16) []*DrawElements {
	elements := make([]*DrawElements, 0, len(elementMap))
	for k, v := range elementMap {
		drawElements := NewDrawElements(v, k)
		elements = append(elements, drawElements)
	}
	return elements
}


func RotationComponent(m glm.Mat4d) glm.Mat3d {
   m2 := glm.Ident3d()
   j := 0
   for i := 0; i < 3; i++ {
      v := glm.Vec4d{}
      v[i] = 1
      vt := m.Mul4x1(v)
      for k := 0; k < 3; k++ {
         m2[j] = vt[k]
         j++
      }
   }
   return m2
}


func Quaternion(m glm.Mat3d) glm.Quatd {
   q := glm.Quatd{}
   m00,m01,m02 := m[0],m[1],m[2]
   m10,m11,m12 := m[3],m[4],m[5]
   m20,m21,m22 := m[6],m[7],m[8]
   
   q.W    = math.Sqrt(math.Max(0,1 + m00 + m11 + m22)) / 2
   q.V[0] = math.Copysign(math.Sqrt(math.Max(0, 1 + m00 - m11 - m22)) / 2, m12 - m21)
   q.V[1] = math.Copysign(math.Sqrt(math.Max(0, 1 - m00 + m11 - m22)) / 2, m20 - m02)
   q.V[2] = math.Copysign(math.Sqrt(math.Max(0, 1 - m00 - m11 + m22)) / 2, m01 - m10)
   return q
}

func LoadSceneAsModel(filename string) (*Model, error) {
   doc, err := collada.LoadDocument(filename)
   if err != nil {
      return nil, err
   }
   index, err := NewIndex(doc)
   if err != nil {
      return nil, err
   }   
   model := EmptyModel()
   switch doc.Asset.UpAxis {
   case collada.Xup:
   case collada.Yup:
   case collada.Zup:
      model.Transform = glm.HomogRotate3DXd(-90).Mul4(glm.HomogRotate3DZd(90))
   }
   
   geometryTemplates := make(map[collada.Id][]*Geometry)
   for id, mesh := range index.Mesh {
      geoms := make([]*Geometry, 0)
      for _, pl := range mesh.Polylist {
         elements := make([]*DrawElements, 0)
         drawElements := NewDrawElements(pl.TriangleElements, gl.TRIANGLES)
         if drawElements != nil {
            elements = append(elements, drawElements)
         }
         geometry := NewGeometry(pl.VertexData, pl.NormalData, elements)
         geoms = append(geoms, geometry)  
      }
      if len(geoms) > 0 {
         geometryTemplates[id] = geoms
      }
   }
   for _, node := range index.VisualScene.Node {
      child, ok := LoadModel(index, node, geometryTemplates)
      if ok {
         model.AddChild(child)
      }
   }
   return model, nil
}

func LoadModel(index *Index, node *collada.Node, geometryTemplates map[collada.Id][]*Geometry) (*Model, bool) {
   transform, ok := index.Transforms[node.Id]
   if !ok {
      panic("no transform for id", node.ID)
   }
   geoms := make([]*Geometry, 0)
   children := make([]*Model, 0)
   for _, geoinstance := range node.InstanceGeometry {
      geoid, _ := geoinstance.Url.Id()
      geoms = append(geoms, geometryTemplates[geoid]...)
   }
   for _, childNode := range node.Node {
      child, ok := LoadModel(index, childNode, geometryTemplates)
      if ok {
         children = append(children, child)
      }
   }
   model := NewModel(children, geoms, transform)
   return model, len(geoms) > 0 || len(children) > 0
}

func DrawModel(mv glm.Mat4d, model *Model, modelview gl.UniformLocation, vertexAttribute gl.AttributeLocation, vao gl.VertexArrayObject) {
   mv2 := mv.Mul4(model.Transform)
   gl.UniformMatrix4fv(modelview, 1, gl.FALSE, MatArray(mv2))
   for _, geo := range model.Geometry {
      DrawGeometry(geo, vertexAttribute, vao)
   }
   for _, child := range model.Children {
      DrawModel(mv2, child, modelview, vertexAttribute, vao)
   }
}

func DrawGeometry(geo *Geometry, vertexAttribute gl.AttributeLocation, vao gl.VertexArrayObject) {
   gl.BindBuffer(gl.ARRAY_BUFFER, geo.VertexBuffer)
   gl.BindVertexArray(vao)
   gl.VertexAttribPointer(vertexAttribute, 3, gl.FLOAT, gl.FALSE, 12, nil)
   gl.EnableVertexAttribArray(vertexAttribute)
   for _, elem := range geo.Elements {
      gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, elem.Buffer)
      gl.DrawElements(elem.DrawType, gl.Sizei(elem.Count), gl.UNSIGNED_SHORT, nil)
      PanicOnError()         
   }
   gl.DisableVertexAttribArray(vertexAttribute)
}

func Grid(n int) *Geometry {
   vs := make([]float64, 0, n*12)
   ns := make([]float64, 0, n*12)
   es := make([]int16, 0, n*4)
   ec := int16(0)
   for i := -n; i <= n; i++ {
      nd := float64(n)
      id := float64(i)
      vs = append(vs, -nd, 0,  id,    nd, 0,  id,   id, 0, -nd,  id, 0,  nd)
      ns = append(ns, 0,1,0,0,1,0,0,1,0,0,1,0)
      es = append(es, ec, ec+1, ec+2, ec+3)
      ec += 4
   }
   return NewGeometry(vs, ns, []*DrawElements{NewDrawElements(es, gl.LINES)})
}