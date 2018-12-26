package main

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/fhs/gompd/mpd"
	"github.com/fsnotify/fsnotify"
	homedir "github.com/mitchellh/go-homedir"
)

func (bar *Bar) clockFun() {
	// Initialize block.
	block := bar.initBlock("clock", "?", 800, 'a', 0, "#445967", "#CCCCCC")

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

func (bar *Bar) musicFun() {
	// Initialize block.
	block := bar.initBlock("music", "»  ", 660, 'r', -12, "#3C4F5B", "#CCCCCC")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Connect to MPD.
	c, err := mpd.Dial("tcp", ":6600")
	if err != nil {
		panic(err)
	}

	// Keep connection alive by pinging ever 45 seconds.
	go func() {
		for {
			time.Sleep(time.Second * 45)

			if err := c.Ping(); err != nil {
				c, err = mpd.Dial("tcp", ":6600")
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	// Show popup on clicking the left mouse button.
	block.actions["button1"] = func() {
		if block.popup == nil {
			var err error
			block.popup, err = bar.initPopup(1920-304-29, 29, 304, 148,
				"#3C4F5B", "#CCCCCC")
			if err != nil {
				panic(err)
			}

			//popup.draw()
		} else {
			block.popup = block.popup.destroy()
		}
	}

	// Toggle play/pause on clicking the right mouse button.
	block.actions["button3"] = func() {
		status, err := c.Status()
		if err != nil {
			panic(err)
		}

		if err := c.Pause(status["state"] != "pause"); err != nil {
			panic(err)
		}
	}

	// Previous song on scrolling up.
	block.actions["button4"] = func() {
		if err := c.Previous(); err != nil {
			panic(err)
		}
	}

	// Next song on on scrolling down..
	block.actions["button5"] = func() {
		if err := c.Next(); err != nil {
			panic(err)
		}
	}

	// Watch MPD for events.
	w, err := mpd.NewWatcher("tcp", ":6600", "", "player")
	if err != nil {
		panic(err)
	}

	for {
		cur, err := c.CurrentSong()
		if err != nil {
			panic(err)
		}
		sts, err := c.Status()
		if err != nil {
			panic(err)
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
		}

		<-w.Event
	}
}

func (bar *Bar) todoFun() {
	// Initialize block.
	block := bar.initBlock("todo", "¢", 29, 'c', 0, "#5394C9", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Find `.todo` file.
	hd, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	fp := path.Join(hd, ".todo")

	// Watch file for events.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	if err := w.Add(fp); err != nil {
		panic(err)
	}
	f, err := os.Open(fp)
	if err != nil {
		panic(err)
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
			panic(err)
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

func (bar *Bar) windowFun() {
	// Initialize blocks.
	bar.initBlock("window", "º", 21, 'l', 12, "#37BF8D", "#FFFFFF")
	block := bar.initBlock("window", "?", 200, 'c', 0, "#37BF8D", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// TODO: This doesn't check for window title changes.
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		// Only listen to `_NET_ACTIVE_WINDOW` events.
		atom, err := xprop.Atm(bar.xu, "_NET_ACTIVE_WINDOW")
		if err != nil {
			panic(err)
		}
		if ev.Atom != atom {
			return
		}

		// Get active window.
		id, err := ewmh.ActiveWindowGet(bar.xu)
		if err != nil {
			panic(err)
		}
		if id == 0 {
			return
		}

		// Compose block text.
		txt, err := ewmh.WmNameGet(bar.xu, id)
		if err != nil || len(txt) == 0 {
			txt, err = icccm.WmNameGet(bar.xu, id)
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
	}).Connect(bar.xu, bar.xu.RootWin())
}

func (bar *Bar) workspaceFun() {
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
	blockWWW.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 0); err != nil {
			panic(err)
		}
	}
	blockIRC.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 1); err != nil {
			panic(err)
		}
	}
	blockSRC.actions["button1"] = func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, 2); err != nil {
			panic(err)
		}
	}

	var owsp uint
	var pwsp, nwsp int
	xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
		PropertyNotifyEvent) {
		// Only listen to `_NET_ACTIVE_WINDOW` events.
		atom, err := xprop.Atm(bar.xu, "_NET_CURRENT_DESKTOP")
		if err != nil {
			panic(err)
		}
		if ev.Atom != atom {
			return
		}

		// Get the current active desktop.
		wsp, err := ewmh.CurrentDesktopGet(bar.xu)
		if err != nil {
			panic(err)
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
	}).Connect(bar.xu, bar.xu.RootWin())

	prevFun := func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, pwsp); err != nil {
			panic(err)
		}
	}
	nextFun := func() {
		if err := ewmh.CurrentDesktopReq(bar.xu, nwsp); err != nil {
			panic(err)
		}
	}

	blockWWW.actions["button4"] = prevFun
	blockWWW.actions["button5"] = nextFun
	blockIRC.actions["button4"] = prevFun
	blockIRC.actions["button5"] = nextFun
	blockSRC.actions["button4"] = prevFun
	blockSRC.actions["button5"] = nextFun
}
