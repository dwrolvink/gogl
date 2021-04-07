package gogl

/* 	TYPES

	General types will be put here. When a type is closely linked to a separate 
	code file, the type will be put there. 
		E.g.: the ShaderFileInfo struct is used exclusively to track changes in the 
		shader files, so that type is put in hotloading.go.
	Use project level search if you can't find a certain type! 

	The uint32 recastings are mostly to create a type-awareness/-safety in the code, 
	this convention can be stripped entirely from the code, and it would all still work
	as intended (if one is thorough).
*/

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