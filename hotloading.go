package gogl

/*	
	HOTLOADING

	This file stores all the code that is used exclusively for hotloading shaders.
	This means that we can change the shader definitions while the program is running, 
	and it will load them in upon saving - without the need of recompiling the entire
	program.

	When the compilation of one or more of the shaders fails, the programs using them will 
	continue running on the previous shader compilations. An error will be logged in the
	terminal.

	Note that the other code in gogl.go (like MakeProgram()) also uses components from this 
	file; notably to register newly created Programs and shader files, so that they are 
	automatically tracked and updated upon change.
*/

import (
	"time"
	"os"
	"io/ioutil"
	"log"
	"github.com/go-gl/gl/v4.5-core/gl"
)

var (
	// Vars to keep track of what we've loaded, 
	// so that we can rebuild upon shader change
	LoadedShaders []ShaderFileInfo					// used by GetChangedShaderFiles()
	LoadedPrograms = make(map[string]*Program)		// used by HotloadShaders()
)

type ShaderFileInfo struct {
	FilePath string
	LastModified time.Time
}

// <toplevel function>
func HotloadShaders(){
	// Check all shader files for changes (by LastModified date)
	// This will update LastModified in LoadedShaders for each
	// ShaderFileInfo struct, and thus will only work once per change. 
	changedShaderFiles := GetChangedShaderFiles()

	// If there are changed files, check for each program if it needs to be recompiled,
	// and if so, recompile it. 
	if len(changedShaderFiles) > 0 {
		for programName, program := range LoadedPrograms {
			err := ReloadProgram(programName, program, changedShaderFiles)
			if err != nil {
				// On error, we just resume using the previous compilation.
				// The only way the user will know hotloading has failed is via
				// the error in the terminal output
				log.Println(err)
			}
		}
	}	
}

func ReloadProgram(programName string, storedProgramPtr *Program, changedShaderFiles []string) error{

	// Check if any changed files are related to our program
	needsRebuilding := false
	for i := range changedShaderFiles {
		if changedShaderFiles[i] == (*storedProgramPtr).VertexShaderFilePath || 
		   changedShaderFiles[i] == (*storedProgramPtr).FragmentShaderFilePath {
			needsRebuilding = true
			log.Printf("Program %s (%d) needs rebuiding", programName, (*storedProgramPtr).ID)
			break
		}
	}

	// Rebuild
	if needsRebuilding {
		// Save old id, so we can remove the old program when the new one is compiled
		oldProgramID := (*storedProgramPtr).ID

		// Try make a new program (this will update the ProgramID in the current struct)
		// So we start using it immediately if the compilation succeeds
		_, err := MakeProgram(programName, (*storedProgramPtr).VertexShaderFilePath, (*storedProgramPtr).FragmentShaderFilePath)
		if err != nil {
			// Handle error, and continue using old program
			log.Printf("Failed to build program %s, continuing to use old compilation (%d). \n", programName, (*storedProgramPtr).ID)
			return err
		}

		// Remove old program
		gl.DeleteProgram(uint32(oldProgramID))
	}

	// Done
	return nil
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

// Used to check if MakeShader() should add the path of the shader
// to the watchlist, or whether it is already present.
func shaderIsInWatchList(path string) bool {
	for _, shaderFileInfo := range LoadedShaders {
		if shaderFileInfo.FilePath == path {
			return true
		}
	}
	return false
}