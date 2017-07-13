package main

import (
	"log"
	"time"
)

func main() {
	bar, err := initBar(0, 0, 1920, 29,
		"/home/onodera/.fonts/cure.ttf", 11)
	if err != nil {
		log.Fatal(err)
	}

	// Run bar block functions. Make sure to sleep a millisecond after
	// each block, else they won't appear in the right order.
	go bar.windowFun()
	time.Sleep(time.Millisecond)

	go bar.workspaceFun()
	time.Sleep(time.Millisecond)

	go bar.clockFun()
	time.Sleep(time.Millisecond)

	go bar.musicFun()
	time.Sleep(time.Millisecond)

	go bar.todoFun()
	time.Sleep(time.Millisecond)

	for {
		if err := bar.paint(<-bar.redraw); err != nil {
			log.Fatal(err)
		}
	}
}
