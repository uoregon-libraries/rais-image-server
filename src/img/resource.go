package img

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"net/url"
	"rais/src/iiif"
	"rais/src/plugins"
	"rais/src/transform"
)

// Resource wraps a streamer and decode function, the two components we must
// have for any image, as well as the image ID and URL.  The actual decoder is
// lazy-loaded when it's needed.
type Resource struct {
	ID         iiif.ID
	URL        *url.URL
	streamer   Streamer
	decoder    Decoder
	decodeFunc DecodeFunc
}

// NewResource initializes and returns an Resource for the given URL
// (translated from a IIIF ID) If the URL doesn't have a streamer, doesn't
// resolve to a valid image, or resolves to an image for which we have no
// decoder, an error is returned.  File type is determined by extension, so
// images will need standard extensions in order to work.
func NewResource(id iiif.ID, u *url.URL) (r *Resource, err error) {
	var openStream OpenStreamFunc
	r = &Resource{ID: id, URL: u}

	// Do we have a streamer for this resource's scheme?
	openStream, err = getStreamOpener(u)
	if err != nil {
		return nil, fmt.Errorf("unable to find streamer for %q: %w", u, err)
	}

	// Streamer exists, so we attempt to open it
	r.streamer, err = openStream()
	if err != nil {
		return nil, fmt.Errorf("unable to open %q: %w", u, err)
	}

	// We have a stream - do we have a decoder for it?
	r.decodeFunc, err = getDecodeFunc(r.streamer)
	if err != nil {
		// No decoder means we have to close the stream before returning
		r.streamer.Close()
		return nil, fmt.Errorf("unable to find decoder for %q: %w", u, err)
	}

	return r, err
}

func getStreamOpener(u *url.URL) (openFunc OpenStreamFunc, err error) {
	for _, streamReader := range streamReaders {
		openFunc, err = streamReader(u)
		if err == nil {
			return openFunc, nil
		}
		if err != plugins.ErrSkipped {
			return nil, err
		}
	}

	// No stream reader was found for this URL
	return nil, ErrNotStreamable
}

func getDecodeFunc(s Streamer) (d DecodeFunc, err error) {
	for _, decodeHandler := range decodeHandlers {
		d, err = decodeHandler(s)
		if err == nil && d != nil {
			return d, nil
		}
		if err == plugins.ErrSkipped {
			continue
		}
		return nil, err
	}

	// No decoder was found for this file type
	return nil, ErrInvalidFiletype
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

// Decoder attempts to initialize the registered decoder.  Because this can
// read from disk, it should only be called when it has to be called.  It may
// return errors if reading the image fails.
func (res *Resource) Decoder() (Decoder, error) {
	var err error
	if res.decoder == nil {
		res.decoder, err = res.decodeFunc()
	}

	return res.decoder, err
}

// Apply runs all image manipulation operations described by the IIIF URL, and
// returns an image.Image ready for encoding to the client
func (res *Resource) Apply(u *iiif.URL, max Constraint) (image.Image, error) {
	// Initialize a decoder if that hasn't already happened
	var decoder, err = res.Decoder()
	if err != nil {
		return nil, err
	}

	// Crop and resize have to be prepared before we can decode
	w, h := decoder.GetWidth(), decoder.GetHeight()
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

	decoder.SetCrop(crop)
	decoder.SetResizeWH(scale.Dx(), scale.Dy())

	img, err := decoder.DecodeImage()
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

// Destroy lets the resource clean up any open streams, etc.  This *must* be
// called to prevent resource leaks!
func (res *Resource) Destroy() {
	res.streamer.Close()
}
