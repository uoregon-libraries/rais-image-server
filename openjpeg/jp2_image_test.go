package openjpeg

import (
	"testing"
	"os"
)

func TestNewJP2Image(t *testing.T) {
	dir, _ := os.Getwd()
	jp2, err := NewJP2Image(dir + "/../test-world.jp2")
	if err != nil {
		t.Log("Error:", err)
		t.FailNow()
	}
	if jp2 == nil {
		t.Log("No jp2 object!")
		t.FailNow()
	}
	t.Log(jp2.image)
}
