package main

import (
	"image"
	"io"
	"math"
	"os"
	"path"
	"strconv"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/fhs/gompd/mpd"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func (popup *Popup) music(c *mpd.Client) error {
	d := &font.Drawer{
		Dst:  popup.img,
		Src:  image.NewUniform(hexToBGRA(popup.fg)),
		Face: face,
	}

	// Color the background.
	popup.img.For(func(cx, cy int) xgraphics.BGRA {
		return hexToBGRA(popup.bg)
	})

	cur, err := c.CurrentSong()
	if err != nil {
		return err
	}
	sts, err := c.Status()
	if err != nil {
		return err
	}

	// Draw album info text.
	d.Dot = fixed.P(10, 20)
	d.DrawString("Album: " + cur["Album"])
	d.Dot = fixed.P(10, 20+16)
	d.DrawString("Artist: " + cur["AlbumArtist"])
	d.Dot = fixed.P(10, 20+16+16)
	d.DrawString("Date: " + cur["Date"])

	// Find album art.
	var f interface{}
	f, err = os.Open(path.Join("/home/onodera/media/music/", path.Dir(
		cur["file"]), "cover_popup.png"))
	if err != nil {
		f, err = box.Open("images/cover.png")
		if err != nil {
			return err
		}
	}

	// Draw album art.
	img, _, err := image.Decode(f.(io.Reader))
	if err != nil {
		return err
	}
	xgraphics.Blend(popup.img, img, image.Point{-166, -10})

	// Calculate progressbar lengths.
	e, err := strconv.ParseFloat(sts["elapsed"], 32)
	if err != nil {
		return err
	}
	t, err := strconv.ParseFloat(sts["duration"], 32)
	if err != nil {
		return err
	}
	pf := int(math.Round(e / t * 29))
	pu := 29 - pf

	// Draw progressbar.
	d.Dot = fixed.P(10, 132)
	d.Src = image.NewUniform(hexToBGRA("#5394C9"))
	for i := 1; i <= pf; i++ {
		d.DrawString("-")
	}
	d.Src = image.NewUniform(hexToBGRA(popup.fg))
	for i := 1; i <= pu; i++ {
		d.DrawString("-")
	}

	// Draw the popup.
	popup.img.XDraw()
	popup.img.XPaint(popup.win.Id)

	return nil
}
