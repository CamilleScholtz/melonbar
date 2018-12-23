package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/fhs/gompd/mpd"
	"github.com/fsnotify/fsnotify"
)

func (bar *Bar) clockFun() {
	block := bar.initBlock("clock", "?", 800, 'a', 0, "#445967", "#CCCCCC")

	init := true
	for {
		if !init {
			time.Sleep(20 * time.Second)
		}
		init = false

		txt := time.Now().Format("Monday, January 2th 03:04 PM")
		if block.txt == txt {
			continue
		}

		block.txt = txt
		bar.redraw <- block
	}
}

func (bar *Bar) musicFun() error {
	block := bar.initBlock("music", "?", 660, 'r', -13, "#3C4F5B", "#CCCCCC")

	block.actions["button1"] = func() {
		if block.popup == nil {
			var err error
			block.popup, err = bar.initPopup(1920-304-29, 29, 304, 148,
				"#3C4F5B", "#CCCCCC")
			if err != nil {
				log.Print(err)
			}

			//popup.draw()
		} else {
			block.popup = block.popup.destroy()
		}
	}
	block.actions["button3"] = func() {
		conn, err := mpd.Dial("tcp", ":6600")
		if err != nil {
			log.Print(err)
		}
		defer conn.Close()

		status, err := conn.Status()
		if err != nil {
			log.Print(err)
		}

		if err := conn.Pause(status["state"] != "pause"); err != nil {
			log.Print(err)
		}
	}
	block.actions["button4"] = func() {
		conn, err := mpd.Dial("tcp", ":6600")
		if err != nil {
			log.Print(err)
		}
		defer conn.Close()

		if err := conn.Previous(); err != nil {
			log.Print(err)
		}
	}
	block.actions["button5"] = func() {
		conn, err := mpd.Dial("tcp", ":6600")
		if err != nil {
			log.Print(err)
		}
		defer conn.Close()

		if err := conn.Next(); err != nil {
			log.Print(err)
		}
	}

	watcher, err := mpd.NewWatcher("tcp", ":6600", "", "player")
	if err != nil {
		log.Fatal(err)
	}
	var conn *mpd.Client
	init := true
	for {
		if !init {
			conn.Close()
			<-watcher.Event
		}
		init = false

		// TODO: Is it maybe possible to not create a new connection each loop?
		conn, err = mpd.Dial("tcp", ":6600")
		if err != nil {
			log.Print(err)
			continue
		}

		cur, err := conn.CurrentSong()
		if err != nil {
			log.Print(err)
			continue
		}

		status, err := conn.Status()
		if err != nil {
			log.Print(err)
			continue
		}

		var state string
		if status["state"] == "pause" {
			state = "[paused] "
		}

		txt := state + cur["Artist"] + " - " + cur["Title"]
		if block.txt == txt {
			continue
		}

		block.txt = txt
		bar.redraw <- block
	}
}

func (bar *Bar) todoFun() {
	block := bar.initBlock("todo", "?", 29, 'c', 0, "#5394C9", "#FFFFFF")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if err := watcher.Add("/home/onodera/.todo"); err != nil {
		log.Fatal(err)
	}
	file, err := os.Open("/home/onodera/.todo")
	if err != nil {
		log.Fatal(err)
	}
	init := true
	for {
		if !init {
			ev := <-watcher.Events
			if ev.Op&fsnotify.Write != fsnotify.Write {
				continue
			}
		}
		init = false

		s := bufio.NewScanner(file)
		s.Split(bufio.ScanLines)
		var c int
		for s.Scan() {
			c++
		}
		if _, err := file.Seek(0, 0); err != nil {
			log.Print(err)
			continue
		}

		txt := strconv.Itoa(c)
		if block.txt == txt {
			continue
		}

		block.txt = txt
		bar.redraw <- block
	}
}

/*
func (bar *Bar) weatherFun() {
	block := bar.initBlock("weather", "?", 29, 'r', 0, "#5394C9",
		"#FFFFFF")

	w, err := owm.NewCurrent("C", "en")
	if err != nil {
		log.Fatalln(err)
	}
	init := true
	for {
		if !init {
			time.Sleep(200 * time.Second)
		}
		init = false

		if err := w.CurrentByID(2758106); err != nil {
			log.Print(err)
			continue
		}

		var state uint
		switch w.Weather[0].Icon[0:2] {
		case "01":
			state = 0
		case "02":
			state = 1
		case "03":
			state = 2
		case "04":
			state = 3
		case "09":
			state = 4
		case "10":
			state = 5
		case "11":
			state = 6
		case "13":
			state = 7
		case "50":
			state = 8
		}

		block.txt = strconv.FormatFloat(w.Main.Temp, 'f', 0, 64) +
			" °C"
		bar.redraw <- block
	}
}
*/

func (bar *Bar) windowFun() {
	block := bar.initBlock("window", "?", 220, 'c', 0, "#37BF8D", "#FFFFFF")

	// TODO: I'm not sure how I can use init (to prevent a black bar) here?
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		atom, err := xprop.Atm(bar.xu, "_NET_ACTIVE_WINDOW")
		if ev.Atom != atom {
			return
		}
		if err != nil {
			log.Print(err)
			return
		}

		id, err := ewmh.ActiveWindowGet(bar.xu)
		if err != nil {
			log.Print(err)
			return
		}

		txt, err := ewmh.WmNameGet(bar.xu, id)
		if err != nil || len(txt) == 0 {
			txt, err = icccm.WmNameGet(bar.xu, id)
			if err != nil || len(txt) == 0 {
				txt = "?"
			}
		}
		if block.txt == txt {
			return
		}

		block.txt = txt
		bar.redraw <- block
	}).Connect(bar.xu, bar.xu.RootWin())
}

func (bar *Bar) workspaceFun() {
	blockWWW := bar.initBlock("www", "www", 74, 'c', 0, "#5394C9", "#FFFFFF")
	blockWWW.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 0); err != nil {
			log.Println(err)
		}
	}

	blockIRC := bar.initBlock("irc", "irc", 67, 'c', 0, "#5394C9", "#FFFFFF")
	blockIRC.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 1); err != nil {
			log.Println(err)
		}
	}

	blockSRC := bar.initBlock("src", "src", 70, 'c', 0, "#5394C9", "#FFFFFF")
	blockSRC.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 2); err != nil {
			log.Println(err)
		}
	}

	// TODO: I'm not sure how I can use init (to prevent a black bar) here?
	var owsp uint
	var pwsp, nwsp int
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		atom, err := xprop.Atm(bar.xu, "_NET_CURRENT_DESKTOP")
		if ev.Atom != atom {
			return
		}
		if err != nil {
			log.Print(err)
			return
		}

		wsp, err := ewmh.CurrentDesktopGet(bar.xu)
		if err != nil {
			log.Print(err)
			return
		}

		switch wsp {
		case 0:
			blockWWW.bg = "#72A7D3"
			blockIRC.bg = "#5394C9"
			blockSRC.bg = "#5394C9"

			pwsp = 2
			nwsp = 1
		case 1:
			blockWWW.bg = "#5394C9"
			blockIRC.bg = "#72A7D3"
			blockSRC.bg = "#5394C9"

			pwsp = 0
			nwsp = 2
		case 2:
			blockWWW.bg = "#5394C9"
			blockIRC.bg = "#5394C9"
			blockSRC.bg = "#72A7D3"

			pwsp = 1
			nwsp = 0
		}

		if owsp == wsp {
			return
		}
		owsp = wsp

		bar.redraw <- blockWWW
		bar.redraw <- blockIRC
		bar.redraw <- blockSRC
	}).Connect(bar.xu, bar.xu.RootWin())

	prevFun := func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, pwsp); err != nil {
			log.Println(err)
		}
	}
	nextFun := func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, nwsp); err != nil {
			log.Println(err)
		}
	}

	blockWWW.actions["button4"] = prevFun
	blockWWW.actions["button5"] = nextFun
	blockIRC.actions["button4"] = prevFun
	blockIRC.actions["button5"] = nextFun
	blockSRC.actions["button4"] = prevFun
	blockSRC.actions["button5"] = nextFun
}
