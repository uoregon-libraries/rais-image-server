package main

// IIIFDecodeFn is a function which takes a filename and returns a IIIFImageDecoder
type IIIFDecodeFn func(string) (IIIFImageDecoder, error)

// ExtDecoders is our list of registered decoders for given file extensions
var ExtDecoders = make(map[string]IIIFDecodeFn)

// RegisterDecoder assigns the given decode function to a file extension
func RegisterDecoder(ext string, fn IIIFDecodeFn) {
	ExtDecoders[ext] = fn
}
