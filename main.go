package main

import (
	"path"
	"time"

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
	runBlock(bar.windowFun)
	runBlock(bar.workspaceFun)
	runBlock(bar.clockFun)
	runBlock(bar.musicFun)
	runBlock(bar.todoFun)

	for {
		if err := bar.draw(<-bar.redraw); err != nil {
			panic(err)
		}
	}
}

func runBlock(f func()) {
	go f()
	time.Sleep(time.Millisecond * 10)
}
