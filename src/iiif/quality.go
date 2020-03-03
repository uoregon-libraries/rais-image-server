package iiif

// Quality is the representation of a IIIF 2.0 quality (color space / depth)
// which a client may request.  We also include "native" for better
// compatibility with older clients, since it's the same as "default".
type Quality string

// All possible qualities for IIIF 2.0 and 1.1
const (
	QUnknown Quality = ""
	QColor   Quality = "color"
	QGray    Quality = "gray"
	QBitonal Quality = "bitonal"
	QDefault Quality = "default"
	QNative  Quality = "native" // For 1.1 compatibility
)

// Qualities is the definitive list of all possible Quality constants
var Qualities = []Quality{QColor, QGray, QBitonal, QDefault, QNative}

// StringToQuality converts val into a Quality constant if val is one of our
// valid Qualities
func StringToQuality(val string) Quality {
	q := Quality(val)
	if q.Valid() {
		return q
	}
	return QUnknown
}

// Valid returns whether a given Quality string is valid.  Since a Quality can be
// created via Quality("blah"), this ensures the quality is, in fact, within the
// list of known qualities.
func (q Quality) Valid() bool {
	for _, valid := range Qualities {
		if valid == q {
			return true
		}
	}

	return false
}
