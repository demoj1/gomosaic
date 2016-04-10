package main

import (
	//"encoding/json"
	//"fmt"
	//"github.com/diman3241/ImageMosaic/img"
	"bytes"
	"flag"
	"fmt"
	"github.com/nfnt/resize"
	. "image"
	"image/jpeg"
	"io/ioutil"
	. "math"
	"math/rand"
	"os"
	"sync"
	"time"
)

type ImageInfo struct {
	FilePaht string  `json:"file"`
	R        float64 `json:"red"`
	G        float64 `json:"green"`
	B        float64 `json:"blue"`
	Img      Image
}

func main() {
	out_width := flag.Uint("out-w", 1000, "Out image width.")
	out_height := flag.Uint("out-h", 1000, "Out image height.")

	grid_width := flag.Uint("grid-w", 50, "The width of the grid cell.")
	grid_height := flag.Uint("grid-h", 50, "The height of the grid cell.")

	src_file := flag.String("src-img", "src.jpg", "Source image file.")
	out_file := flag.String("out-img", "out.jpg", "Name out file.")

	find_dir := flag.String("img-dir", "images", "The directory in which to search for photos mosaicking.")

	flag.Parse()

	files := getFiles(*find_dir)
	out_files := make(chan string, 10)
	proc_image := make(chan ImageInfo, 10)

	go transferFiles(out_files, *find_dir, files)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		go proccessImgFile(out_files, proc_image, *grid_width, *grid_height, &wg)
	}

	src_img := loadImg(*src_file)
	src_img = resize.Resize(*out_width, *out_height, src_img, resize.Lanczos3)

	out_image := NewRGBA(Rect(0, 0, int(*out_width), int(*out_height)))

	img_infos := make([]ImageInfo, 1)
	go func(in <-chan ImageInfo) {
		for info := range in {
			img_infos = append(img_infos, info)
		}
	}(proc_image)

	wg.Wait()
	createMosaic(src_img, out_image, int(*grid_width), int(*grid_height), img_infos)
	save_file, _ := os.Create(*out_file)
	jpeg.Encode(save_file, out_image, &jpeg.Options{jpeg.DefaultQuality})
}

func createMosaic(src Image, out *RGBA, gw, gh int, info []ImageInfo) {
	fmt.Println("Start create mosaic")
	all_S := src.Bounds().Size().Y

	for y := 0; y < src.Bounds().Size().Y; y += gh {
		for x := 0; x < src.Bounds().Size().X; x += gw {
			min_dist := minDist(src, Rectangle{
				Min: Point{
					X: x,
					Y: y},
				Max: Point{
					X: x + gw,
					Y: y + gh}}, info)

			ok_img := min_dist.Img
			if ok_img == nil {
				continue
			}

			for yy := y; yy < y+gh; yy++ {
				for xx := x; xx < x+gw; xx++ {
					out.Set(xx, yy, ok_img.At(xx-x, yy-y))
				}
			}
		}

		cur_S := y
		fmt.Printf("Progress: %-10v\r", (cur_S*100)/all_S)
	}
}

func dist(r1, g1, b1, r2, g2, b2 float64) float64 {
	return Sqrt(Pow(r2-r1, 2) + Pow(b2-b1, 2) + Pow(g2-g1, 2))
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func minDist(src Image, rect Rectangle, info []ImageInfo) ImageInfo {
	sr, sg, sb := avgImg(src, rect)

	min_d := 100000.0
	i1 := ImageInfo{}

	min_d2 := 100000.0
	i2 := ImageInfo{}

	for _, v := range info {
		d := dist(sr, sg, sb, v.R, v.G, v.B)

		if d < min_d {
			min_d = d
			i1 = v
		}

		if d < min_d2 && v != i1 {
			min_d2 = d
			i2 = v
		}
	}

	num := randInt(0, 2)
	if num == 0 {
		return i1
	} else {
		return i2
	}
}

func avgImg(img Image, rect Rectangle) (r, g, b float64) {
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

func loadImg(file string) Image {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	return img
}

func proccessImgFile(in <-chan string, out chan<- ImageInfo, w, h uint, wg *sync.WaitGroup) {
	wg.Add(1)
	for file := range in {
		img := loadImg(file)
		img = resize.Resize(w, h, img, resize.Lanczos3)

		r, g, b := avgImg(img, img.Bounds())

		out <- ImageInfo{
			FilePaht: file,
			Img:      img,
			R:        r,
			G:        g,
			B:        b}
	}
	wg.Done()
}

func transferFiles(out chan<- string, path string, files []os.FileInfo) {
	for _, v := range files {
		out <- path + "/" + v.Name()
	}

	close(out)
}

func getFiles(dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	return files
}
