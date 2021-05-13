package gogl

import (
	"log"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type ProgramID uint32
type Program struct {
	ID                     ProgramID
	ProgramName            string
	VertexShaderFilePath   string
	FragmentShaderFilePath string
}

// Loads the given value as a Uniform1f uniform to be consumed by a shader
func (program *Program) SetFloat(name string, value float32) {
	name_cstr := gl.Str(name + "\x00")
	location := gl.GetUniformLocation(uint32(program.ID), name_cstr)
	gl.Uniform1f(location, value)
}

// Loads the given value as a Uniform1f uniform to be consumed by a shader
func (program *Program) SetInt(name string, value int32) {
	name_cstr := gl.Str(name + "\x00")
	location := gl.GetUniformLocation(uint32(program.ID), name_cstr)
	gl.Uniform1i(location, value)
}

/*
Creates a Program, builds shaders, links shaders, and adds program
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
	programID := ProgramID(gl.CreateProgram())
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
		LoadedPrograms[programName] = &Program{
			ID:                     programID,
			VertexShaderFilePath:   vertexShaderPath,
			FragmentShaderFilePath: fragmentShaderPath,
		}
	} else {
		// If it already exists, update the id
		(*programPtr).ID = programID
	}

	log.Printf("Program %s (%d) compiled succesfully. \n", programName, programID)

	return LoadedPrograms[programName], nil
}
