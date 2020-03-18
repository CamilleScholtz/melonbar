package main

import (
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Bar is a struct with information about the bar.
type Bar struct {
	// Bar window, and bar image.
	win *xwindow.Window
	img *xgraphics.Image

	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block to the
	// right of the last block.
	xsum int

	// Text drawer.
	drawer *font.Drawer

	// A map with information about the block, see the `Block` type.
	blocks *sync.Map

	// A channel where the block should be send to to once its ready to be
	// redrawn.
	redraw chan *Block

	// A channel where a boolean should be send once a block has initizalized,
	// notifying that the next block can intialize.
	ready chan bool
}

func initBar(x, y, w, h int) (*Bar, error) {
	bar := new(Bar)
	var err error

	// Create a window for the bar. This window listens to button press events
	// in order to respond to them.
	bar.win, err = xwindow.Generate(X)
	if err != nil {
		return nil, err
	}
	bar.win.Create(X.RootWin(), x, y, w, h, xproto.CwBackPixel|xproto.
		CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff to make the window behave like an actual bar.
	if err := initEWMH(bar.win.Id); err != nil {
		return nil, err
	}

	// Map window.
	bar.win.Map()

	// XXX: Moving the window is again a hack to keep OpenBox happy.
	bar.win.Move(x, y)

	// Create the bar image.
	bar.img = xgraphics.New(X, image.Rect(0, 0, w, h))
	if err := bar.img.XSurfaceSet(bar.win.Id); err != nil {
		return nil, err
	}
	bar.img.XDraw()

	bar.w = w
	bar.h = h

	bar.drawer = &font.Drawer{
		Dst:  bar.img,
		Face: face,
	}

	bar.blocks = new(sync.Map)
	bar.redraw = make(chan *Block)

	// Listen to mouse events and execute the required function.
	xevent.ButtonPressFun(func(_ *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
		// Determine what block the cursor is in.
		var block *Block
		bar.blocks.Range(func(name, i interface{}) bool {
			block = i.(*Block)

			// XXX: Hack for music block.
			if name == "music" {
				tw := font.MeasureString(face, block.txt).Round()
				if ev.EventX >= int16(block.x+(block.w-tw+(block.xoff*2))) &&
					ev.EventX < int16(block.x+block.w) {
					return false
				}
				block = nil
				return true
			}

			// XXX: Hack for clock block.
			if name == "clock" {
				tw := bar.drawer.MeasureString(block.txt).Ceil()
				if ev.EventX >= int16(((bar.w/2)-(tw/2))-13) && ev.
					EventX < int16(((bar.w/2)+(tw/2))+13) {
					return false
				}
				block = nil
				return true
			}

			if ev.EventX >= int16(block.x) && ev.EventX < int16(block.x+block.
				w) {
				return false
			}
			block = nil
			return true
		})

		// Execute the function as specified.
		if _, ok := block.actions[ev.Detail]; ok {
			if err := block.actions[ev.Detail](); err != nil {
				log.Println(err)
			}
		}
	}).Connect(X, bar.win.Id)

	// Initialize keybind package, needed to grab key press events.
	keybind.Initialize(X)

	return bar, nil
}

func (bar *Bar) draw(block *Block) error {
	// Calculate the required x coordinate for the different aligments.
	var x int
	tw := bar.drawer.MeasureString(block.txt).Round()
	switch block.align {
	case 'l':
		x = block.x
	case 'c':
		x = block.x + ((block.w / 2) - (tw / 2))
	case 'r':
		x = (block.x + block.w) - tw
	case 'a':
		x = (bar.w / 2) - (tw / 2)
	default:
		return fmt.Errorf("draw %#U: Not a valid aligment rune", block.align)
	}
	x += block.xoff
	x += 2

	// Color the background.
	block.img.For(func(cx, cy int) xgraphics.BGRA {
		// XXX: Hack for music block.
		if block.w == 660 {
			if cx < x+block.xoff {
				return hexToBGRA("#445967")
			}
			return hexToBGRA(block.bg)
		}

		return hexToBGRA(block.bg)
	})

	// Set text color.
	bar.drawer.Src = image.NewUniform(hexToBGRA(block.fg))

	// Draw the text.
	bar.drawer.Dot = fixed.P(x, 16)
	bar.drawer.DrawString(block.txt)

	// Redraw the bar.
	block.img.XDraw()
	bar.img.XPaint(bar.win.Id)

	return nil
}

func (bar *Bar) initBlocks(blocks []func()) {
	bar.ready = make(chan bool)

	for _, f := range blocks {
		go f()
		<-bar.ready
	}

	close(bar.ready)
}

func (bar *Bar) listen() {
	for {
		if err := bar.draw(<-bar.redraw); err != nil {
			log.Fatalln(err)
		}
	}
}
