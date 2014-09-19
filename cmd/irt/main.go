package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
	"io/ioutil"
	"path/filepath"
	"sync"
	"runtime"
)

var (
	width  = flag.String("w", "", "width,  example: 80%, 200px")
	height = flag.String("h", "", "height, example: 80%, 100px")
	length = flag.String("l", "", "length, example: 80%, 100px")
	out    = flag.String("o", "", "output file name and path, example: ./out.jpg")
)

const (
	JPG = "jpg"
	GIF = "gif"
	PNG = "png"
)

func main() {
	log.Println("image resize start!")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	inputName := flag.Arg(0)
	if len(inputName) == 0 {
		log.Fatalln("Please input filename or directory!")
	}
	if inputName[len(inputName)-1:] == "/" {
		if len(*length) == 0 {
			log.Fatalln("Please input length")
		}
		execFiles(inputName)
	} else {
		if len(*width) == 0 && len(*height) == 0 {
			log.Fatalln("Please input width or height")
		}
		if len(*width) > 0 && len(*height) > 0 {
			log.Fatalln("Please input width or height")
		}
		execFile(inputName, false, "")
	}

}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func execFiles(inputName string) {
	outDir := inputName + "resized/"
	outDirExists, err := exists(outDir)
	if err != nil{
		log.Fatalln(err)
	}
	if !outDirExists{
		err := os.Mkdir(outDir, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}
	files, err := ioutil.ReadDir(inputName)
	if err != nil {
		log.Fatalln(err)
	}
	var wg sync.WaitGroup
	for _, f := range files {
		if !f.IsDir() {
			wg.Add(1)
			filename := inputName + f.Name()
			go func(fn, od string){
				defer wg.Done()
				execFile(fn, true, od)
			}(filename, outDir)
		}
	}
	wg.Wait()
}

func execFile(inputName string, dir bool, outDir string) {
	file, err := os.Open(inputName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if err != nil {
		log.Fatalln(err)
	}
	format := getFormat(file)
	img, err := getDecodedImage(file, format)
	if err != nil {
		log.Printf("%s:%s", file.Name(), format)
		//log.Fatalln(err)
		//return false
	}
	realWidth := img.Bounds().Size().X
	realHeight := img.Bounds().Size().Y

	var setWidth, setHeight uint
	if dir {
		if realWidth < realHeight {
			setWidth = 0
			setHeight = getLength(*length, realHeight)
		} else {
			setHeight = 0
			setWidth = getLength(*length, realWidth)
		}
	} else {
		setWidth, setHeight = getSize(*width, *height, realWidth, realHeight)
	}

	resizedImg := resize.Resize(setWidth, setHeight, img, resize.NearestNeighbor)

	var filename string
	var output *os.File
	if dir {
		_, f := filepath.Split(file.Name())
		filename = outDir + f
		output, err = os.Create(filename)
	} else {
		if len(*out) > 0 {
			filename = *out
			output, err = os.Create(*out)
		} else {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}
			filename = fmt.Sprintf("%s/resized.%s", pwd, format)
			output, err = os.Create(filename)
		}
	}

	if err != nil {
		log.Fatalln(err)
	}
	defer output.Close()
	createEncodeImage(format, output, resizedImg)

	log.Printf("input file: %s, width:%d, height:%d\n", inputName, realWidth, realHeight)
	log.Printf("output file: %s, width:%d, height:%d\n", filename, setWidth, setHeight)
}

func getLength(length string, realLength int) (setLength uint) {
	setLength = str2uint(length, realLength)
	return
}

func getSize(width, height string, realWidth, realHeight int) (setWidth, setHeight uint) {
	if len(width) > 0 {
		setHeight = 0
		setWidth = str2uint(width, realWidth)
	} else {
		setWidth = 0
		setHeight = str2uint(height, realHeight)
	}
	return
}

func str2uint(numStr string, imgSize int) (res uint) {
	if strings.Contains(numStr, "px") {
		str := strings.Trim(numStr, "px")
		val, _ := strconv.Atoi(str)
		res = uint(val)
	} else if strings.Contains(numStr, "%") {
		str := strings.Trim(numStr, "%")
		per, _ := strconv.Atoi(str)
		res = uint(imgSize * per / 100)
	} else {
		val, _ := strconv.Atoi(numStr)
		res = uint(val)
	}
	return
}

func getDecodedImage(file io.Reader, format string) (img image.Image, err error) {
	switch format {
	case JPG:
		img, err = jpeg.Decode(file)
	case GIF:
		img, err = gif.Decode(file)
	case PNG:
		img, err = png.Decode(file)
	default:
		img = nil
		err = &otherImageError{"other image format"}
	}
	return

}

type otherImageError struct {
	s string
}

func (self *otherImageError) Error() string {
	return self.s
}

func createEncodeImage(format string, file *os.File, image image.Image) {
	switch format {
	case JPG:
		jpeg.Encode(file, image, nil)
	case GIF:
		gif.Encode(file, image, nil)
	case PNG:
		png.Encode(file, image)
	}
}

func getFormat(file *os.File) string {
	bytes := make([]byte, 4)
	n, _ := file.ReadAt(bytes, 0)
	if n < 4 {
		return ""
	}
	if bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47 {
		return PNG
	}
	if bytes[0] == 0xFF && bytes[1] == 0xD8 {
		return JPG
	}
	if bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x38 {
		return GIF
	}
	if bytes[0] == 0x42 && bytes[1] == 0x4D {
		return "bmp"
	}
	return ""
}
