package gogl

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type DataObject struct {
	VAOID                VAOID                // id of the vertex array object
	VBOID                BufferID             // id of the vertex buffer object
	EBOID                BufferID             // element buffer object for quads
	Type                 int                  // Lets us know in what format the raw vertex data is defined. GOGL_TRIANGLES, GOGL_QUADS
	Vertices             []float32            // raw vertex data
	Indices              []uint32             // when giving the data in quad format, this value should indicate which vertices make a triangle together
	ProgramName          string               // Used for keeping track of the program, and hotloading the shaders when they change.
	Program              *Program             // Contains the id of the GL program, and other data to simplify hotloading shaders.
	VertexShaderSource   string               // Filepath of the .vert shader. Can be relative.
	FragmentShaderSource string               // Filepath of the .frag shader. Can be relative.
	Textures             map[string]TextureID // Map used to avoid loading in textures more than once.
	Sprites              []Sprite             // List of Sprites that belong to this DataObject.
}

/*
This function makes sure that the filled in DataObject is made ready to be used with OpenGL.
This function should only be called once.
To actually get ready to draw using a DataObject, call DataObject.Enable() after calling this function
to select it as your current active DataObject.
*/
func (data *DataObject) ProcessData() {

	// Link Program
	program, err := MakeProgram(data.ProgramName, data.VertexShaderSource, data.FragmentShaderSource)
	if err != nil {
		panic(err)
	}
	data.Program = program

	// Create VAO, VBO
	data.VAOID = GenVertexArray()
	data.VBOID = GenBuffer(gl.ARRAY_BUFFER)

	if data.Type == GOGL_QUADS {
		// Create Element Buffer Object
		data.EBOID = GenBuffer(gl.ELEMENT_ARRAY_BUFFER)
	}

	// Unbind
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	if data.Type == GOGL_QUADS {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	}
}

/*
After a DataObject has been build through DataObject.ProcessData(),
you can enable it to start drawing to the screen. It binds all the things that need to be bound
for the DataObject to be active. If you want to use attached Sprites, activate them separately: `sp := data.SelectSprite(0); sp.SetUniforms()`
This function can be called as often as you want, to switch between multiple DataObjects.
*/
func (data *DataObject) Enable() {

	// Use Program
	UseProgram((*data.Program).ID)

	// Bind VAO
	gl.BindVertexArray(uint32(data.VAOID))

	// Bind VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, uint32(data.VBOID))
	BufferDataFloat32(data.Vertices, gl.ARRAY_BUFFER, gl.STATIC_DRAW)

	if data.Type == GOGL_QUADS {
		// Bind EBO
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, uint32(data.EBOID))
		BufferDataUint32(data.Indices, gl.ELEMENT_ARRAY_BUFFER, gl.STATIC_DRAW)

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
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
		gl.EnableVertexAttribArray(0)
	}
}

// Calls Update on all the Sprites in the Sprite list.
func (data *DataObject) Update() {
	for i := range data.Sprites {
		data.Sprites[i].Update()
	}
}
