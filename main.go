package main

import (
	"log"

	"github.com/AndreKR/multiface"
	"github.com/BurntSushi/xgbutil"
	"github.com/gobuffalo/packr/v2"
)

var (
	box = packr.New("box", "./box")

	// Connection to the X server.
	X *xgbutil.XUtil

	// The multifont face that should be used.
	face *multiface.Face
)

func main() {
	// Initialize X.
	if err := initX(); err != nil {
		log.Fatalln(err)
	}

	// Initialize font face.
	if err := initFace(); err != nil {
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
