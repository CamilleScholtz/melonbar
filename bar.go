package main

import (
	"image"
	"log"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/freetype-go/freetype/truetype"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
)

// Bar is a struct with information about the bar.
type Bar struct {
	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block
	// to the right of the last block.
	xsum int

	// Connection to the X server, the abe window, and the bar image.
	xu  *xgbutil.XUtil
	win *xwindow.Window
	img *xgraphics.Image

	// The font and fontsize that should be used.
	font     *truetype.Font
	fontSize float64

	// A map with information about the block, see `Block`.
	block *sync.Map

	// The channel where the block name should be written to once the
	// block is ready to be redrawn.
	redraw chan string
}

// Block is a struct with information about a block.
type Block struct {
	// The text the block should display.
	txt string

	// The x coordinate and width of the bar.
	x, w int

	// How to align the text, 'l' for left, 'c' for center and 'r' for
	// right.
	align rune

	// The foreground and background colors.
	bg, fg xgraphics.BGRA

	// The sub-image that represents the block.
	img *xgraphics.Image
}

func initBar(x, y, w, h int, font string, fontSize float64) (*Bar,
	error) {
	bar := new(Bar)
	var err error

	bar.w = w
	bar.h = h

	// Connect to X.
	bar.xu, err = xgbutil.NewConn()
	if err != nil {
		return nil, err
	}
	// TODO: Is this needed when I have a bar?
	xwindow.New(bar.xu, bar.xu.RootWin()).Listen(
		xproto.EventMaskPropertyChange)

	// Crate and map window.
	bar.win, err = xwindow.Generate(bar.xu)
	if err != nil {
		return nil, err
	}
	bar.win.Create(bar.xu.RootWin(), x, y, w, h, xproto.CwBackPixel|
		xproto.CwEventMask, 0xEEEEEE, xproto.EventMaskPropertyChange)
	if err := ewmh.WmWindowTypeSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return nil, err
	}
	bar.win.Map()

	// Create bar image.
	bar.img = xgraphics.New(bar.xu, image.Rect(0, 0, w, h))
	bar.img.XSurfaceSet(bar.win.Id)

	// TODO: I don't *really* want to use `ttf` fonts but there
	// doesn't seem to be any `pcf` Go library at the moment.
	// Load font.
	f, err := os.Open(font)
	if err != nil {
		return nil, err
	}
	bar.font, err = xgraphics.ParseFont(f)
	if err != nil {
		return nil, err
	}
	bar.fontSize = fontSize

	bar.block = new(sync.Map)
	bar.redraw = make(chan string)

	return bar, nil
}

func (bar *Bar) initBlock(name string, width int, align rune, txt, bg,
	fg string) {
	bar.block.Store(name, &Block{txt, bar.xsum, width, align,
		hexToBGRA(bg), hexToBGRA(fg), bar.img.SubImage(image.Rect(
			bar.xsum, 0, bar.xsum+width, bar.h)).(*xgraphics.Image)})
	bar.xsum += width
}

func (bar *Bar) updateBlockBg(name, bg string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	block.bg = hexToBGRA(bg)
}

func (bar *Bar) updateBlockFg(name, fg string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	block.fg = hexToBGRA(fg)
}

func (bar *Bar) updateBlockTxt(name, txt string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	block.txt = txt
}

func (bar *Bar) draw(name string) {
	// Needed to prevent a `interface conversion: interface {} is nil,
	// not *main.Block` panic for some reason...
	time.Sleep(time.Nanosecond)

	i, _ := bar.block.Load(name)
	block := i.(*Block)

	// Color the backround.
	block.img.For(func(x, y int) xgraphics.BGRA {
		return block.bg
	})

	// Calculate the required x for the different aligments.
	var x int
	switch block.align {
	case 'l':
		x = block.x
	case 'c':
		tw, _ := xgraphics.Extents(bar.font, bar.fontSize, block.txt)
		x = block.x + ((block.w / 2) - (tw / 2))
	case 'r':
		tw, _ := xgraphics.Extents(bar.font, bar.fontSize, block.txt)
		x = (block.x + block.w) - tw
	}

	// TODO: Center vertically automatically.
	// Draw the text.
	if _, _, err := block.img.Text(x, 6, block.fg, bar.fontSize,
		bar.font, block.txt); err != nil {
		log.Fatal(err)
	}

	block.img.XDraw()
}

func (bar *Bar) paint() {
	bar.img.XPaint(bar.win.Id)
}
