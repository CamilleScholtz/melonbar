package main

import (
	"log"
	"embed"

	"github.com/AndreKR/multiface"
	"github.com/BurntSushi/xgbutil"
)

var (
	//go:embed runtime/*
	runtime embed.FS

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

	// Initialize blocks and popups.
	bar.initBlocks()
	bar.initPopups()

	// Draw blocks.
	go bar.drawBlocks()

	// Listen for redraw events.
	bar.listen()
}
