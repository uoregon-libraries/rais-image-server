package main

func init() {
	// Known extensions Go handles internally - note that if we want to properly
	// deal with pyramidal TIFFs, that's going to require something beyond a per-
	// extension registry.
	extList := []string{".tif", ".tiff", ".png", ".jpg", "jpeg", ".gif"}
	for _, ext := range extList {
		RegisterDecoder(ext, decodeCommonFile)
	}
}

func decodeCommonFile(path string) (IIIFImage, error) {
	return NewSimpleImage(path)
}
