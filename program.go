package gogl

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type ProgramID uint32
type Program struct {
	ID                     ProgramID
	ProgramName            string
	VertexShaderFilePath   string
	FragmentShaderFilePath string
}

func (program *Program) SetFloat(name string, value float32) {
	name_cstr := gl.Str(name + "\x00")
	location := gl.GetUniformLocation(uint32(program.ID), name_cstr)
	gl.Uniform1f(location, value)
}

type DataObject struct {
	VAOID                VAOID
	Type                 int // GOGL_TRIANGLES, GOGL_QUADS
	Vertices             []float32
	Indices              []uint32
	Program              *Program
	VertexShaderSource   string
	FragmentShaderSource string
}

func (data *DataObject) ProcessData() {
	// Link Program
	program, err := MakeProgram("program", data.VertexShaderSource, data.FragmentShaderSource)
	if err != nil {
		panic(err)
	}
	data.Program = program

	// Bind data to OpenGL
	// ------------------------------------------------
	// Create VAO
	gl.BindVertexArray(uint32(data.VAOID))
	GenBindBuffer(gl.ARRAY_BUFFER)
	BufferData(data.Vertices, gl.ARRAY_BUFFER, gl.STATIC_DRAW)

	if data.Type == GOGL_QUADS {
		// Create Element Buffer Object
		GenBindBuffer(gl.ELEMENT_ARRAY_BUFFER)
		BufferDataInt(data.Indices, gl.ELEMENT_ARRAY_BUFFER, gl.STATIC_DRAW)

		// Configure VAO
		// ---------------------------
		// - x,y,z data starts at index 0, and is 3 values long (0,3)
		// - Each vertex is 5 values long, and a float32 is 4 bytes long, so
		//   the stride is 5*4
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(0)

		// - texcoord is two values long (2), and starts at index 3 (gl.PtrOffset(3*4))
		// - this is the second attribpointer (1), non-normalized data (false)
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		gl.EnableVertexAttribArray(1)

	} else if data.Type == GOGL_TRIANGLES {
		// Configure VAO
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
		gl.EnableVertexAttribArray(0)
	}

	// Unbind
	UnbindVertexArray()
}
