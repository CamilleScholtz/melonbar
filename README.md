[![Go Report Card](https://goreportcard.com/badge/github.com/onodera-punpun/melonbar)](https://goreportcard.com/report/github.com/onodera-punpun/melonbar)

melonbar - A concurrent, hackable bar/panel for X written in Go.

![](https:/camille.sh/LGxO.png)


## INSTALLATION

`go get github.com/onodera-punpun/melonbar`

Or for a binary that includes embedded static files:

`packr2 get github.com/onodera-punpun/melonbar`

`melonbar` depends on Go 1.9 or newer, gnuplot, and
[packr2](https://github.com/gobuffalo/packr/tree/master/v2).


## USAGE

The idea is that this bar is very "simple" to configure by just modifying the
source code, à la suckless.

Users can configure the position, width, height and font of the bar in
`main.go`. A bar consist of various blocks that display info, these blocks are 
definded in `blocks.go`.


## AUTHORS

Camille Scholtz
