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

var rotate90 = rotation{
	Method:         "Rotate90",
	Comment:        "does a simple 90-degree clockwise rotation",
	getDstXBase:    "(maxY - 1 - y)",
	GetDstY:        "x",
	DimensionOrder: "srcHeight, srcWidth",
}

var rotate180 = rotation{
	Method:         "Rotate180",
	Comment:        "does a simple 180-degree clockwise rotation",
	getDstXBase:    "(maxX - 1 - x)",
	GetDstY:        "(maxY - 1 - y)",
	DimensionOrder: "srcWidth, srcHeight",
}

var rotate270 = rotation{
	Method:         "Rotate270",
	Comment:        "does a simple 270-degree clockwise rotation",
	getDstXBase:    "y",
	GetDstY:        "(maxX - 1 - x)",
	DimensionOrder: "srcHeight, srcWidth",
}

var rotateMirror = rotation{
	Method:         "Mirror",
	Comment:        "flips the image around its vertical axis",
	getDstXBase:    "(maxX - 1 - x)",
	GetDstY:        "y",
	DimensionOrder: "srcWidth, srcHeight",
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
	CopyStatement:     "dstPix[dstIdx] = srcPix[srcIdx]",
	ByteSize:          1,
}

var typeRGBA = imageType{
	String:            "*image.RGBA",
	Shortstring:       "RGBA",
	ConstructorMethod: "image.NewRGBA",
	CopyStatement:     "copy(dstPix[dstIdx:dstIdx+4], srcPix[srcIdx:srcIdx+4])",
	ByteSize:          4,
}

type page struct {
	Rotations []rotation
	Types     []imageType
}

func main() {
	t := template.Must(template.ParseFiles("src/transform/template.txt"))
	f, err := os.Create("src/transform/rotation.go")
	if err != nil {
		fmt.Println("ERROR creating file:", err)
	}

	p := page{
		Rotations: []rotation{rotate90, rotate180, rotate270, rotateMirror},
		Types:     []imageType{typeGray, typeRGBA},
	}

	err = t.Execute(f, p)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}
