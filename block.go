package main

import (
	"image"

	"github.com/BurntSushi/xgbutil/xgraphics"
)

// Block is a struct with information about a block.
type Block struct {
	// The sub-image that represents the block.
	img *xgraphics.Image

	// The x coordinate and width of the block.
	x, w int

	// Additional x offset to further tweak the location of the text.
	xoff int

	// The text the block should display.
	txt string

	// The aligment of the text, this can be `l` for left aligment, `c` for
	// center aligment, `r` for right aligment and `a` for absolute center
	// aligment.
	align rune

	// The foreground and background colors in hex.
	bg, fg string

	// A map with functions to execute on button events. Accepted button strings
	// are `button0` to `button5`
	actions map[string]func() error

	// Block popup..
	popup *Popup
}

func (bar *Bar) initBlock(name, txt string, w int, align rune, xoff int, bg,
	fg string) *Block {
	block := new(Block)

	block.img = bar.img.SubImage(image.Rect(bar.xsum, 0, bar.xsum+w, bar.
		h)).(*xgraphics.Image)
	block.x = bar.xsum
	block.w = w
	block.xoff = xoff
	block.txt = txt
	block.align = align
	block.bg = bg
	block.fg = fg
	block.actions = map[string]func() error{
		"button1": func() error { return nil },
		"button2": func() error { return nil },
		"button3": func() error { return nil },
		"button4": func() error { return nil },
		"button5": func() error { return nil },
	}

	// Add the width of this block to the xsum.
	bar.xsum += w

	// Store the block in map.
	bar.blocks.Store(name, block)

	// Draw block.
	bar.redraw <- block

	return block
}

// TODO: Make this function more versatile by allowing different and multiple
// properties to be checked.
func (block *Block) diff(txt string) bool {
	if block.txt == txt {
		return false
	}
	block.txt = txt
	return true
}
