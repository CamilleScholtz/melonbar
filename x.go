package main

import (
	"io/ioutil"
	"log"

	"github.com/AndreKR/multiface"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/zachomedia/go-bdf"
)

func initX() error {
	// Disable logging messages.
	xgb.Logger = log.New(ioutil.Discard, "", 0)

	// Set up a connection to the X server.
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		return err
	}

	// Run the main X event loop, this is used to catch events.
	go xevent.Main(X)

	// Listen to the root window for property change events, used to check if
	// the user changed the focused window or active workspace for example.
	return xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskPropertyChange)
}

func initEWMH(w xproto.Window) error {
	// TODO: `WmStateSet` and `WmDesktopSet` are basically here to keep OpenBox
	// happy, can I somehow remove them and just use `_NET_WM_WINDOW_TYPE_DOCK`
	// like I can with WindowChef?
	if err := ewmh.WmWindowTypeSet(X, w, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return err
	}
	if err := ewmh.WmStateSet(X, w, []string{
		"_NET_WM_STATE_STICKY"}); err != nil {
		return err
	}
	if err := ewmh.WmDesktopSet(X, w, ^uint(0)); err != nil {
		return err
	}
	return ewmh.WmNameSet(X, w, "melonbar")
}

func initFace() error {
	face = new(multiface.Face)

	fpl := []string{
		"/fonts/cure.punpun.bdf",
		"/fonts/kochi.small.bdf",
		"/fonts/baekmuk.small.bdf",
	}

	for _, fp := range fpl {
		f, err := runtime.Open(fp)
		if err != nil {
			return err
		}
		fb, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		ff, err := bdf.Parse(fb)
		if err != nil {
			return err
		}

		face.AddFace(ff.NewFace())

		f.Close()
	}

	return nil
}
