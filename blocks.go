package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/fhs/gompd/mpd"
	"github.com/rkoesters/xdg/basedir"
)

func (bar *Bar) clock() {
	// Initialize block.
	block := bar.initBlock("clock", "?", 799, 'a', 0, "#445967", "#CCCCCC")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Show popup on clicking the left mouse button.
	block.actions[1] = func() error {
		if block.popup != nil {
			block.popup = block.popup.destroy()
			return nil
		}

		var err error
		block.popup, err = initPopup((bar.w/2)-(178/2), bar.h, 178, 129,
			"#EEEEEE", "#021B21")
		if err != nil {
			return err
		}

		return block.popup.clock()
	}

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
	block := bar.initBlock("music", " Ƅ  ", 660, 'r', -12, "#3C4F5B", "#CCCCCC")

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
	block.actions[2] = func() error {
		if block.popup != nil {
			block.popup = block.popup.destroy()
			return nil
		}

		block.popup, err = initPopup(bar.w-304-29, bar.h, 304, 148, "#EEEEEE",
			"#021B21")
		if err != nil {
			return err
		}

		return block.popup.music(c)
	}

	// Toggle play/pause on clicking the right mouse button.
	block.actions[3] = func() error {
		status, err := c.Status()
		if err != nil {
			return err
		}

		return c.Pause(status["state"] != "pause")
	}

	// Previous song on scrolling up.
	block.actions[4] = func() error {
		return c.Previous()
	}

	// Next song on on scrolling down..
	block.actions[5] = func() error {
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
		txt := " Ƅ  " + s + cur["Artist"] + " - " + cur["Title"]

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
	block := bar.initBlock("todo", "ƅ", bar.h, 'c', 1, "#5394C9", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Show popup on clicking the left mouse button.
	block.actions[1] = func() error {
		cmd := exec.Command("st", "micro", "-savecursor", "false", path.Join(
			basedir.Home, ".todo"))
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	// Show count popup.
	var err error
	block.popup, err = initPopup(bar.w-19-16, bar.h-8, 16, 12, "#EB6084",
		"#FFFFFF")
	if err != nil {
		log.Fatalln(err)
	}
	if err := block.popup.todo(); err != nil {
		log.Fatalln(err)
	}
}

func (bar *Bar) window() {
	// Initialize blocks.
	bar.initBlock("windowIcon", "ƀ", 21, 'r', 3, "#37BF8D", "#FFFFFF")
	block := bar.initBlock("window", "?", 200, 'c', 0, "#37BF8D", "#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Enable input on clicking the left mouse button.
	block.actions[1] = func() error {
		str := ""

		xevent.KeyPressFun(func(_ *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			k := keybind.LookupString(X, ev.State, ev.Detail)
			// TODO: Is there a better way to ignore modifiers?
			if len(k) == 1 {
				str += k
			}

			if keybind.KeyMatch(X, "Return", ev.State, ev.Detail) {
				keybind.SmartUngrab(X)

				fmt.Println(str)

				cmd := exec.Command(str)
				cmd.Stdout = os.Stdout
				// TODO: Return this in the super function.
				cmd.Run()
			}
		}).Connect(X, bar.win.Id)

		if err := keybind.SmartGrab(X, bar.win.Id); err != nil {
			log.Println(err)
		}

		return nil
	}

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
		txt = trim(txt, 34)

		// Redraw block.
		if block.diff(txt) {
			bar.redraw <- block
		}
	}).Connect(X, X.RootWin())
}

func (bar *Bar) workspace() {
	// Initialize block.
	blockWWW := bar.initBlock("www", "Ɓ   www", 74, 'l', 12, "#5394C9",
		"#FFFFFF")
	blockIRC := bar.initBlock("irc", "Ƃ   irc", 67, 'l', 12, "#5394C9",
		"#FFFFFF")
	blockSRC := bar.initBlock("src", "ƃ   src", 70, 'l', 12, "#5394C9",
		"#FFFFFF")

	// Notify that the next block can be initialized.
	bar.ready <- true

	// Change active workspace on clicking on one of the blocks.
	blockWWW.actions[1] = func() error {
		return ewmh.CurrentDesktopReq(X, 0)
	}
	blockIRC.actions[1] = func() error {
		return ewmh.CurrentDesktopReq(X, 1)
	}
	blockSRC.actions[1] = func() error {
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

	blockWWW.actions[4] = prevFun
	blockWWW.actions[5] = nextFun
	blockIRC.actions[4] = prevFun
	blockIRC.actions[5] = nextFun
	blockSRC.actions[4] = prevFun
	blockSRC.actions[5] = nextFun
}
