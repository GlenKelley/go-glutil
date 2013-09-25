package render

import (
    "os"
    "fmt"
    "image"
    "errors"
    "unsafe"
    "reflect"
    "io/ioutil"
    _ "image/png"
    _ "image/jpeg"
    gl "github.com/GlenKelley/go-gl32"
)

func ArrayPtr(data interface{}) (gl.Pointer, gl.Sizeiptr) {
    var size gl.Sizeiptr
    var ptr gl.Pointer 
    switch data := data.(type) {
    case []gl.Float:
        var v gl.Float
        size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
        ptr = gl.Pointer(&data[0])
    case []gl.Ushort:
        var v gl.Ushort
        size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
        ptr = gl.Pointer(&data[0])
    case []gl.Int:
        var v gl.Int
        size = gl.Sizeiptr(len(data) * int(unsafe.Sizeof(v)))
        ptr = gl.Pointer(&data[0])
    default:
        panic("unknown data type")
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
    return 0,0,gl.RGB,gl.UNSIGNED_BYTE,nil
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
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S,     gl.CLAMP_TO_EDGE)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T,     gl.CLAMP_TO_EDGE)
    gl.TexImage2D(
          gl.TEXTURE_2D, 0,           /* target, level of detail */
          gl.RGB8,                    /* internal format */
          width, height, 0,           /* width, height, border */
          format, channelType,          /* external format, type */
          pixels,                      /* pixels */
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
    if (ok == 0) {
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
     if (ok == 0) {
        fmt.Fprintln(os.Stderr, gl.GetProgramInfoLog(program))
         gl.DeleteProgram(program)
         return errors.New("Failed to link shader program")
     }
     return nil
}

var bufferType          = reflect.TypeOf(gl.Buffer(0))
var textureType         = reflect.TypeOf(gl.Texture(0))
var vaoType             = reflect.TypeOf(gl.VertexArrayObject(0))
var vertexShaderType    = reflect.TypeOf(gl.VertexShader(0))
var fragmentShaderType  = reflect.TypeOf(gl.FragmentShader(0))
var programType         = reflect.TypeOf(gl.Program(0))
var uniformLocationType = reflect.TypeOf(gl.UniformLocation(0))
var attributeLocationType = reflect.TypeOf(gl.AttributeLocation(0))


func Bind (bindings interface{}) {
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
    // VertexBuffer gl.Buffer
    // ElementBuffer gl.Buffer
    // Tex [2]gl.Texture
    // VAO gl.VertexArrayObject    
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
            case attributeLocationType:
                location := gl.GetAttribLocation(program, name)
                field.Set(reflect.ValueOf(location))
            }
        }
    }
}

func AttachTexture(location gl.UniformLocation, textureEnum gl.Enum, target gl.Enum, texture gl.Texture) {
    gl.ActiveTexture(textureEnum)
    gl.BindTexture(target, texture)
    gl.Uniform1i(location, gl.Int(textureEnum - gl.TEXTURE0))
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
        panic(err)
    }
}
