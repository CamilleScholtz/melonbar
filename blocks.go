package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/fhs/gompd/mpd"
)

func (bar *Bar) initBlocks() {
	bar.blocks.Set("window-icon", &Block{
		txt:   "ƀ",
		w:     21,
		align: 'r',
		xoff:  3,
		bg:    xgraphics.BGRA{B: 141, G: 191, R: 55, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {},
	})

	bar.blocks.Set("window", &Block{
		txt:   "?",
		w:     200,
		align: 'c',
		xoff:  0,
		bg:    xgraphics.BGRA{B: 141, G: 191, R: 55, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {
			block := bar.block("window")

			// Redraw block function.
			t := func(id xproto.Window) {
				// Set new block text.
				txt, err := ewmh.WmNameGet(X, id)
				if err != nil || len(txt) == 0 {
					txt, err = icccm.WmNameGet(X, id)
					if err != nil || len(txt) == 0 {
						txt = "?"
					}
				}
				txt = trim(txt, 34)

				// Return if the text is the same.
				if txt == block.txt {
					return
				}
				block.txt = txt

				// Redraw block.
				bar.redraw <- block
			}

			// Variable where we store the (previous) xwindow.
			var xid *xwindow.Window

			// Get window ID function.
			f := func() {
				// Get active window.
				id, err := ewmh.ActiveWindowGet(X)
				if err != nil {
					log.Println(err)
					return
				}
				if id == 0 {
					return
				}

				// Stop listening to the previous window.
				if xid != nil {
					xid.Detach()
				}

				// Create xwindow from active window.
				xid = xwindow.New(X, id)

				// Listen to this window for window name changes.
				xid.Listen(xproto.EventMaskPropertyChange)
				xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
					PropertyNotifyEvent) {
					// Only listen to `_NET_WM_NAME` events.
					atom, err := xprop.Atm(X, "_NET_WM_NAME")
					if err != nil {
						log.Println(err)
						return
					}
					if ev.Atom != atom {
						return
					}

					t(id)
				}).Connect(X, id)

				t(id)
			}

			// Listen for window change event, execute `f()` accordingly.
			xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
				PropertyNotifyEvent) {
				// Only listen to `_NET_ACTIVE_WINDOW` events.
				atom, err := xprop.Atm(X, "_NET_ACTIVE_WINDOW")
				if err != nil {
					log.Println(err)
					return
				}
				if ev.Atom != atom {
					return
				}

				f()
			}).Connect(X, X.RootWin())

			// Execute `f()` one time initially.
			f()
		},
	})

	bar.blocks.Set("workspace-www", &Block{
		txt:   "Ɓ   www",
		w:     74,
		align: 'l',
		xoff:  12,
		bg:    xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {},

		actions: map[xproto.Button]func() error{
			1: func() error {
				return ewmh.CurrentDesktopReq(X, 0)
			},
			4: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["prev"].(int))
			},
			5: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["next"].(int))
			},
		},
	})

	bar.blocks.Set("workspace-irc", &Block{
		txt:   "Ƃ   irc",
		w:     67,
		align: 'l',
		xoff:  12,
		bg:    xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {},

		actions: map[xproto.Button]func() error{
			1: func() error {
				return ewmh.CurrentDesktopReq(X, 1)
			},
			4: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["prev"].(int))
			},
			5: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["next"].(int))
			},
		},
	})

	bar.blocks.Set("workspace-src", &Block{
		txt:   "ƃ   src",
		w:     70,
		align: 'l',
		xoff:  12,
		bg:    xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {},

		actions: map[xproto.Button]func() error{
			1: func() error {
				return ewmh.CurrentDesktopReq(X, 2)
			},
			4: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["prev"].(int))
			},
			5: func() error {
				return ewmh.CurrentDesktopReq(X, bar.store["next"].(int))
			},
		},
	})

	bar.blocks.Set("workspace", &Block{
		script: true,

		update: func() {
			www := bar.block("workspace-www")
			irc := bar.block("workspace-irc")
			src := bar.block("workspace-src")

			f := func() {
				// Get the current active desktop.
				wsp, err := ewmh.CurrentDesktopGet(X)
				if err != nil {
					log.Println(err)
					return
				}

				// Set new block colors and the previous/next workspaces.
				switch wsp {
				case 0:
					www.bg = xgraphics.BGRA{B: 211, G: 167, R: 114, A: 0xFF}
					irc.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}
					src.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}

					bar.store["prev"] = 2
					bar.store["next"] = 1
				case 1:
					www.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}
					irc.bg = xgraphics.BGRA{B: 211, G: 167, R: 114, A: 0xFF}
					src.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}

					bar.store["prev"] = 0
					bar.store["next"] = 2
				case 2:
					www.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}
					irc.bg = xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF}
					src.bg = xgraphics.BGRA{B: 211, G: 167, R: 114, A: 0xFF}

					bar.store["prev"] = 1
					bar.store["next"] = 0
				}

				// Redraw block.
				bar.redraw <- www
				bar.redraw <- irc
				bar.redraw <- src
			}

			// Listen for workspace change event, execute `f()` accordingly.
			xevent.PropertyNotifyFun(func(_ *xgbutil.XUtil, ev xevent.
				PropertyNotifyEvent) {
				// Only listen to `_NET_ACTIVE_WINDOW` events.
				atom, err := xprop.Atm(X, "_NET_CURRENT_DESKTOP")
				if err != nil {
					log.Println(err)
					return
				}
				if ev.Atom != atom {
					return
				}

				f()
			}).Connect(X, X.RootWin())

			// Execute `f()` one time initially.
			f()
		},
	})

	bar.blocks.Set("clock", &Block{
		txt:   "?",
		w:     799,
		align: 'a',
		xoff:  0,
		bg:    xgraphics.BGRA{B: 103, G: 89, R: 68, A: 0xFF},
		fg:    xgraphics.BGRA{B: 204, G: 204, R: 204, A: 0xFF},

		update: func() {
			for {
				block := bar.block("clock")

				// Set new block text.
				block.txt = time.Now().Format("Monday, January 2th 03:04 PM")

				// Redraw block.
				bar.redraw <- block

				// Update every 45 seconds.
				time.Sleep(45 * time.Second)
			}
		},

		actions: map[xproto.Button]func() error{
			1: func() error {
				return bar.drawPopup("clock")
			},
		},
	})

	bar.blocks.Set("music", &Block{
		txt:   " Ƅ  ",
		w:     660,
		align: 'r',
		xoff:  -12,
		bg:    xgraphics.BGRA{B: 91, G: 79, R: 60, A: 0xFF},
		fg:    xgraphics.BGRA{B: 204, G: 204, R: 204, A: 0xFF},

		update: func() {
			block := bar.block("music")
			popup := bar.popup("music")

			// Connect to MPD.
			var err error
			bar.store["mpd"], err = mpd.Dial("tcp", ":6600")
			if err != nil {
				log.Fatalln(err)
			}

			// Keep connection alive by pinging ever 45 seconds.
			go func() {
				for {
					time.Sleep(time.Second * 45)

					if err := bar.store["mpd"].(*mpd.Client).
						Ping(); err != nil {
						bar.store["mpd"], err = mpd.Dial("tcp", ":6600")
						if err != nil {
							log.Fatalln(err)
						}
					}
				}
			}()

			// Watch MPD for events.
			w, err := mpd.NewWatcher("tcp", ":6600", "", "player")
			if err != nil {
				log.Fatalln(err)
			}
			for {
				cur, err := bar.store["mpd"].(*mpd.Client).CurrentSong()
				if err != nil {
					log.Println(err)
					return
				}
				sts, err := bar.store["mpd"].(*mpd.Client).Status()
				if err != nil {
					log.Println(err)
					return
				}

				// Set new block text.
				var s string
				if sts["state"] == "pause" {
					s = "[paused] "
				}
				block.txt = " Ƅ  " + s + cur["Artist"] + " - " + cur["Title"]

				// Redraw block.
				bar.redraw <- block

				// Update popup if open.
				if popup.open {
					popup.update()
				}

				// Wait for next event.
				<-w.Event
			}
		},

		actions: map[xproto.Button]func() error{
			1: func() error {
				return bar.drawPopup("music")
			},
			3: func() error {
				s, err := bar.store["mpd"].(*mpd.Client).Status()
				if err != nil {
					return err
				}

				return bar.store["mpd"].(*mpd.Client).
					Pause(s["state"] != "pause")
			},
			4: func() error {
				return bar.store["mpd"].(*mpd.Client).Previous()
			},
			5: func() error {
				return bar.store["mpd"].(*mpd.Client).Next()
			},
		},
	})

	bar.blocks.Set("todo", &Block{
		txt:   "ƅ",
		w:     bar.h,
		align: 'c',
		xoff:  1,
		bg:    xgraphics.BGRA{B: 201, G: 148, R: 83, A: 0xFF},
		fg:    xgraphics.BGRA{B: 255, G: 255, R: 255, A: 0xFF},

		update: func() {},

		actions: map[xproto.Button]func() error{
			1: func() error {
				u, err := user.Current()
				if err != nil {
					return err
				}

				cmd := exec.Command("st", "micro", "-savecursor", "false", path.
					Join(u.HomeDir, ".todo"))
				cmd.Stdout = os.Stdout
				return cmd.Start()
			},
		},
	})
}
