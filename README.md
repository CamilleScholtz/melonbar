[![Go Report Card](https://goreportcard.com/badge/github.com/onodera-punpun/melonbar)](https://goreportcard.com/report/github.com/onodera-punpun/melonbar)

melonbar - A concurrent, hackable bar/panel for X written in Go.

![](https://punpun.xyz/54c7.png)


## INSTALLATION

`go get github.com/onodera-punpun/melonbar`

`melonbar` depends on Go 1.9 or newer.


## USAGE

The idea is that this bar is very "simple" to configure by just modifying the
source code, à la suckless.

Users can configure the position, width, height and font of the bar in
`main.go`. A bar consist of various blocks that display info, these blocks are
functions definded in `blocks.go` and exectured in a goroutine in `main.go`.


## CREATING BLOCK FUNCTIONS

On top of each block function you should run `bar.initBlock()`, this generates a
block object, and is where most of the configuration happens. Here is a short
explanation of each parameter from left to right:

* The name of the block, this is gets used as the name of the block  map key.
  (`string`)
* The initial string the block should display. (`string`)
* The width of the block. (`int`)
* The aligment of the text, this can be `'l'` for left aligment, `'c'` for
  center aligment `'r'` for right aligment and `'a'` for absolute center
  aligment. (`rune`)
* Additional x offset to further tweak the location of the text. (`int`)
* The foreground color of the block in hexadecimal. (`string`)
* The background color of the block in hexadecimal. (`string`)

It is possible to bind mousebindings to a block using using:

```go
block.actions["buttonN"] = func() {
	// Do stuff.
}
```

---

When you've gathered all information you can update the block values using for
example `block.bg = "#FF0000"` or `block.txt = "Hello World!"` and executing
`bar.redraw <- block`.


## TODO

* Add popups.


## AUTHORS

Camille Scholtz
