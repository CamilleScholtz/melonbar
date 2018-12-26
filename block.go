package main

import (
	"image"

	"github.com/BurntSushi/xgbutil/xgraphics"
)

// Block is a struct with information about a block.
type Block struct {
	// The text the block should display.
	txt string

	// The x coordinate and width of the block.
	x, w int

	// The aligment of the text, this can be `l` for left aligment, `c` for
	// center aligment, `r` for right aligment and `a` for absolute center
	// aligment.
	align rune

	// Additional x offset to further tweak the location of the text.
	xoff int

	// The foreground and background colors in hex.
	bg, fg string

	// A map with functions to execute on button events. Accepted button strings
	// are `button0` to `button5`
	actions map[string]func()

	// The popup.
	popup *Popup

	// The sub-image that represents the block.
	img *xgraphics.Image
}

func (bar *Bar) initBlock(name, txt string, w int, align rune, xoff int, bg,
	fg string) *Block {
	block := new(Block)

	block.txt = txt
	block.x = bar.xsum
	block.w = w
	block.align = align
	block.xoff = xoff
	block.bg = bg
	block.fg = fg
	block.actions = map[string]func(){
		"button1": func() {},
		"button2": func() {},
		"button3": func() {},
		"button4": func() {},
		"button5": func() {},
	}
	block.img = bar.img.SubImage(image.Rect(bar.xsum, 0, bar.xsum+w, bar.
		h)).(*xgraphics.Image)

	// Add the width of this block to the xsum.
	bar.xsum += w

	// Store the block in map.
	bar.block.Store(name, block)

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
