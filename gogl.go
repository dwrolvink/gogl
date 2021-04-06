package gogl

import (
	"os"
	"log"
	"strings"
	"runtime"
	//"path/filepath"
	"io/ioutil"
	"errors"
	"time"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	// Watch list for hotloading shaders
	LoadedShaders []ShaderFileInfo
	LoadedPrograms []Program
)

type ShaderFileInfo struct {
	FilePath string
	LastModified time.Time
}

type Program struct {
	ID ProgramID
	ProgramName string
	VertexShaderFilePath string
	FragmentShaderFilePath string
}

type ShaderID uint32
type ProgramID uint32

type VAOID uint32
type VBOID uint32

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
func MakeProgram(programName string, vertexShaderPath string, fragmentShaderPath string) (ProgramID, error) {
	// Create shaders
	vertexShaderID, err := LoadShader(vertexShaderPath, gl.VERTEX_SHADER)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	fragmentShaderID, err2 := LoadShader(fragmentShaderPath, gl.FRAGMENT_SHADER)
	if err2 != nil {
		log.Println(err2)
		return 0, err2
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

	// Add program to watchlist so we can recreate it when one of the shaders
	// needs to be recompiled
	if programIsInWatchList(programName) == false {
		LoadedPrograms = append(LoadedPrograms, Program{
			ID: programID,
			ProgramName: programName,
			VertexShaderFilePath: vertexShaderPath,
			FragmentShaderFilePath: fragmentShaderPath,
		})
	}

	log.Printf("Program %s (%d) compiled succesfully. \n", programName, programID)

	return programID, nil
}

/* Hotloading shader interface for program */
func ReloadProgram(programName string, changedShaderFiles []string) (ProgramID, error){
	var currentProgramID ProgramID
	var newProgramID ProgramID
	var vertexShaderPath, fragmentShaderPath string
	
	// Find stored Program definition, and fetch shader file names
	programFound := false
	programIndex := 0
	for i := range LoadedPrograms {
		if LoadedPrograms[i].ProgramName == programName {
			programName        = LoadedPrograms[i].ProgramName
			currentProgramID   = LoadedPrograms[i].ID
			vertexShaderPath   = LoadedPrograms[i].VertexShaderFilePath
			fragmentShaderPath = LoadedPrograms[i].FragmentShaderFilePath
			programFound       = true
			programIndex       = i
		}
	}
	if programFound == false {
		return 0, errors.New("Could not find program " + programName)
	}

	// Check if any changed files are related to our program
	needsRebuilding := false
	for i := range changedShaderFiles {
		if changedShaderFiles[i] == vertexShaderPath || changedShaderFiles[i] == fragmentShaderPath {
			needsRebuilding = true
			log.Printf("Program %s (%d) needs rebuiding", programName, currentProgramID)
			break
		}
	}

	// Rebuild
	if needsRebuilding {
		var err error
		newProgramID, err = MakeProgram(programName, vertexShaderPath, fragmentShaderPath)
		if err != nil {
			log.Printf("Failed to build program %s, continuing to use old compilation (%d). \n", programName, currentProgramID)
			return currentProgramID, err
		}

		// Update programID in watchlist
		LoadedPrograms[programIndex].ID = newProgramID
		return newProgramID, nil
	}

	// Nothing has changed, return current ID
	return currentProgramID, nil
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

func shaderIsInWatchList(path string) bool {
	for _, shaderFileInfo := range LoadedShaders {
		if shaderFileInfo.FilePath == path {
			return true
		}
	}
	return false
}

func programIsInWatchList(programName string) bool {
	for i := range LoadedPrograms {
		if LoadedPrograms[i].ProgramName == programName {
			return true
		}
	}
	return false
}

func GetChangedShaderFiles() []string{
	changedFiles := []string{}
	for i := range LoadedShaders {
		file, err := os.Stat(LoadedShaders[i].FilePath)
		if err != nil {
			panic(err)
		}
		// Check if the file has been changed since last import
		changed := !file.ModTime().Equal(LoadedShaders[i].LastModified)
		if changed {
			log.Printf("Shader %s has changed! \n", LoadedShaders[i].FilePath)
			// Update LastModified time
			LoadedShaders[i].LastModified = file.ModTime()
			// Add to output
			changedFiles = append(changedFiles, LoadedShaders[i].FilePath)
		}
	}

	return changedFiles
}