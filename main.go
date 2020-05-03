package main

// requires libxdo-dev

import (
	"flag"
	"fmt"
	_ "image/png"
	"os"

	"github.com/DanieleDaccurso/goxdo"
	"github.com/karalabe/hid"
)

var (
	vendorID  = uint16(0x20A0)
	productID = uint16(0x422d)

	debug   = func(s string, args ...interface{}) {}
	debugln = func(s string) { debug(s + "\n") }

	flagNoKeypress = flag.Bool("n", false, "do not send unicode keypresses")
	flagDoDebug    = flag.Bool("d", false, "print debug output")
)

type keyStatus [8]byte

func (ks *keyStatus) Slice() []byte {
	return ks[:]
}

var keyMap = map[byte]string{
	30: "1",
	31: "2",
	32: "3",
	33: "4",
	34: "5",
	35: "6",
	36: "7",
	37: "8",
	38: "9",
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

func keypress(s string) {
	var ucode string
	switch s {
	case "1", "3", "7", "9":
		// angery
		ucode = "U1F620"
	case "2":
		// haha
		ucode = "U1F923"
	case "4":
		// wow
		ucode = "U1F62E"
	case "5":
		// like
		ucode = "U1F44D"
	case "6":
		// love
		ucode = "U2764"
	case "8":
		// rainbow
		ucode = "U1F308"
	}
	// no specific window
	window := goxdo.Window(0)
	xdo := goxdo.NewXdo()
	// equivalent to "xdotool key <ucode>"
	xdo.SendKeysequenceWindow(window, ucode, 0)
}

func run(noKeypress bool) error {
	devices := hid.Enumerate(vendorID, productID)
	if len(devices) == 0 {
		return fmt.Errorf("No device found with vendor:product ID %04x:%04xx", vendorID, productID)
	}
	debug("Found %d device(s), using the first one\n", len(devices))
	dev, err := devices[0].Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := dev.Close(); err != nil {
			debug("Warning: failed to close device: %v\n", err)
		}
	}()

	var buf keyStatus
	for {
		n, err := dev.Read(buf.Slice())
		if err != nil {
			return err
		}
		if n == 0 {
			debugln("Finished")
			return nil
		}
		debug("raw: %v\n", buf[:n])
		s := buf.String()
		if s == "" {
			debugln("<released>")
		} else {
			debugln(s)
			if !noKeypress {
				go keypress(s)
			}
		}
	}
}

func main() {
	flag.Parse()
	if *flagDoDebug {
		debug = func(s string, args ...interface{}) {
			fmt.Printf(s, args...)
		}
	}
	if err := run(*flagNoKeypress); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
