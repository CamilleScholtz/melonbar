package main

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"golang.org/x/image/font"
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

	// The foreground and background colors.
	bg, fg xgraphics.BGRA

	// This boolean decides if the block is an invisible "script block", that
	// doesn't draw anything to the bar, only executes the `update` function.
	script bool

	// The fuction that updates the block, this will be executes as a goroutine.
	update func()

	// A map with functions to execute on button events.
	actions map[xproto.Button]func() error
}

func (bar *Bar) drawBlocks() {
	for _, key := range bar.blocks.Keys() {
		block := bar.block(key.(string))

		// Run update function.
		go block.update()

		// Check if the block is a script block, if that is the case, we only
		// execute the `upload` function.
		if block.script {
			continue
		}

		// Initialize block image.
		block.img = bar.img.SubImage(image.Rect(bar.xsum, 0, bar.xsum+block.
			w, bar.h)).(*xgraphics.Image)

		// set the block location.
		block.x = bar.xsum

		// Add the width of this block to the xsum.
		bar.xsum += block.w

		// Draw block.
		bar.redraw <- block
	}

	// Listen to mouse events and execute the required function.
	xevent.ButtonPressFun(func(_ *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
		for _, k := range bar.blocks.Keys() {
			block := bar.block(k.(string))

			// Check if clicked inside the block, if not, return.
			switch k {
			// XXX: Hack for music block.
			case "music":
				tw := font.MeasureString(face, block.txt).Round()
				if ev.EventX < int16(block.x+(block.w-tw+(block.xoff*2))) || ev.
					EventX > int16(block.x+block.w) {
					continue
				}
			// XXX: Hack for clock block.
			case "clock":
				tw := bar.drawer.MeasureString(block.txt).Ceil()
				if ev.EventX < int16(((bar.w/2)-(tw/2))-13) || ev.
					EventX > int16(((bar.w/2)+(tw/2))+13) {
					continue
				}
			default:
				if ev.EventX < int16(block.x) || ev.EventX > int16(block.x+block.
					w) {
					continue
				}
			}

			// Execute the function as specified.
			if _, ok := block.actions[ev.Detail]; ok {
				go block.actions[ev.Detail]()
			}
		}
	}).Connect(X, bar.win.Id)
}

func (bar *Bar) block(key string) *Block {
	i, _ := bar.blocks.Get(key)
	return i.(*Block)
}
