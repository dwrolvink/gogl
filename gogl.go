package gogl

import (
	"os"
	"log"
	"strings"
	"runtime"
	//"path/filepath"
	"io/ioutil"
	"errors"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)





// ------------------------------------------------------------------------------------------
// [ Init functions ]

/* Inits GL and GLFW */
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
// [ Main functions ]

func Draw(window *glfw.Window, programPtr *Program, data []float32, dataType uint32) {
	// dataType: gl.TRIANGLES

	vaoID, _ := MakeVao(data)		// Recalc vao

	// Clear buffer
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Activate program
	UseProgram((*programPtr).ID)

	// Compile image
	gl.BindVertexArray(uint32(vaoID))
	gl.DrawArrays(dataType, 0, int32(len(data)/3))

	// Handle window events
	glfw.PollEvents()

	// Put buffer that we painted on on the foreground
	window.SwapBuffers()
}

// [ Main functions ]
// ------------------------------------------------------------------------------------------
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

/* Creates a Program, builds shaders, links shaders, and adds program 
   to custom watchlist "LoadedPrograms", which allows us to use ReloadProgram()
   when one of the shaderfiles get modified.
*/
func MakeProgram(programName string, vertexShaderPath string, fragmentShaderPath string) (*Program, error) {
	// Create shaders
	vertexShaderID, err := LoadShader(vertexShaderPath, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	fragmentShaderID, err2 := LoadShader(fragmentShaderPath, gl.FRAGMENT_SHADER)
	if err2 != nil {
		return nil, err2
	}

	// Create program & link shaders
	programID := ProgramID( gl.CreateProgram() )
	AttachShader(programID, vertexShaderID)
	AttachShader(programID, fragmentShaderID)
	LinkProgram(programID)

	// Log error and stop execution if failed
	err = CheckProgramLinkSuccess(programID)
	if err != nil {
		panic(err)
	}

	// After linking, we can delete the shaders
	gl.DeleteShader(uint32(vertexShaderID))
	gl.DeleteShader(uint32(fragmentShaderID))

	// Keep track of the program in a watchlist, so we can update it when the shaders change
	programPtr, ok := LoadedPrograms[programName] 
	if ok == false {
		// Add to the list
		LoadedPrograms[programName] = &Program {
			ID: programID,
			VertexShaderFilePath: vertexShaderPath,
			FragmentShaderFilePath: fragmentShaderPath,
		}
	} else {
		// If it already exists, update the id
		(*programPtr).ID = programID
	}

	log.Printf("Program %s (%d) compiled succesfully. \n", programName, programID)

	return LoadedPrograms[programName], nil
}



// [/ Makers ]
// ------------------------------------------------------------------------------------------
// [ Status checkers ]

func CheckProgramLinkSuccess(programID ProgramID) error{
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

func CheckShaderCompileSuccess(shaderID ShaderID, shaderSource string) error{
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

func AttachShader(programID ProgramID, shaderID ShaderID){
	gl.AttachShader(uint32(programID), uint32(shaderID))
} 

func LinkProgram(programID ProgramID) {
	gl.LinkProgram(uint32(programID))
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

func UseProgram(programID ProgramID) {
	gl.UseProgram(uint32(programID))
}
// [/ Log functions ]
// ------------------------------------------------------------------------------------------
// [ Hotloading Shaders ]

func LoadShader(path string, shaderType uint32) (ShaderID, error){
	shaderFileData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	shaderFileStr := string(shaderFileData)
	shaderID, err := MakeShader(shaderFileStr, shaderType)
	if err != nil {
		return 0, err
	}

	// Add to watchlist if not yet a member
	if shaderIsInWatchList(path) == false {
		// Get Last Modified time
		file, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		// Add to list
		shaderFileInfo := ShaderFileInfo{
			FilePath: path,
			LastModified: file.ModTime(),
		}
		LoadedShaders = append(LoadedShaders, shaderFileInfo)
	}

	return shaderID, nil
}

