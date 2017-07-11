package main

import (
	"image"
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
	// Connection to the X server, the abe window, and the bar image.
	xu  *xgbutil.XUtil
	win *xwindow.Window
	img *xgraphics.Image

	// The font and fontsize that should be used.
	font     *truetype.Font
	fontSize float64

	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block
	// to the right of the last block.
	xsum int

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

	/// The aligment of the text, this can be `'l'` for left aligment,
	// `'c'` for center aligment and `'r'` for right aligment.
	align rune

	// Additional x offset to further tweak the location of the text.
	xoff int

	// The foreground and background colors.
	bg, fg xgraphics.BGRA

	// The sub-image that represents the block.
	img *xgraphics.Image
}

func initBar(x, y, w, h int, font string, fontSize float64) (*Bar,
	error) {
	bar := new(Bar)
	var err error

	// Connect to X.
	bar.xu, err = xgbutil.NewConn()
	if err != nil {
		return nil, err
	}

	// Listen to the root window for events.
	xwindow.New(bar.xu, bar.xu.RootWin()).Listen(
		xproto.EventMaskPropertyChange)

	// Create a window.
	bar.win, err = xwindow.Generate(bar.xu)
	if err != nil {
		return nil, err
	}
	bar.win.Create(bar.xu.RootWin(), x, y, w, h, xproto.CwBackPixel,
		0x000000)

	// EWMH stuff.
	if err := ewmh.WmWindowTypeSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return nil, err
	}

	// Map window.
	bar.win.Map()

	// Create bar image.
	bar.img = xgraphics.New(bar.xu, image.Rect(0, 0, w, h))
	bar.img.XSurfaceSet(bar.win.Id)
	bar.img.XDraw()

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

	bar.w = w
	bar.h = h

	bar.block = new(sync.Map)
	bar.redraw = make(chan string)

	return bar, nil
}

func (bar *Bar) initBlock(name, txt string, width int, align rune,
	xoff int, bg, fg string) {
	bar.block.Store(name, &Block{txt, bar.xsum, width, align, xoff,
		hexToBGRA(bg), hexToBGRA(fg), bar.img.SubImage(image.Rect(
			bar.xsum, 0, bar.xsum+width, bar.h)).(*xgraphics.Image)})
	bar.xsum += width
}

func (bar *Bar) updateBlockBg(name, bg string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	nbg := hexToBGRA(bg)
	if block.bg == nbg {
		return
	}

	block.bg = nbg
	bar.redraw <- name
	return
}

func (bar *Bar) updateBlockFg(name, fg string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	nfg := hexToBGRA(fg)
	if block.fg == nfg {
		return
	}

	block.fg = nfg
	bar.redraw <- name
	return
}

func (bar *Bar) updateBlockTxt(name, txt string) {
	i, _ := bar.block.Load(name)
	block := i.(*Block)

	if block.txt == txt {
		return
	}

	block.txt = txt
	bar.redraw <- name
	return
}

func (bar *Bar) draw(name string) error {
	// Needed to prevent an `interface conversion: interface {} is
	// nil, not *main.Block` panic for some reason...
	time.Sleep(time.Nanosecond)

	i, _ := bar.block.Load(name)
	block := i.(*Block)

	// Color the backround.
	tw, _ := xgraphics.Extents(bar.font, bar.fontSize, block.txt)
	block.img.For(func(x, y int) xgraphics.BGRA {
		return block.bg
	})

	// Calculate the required x coordinate for the different
	// aligments.
	var x int
	switch block.align {
	case 'l':
		x = block.x + block.xoff
	case 'c':
		x = block.x + ((block.w / 2) - (tw / 2)) + block.xoff
	case 'r':
		x = (block.x + block.w) - tw + block.xoff
	}

	// TODO: Center vertically automatically.
	// Draw the text.
	if _, _, err := block.img.Text(x, 6, block.fg, bar.fontSize,
		bar.font, block.txt); err != nil {
		return err
	}

	block.img.XDraw()
	return nil
}

func (bar *Bar) paint() {
	bar.img.XPaint(bar.win.Id)
}
