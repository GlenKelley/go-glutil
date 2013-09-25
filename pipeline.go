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

func BindBufferData(buffer gl.Uint, target gl.Enum, data interface{}) {
    ptr, size := ArrayPtr(data)
    gl.BindBuffer(target, buffer)
    gl.BufferData(target, size, ptr, gl.STATIC_DRAW)
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

func LoadTexture(filename string) (gl.Uint, error) {
    var texture gl.Uint
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()
    img, _, err := image.Decode(file)
    if err != nil {
        return 0, err
    }
    width, height, format, channelType, pixels := ImageData(img)
    gl.GenTextures(1, &texture)
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
    return texture, nil
}

func LoadShader(shaderType gl.Enum, filename string) (gl.Uint, error){
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return 0, err
    }
    shader := gl.CreateShader(shaderType)
    source := string(bytes)
    gl.ShaderSource(shader, []string{source})
    gl.CompileShader(shader)
    var ok gl.Int
    gl.GetShaderiv(shader, gl.COMPILE_STATUS, &ok)
    if (ok == 0) {
        fmt.Fprintln(os.Stderr, gl.GetShaderInfoLog(shader))
        gl.DeleteShader(shader)
        return 0, errors.New("Failed to compile " + filename + "\n")
    }
    return shader, nil
}

func CreateProgram(vertexShader, fragmentShader gl.Uint) (gl.Uint, error) {
    program := gl.CreateProgram()
    gl.AttachShader(program, vertexShader)
    gl.AttachShader(program, fragmentShader)
    gl.LinkProgram(program)
    var ok gl.Int
    gl.GetProgramiv(program, gl.LINK_STATUS, &ok)
     if (ok == 0) {
        fmt.Fprintln(os.Stderr, gl.GetProgramInfoLog(program))
         gl.DeleteProgram(program)
         return 0, errors.New("Failed to link shader program")
     }
     return program, nil
}
