package openjpeg

import (
	"testing"
	"os"
)

func TestNewJP2Image(t *testing.T) {
	dir, _ := os.Getwd()
	jp2 := NewJP2Image(dir + "/../test-world.jp2")
	if jp2 == nil {
		t.Log("No jp2 object!")
		t.FailNow()
	}
	t.Log(jp2.image)
}
