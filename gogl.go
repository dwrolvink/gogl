package gogl

import (
	"log"
	"runtime"
	"strings"

	//"path/filepath"

	"errors"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// ------------------------------------------------------------------------------------------
// [ Init functions ]

/* Inits GL and GLFW. Creates a window in the process with given dimensions. */
func Init(windowTitle string, width, height int) *glfw.Window {
	runtime.LockOSThread()

	window := InitGlfw(windowTitle, width, height)

	// init OpenGL
	if err := gl.Init(); err != nil {
		panic(err)
	}

	PrintGLVersion()
	PrintGLFWVersion()

	return window
}

/* initializes glfw and returns a Window to use. */
func InitGlfw(windowTitle string, width, height int) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, windowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// [ / Init functions ]
// ------------------------------------------------------------------------------------------
// [ Makers ]

// Creates a generic Buffer Object in GL, returns its ID.
// Can be used both as a VBO and EBO
func GenBuffer(target uint32) BufferID {
	// target: gl.ARRAY_BUFFER
	var id uint32
	gl.GenBuffers(1, &id)
	return BufferID(id)
}

// Creates a VertexArray Object in GL, returns its ID
func GenVertexArray() VAOID {
	var vaoID uint32
	gl.GenVertexArrays(1, &vaoID)
	return VAOID(vaoID)
}

// A slightly more intelligent/go version of gl.BufferData.
// Typical target: gl.ARRAY_BUFFER
// Typical usage: gl.STATIC_DRAW
func BufferDataFloat32(data []float32, target uint32, usage uint32) {
	gl.BufferData(target, 4*len(data), gl.Ptr(data), usage)
}

// A slightly more intelligent/go version of gl.BufferData.
// Typical target: gl.ELEMENT_ARRAY_BUFFER
// Typical usage: gl.STATIC_DRAW
func BufferDataUint32(data []uint32, target uint32, usage uint32) {
	gl.BufferData(target, 4*len(data), gl.Ptr(data), usage)
}

// Creates shadersource, compiles it, and checks for errors in that process.
func MakeShader(shaderSourceCode string, shaderType uint32) (ShaderID, error) {
	// We need to convert the shaderSource from a Go string to
	// a C string. C strings need a null byte at the end, and
	// they need to be freed after they are no longer needed
	shaderSourceCode = shaderSourceCode + "\x00"
	c_shaderSourcePtr, free := gl.Strs(shaderSourceCode) // c_shaderSource is of type **uint8, so a pointer to a pointer

	// Create shader
	shaderId := gl.CreateShader(shaderType)
	gl.ShaderSource(shaderId, 1, c_shaderSourcePtr, nil)

	// Clean up C string
	free()

	// Compile
	gl.CompileShader(shaderId)

	// Check for error
	err := CheckShaderCompileSuccess(ShaderID(shaderId), shaderSourceCode)
	if err != nil {
		return 0, err
	}

	return ShaderID(shaderId), nil
}

// [/ Makers ]
// ------------------------------------------------------------------------------------------
// [ Status checkers ]

// Return an error when errors are found in linking shaders to given program.
func CheckProgramLinkSuccess(programID ProgramID) error {
	var success int32
	gl.GetProgramiv(uint32(programID), gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		// Set log length
		var logLength int32
		gl.GetShaderiv(uint32(programID), gl.INFO_LOG_LENGTH, &logLength)

		// Make log variable with correct length
		log := strings.Repeat("\x00", int(logLength+1))

		// Fetch log data (put it in log)
		gl.GetShaderInfoLog(uint32(programID), logLength, nil, gl.Str(log))

		return errors.New("failed to link program: \n" + log)
	}
	return nil
}

// Return an error when errors are found in compiling given shader.
func CheckShaderCompileSuccess(shaderID ShaderID, shaderSource string) error {
	var success int32
	gl.GetShaderiv(uint32(shaderID), gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		// Set log length
		var logLength int32
		gl.GetShaderiv(uint32(shaderID), gl.INFO_LOG_LENGTH, &logLength)

		// Make log variable with correct length
		log := strings.Repeat("\x00", int(logLength+1))

		// Fetch log data (put it in log)
		gl.GetShaderInfoLog(uint32(shaderID), logLength, nil, gl.Str(log))

		return errors.New("failed to compile " + shaderSource + ", " + log)
	}
	return nil
}

// [/ Status checkers ]
// ------------------------------------------------------------------------------------------
// [ Type-Aware Wrappers ]

// Simple type aware wrapper for gl.AttachShader
func AttachShader(programID ProgramID, shaderID ShaderID) {
	gl.AttachShader(uint32(programID), uint32(shaderID))
}

// Simple type aware wrapper for gl.LinkProgram
func LinkProgram(programID ProgramID) {
	gl.LinkProgram(uint32(programID))
}

// Simple type aware wrapper for gl.UseProgram
func UseProgram(programID ProgramID) {
	gl.UseProgram(uint32(programID))
}

// [/ Type-Aware Wrappers ]
// ------------------------------------------------------------------------------------------
// [ Log functions ]

func GetVersion() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}
func PrintGLVersion() {
	log.Println("OpenGL version", GetVersion())
}

func PrintGLFWVersion() {
	major, minor, rev := glfw.GetVersion()
	log.Printf("GLFW version: %d.%d.%d\n", major, minor, rev)
}

// [/ Log functions ]
// ------------------------------------------------------------------------------------------
