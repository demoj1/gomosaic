package main

import (
	"encoding/json"
	//"fmt"
	"github.com/diman3241/ImageMosaic/img"
	"io/ioutil"
	//"log"
)

func main() {
	//	f := img.StartAverageImage()
	b, _ := ioutil.ReadFile("json_images")
	f := &img.Files{}

	json.Unmarshal(b, f)

	img.CreatePicture("ia.jpg", "new.jpg", f)
	//var a int
	//fmt.Scanf("%v", &a)

	// b, err := json.Marshal(f)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = ioutil.WriteFile("json_images", b, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
