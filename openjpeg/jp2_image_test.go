package openjpeg

import (
	"os"
	"testing"
)

func TestNewJP2Image(t *testing.T) {
	dir, _ := os.Getwd()
	jp2, err := NewJP2Image(dir + "/../test-world.jp2")
	if err != nil {
		t.Error("Error reading JP2:", err)
	}

	if jp2 == nil {
		t.Error("No JP2 object!")
	}

	t.Log(jp2.image)
}
