package main

import (
	"log"
	"time"
)

func main() {
	b, err := initBar(0, 0, 1920, 29, "/home/onodera/.fonts/cure.ttf",
		11)
	if err != nil {
		log.Fatal(err)
	}

	// Run bar block functions. Make sure to sleep a millisecond after
	// each block, else they won't appear in the right order.
	go b.windowFun(220, "#37BF8D", "#FFFFFF")
	time.Sleep(time.Millisecond)
	go b.workspaceFun(74, 67, 70, "#5394C9", "#72A7D3", "#FFFFFF")
	time.Sleep(time.Millisecond * 3)
	go b.clockFun(900, "#445967", "#CCCCCC")
	time.Sleep(time.Millisecond)
	go b.musicFun(600, "#3C4F5B", "#CCCCCC")
	time.Sleep(time.Millisecond)
	//go b.todoFun(200)
	//time.Sleep(time.Millisecond)

	for {
		b.draw(<-b.redraw)
		b.paint()
	}
}
