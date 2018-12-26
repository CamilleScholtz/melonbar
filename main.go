package main

import "time"

func main() {
	bar, err := initBar(0, 0, 1920, 29, "./vendor/font/cure.font")
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
