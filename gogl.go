package gogl

import (
	"log"
	"strings"
	"runtime"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type ShaderID uint32
type ProgramID uint32
type VAOID uint32
type VBOID uint32


// [ Init functions ]
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

// initGlfw initializes glfw and returns a Window to use.
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

// [ Main functions ]
func Draw(window *glfw.Window, programID ProgramID, data []float32, dataType uint32) {
	// dataType: gl.TRIANGLES

	vaoID, _ := MakeVao(data)		// Recalc vao

	// Clear buffer
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Activate program
	UseProgram(programID)

	// Compile image
	gl.BindVertexArray(uint32(vaoID))
	gl.DrawArrays(dataType, 0, int32(len(data)/3))

	// Handle window events
	glfw.PollEvents()

	// Put buffer that we painted on on the foreground
	window.SwapBuffers()
}

// [ Main functions ]

// [ Makers ]

func MakeVao(data []float32) (VAOID, VBOID) {
	// Create Buffer object
	vboID := GenBindBuffer(gl.ARRAY_BUFFER)

	// Buffer our data
	BufferData(data, gl.ARRAY_BUFFER, gl.STATIC_DRAW)

	// Create Vertex Array Object (VAO)
	vaoID := GenBindVertexArray()

	// Configure VAO
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(0)

	// Unbind 
	UnbindVertexArray()

	return VAOID(vaoID), VBOID(vboID)
}

func GenBindBuffer(target uint32) VBOID {
	// target: gl.ARRAY_BUFFER
	var vboID uint32
	gl.GenBuffers(1, &vboID)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboID)

	return VBOID(vboID)
}

func GenBindVertexArray() VAOID {
	var vaoID uint32
	gl.GenVertexArrays(1, &vaoID)
	gl.BindVertexArray(vaoID)	
	return VAOID(vaoID)
}

func UnbindVertexArray() {
	gl.BindVertexArray(0)
}

func BufferData(data []float32, target uint32, usage uint32) {
	// target: gl.ARRAY_BUFFER
	// usage: gl.STATIC_DRAW
	gl.BufferData(target, 4*len(data), gl.Ptr(data), usage)
}


func MakeShader(shaderSourceCode string, shaderType uint32) ShaderID{
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

	// Log error and stop execution if failed
	CheckShaderCompileSuccess(ShaderID(shaderId), shaderSourceCode)

	return ShaderID(shaderId)
}

func MakeProgram(vertexShaderID ShaderID, fragmentShaderID ShaderID) ProgramID {
	programID := ProgramID( gl.CreateProgram() )

	AttachShader(programID, vertexShaderID)
	AttachShader(programID, fragmentShaderID)
	LinkProgram(programID)

	// Log error and stop execution if failed
	CheckProgramLinkSuccess(programID)

	// After linking, we can delete the shaders
	gl.DeleteShader(uint32(vertexShaderID))
	gl.DeleteShader(uint32(fragmentShaderID))

	return programID
}

// [/ Makers ]
// [ Status checkers ]

func CheckProgramLinkSuccess(programID ProgramID){
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

		panic("failed to link program: \n" + log)
	}	
}

func CheckShaderCompileSuccess(shaderID ShaderID, shaderSource string) {
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

		panic("failed to compile " + shaderSource + ", " + log)
	}
}

// [/ Status checkers ]
// [ Type-Aware Wrappers ]

func AttachShader(programID ProgramID, shaderID ShaderID){
	gl.AttachShader(uint32(programID), uint32(shaderID))
} 

func LinkProgram(programID ProgramID) {
	gl.LinkProgram(uint32(programID))
}

// [/ Type-Aware Wrappers ]

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

func UseProgram(programID ProgramID) {
	gl.UseProgram(uint32(programID))
}
// [/ Log functions ]