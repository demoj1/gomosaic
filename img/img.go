package img

import (
	"bytes"
	"encoding/base64"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	. "math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Files struct {
	Infos []ImageInfo `json:"files"`
}

type ImageInfo struct {
	FilePaht string  `json:"file"`
	R        float64 `json:"red"`
	G        float64 `json:"green"`
	B        float64 `json:"blue"`
	W        int     `json:"width"`
	H        int     `json:"height"`
}

func saveImage(in <-chan string) {
	for {
		url := <-in
		log.Println("Download url... " + url)
		defer log.Println("Succesfully.")

		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		file_name := base64.StdEncoding.EncodeToString([]byte(url))
		file, err := os.Create("images/" + file_name)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(file, response.Body)
		if err != nil {
			log.Fatal(err)
		}

		response.Body.Close()
		file.Close()
	}
}

func generateUrl(out chan<- string) {
	text, err := ioutil.ReadFile("photo_urls")
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(text), "\n")

	for _, val := range lines {
		out <- val
	}
}

func StartImageDownload() {
	var out chan string = make(chan string, 100)
	go generateUrl(out)
	for i := 0; i < 100; i++ {
		go saveImage(out)
	}
}

func countChan(in <-chan string, out chan<- ImageInfo) {
	for {
		file_name := <-in
		text, err := ioutil.ReadFile(file_name)
		if err != nil {
			log.Fatal(err)
		}

		img, err := jpeg.Decode(bytes.NewReader(text))
		if err != nil {
			log.Fatal(err)
		}

		h := img.Bounds().Size().Y
		w := img.Bounds().Size().X

		_r, _g, _b := avg(img, image.Rect(0, 0, w, h))

		info := ImageInfo{
			FilePaht: file_name,
			R:        _r,
			G:        _g,
			B:        _b,
			W:        w,
			H:        h}

		out <- info
		log.Println("Processed: " + info.FilePaht)
	}
}

func StartAverageImage() *Files {
	var files_chan chan string = make(chan string, 100)
	var info_chan chan ImageInfo = make(chan ImageInfo, 100)

	files := new(Files)

	text, err := ioutil.ReadFile("photo_urls")
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(text), "\n")

	go func() {
		for _, val := range lines {
			file_name := base64.StdEncoding.EncodeToString([]byte(val))
			files_chan <- "images/" + file_name
		}
	}()

	go func(f *Files, c <-chan ImageInfo) {
		for {
			f.Infos = append(f.Infos, <-c)
		}
	}(files, info_chan)

	for i := 0; i < 50; i++ {
		go countChan(files_chan, info_chan)
	}

	return files
}

func avg(img image.Image, rect image.Rectangle) (r, g, b float64) {
	s := float64(rect.Bounds().Size().X * rect.Bounds().Size().Y)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			_r, _g, _b, _ := img.At(x, y).RGBA()

			r += float64(_r)
			g += float64(_g)
			b += float64(_b)
		}
	}

	r /= s
	g /= s
	b /= s
	return
}

func dist(r1, g1, b1, r2, g2, b2 float64) float64 {
	return Sqrt(Pow(r2-r1, 2) + Pow(b2-b1, 2) + Pow(g2-g1, 2))
}

func loadImg(file string) image.Image {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func CreatePicture(s string, o string, f *Files) {
	src_img := loadImg(s)

	out_img := image.NewRGBA(image.Rect(0, 0, src_img.Bounds().Size().X, src_img.Bounds().Size().Y))
	for i := 0; i < src_img.Bounds().Size().X; i++ {
		for j := 0; j < src_img.Bounds().Size().Y; j++ {
			out_img.SetRGBA(i, j, color.RGBA{
				R: 0,
				G: 0,
				B: 0})
		}
	}

	for gy := 0; gy < src_img.Bounds().Size().Y; gy += 75 {
		for gx := 0; gx < src_img.Bounds().Size().X; gx += 75 {
			sr, sg, sb := avg(src_img, image.Rect(gx, gy, gx+75, gy+75))

			min_l := 100000.0
			min_file := ""

			min_l2 := 100000.0
			min_file2 := ""

			// find good image
			for _, v := range f.Infos {
				l := dist(sr, sg, sb, v.R, v.G, v.B)
				if l < min_l {
					min_l = l
					min_file = v.FilePaht
				}

				if l < min_l2 && v.FilePaht != min_file {
					min_l2 = l
					min_file2 = v.FilePaht
				}
			}

			num := randInt(0, 2)

			var ok_img image.Image

			//Fill rect
			if num == 0 {
				ok_img = loadImg(min_file)
			} else {
				ok_img = loadImg(min_file2)
			}

			ok_img = resize.Resize(75, 75, ok_img, resize.Lanczos3)

			offset_y := 75.0 - float64(ok_img.Bounds().Size().Y)
			offset_y /= 2.0

			int_offset := int(offset_y)
			for fy := gy + int_offset; fy < gy+ok_img.Bounds().Size().Y; fy++ {
				for fx := gx; fx < gx+ok_img.Bounds().Size().X; fx++ {
					out_img.Set(fx, fy, ok_img.At(fx-gx, fy-gy-int_offset))
				}
			}
		}
	}

	save_file, _ := os.Create(o)
	defer save_file.Close()

	jpeg.Encode(save_file, out_img, &jpeg.Options{jpeg.DefaultQuality})
}
