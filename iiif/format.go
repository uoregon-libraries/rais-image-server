package iiif

type Format string

const (
	FmtJPG  Format = "jpg"
	FmtTIF  Format = "tif"
	FmtPNG  Format = "png"
	FmtGIF  Format = "gif"
	FmtJP2  Format = "jp2"
	FmtPDF  Format = "pdf"
	FmtWEBP Format = "webp"
)

var Formats = []Format{FmtJPG, FmtTIF, FmtPNG, FmtJP2, FmtPDF, FmtWEBP}

func (f Format) Valid() bool {
	for _, valid := range Formats {
		if valid == f {
			return true
		}
	}

	return false
}

