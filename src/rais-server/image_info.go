package main

// ImageInfo holds just enough data to reproduce the dynamic portions of
// info.json
type ImageInfo struct {
	Width, Height         int
	TileWidth, TileHeight int
}
