package openjpeg

import (
	"crypto/md5"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

var JP2_URLS = []string{"http://chroniclingamerica.loc.gov/lccn/sn83030214/1903-05-14/ed-1/seq-1.jp2", "http://chroniclingamerica.loc.gov/lccn/sn83045433/1913-02-20/ed-1/seq-1.jp2"}

func TestTile(t *testing.T) {
	for _, url := range JP2_URLS {
		h := md5.New()
		io.WriteString(h, url)
		path := fmt.Sprintf("/tmp/brikker-%x.jp2", string(h.Sum(nil)))
		_, err := os.Stat(path)
		if err != nil {
			f, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			resp, err := http.Get(url)
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
		if i, err := NewImageTile(path, r, width, height); err == nil {
			log.Println(i.Bounds())
			if i.Bounds().Max.X < 10 {
				t.FailNow()
			}

		} else {
			log.Fatal("error creating image tile:", err)
		}
	}
}
