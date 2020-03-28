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
	"github.com/rkoesters/xdg/userdirs"
	"github.com/thedevsaddam/gojsonq"
	"golang.org/x/image/math/fixed"
)

func (bar *Bar) initPopups() {
	bar.popups.Set("clock", &Popup{
		x: (bar.w / 2) - (178 / 2),
		y: bar.h,
		w: 178,
		h: 129,

		update: func() {
			popup := bar.popup("clock")

			// Color the background.
			f, err := box.Open("images/clock-popup-bg.png")
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()
			bg, _, err := image.Decode(f.(io.Reader))
			if err != nil {
				log.Println(err)
				return
			}
			xgraphics.Blend(popup.img, xgraphics.NewConvert(X, bg), image.Point{
				0, 0})

			// Redraw the popup.
			popup.draw()

			// Set foreground color.
			popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 33, G: 27,
				R: 2, A: 0xFF})

			// Create http client with a timeout.
			c := &http.Client{Timeout: time.Duration(2 * time.Second)}

			// Get weather information.
			r, err := c.Get(
				"https://api.buienradar.nl/data/public/2.0/jsonfeed")
			if err != nil {
				log.Println(err)
				return
			}
			defer r.Body.Close()
			jq := gojsonq.New().Reader(r.Body).From(
				"actual.stationmeasurements.[0]").WhereEqual("stationid",
				"6260")

			// Draw weather information.
			popup.drawer.Dot = fixed.P(11, 101)
			popup.drawer.DrawString("Rainfall graph, it's " + fmt.Sprint(jq.
				Find("feeltemperature")) + "°C.")

			// Redraw the popup.
			popup.draw()

			// Set location.
			lat := "52.0646"
			lon := "5.2065"

			// Get rainfall information.
			r, err = c.Get(
				"https://gpsgadget.buienradar.nl/data/raintext?lat=" + lat +
					"&lon=" + lon)
			if err != nil {
				log.Println(err)
				return
			}
			defer r.Body.Close()

			// Create rainfall tmp files.
			td, err := ioutil.TempFile(os.TempDir(), "melonbar-rain-*.dat")
			if err != nil {
				log.Println(err)
				return
			}
			defer os.Remove(td.Name())
			ti, err := ioutil.TempFile(os.TempDir(), "melonbar-rain-*.png")
			if err != nil {
				log.Println(err)
				return
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
				log.Println(err)
				return
			}
			if err := td.Close(); err != nil {
				log.Println(err)
				return
			}

			// Create rainfall graph.
			cmd := exec.Command("gnuplot", "-e", `
				set terminal png transparent size 205,107;
				set output '`+ti.Name()+`';
				set yrange [0:255];
				set noborder;
				set nolabel;
				set nokey;
				set notics;
				set notitle;
				set style fill solid border rgb '#5394C9';
				plot '`+td.Name()+`' smooth csplines with filledcurve lc rgb '#72A7D3';
			`)
			if err := cmd.Run(); err != nil {
				log.Println(err)
				return
			}

			// Draw rainfall graph.
			img, _, err := image.Decode(ti)
			if err != nil {
				log.Println(err)
				return
			}
			xgraphics.Blend(popup.img, xgraphics.NewConvert(X, img), image.
				Point{11, 3})

			// Redraw the popup.
			popup.draw()
		},
	})

	bar.popups.Set("music", &Popup{
		x: bar.w - 304 - bar.h,
		y: bar.h,
		w: 304,
		h: 148,

		update: func() {
			popup := bar.popup("music")

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

			// Color the background.
			popup.img.For(func(cx, cy int) xgraphics.BGRA {
				return xgraphics.BGRA{B: 238, G: 238, R: 238, A: 0xFF}
			})

			// Redraw the popup.
			popup.draw()

			// Set foreground color.
			popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 33, G: 27,
				R: 2, A: 0xFF})

			// Draw album text.
			album := trim(cur["Album"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(album).
				Ceil()/2)+82, 48)
			popup.drawer.DrawString(album)

			// Draw artist text.
			artist := trim("Artist: "+cur["AlbumArtist"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(artist).
				Ceil()/2)+82, 58+16)
			popup.drawer.DrawString(artist)

			// Draw rlease date text.
			date := trim("Release date: "+cur["Date"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(date).
				Ceil()/2)+82, 58+16+16)
			popup.drawer.DrawString(date)

			// Check if the cover art file exists.
			fp := path.Join(userdirs.Music, path.Dir(cur["file"]),
				"cover_popup.png")
			if _, err := os.Stat(fp); !os.IsNotExist(err) {
				f, err := os.Open(fp)
				if err != nil {
					log.Println(err)
					return
				}
				defer f.Close()

				// Draw cover art.
				img, _, err := image.Decode(f)
				if err != nil {
					log.Println(err)
					return
				}
				xgraphics.Blend(popup.img, xgraphics.NewConvert(X, img), image.
					Point{-166, -10})
			} else {
				popup.drawer.Dot = fixed.P(200, 78)
				popup.drawer.DrawString("No cover found!")
			}

			// Calculate progressbar lengths.
			e, err := strconv.ParseFloat(sts["elapsed"], 32)
			if err != nil {
				log.Println(err)
				return
			}
			t, err := strconv.ParseFloat(sts["duration"], 32)
			if err != nil {
				log.Println(err)
				return
			}
			pf := int(math.Round(e / t * 29))
			pu := 29 - pf

			// Draw progressbar.
			popup.drawer.Dot = fixed.P(10, 132)
			for i := 1; i <= pf; i++ {
				popup.drawer.DrawString("-")
			}
			popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 211, G: 167,
				R: 114, A: 0xFF})
			for i := 1; i <= pu; i++ {
				popup.drawer.DrawString("-")
			}

			// Redraw the popup.
			popup.draw()
		},
	})
}
