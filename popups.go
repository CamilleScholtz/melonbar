package main

import (
	"image"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/IvanMenshykov/MoonPhase"
	"github.com/RadhiFadlillah/go-prayer"
	"github.com/elliotchance/orderedmap"
	"github.com/fhs/gompd/mpd"
	"github.com/rkoesters/xdg/userdirs"
	"golang.org/x/image/math/fixed"
)

func (bar *Bar) initPopups() {
	bar.popups.Set("clock", &Popup{
		x: (bar.w / 2) - (184 / 2),
		y: bar.h,
		w: 184,
		h: 118,

		update: func() {
			popup := bar.popup("clock")

			// Color the background.
			popup.img.For(func(cx, cy int) xgraphics.BGRA {
				return xgraphics.BGRA{B: 238, G: 238, R: 238, A: 0xFF}
			})

			// Set foreground color.
			popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 33, G: 27,
				R: 2, A: 0xFF})

			// Get the current time. Present Day, heh... Present Time! Hahahaha!
			n := time.Now()

			// Get moon phase.
			mc := MoonPhase.New(n)
			mn := map[int]string{
				0: "Ɔ",
				1: "Ƈ",
				2: "ƈ",
				3: "Ɖ",
				4: "Ɗ",
				5: "Ƌ",
				6: "ƌ",
				7: "ƍ",
				8: "Ƈ",
			}

			// Draw moon text.
			popup.drawer.Dot = fixed.P(19, 48)
			popup.drawer.DrawString("The moon currently looks like: " + mn[int(
				math.Floor((mc.Phase()+0.0625)*8))])

			// Get the prayers.
			pm := (&prayer.Calculator{
				Latitude:          52.1277,
				Longitude:         5.6686,
				Elevation:         21,
				CalculationMethod: prayer.MWL,
				AsrConvention:     prayer.Hanafi,
				PreciseToSeconds:  false,
			}).Init().SetDate(n).Calculate()

			// The prayers we want to track.
			pom := orderedmap.NewOrderedMap()
			pom.Set("Fajr", pm[prayer.Fajr])
			pom.Set("Zuhr", pm[prayer.Zuhr])
			pom.Set("Asr", pm[prayer.Asr])
			pom.Set("Maghrib", pm[prayer.Maghrib])
			pom.Set("Isha", pm[prayer.Isha])

			// Calculate the dot lenght, this is the length of the line (164px)
			// divided by 2400 (minutes in a day).
			d := 164.00 / 2400.00

			// Calculate elapsed line length.
			e := int(math.Round(d*float64(n.Hour()*100+n.Minute()))) + 10

			// Draw line.
			popup.img.SubImage(image.Rect(10, 101, 10+164, 102)).(*xgraphics.
				Image).For(func(x, y int) xgraphics.BGRA {
				// Make the line look dashed.
				if x%5 == 4 {
					return xgraphics.BGRA{B: 238, G: 238, R: 238, A: 0xFF}
				}

				if x < e {
					return xgraphics.BGRA{B: 33, G: 27, R: 2, A: 0xFF}
				}
				return xgraphics.BGRA{B: 211, G: 167, R: 114, A: 0xFF}
			})

			// Loop over these prayers and draw stuff for each one.
			tm := false
			np := false
			for p := pom.Front(); p != nil; p = p.Next() {
				k := p.Key.(string)
				v := p.Value.(time.Time)

				// Calculate arrow position.
				pd := int(math.Round(d*float64(v.Hour()*100+v.Minute()))) + 9

				// Set arrow color.
				popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 211,
					G: 167, R: 114, A: 0xFF})

				if tm || (!np && v.Unix() > n.Add(-time.Hour).Unix()) {
					np = true

					// Set arrow color for next prayer.
					popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 33,
						G: 27, R: 2, A: 0xFF})

					// Compose arrow text.
					s := k + ", " + v.Format("03:04 PM")

					// Calculate X offset, we use some smart logic in order to
					// always have nice padding, even with longer strings.
					sl := popup.drawer.MeasureString(s).Round()
					x := pd + 2 - (sl / 2)
					if x < 10 {
						x = 10
					} else if x > 184-8-sl {
						x = 184 - 8 - sl
					}

					// Draw arrow text.
					popup.drawer.Dot = fixed.P(x, 85)
					popup.drawer.DrawString(s)
				}

				// Workaround if the next prayer is tomorrow.
				if p.Next() == nil && !np {
					tm = true

					pom.Delete(prayer.Fajr)
					pom.Set(prayer.Fajr, pm[prayer.Fajr])
				}

				// Draw arrow.
				popup.drawer.Dot = fixed.P(pd+2, 98)
				popup.drawer.DrawString("↓")
			}

			// Redraw the popup.
			popup.draw()
		},
	})

	bar.popups.Set("music", &Popup{
		x: bar.w - 327 - bar.h,
		y: bar.h,
		w: 327,
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

			// Set foreground color.
			popup.drawer.Src = image.NewUniform(xgraphics.BGRA{B: 33, G: 27,
				R: 2, A: 0xFF})

			// Draw album text.
			album := trim(cur["Album"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(album).
				Ceil()/2)+90, 48)
			popup.drawer.DrawString(album)

			// Draw artist text.
			artist := trim("Artist: "+cur["AlbumArtist"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(artist).
				Ceil()/2)+90, 58+16)
			popup.drawer.DrawString(artist)

			// Draw rlease date text.
			date := trim("Release date: "+cur["Date"], 32)
			popup.drawer.Dot = fixed.P(-(popup.drawer.MeasureString(date).
				Ceil()/2)+90, 58+16+16)
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
					Point{-179, -0})
			} else {
				popup.drawer.Dot = fixed.P(218, 78)
				popup.drawer.DrawString("No cover found!")
			}

			// Get elapsed and duration times.
			se, err := strconv.ParseFloat(sts["elapsed"], 32)
			if err != nil {
				log.Println(err)
				return
			}
			//e := int(math.Round(se))

			// Calculate the dot lenght, this is the length of the line divided
			// by the length of the song.
			sd, err := strconv.ParseFloat(sts["duration"], 32)
			if err != nil {
				log.Println(err)
				return
			}
			d := 159.00 / sd

			// Calculate elapsed line length.
			e := int(math.Round(d*se)) + 10

			// Draw line.
			popup.img.SubImage(image.Rect(10, 131, 10+159, 132)).(*xgraphics.
				Image).For(func(x, y int) xgraphics.BGRA {
				// Make the line look dashed.
				if x%5 == 4 {
					return xgraphics.BGRA{B: 238, G: 238, R: 238, A: 0xFF}
				}

				if x < e {
					return xgraphics.BGRA{B: 33, G: 27, R: 2, A: 0xFF}
				}
				return xgraphics.BGRA{B: 211, G: 167, R: 114, A: 0xFF}
			})

			// Redraw the popup.
			popup.draw()
		},
	})

	/*bar.popups.Set("clock", &Popup{
		x: (bar.w / 2) - (178 / 2),
		y: bar.h,
		w: 178,
		h: 129,

		update: func() {
			popup := bar.popup("clock")

			// Color the background.
			f, err := runtime.Open("/images/clock-popup-bg.png")
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
				plot '`+td.Name()+`' smooth csplines w filledcu x1 fc rgb '#72A7D3', '`+td.Name()+`' smooth csplines w line lc rgb '#5394C9';
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
				Point{11, 2})

			// Redraw the popup.
			popup.draw()
		},
	})*/
}
