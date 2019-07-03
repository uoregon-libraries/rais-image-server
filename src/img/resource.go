package img

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"rais/src/iiif"
	"rais/src/transform"
)

// Resource wraps a decoder, IIIF ID, and the path to the image
type Resource struct {
	Decoder  Decoder
	ID       iiif.ID
	FilePath string
}

// NewResource initializes and returns an Resource for the given id
// and path.  If the path doesn't resolve to a valid file, or resolves to a
// file type that isn't supported, an error is returned.  File type is
// determined by extension, so images will need standard extensions in order to
// work.
func NewResource(id iiif.ID, filepath string) (*Resource, error) {
	var err error

	// First, does the file exist?
	if _, err = os.Stat(filepath); err != nil {
		return nil, ErrDoesNotExist
	}

	// File exists - is a decoder registered for it?
	var d Decoder
	for _, decodeFn := range fns {
		d, err = decodeFn(filepath)
		if err == ErrNotHandled {
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	if d == nil {
		return nil, ErrInvalidFiletype
	}

	img := &Resource{ID: id, Decoder: d, FilePath: filepath}
	return img, nil
}

// getResizeWithConstraints returns a scaled rectangle, computing the best fit
// for the given dimensions combined with our local constraints
func getResizeWithConstraints(crop image.Rectangle, max Constraint) image.Rectangle {
	// First figure out the ideal width and height within our max width and height
	cx := crop.Dx()
	cy := crop.Dy()

	// Sanity - we don't actually want any upscaling
	if max.Width > cx {
		max.Width = cx
	}
	if max.Height > cy {
		max.Height = cy
	}

	s := iiif.Size{Type: iiif.STBestFit, W: max.Width, H: max.Height}
	scale := s.GetResize(crop)
	sx := scale.Dx()
	sy := scale.Dy()

	// If this is within the bounds of our max area, we can return, otherwise we
	// have to scale further
	area := int64(sx) * int64(sy)
	if area <= max.Area {
		return scale
	}

	mult := math.Sqrt(float64(max.Area) / float64(area))
	xf := mult * float64(sx)
	yf := mult * float64(sy)
	return image.Rect(0, 0, int(xf), int(yf))
}

// Apply runs all image manipulation operations described by the IIIF URL, and
// returns an image.Image ready for encoding to the client
func (res *Resource) Apply(u *iiif.URL, max Constraint) (image.Image, error) {
	// Crop and resize have to be prepared before we can decode
	w, h := res.Decoder.GetWidth(), res.Decoder.GetHeight()
	crop := u.Region.GetCrop(w, h)

	// If size is "max", we actually want the "best fit" size type, but with our
	// constraints used instead of a user-supplied value.
	var scale image.Rectangle
	if u.Size.Type == iiif.STMax {
		scale = getResizeWithConstraints(crop, max)
	} else {
		scale = u.Size.GetResize(crop)
	}

	// Determine the final image output dimensions to test size constraints
	sw, sh := scale.Dx(), scale.Dy()
	if u.Rotation.Degrees == 90 || u.Rotation.Degrees == 270 {
		sw, sh = sh, sw
	}
	if max.SmallerThanAny(sw, sh) {
		return nil, ErrDimensionsExceedLimits
	}

	res.Decoder.SetCrop(crop)
	res.Decoder.SetResizeWH(scale.Dx(), scale.Dy())

	img, err := res.Decoder.DecodeImage()
	if err != nil {
		return nil, errors.New("unable to decode image: " + err.Error())
	}

	if u.Rotation.Mirror || u.Rotation.Degrees != 0 {
		img = rotate(img, u.Rotation)
	}

	// Unless I'm missing something, QColor doesn't actually change an image -
	// e.g., if it's already color, nothing happens.  If it's grayscale, there's
	// nothing to do (obviously we shouldn't report it, but oh well)
	switch u.Quality {
	case iiif.QGray:
		img = grayscale(img)
	case iiif.QBitonal:
		img = bitonal(img)
	}

	return img, nil
}

func rotate(img image.Image, rot iiif.Rotation) image.Image {
	var r transform.Rotator
	switch img0 := img.(type) {
	case *image.Gray:
		r = &transform.GrayRotator{Img: img0}
	case *image.RGBA:
		r = &transform.RGBARotator{Img: img0}
	}

	if rot.Mirror {
		r.Mirror()
	}

	switch rot.Degrees {
	case 90:
		r.Rotate90()
	case 180:
		r.Rotate180()
	case 270:
		r.Rotate270()
	}

	return r.Image()
}

func grayscale(img image.Image) image.Image {
	cm := img.ColorModel()
	if cm == color.GrayModel || cm == color.Gray16Model {
		return img
	}

	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, b, img, b.Min, draw.Src)
	return dst
}

func bitonal(img image.Image) image.Image {
	// First turn the image into 8-bit grayscale for easier manipulation
	imgGray := grayscale(img).(*image.Gray)
	b := imgGray.Bounds()
	imgBitonal := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	for i, pixel := range imgGray.Pix {
		if pixel > 190 {
			imgBitonal.Pix[i] = 255
		}
	}

	return imgBitonal
}
