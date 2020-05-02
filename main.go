package main

import (
	"context"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/karalabe/hid"
	"golang.org/x/image/colornames"
)

var (
	vendorID  = uint16(0x20A0)
	productID = uint16(0x422d)
)

type keyStatus [8]byte

func (ks *keyStatus) Slice() []byte {
	return ks[:]
}

var keyMap = map[byte]string{
	30: "1",
	31: "2",
	32: "3",
	33: "3",
	34: "4",
	35: "5",
	36: "6",
	37: "7",
	38: "8",
	39: "9",
}

func (ks keyStatus) String() string {
	var ret string
	for _, b := range ks[2:] {
		if b == 0 {
			continue
		}
		if k, ok := keyMap[b]; ok {
			ret += k
		} else {
			ret += fmt.Sprintf("<%d>", b)
		}
	}
	return ret
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func runimg(ctx context.Context, imgCh <-chan string) {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 250, 250),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case imgName := <-imgCh:
			log.Printf("Got img %s", imgName)
			pic, err := loadPicture(imgName)
			if err != nil {
				panic(err)
			}
			sprite := pixel.NewSprite(pic, pic.Bounds())
			win.Clear(colornames.White)
			sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			win.Update()
			time.Sleep(200 * time.Millisecond)
			win.SetClosed(true)
		case <-ctx.Done():
			break
		}
	}
}

func run(imgCh chan string) error {
	devices := hid.Enumerate(vendorID, productID)
	if len(devices) == 0 {
		return fmt.Errorf("No device found with vendor:product ID %04x:%04xx", vendorID, productID)
	}
	log.Printf("Found %d device(s), using the first one", len(devices))
	dev, err := devices[0].Open()
	if err != nil {
		return err
	}
	defer dev.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pixelgl.Run(func() { runimg(ctx, imgCh) })
	var buf keyStatus
	for {
		n, err := dev.Read(buf.Slice())
		if err != nil {
			return err
		}
		if n == 0 {
			log.Printf("Finished")
			return nil
		}
		s := buf.String()
		if s == "" {
			fmt.Println("<no key is pressed>")
		} else {
			fmt.Println(s)
			switch s[len(s)-1] {
			case '1', '3', '7', '9':
				imgCh <- "angery.png"
			default:
				log.Printf("No image bound for %s", s[len(s)-1])
			}
		}
	}
}

func main() {
	imgCh := make(chan string, 1)
	if err := run(imgCh); err != nil {
		log.Fatal(err)
	}
}
