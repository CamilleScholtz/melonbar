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

func (bar *Bar) clock() {
	// Initialize block.
	block := bar.initBlock("clock", "?", 799, 'a', 0, "#445967", "#CCCCCC")

	// Notify that the next block can be initialized.
	bar.ready <- true

	for {
		// Compose block text.
		txt := time.Now().Format("Monday, January 2th 03:04 PM")

		// Redraw block.
		if block.diff(txt) {
			bar.redraw <- block
		}

		// Update every 45 seconds.
		time.Sleep(45 * time.Second)
	}
}

func (bar *Bar) music() {
	// Initialize block.
	block := bar.initBlock("music", "»  ", 660, 'r', -12, "#3C4F5B", "#CCCCCC")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Connect to MPD.
	c, err := mpd.Dial("tcp", ":6600")
	if err != nil {
		log.Fatalln(err)
	}

	// Keep connection alive by pinging ever 45 seconds.
	go func() {
		for {
			time.Sleep(time.Second * 45)

			if err := c.Ping(); err != nil {
				c, err = mpd.Dial("tcp", ":6600")
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}()

	// Show popup on clicking the left mouse button.
	block.actions["button1"] = func() error {
		if block.popup != nil {
			block.popup = block.popup.destroy()
			return nil
		}

		block.popup, err = initPopup(1920-304-29, 29, 304, 148, "#EEEEEE",
			"#021B21")
		if err != nil {
			return err
		}

		return block.popup.music(c)
	}

	// Toggle play/pause on clicking the right mouse button.
	block.actions["button3"] = func() error {
		status, err := c.Status()
		if err != nil {
			return err
		}

		return c.Pause(status["state"] != "pause")
	}

	// Previous song on scrolling up.
	block.actions["button4"] = func() error {
		return c.Previous()
	}

	// Next song on on scrolling down..
	block.actions["button5"] = func() error {
		return c.Next()
	}

	// Watch MPD for events.
	w, err := mpd.NewWatcher("tcp", ":6600", "", "player")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		cur, err := c.CurrentSong()
		if err != nil {
			log.Println(err)
		}
		sts, err := c.Status()
		if err != nil {
			log.Println(err)
		}

		// Compose text.
		var s string
		if sts["state"] == "pause" {
			s = "[paused] "
		}
		txt := "»      " + s + cur["Artist"] + " - " + cur["Title"]

		// Redraw block.
		if block.diff(txt) {
			bar.redraw <- block

			// Redraw popup.
			if block.popup != nil {
				if err := block.popup.music(c); err != nil {
					log.Println(err)
				}
			}
		}

		<-w.Event
	}
}

func (bar *Bar) todo() {
	// Initialize block.
	block := bar.initBlock("todo", "¢", 29, 'c', 0, "#5394C9", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Watch file for events.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	if err := w.Add("/home/onodera/.todo"); err != nil {
		log.Fatalln(err)
	}
	f, err := os.Open("/home/onodera/.todo")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		// Count file lines.
		s := bufio.NewScanner(f)
		s.Split(bufio.ScanLines)
		var c int
		for s.Scan() {
			c++
		}

		// Rewind file.
		if _, err := f.Seek(0, 0); err != nil {
			log.Println(err)
		}

		// Compose block text.
		txt := "¢ " + strconv.Itoa(c)

		// Redraw block.
		if block.diff(txt) {
			bar.redraw <- block
		}

		// Listen for next write event.
		ev := <-w.Events
		if ev.Op&fsnotify.Write != fsnotify.Write {
			continue
		}
	}
}

func (bar *Bar) window() {
	// Initialize blocks.
	bar.initBlock("windowIcon", "º", 21, 'l', 12, "#37BF8D", "#FFFFFF")
	block := bar.initBlock("window", "?", 200, 'c', 0, "#37BF8D", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// TODO: This doesn't check for window title changes.
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		// Only listen to `_NET_ACTIVE_WINDOW` events.
		atom, err := xprop.Atm(X, "_NET_ACTIVE_WINDOW")
		if err != nil {
			log.Println(err)
		}
		if ev.Atom != atom {
			return
		}

		// Get active window.
		id, err := ewmh.ActiveWindowGet(X)
		if err != nil {
			log.Println(err)
		}
		if id == 0 {
			return
		}

		// Compose block text.
		txt, err := ewmh.WmNameGet(X, id)
		if err != nil || len(txt) == 0 {
			txt, err = icccm.WmNameGet(X, id)
			if err != nil || len(txt) == 0 {
				txt = "?"
			}
		}
		if len(txt) > 34 {
			txt = txt[0:34] + "..."
		}

		// Redraw block.
		if block.diff(txt) {
			bar.redraw <- block
		}
	}).Connect(X, X.RootWin())
}

func (bar *Bar) workspace() {
	// Initialize block.
	blockWWW := bar.initBlock("www", "¼      www", 74, 'l', 10, "#5394C9",
		"#FFFFFF")
	blockIRC := bar.initBlock("irc", "½      irc", 67, 'l', 10, "#5394C9",
		"#FFFFFF")
	blockSRC := bar.initBlock("src", "¾      src", 70, 'l', 10, "#5394C9",
		"#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Change active workspace on clicking on one of the blocks.
	blockWWW.actions["button1"] = func() error {
		return ewmh.CurrentDesktopReq(X, 0)
	}
	blockIRC.actions["button1"] = func() error {
		return ewmh.CurrentDesktopReq(X, 1)
	}
	blockSRC.actions["button1"] = func() error {
		return ewmh.CurrentDesktopReq(X, 2)
	}

	var owsp uint
	var pwsp, nwsp int
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		// Only listen to `_NET_ACTIVE_WINDOW` events.
		atom, err := xprop.Atm(X, "_NET_CURRENT_DESKTOP")
		if err != nil {
			log.Println(err)
		}
		if ev.Atom != atom {
			return
		}

		// Get the current active desktop.
		wsp, err := ewmh.CurrentDesktopGet(X)
		if err != nil {
			log.Println(err)
		}

		// Set colors accordingly.
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

		if owsp != wsp {
			bar.redraw <- blockWWW
			bar.redraw <- blockIRC
			bar.redraw <- blockSRC

			owsp = wsp
		}
	}).Connect(X, X.RootWin())

	prevFun := func() error {
		return ewmh.CurrentDesktopReq(X, pwsp)
	}
	nextFun := func() error {
		return ewmh.CurrentDesktopReq(X, nwsp)
	}

	blockWWW.actions["button4"] = prevFun
	blockWWW.actions["button5"] = nextFun
	blockIRC.actions["button4"] = prevFun
	blockIRC.actions["button5"] = nextFun
	blockSRC.actions["button4"] = prevFun
	blockSRC.actions["button5"] = nextFun
}
