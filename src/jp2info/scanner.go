package jp2info

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// JP2HEADER contains the raw bytes for the only JP2 header we currently
// respect
var JP2HEADER = []byte{
	0x00, 0x00, 0x00, 0x0c,
	0x6a, 0x50, 0x20, 0x20,
	0x0d, 0x0a, 0x87, 0x0a,
}

// Various hard-coded byte values for finding JP2 boxes
var (
	IHDR   = []byte{0x69, 0x68, 0x64, 0x72} // "ihdr"
	COLR   = []byte{0x63, 0x6f, 0x6c, 0x72} // "colr"
	SOCSIZ = []byte{0xFF, 0x4F, 0xFF, 0x51}
	COD    = []byte{0xFF, 0x52}
)

// Scanner reads a Jpeg2000 header and parsing its data into an Info structure
type Scanner struct {
	r *bufio.Reader
	e error
	i *Info
}

// Scan reads the file and populates an Info pointer
func (s *Scanner) Scan(filename string) (*Info, error) {
	var f, err = os.Open(filename)
	if err != nil {
		return nil, err
	}

	s.readInfo(f)
	return s.i, s.e
}

func (s *Scanner) readInfo(ior io.Reader) {
	s.i = &Info{}
	s.r = bufio.NewReader(ior)

	// Make sure the header bytes are legit - this doesn't cover all types of
	// JP2, but it works for what RAIS needs
	var header = make([]byte, 12)
	_, s.e = s.r.Read(header)
	if !bytes.Equal(header, JP2HEADER) {
		s.e = fmt.Errorf("unknown file format")
		return
	}

	// Find IHDR for basic information
	s.scanUntil(IHDR)
	s.readBE(&s.i.Height, &s.i.Width, &s.i.Comps, &s.i.BPC)

	// Find COLR to get colorspace data
	s.scanUntil(COLR)
	s.readColor()

	// Find various SIZ data
	s.scanUntil(SOCSIZ)
	s.readBE(&s.i.LSiz, &s.i.RSiz, &s.i.XSiz, &s.i.YSiz, &s.i.XOSiz,
		&s.i.YOSiz, &s.i.XTSiz, &s.i.YTSiz, &s.i.XTOSiz, &s.i.YTOSiz, &s.i.CSiz)

	// Find COD, primarily to get resolution levels
	s.scanUntil(COD)
	s.readBE(&s.i.LCod, &s.i.SCod, &s.i.SGCod, &s.i.Levels)
}

func (s *Scanner) readColor() {
	s.readBE(&s.i.ColorMethod, &s.i.Prec, &s.i.Approx)
	if s.i.ColorMethod == CMEnumerated {
		s.readEnumeratedColor()
	} else {
		s.readColorProfile()
	}
}

func (s *Scanner) readEnumeratedColor() {
	s.r.Discard(2)
	var colorSpace uint16

	s.readBE(&colorSpace)
	switch colorSpace {
	case 16:
		s.i.ColorSpace = CSRGB
	case 17:
		s.i.ColorSpace = CSGrayScale
	case 18:
		s.i.ColorSpace = CSYCC
	default:
		s.i.ColorSpace = CSUnknown
	}
}

func (s *Scanner) readColorProfile() {
	// TODO: make this a bit more useful
	s.i.ColorSpace = CSUnknown
}

// scanUntil reads until the given token has been found and fully read
// in, leaving the io pointer exactly one byte past the token
func (s *Scanner) scanUntil(token []byte) {
	if s.e != nil {
		return
	}

	var matchOffset int
	var tokenLen = len(token)
	var b byte

	for {
		b, s.e = s.r.ReadByte()
		if s.e != nil {
			return
		}

		if b == token[matchOffset] {
			matchOffset++
		}

		if matchOffset == tokenLen {
			return
		}
	}
}

// readBE wraps binary.Read for reading any arbitrary amount of BigEndian data
func (s *Scanner) readBE(data ...interface{}) {
	if s.e != nil {
		return
	}

	var datum interface{}
	for _, datum = range data {
		s.e = binary.Read(s.r, binary.BigEndian, datum)
		if s.e != nil {
			return
		}
	}
}
