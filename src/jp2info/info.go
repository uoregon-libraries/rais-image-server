package jp2info

type ColorMethod uint8

const (
	CMEnumerated    ColorMethod = 1
	CMRestrictedICC             = 2
)

type ColorSpace uint8

const (
	CSUnknown ColorSpace = iota
	CSRGB
	CSGrayScale
	CSYCC
)

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
	LCod           uint16
	SCod           uint8
	SGCod          uint32
	Levels         uint8
}

func (i *Info) TileWidth() uint32 {
	return i.XTSiz - i.XTOSiz
}

func (i *Info) TileHeight() uint32 {
	return i.YTSiz - i.YTOSiz
}

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
