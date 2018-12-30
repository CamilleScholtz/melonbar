package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/antchfx/xmlquery"
	"github.com/fhs/gompd/mpd"
	"github.com/rkoesters/xdg/userdirs"
	"golang.org/x/image/math/fixed"
)

// TODO: Make progressbar clickable.
// TODO: Make progressbar update every X milliseconds.
func (popup *Popup) music(c *mpd.Client) error {
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

	// Set text color.
	popup.drawer.Src = image.NewUniform(hexToBGRA(popup.fg))

	// Draw album text.
	album := trim(cur["Album"], 32)
	popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(album).Ceil()/2)+82,
		48)
	popup.drawer.DrawString(album)

	// Draw artist text.
	artist := trim("Artist: "+cur["AlbumArtist"], 32)
	popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(artist).Ceil()/2)+
		82, 58+16)
	popup.drawer.DrawString(artist)

	// Draw rlease date text.
	date := trim("Release date: "+cur["Date"], 32)
	popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(date).Ceil()/2)+82,
		58+16+16)
	popup.drawer.DrawString(date)

	// Check if album art file exists.
	fp := path.Join(userdirs.Music, path.Dir(cur["file"]), "cover_popup.png")
	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		f, err := os.Open(fp)
		if err != nil {
			return err
		}
		defer f.Close()

		// Draw album art.
		img, _, err := image.Decode(f)
		if err != nil {
			return err
		}
		xgraphics.Blend(popup.img, xgraphics.NewConvert(X, img), image.Point{
			-166, -10})
	} else {
		popup.drawer.Dot = fixed.P(200, 78)
		popup.drawer.DrawString("No cover found!")
	}

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
	popup.drawer.Dot = fixed.P(10, 132)
	for i := 1; i <= pf; i++ {
		popup.drawer.DrawString("-")
	}
	popup.drawer.Src = image.NewUniform(hexToBGRA("#72A7D3"))
	for i := 1; i <= pu; i++ {
		popup.drawer.DrawString("-")
	}

	// Redraw the bar.
	popup.draw()

	return nil
}

func (popup *Popup) clock() error {
	// Color the background.
	f, err := box.Open("images/clock-popup-bg.png")
	if err != nil {
		return err
	}
	defer f.Close()

	// Draw album art.
	bg, _, err := image.Decode(f.(io.Reader))
	if err != nil {
		return err
	}
	xgraphics.Blend(popup.img, xgraphics.NewConvert(X, bg), image.Point{0, 0})

	// Redraw the popup.
	popup.draw()

	// Set location.
	lat := "52.0646"
	lon := "5.2065"

	// Get rainfall information.
	r, err := http.Get("https://gps.buienradar.nl/getrr.php?lat=" + lat +
		"&lon=" + lon)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Create rainfall tmp files.
	td, err := ioutil.TempFile(os.TempDir(), "melonbar-rain-*.dat")
	if err != nil {
		return err
	}
	defer os.Remove(td.Name())
	ti, err := ioutil.TempFile(os.TempDir(), "melonbar-rain-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(ti.Name())

	// Compose rainfall data tmp file contents.
	var d []byte
	s := bufio.NewScanner(r.Body)
	for s.Scan() {
		d = append(d, bytes.Split(s.Bytes(), []byte("|"))[0]...)
		d = append(d, []byte("\n")...)
	}

	// Write rainfall data tmp file.
	if _, err = td.Write(d); err != nil {
		return err
	}
	if err := td.Close(); err != nil {
		return err
	}

	// Create rainfall graph.
	cmd := exec.Command("gnuplot", "-e", `
		set terminal png transparent size 211,107;
		set output '`+ti.Name()+`';
		set yrange [0:255];
		set noborder;
		set nolabel;
		set nokey;
		set notics;
		set notitle;
		set style fill solid border rgb '#5394C9';
		plot '`+td.Name()+`' smooth csplines with filledcurve x1 lc rgb '#72A7D3'
	`)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Draw rainfall graph.
	img, _, err := image.Decode(ti)
	if err != nil {
		return err
	}
	xgraphics.Blend(popup.img, xgraphics.NewConvert(X, img), image.Point{14, 0})

	// Redraw the popup.
	popup.draw()

	// Get weather information.
	x, err := xmlquery.LoadURL("https://xml.buienradar.nl")
	if err != nil {
		return err
	}
	w := xmlquery.FindOne(x, "//weerstation[@id=6260]")
	fmt.Println()
	popup.drawer.Src = image.NewUniform(hexToBGRA(popup.fg))
	popup.drawer.Dot = fixed.P(10, 100)
	popup.drawer.DrawString("Rainfall graph, it's " + w.SelectElement(
		"temperatuurGC").InnerText() + "ºC")

	// Redraw the popup.
	popup.draw()

	return nil
}
