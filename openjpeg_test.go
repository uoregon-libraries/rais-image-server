package main

import (
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestTile(t *testing.T) {
	path := "/tmp/tile.jp2"
	_, err := os.Stat(path)
	if err != nil {
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		resp, err := http.Get("http://chroniclingamerica.loc.gov/lccn/sn83030214/1903-05-14/ed-1/seq-1.jp2")
		if err == nil {
			log.Println("Status:", resp.Status)
			io.Copy(f, resp.Body)
		} else {
			log.Println("Get error:", err)
		}
		f.Close()
	}

	r := image.Rect(2547, 447, 4298, 1559)
	width := 681
	height := 432
	if err, i := NewImageTile(path, r, width, height); err == nil {
		log.Println(i.Bounds())
		if i.Bounds().Max.X < 10 {
			t.FailNow()
		}

	} else {
		log.Fatal("error creating image tile:", err)
	}
}
