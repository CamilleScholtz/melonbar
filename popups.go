package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/fhs/gompd/mpd"
	"github.com/fsnotify/fsnotify"
	"github.com/rkoesters/xdg/basedir"
	"github.com/rkoesters/xdg/userdirs"
	"github.com/thedevsaddam/gojsonq"
	"golang.org/x/image/math/fixed"
)

func (popup *Popup) clock() error {
	// Color the background.
	f, err := box.Open("images/clock-popup-bg.png")
	if err != nil {
		return err
	}
	defer f.Close()

	// Draw the background.
	bg, _, err := image.Decode(f.(io.Reader))
	if err != nil {
		return err
	}
	xgraphics.Blend(popup.img, xgraphics.NewConvert(X, bg), image.Point{0, 0})

	// Redraw the popup.
	popup.draw()

	// Creat http client with a timeout.
	c := &http.Client{Timeout: time.Duration(2 * time.Second)}

	// Get weather information.
	r, err := c.Get("https://api.buienradar.nl/data/public/2.0/jsonfeed")
	if err != nil {
		return err
	}
	defer r.Body.Close()
	jq := gojsonq.New().Reader(r.Body).From("actual.stationmeasurements.[0]").
		WhereEqual("stationid", "6260")

	// Draw weather information.
	popup.drawer.Src = image.NewUniform(hexToBGRA(popup.fg))
	popup.drawer.Dot = fixed.P(11, 101)
	popup.drawer.DrawString("Rainfall graph, it's " + fmt.Sprint(jq.Find(
		"feeltemperature")) + "°C.")

	// Redraw the popup.
	popup.draw()

	// Set location.
	lat := "52.0646"
	lon := "5.2065"

	// Get rainfall information.
	r, err = c.Get("https://gpsgadget.buienradar.nl/data/raintext?lat=" + lat +
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

	return nil
}

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

	// Redraw the popup.
	popup.draw()

	return nil
}

func (popup *Popup) todo() error {
	// Watch file for events.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := w.Add(path.Join(basedir.Home, ".todo")); err != nil {
		return err
	}
	f, err := os.Open(path.Join(basedir.Home, ".todo"))
	if err != nil {
		return err
	}

	// Set text color.
	popup.drawer.Src = image.NewUniform(hexToBGRA(popup.fg))

	for {
		// Count file lines.
		s := bufio.NewScanner(f)
		var c int
		for s.Scan() {
			c++
		}

		// Rewind file.
		if _, err := f.Seek(0, 0); err != nil {
			log.Println(err)
		}

		// Color the background.
		popup.img.For(func(cx, cy int) xgraphics.BGRA {
			return hexToBGRA(popup.bg)
		})

		// Draw text.
		popup.drawer.Dot = fixed.P(5, 11)
		popup.drawer.DrawString(strconv.Itoa(c))

		// Redraw block.
		popup.draw()

		// Listen for next write event.
		ev := <-w.Events
		if ev.Op&fsnotify.Write != fsnotify.Write {
			continue
		}
	}
}
