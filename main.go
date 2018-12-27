package main

import (
	"log"
	"path"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/gobuffalo/packr/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/plan9font"
)

var (
	box = packr.New("box", "./box")

	// Connection to the X server.
	X *xgbutil.XUtil

	// The font that should be used.
	face font.Face
)

func main() {
	// Initialize X.
	if err := initX(); err != nil {
		log.Fatalln(err)
	}

	// Initialize font.
	if err := initFont(); err != nil {
		log.Fatalln(err)
	}

	// Initialize bar.
	bar, err := initBar(0, 0, 1920, 29)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize blocks.
	go bar.initBlocks([]func(){
		bar.window,
		bar.workspace,
		bar.clock,
		bar.music,
		bar.todo,
	})

	// Listen for redraw events.
	bar.listen()
}

func initX() error {
	// Set up a connection to the X server.
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		return err
	}

	// Run the main X event loop, this is used to catch events.
	go xevent.Main(X)

	// Listen to the root window for property change events, used to check if
	// the user changed the focused window or active workspace for example.
	return xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskPropertyChange)
}

func initFont() error {
	fr := func(name string) ([]byte, error) {
		return box.Find(path.Join("fonts", name))
	}
	font, err := box.Find("fonts/cure.font")
	if err != nil {
		return err
	}
	face, err = plan9font.ParseFont(font, fr)
	if err != nil {
		return err
	}

	return nil
}
