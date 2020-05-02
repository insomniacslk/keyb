package main

import (
	"fmt"
	_ "image/png"
	"log"

	"github.com/karalabe/hid"
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

func run() error {
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
			fmt.Println("<released>")
		} else {
			fmt.Println(s)
		}
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
