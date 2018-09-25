package iiif

// Format represents an IIIF 2.0 file format a client may request
type Format string

// All known file formats for IIIF 2.0
const (
	FmtUnknown Format = ""
	FmtJPG     Format = "jpg"
	FmtTIF     Format = "tif"
	FmtPNG     Format = "png"
	FmtGIF     Format = "gif"
	FmtJP2     Format = "jp2"
	FmtPDF     Format = "pdf"
	FmtWEBP    Format = "webp"
)

// Formats is the definitive list of all possible Format constants
var Formats = []Format{FmtJPG, FmtTIF, FmtPNG, FmtGIF, FmtJP2, FmtPDF, FmtWEBP}

func StringToFormat(val string) Format {
	f := Format(val)
	if f.Valid() {
		return f
	}
	return FmtUnknown
}

// Valid returns whether a given Format string is valid.  Since a Format can be
// created via Format("blah"), this ensures the format is, in fact, within the
// list of known formats.
func (f Format) Valid() bool {
	for _, valid := range Formats {
		if valid == f {
			return true
		}
	}

	return false
}
