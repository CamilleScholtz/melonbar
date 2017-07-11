package main

import (
	"encoding/hex"
	"strings"

	"github.com/BurntSushi/xgbutil/xgraphics"
)

func hexToBGRA(h string) xgraphics.BGRA {
	h = strings.Replace(h, "#", "", 1)
	d, _ := hex.DecodeString(h)

	return xgraphics.BGRA{B: d[2], G: d[1], R: d[0], A: 0xFF}
}
