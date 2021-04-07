package gogl

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