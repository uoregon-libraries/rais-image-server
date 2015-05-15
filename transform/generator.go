// +build ignore
//
// This file builds the rotation code: `go run transform/generator.go`

package main

import (
	"fmt"
	"os"
	"text/template"
)

type rotation struct {
	Method         string
	Comment        string
	getDstXBase    string
	GetDstY        string
	DimensionOrder string
}

func valTimesInt(s string, i int) string {
	switch i {
	case 1:
		return s
	case 2:
		return fmt.Sprintf("(%s << 1)", s)
	case 4:
		return fmt.Sprintf("(%s << 2)", s)
	case 8:
		return fmt.Sprintf("(%s << 3)", s)
	default:
		return fmt.Sprintf("(%s * %d)", s, i)
	}
}

func (r rotation) GetSrcX(i int) string {
	return valTimesInt("x", i)
}

func (r rotation) GetDstX(i int) string {
	return valTimesInt(r.getDstXBase, i)
}

var rotate90 rotation = rotation{
	Method:         "Rotate90",
	Comment:        "does a simple 90-degree clockwise rotation, returning a new image.Image",
	getDstXBase:    "(maxY - 1 - y)",
	GetDstY:        "x",
	DimensionOrder: "srcHeight, srcWidth",
}

var rotate180 rotation = rotation{
	Method:         "Rotate180",
	Comment:        "does a simple 180-degree clockwise rotation, returning a new image.Image",
	getDstXBase:    "(maxX - 1 - x)",
	GetDstY:        "(maxY - 1 - y)",
	DimensionOrder: "srcWidth, srcHeight",
}

var rotate270 rotation = rotation{
	Method:         "Rotate270",
	Comment:        "does a simple 270-degree clockwise rotation, returning a new image.Image",
	getDstXBase:    "y",
	GetDstY:        "(maxX - 1 - x)",
	DimensionOrder: "srcHeight, srcWidth",
}

type imageType struct {
	String            string
	Shortstring       string
	ConstructorMethod string
	CopyStatement     string
	ByteSize          int
}

var typeGray = imageType{
	String:            "*image.Gray",
	Shortstring:       "Gray",
	ConstructorMethod: "image.NewGray",
	CopyStatement:     "dst.Pix[dstPix] = src.Pix[srcPix]",
	ByteSize:          1,
}

var typeRGBA = imageType{
	String:            "*image.RGBA",
	Shortstring:       "RGBA",
	ConstructorMethod: "image.NewRGBA",
	CopyStatement:     "copy(dst.Pix[dstPix:dstPix+4], src.Pix[srcPix:srcPix+4])",
	ByteSize:          4,
}

type Page struct {
	Rotations []rotation
	Types     []imageType
}

func main() {
	t := template.Must(template.ParseFiles("transform/template.txt"))
	f, err := os.Create("transform/rotation.go")
	if err != nil {
		fmt.Println("ERROR creating file:", err)
	}

	p := Page{
		Rotations: []rotation{rotate90, rotate180, rotate270},
		Types:     []imageType{typeGray, typeRGBA},
	}

	err = t.Execute(f, p)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}
