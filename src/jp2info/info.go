package jp2info

// ColorMethod tells us how to determine the colorspace
type ColorMethod uint8

// Known color methods
const (
	CMEnumerated    ColorMethod = 1
	CMRestrictedICC             = 2
)

// ColorSpace tells us how to parse color data coming from openjpeg
type ColorSpace uint8

// Known color spaces
const (
	CSUnknown ColorSpace = iota
	CSRGB
	CSGrayScale
	CSYCC
)

// Info stores a variety of data we can easily scan from a jpeg2000 header
type Info struct {
	// Main header info
	Width, Height uint32
	Comps         uint16
	BPC           uint8

	// Color data
	ColorMethod  ColorMethod
	ColorSpace   ColorSpace
	Prec, Approx uint8

	// From SIZ box - this data can replace the main header data and
	// some of the colorspace data if necessary
	LSiz, RSiz     uint16
	XSiz, YSiz     uint32
	XOSiz, YOSiz   uint32
	XTSiz, YTSiz   uint32
	XTOSiz, YTOSiz uint32
	CSiz           uint16

	// From COD box
	LCod   uint16
	SCod   uint8
	SGCod  uint32
	Levels uint8
}

// TileWidth computes width of tiles
func (i *Info) TileWidth() uint32 {
	return i.XTSiz - i.XTOSiz
}

// TileHeight computes height of tiles
func (i *Info) TileHeight() uint32 {
	return i.YTSiz - i.YTOSiz
}

// String reports the ColorSpace in a human-readable way
func (cs ColorSpace) String() string {
	switch cs {
	case CSRGB:
		return "RGB"
	case CSGrayScale:
		return "Grayscale"
	case CSYCC:
		return "YCC"
	}
	return "Unknown"
}
