package gogl

import (
	"github.com/go-gl/gl/v4.5-core/gl"
)

type Sprite struct {
	Name            string      // Descriptive name, might be used in debug logging.
	TextureSource   string      // The filepath of the image that will be loaded in as a texture. Can be a relative path. Texture is loaded in AddSprite().
	Divisions       int         // How many tiles the spritesheet is divided up in
	Texture         TextureID   // ID of the texture that serves as the spritesheet
	AnimationFrames [][]float32 // In which part of the sprite sheet is each animation frame located?
	AnimationSpeed  int         // How many ticks does it take to advance a frame?
	TickCount       int         // Keeps track of the game loops that have passed. Is reset to 0 when TickCount==AnimationSpeed
	CurrentFrame    int         // Index of a frame in sprite.AnimationFrames
	Xn              float32     // X location of sprite tile on the screen (normalized values)
	Yn              float32     // Y location of sprite tile on the screen (normalized values)
	Scale           float32     // Weird way to scale up/down the sprite :)
	FlipHorizontal  float32     // 1.0 for flip horizontal, 0.0 for no flip
}

// Initializes and adds Sprite to the DataObject for later use.
// Also loads Texture from source, if it wasn't already loaded.
func (data *DataObject) AddSprite(sprite Sprite) {
	// initialize map
	if data.Textures == nil {
		data.Textures = make(map[string]TextureID)
	}

	// load texture
	textureID := data.Textures[sprite.TextureSource]
	if textureID == 0 {
		textureID = LoadImageToTexture(sprite.TextureSource)
		data.Textures[sprite.TextureSource] = textureID
	}
	sprite.Texture = textureID

	// add sprite to DataObject
	data.Sprites = append(data.Sprites, sprite)
}

// Return the requested sprite from the sprite list, and bind its texture.
// When ready to draw, don't forget to also call sprite.SetUniforms(&data).
func (data *DataObject) SelectSprite(spriteIndex int) *Sprite {
	// Get Sprite as pointer
	sprite := &data.Sprites[spriteIndex]

	// Bind the Sprite's texture to TEXTURE_2D
	gl.BindTexture(gl.TEXTURE_2D, uint32(sprite.Texture))

	return sprite
}

// Advances the TickCounter, which causes looping through AnimationFrames,
// thus animating the Sprite.
func (sprite *Sprite) Update() {
	// Tick up
	sprite.TickCount++

	// Advance frame if tick count reaches cap
	if sprite.TickCount >= sprite.AnimationSpeed {
		sprite.TickCount = 0
		sprite.CurrentFrame++

		// Loop frames
		if sprite.CurrentFrame >= len(sprite.AnimationFrames) {
			sprite.CurrentFrame = 0
		}
	}
}

// Sets all the uniforms that apply to the Sprite, so that the shaders know what to do.
func (sprite *Sprite) SetUniforms(data *DataObject) {

	// Set the divisions uniform (used to locate the correct tile on the texture)
	data.Program.SetFloat("tex_divisions", float32(sprite.Divisions))

	// Set the position of the Sprite tile on the Texture
	data.Program.SetFloat("tex_x", sprite.AnimationFrames[sprite.CurrentFrame][0])
	data.Program.SetFloat("tex_y", sprite.AnimationFrames[sprite.CurrentFrame][1])

	// Set the (normalized) position of the Sprite on the screen
	data.Program.SetFloat("x", sprite.Xn)
	data.Program.SetFloat("y", sprite.Yn)

	// Used for zooming, a bit hacky, should rewrite with matrix manipulation or something.
	data.Program.SetFloat("scale", sprite.Scale)

	// Flip the texture tile horizontally or not (1.0 for yes, 0.0 for no)
	data.Program.SetFloat("tex_fliph", sprite.FlipHorizontal)
}
