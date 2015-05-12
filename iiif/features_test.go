package iiif

import (
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"testing"
)

var FeaturesLevel0 = FeatureSet0()
var FeaturesLevel1 = FeatureSet1()
var FeaturesLevel2 = FeatureSet2()

func TestRegionSupport(t *testing.T) {
	r := Region{Type: RTFull}
	assert.True(FeaturesLevel0.SupportsRegion(r), "RTFull supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsRegion(r), "RTFull supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTFull supported by FL2", t)

	r.Type = RTPixel
	assert.False(FeaturesLevel0.SupportsRegion(r), "RTPixel NOT supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsRegion(r), "RTPixel supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTPixel supported by FL2", t)

	r.Type = RTPercent
	assert.False(FeaturesLevel0.SupportsRegion(r), "RTPercent NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsRegion(r), "RTPercent NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTPercent supported by FL2", t)
}

func TestSizeSupport(t *testing.T) {
	s := Size{Type: STFull}
	assert.True(FeaturesLevel0.SupportsSize(s), "STFull supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsSize(s), "STFull supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STFull supported by FL2", t)

	s.Type = STScaleToWidth
	assert.False(FeaturesLevel0.SupportsSize(s), "STScaleToWidth NOT supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsSize(s), "STScaleToWidth supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STScaleToWidth supported by FL2", t)

	s.Type = STScaleToHeight
	assert.False(FeaturesLevel0.SupportsSize(s), "STScaleToHeight NOT supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsSize(s), "STScaleToHeight supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STScaleToHeight supported by FL2", t)

	s.Type = STScalePercent
	assert.False(FeaturesLevel0.SupportsSize(s), "STScalePercent NOT supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsSize(s), "STScalePercent supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STScalePercent supported by FL2", t)

	s.Type = STExact
	assert.False(FeaturesLevel0.SupportsSize(s), "STExact NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsSize(s), "STExact NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STExact supported by FL2", t)

	s.Type = STBestFit
	assert.False(FeaturesLevel0.SupportsSize(s), "STBestFit NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsSize(s), "STBestFit NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsSize(s), "STBestFit supported by FL2", t)
}

func TestRotationSupport(t *testing.T) {
	r := Rotation{Degrees: 0}
	assert.True(FeaturesLevel0.SupportsRotation(r), "0 degrees supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsRotation(r), "0 degrees supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRotation(r), "0 degrees supported by FL2", t)

	r.Degrees = 90
	assert.False(FeaturesLevel0.SupportsRotation(r), "90 degrees NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsRotation(r), "90 degrees NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRotation(r), "90 degrees supported by FL2", t)

	r.Degrees = 90.01
	assert.False(FeaturesLevel0.SupportsRotation(r), "90.01 degrees NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsRotation(r), "90.01 degrees NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsRotation(r), "90.01 degrees NOT supported by FL2", t)

	r.Degrees = 0
	r.Mirror = true
	assert.False(FeaturesLevel0.SupportsRotation(r), "Mirroring NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsRotation(r), "Mirroring NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsRotation(r), "Mirroring NOT supported by FL2", t)
}

func TestQualitySupport(t *testing.T) {
	assert.True(FeaturesLevel0.SupportsQuality(QDefault), "QDefault supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsQuality(QDefault), "QDefault supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsQuality(QDefault), "QDefault supported by FL2", t)

	assert.True(FeaturesLevel0.SupportsQuality(QNative), "QNative supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsQuality(QNative), "QNative supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsQuality(QNative), "QNative supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsQuality(QColor), "QColor NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsQuality(QColor), "QColor NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsQuality(QColor), "QColor supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsQuality(QGray), "QGray NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsQuality(QGray), "QGray NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsQuality(QGray), "QGray supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsQuality(QBitonal), "QBitonal NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsQuality(QBitonal), "QBitonal NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsQuality(QBitonal), "QBitonal supported by FL2", t)
}

func TestFormatSupport(t *testing.T) {
	assert.True(FeaturesLevel0.SupportsFormat(FmtJPG), "FmtJPG supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsFormat(FmtJPG), "FmtJPG supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsFormat(FmtJPG), "FmtJPG supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtPNG), "FmtPNG NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtPNG), "FmtPNG NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsFormat(FmtPNG), "FmtPNG supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtTIF), "FmtTIF NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtTIF), "FmtTIF NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsFormat(FmtTIF), "FmtTIF NOT supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtGIF), "FmtGIF NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtGIF), "FmtGIF NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsFormat(FmtGIF), "FmtGIF NOT supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtJP2), "FmtJP2 NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtJP2), "FmtJP2 NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsFormat(FmtJP2), "FmtJP2 NOT supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtPDF), "FmtPDF NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtPDF), "FmtPDF NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsFormat(FmtPDF), "FmtPDF NOT supported by FL2", t)

	assert.False(FeaturesLevel0.SupportsFormat(FmtWEBP), "FmtWEBP NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsFormat(FmtWEBP), "FmtWEBP NOT supported by FL1", t)
	assert.False(FeaturesLevel2.SupportsFormat(FmtWEBP), "FmtWEBP NOT supported by FL2", t)
}

func TestInclusion(t *testing.T) {
	assert.True(FeaturesLevel2.includes(FeaturesLevel2), "FeaturesLevel2.includes(FeaturesLevel2)", t)
	assert.True(FeaturesLevel2.includes(FeaturesLevel1), "FeaturesLevel2.includes(FeaturesLevel1)", t)
	assert.True(FeaturesLevel2.includes(FeaturesLevel0), "FeaturesLevel2.includes(FeaturesLevel0)", t)

	assert.False(FeaturesLevel1.includes(FeaturesLevel2), "FeaturesLevel1.includes(FeaturesLevel2)", t)
	assert.True(FeaturesLevel1.includes(FeaturesLevel1), "FeaturesLevel1.includes(FeaturesLevel1)", t)
	assert.True(FeaturesLevel1.includes(FeaturesLevel0), "FeaturesLevel1.includes(FeaturesLevel0)", t)

	assert.False(FeaturesLevel0.includes(FeaturesLevel2), "FeaturesLevel0.includes(FeaturesLevel2)", t)
	assert.False(FeaturesLevel0.includes(FeaturesLevel1), "FeaturesLevel0.includes(FeaturesLevel1)", t)
	assert.True(FeaturesLevel0.includes(FeaturesLevel0), "FeaturesLevel0.includes(FeaturesLevel0)", t)
}
