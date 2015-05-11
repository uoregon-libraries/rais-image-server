package iiif

type Quality string

const (
	QColor   Quality = "color"
	QGray    Quality = "gray"
	QBitonal Quality = "bitonal"
	QDefault Quality = "default"
	QNative  Quality = "native" // For 1.1 compatibility
)

var Qualities = []Quality{QColor, QGray, QBitonal, QDefault, QNative}

func (q Quality) Valid() bool {
	for _, valid := range Qualities {
		if valid == q {
			return true
		}
	}

	return false
}
