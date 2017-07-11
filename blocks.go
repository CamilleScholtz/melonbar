package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	owm "github.com/briandowns/openweathermap"
	"github.com/fhs/gompd/mpd"
	"github.com/fsnotify/fsnotify"
)

func (bar *Bar) clockFun() {
	bar.initBlock("clock", "?", 800, 'c', 300, "#445967", "#CCCCCC")

	init := true
	for {
		if !init {
			time.Sleep(20 * time.Second)
		}
		init = false

		t := time.Now()

		bar.updateBlockTxt("clock", t.Format(
			"Monday, January 2th 02:04 PM"))
		bar.redraw <- "clock"
	}
}

func (bar *Bar) musicFun() error {
	bar.initBlock("music", "?", 660, 'r', -10, "#3C4F5B", "#CCCCCC")

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

		// TODO: Is it maybe possible to not create a new conn each
		// loop?
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

		s, err := conn.Status()
		if err != nil {
			log.Print(err)
			continue
		}
		var state string
		if s["state"] == "play" {
			state = "[playing] "
		} else {
			state = "[paused] "
		}

		bar.updateBlockTxt("music", state+cur["Artist"]+" - "+
			cur["Title"])
		bar.redraw <- "music"
	}
}

func (bar *Bar) todoFun() {
	bar.initBlock("todo", "?", 29, 'c', 0, "#5394C9", "#FFFFFF")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if err := watcher.Add("/home/onodera/todo"); err != nil {
		log.Fatal(err)
	}
	file, err := os.Open("/home/onodera/todo")
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

		bar.updateBlockTxt("todo", strconv.Itoa(c))
		bar.redraw <- "todo"
	}
}

func (bar *Bar) weatherFun() {
	bar.initBlock("weather", "?", 29, 'r', 0, "#5394C9", "#FFFFFF")

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

		/*
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
		*/

		bar.updateBlockTxt("weather", strconv.FormatFloat(w.Main.Temp,
			'f', 0, 64)+" °C")
		bar.redraw <- "weather"
	}
}

func (bar *Bar) windowFun() {
	bar.initBlock("window", "?", 220, 'c', 0, "#37BF8D", "#FFFFFF")

	init := true
	var Owin string
	for {
		if !init {
			ev, xgbErr := bar.xu.Conn().WaitForEvent()
			if xgbErr != nil {
				log.Print(xgbErr)
				continue
			}

			atom, err := xprop.Atm(bar.xu, "_NET_ACTIVE_WINDOW")
			if ev.(xproto.PropertyNotifyEvent).Atom != atom {
				continue
			}
			if err != nil {
				log.Print(err)
				continue
			}
		}
		init = false

		id, err := ewmh.ActiveWindowGet(bar.xu)
		if err != nil {
			log.Print(err)
			continue
		}

		win, err := ewmh.WmNameGet(bar.xu, id)
		if err != nil {
			log.Print(err)
			continue
		}
		if Owin == win {
			continue
		}
		Owin = win

		bar.updateBlockTxt("window", win)
		bar.redraw <- "window"
	}
}

func (bar *Bar) workspaceFun() {
	bar.initBlock("www", "www", 74, 'c', 0, "#5394C9", "#FFFFFF")
	bar.initBlock("irc", "irc", 67, 'c', 0, "#5394C9", "#FFFFFF")
	bar.initBlock("src", "src", 70, 'c', 0, "#5394C9", "#FFFFFF")

	init := true
	var Owsp uint
	for {
		if !init {
			ev, xgbErr := bar.xu.Conn().WaitForEvent()
			if xgbErr != nil {
				log.Print(xgbErr)
				continue
			}

			atom, err := xprop.Atm(bar.xu, "WINDOWCHEF_ACTIVE_GROUPS")
			if ev.(xproto.PropertyNotifyEvent).Atom != atom {
				continue
			}
			if err != nil {
				log.Print(err)
				continue
			}
		}
		init = false

		wsp, err := ewmh.CurrentDesktopGet(bar.xu)
		if err != nil {
			log.Print(err)
			continue
		}

		if Owsp == wsp {
			continue
		}
		Owsp = wsp

		switch wsp {
		case 0:
			bar.updateBlockBg("www", "#72A7D3")
			bar.updateBlockBg("irc", "#5394C9")
			bar.updateBlockBg("src", "#5394C9")
		case 1:
			bar.updateBlockBg("www", "#5394C9")
			bar.updateBlockBg("irc", "#72A7D3")
			bar.updateBlockBg("src", "#5394C9")
		case 2:
			bar.updateBlockBg("www", "#5394C9")
			bar.updateBlockBg("irc", "#5394C9")
			bar.updateBlockBg("src", "#72A7D3")
		}
		bar.redraw <- "www"
		bar.redraw <- "irc"
		bar.redraw <- "src"
	}
}
