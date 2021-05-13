package gogl

import (
	//"time"
	"os"
	//"io/ioutil"
	//"log"
	"image/png"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type TextureID uint32

func LoadImageToTexture(filename string) TextureID {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	byteIndex := 0

	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[byteIndex] = byte(r / 256)
			byteIndex++
			pixels[byteIndex] = byte(g / 256)
			byteIndex++
			pixels[byteIndex] = byte(b / 256)
			byteIndex++
			pixels[byteIndex] = byte(a / 256)
			byteIndex++
		}
	}

	texId := GenTexture()
	BindTexture(texId)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Load image in texture
	// target, level, colormode, width, heigth, border, format, xtype, *pixels
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	// Prerender smaller versions of texture at runtime for performance reasons
	gl.GenerateMipmap(gl.TEXTURE_2D)

	return texId
}

func GenTexture() TextureID {
	var texId uint32
	gl.GenTextures(1, &texId)
	return TextureID(texId)
}

func BindTexture(TexId TextureID) {
	gl.BindTexture(gl.TEXTURE_2D, uint32(TexId))
}
