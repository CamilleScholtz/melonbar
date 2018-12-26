package main

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	hd, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	bar, err := initBar(0, 0, 1920, 29, path.Join(hd, ".fonts/plan9/cure.font"))
	if err != nil {
		panic(err)
	}

	// Run bar block functions.
	go bar.initBlocks([]func(){
		bar.windowFun,
		bar.workspaceFun,
		bar.clockFun,
		bar.musicFun,
		bar.todoFun,
	})

	// Listen for redraw events.
	bar.listen()
}
