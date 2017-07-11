[![Go Report Card](https://goreportcard.com/badge/github.com/onodera-punpun/melonbar)](https://goreportcard.com/report/github.com/onodera-punpun/melonbar)

melonbar - A concurrent, hackable bar/panel for X written in Go.


## INSTALLATION

`go get github.com/onodera-punpun/melonbar`

`melonbar` depends on Go 1.9 or newer.


## USAGE

So the idea is that this bar is very "easy" to configure by just
modifying the source code, à la suckless.

Users can modify, configure and create new "blocks" in `blocks.go`.
and configure the position, width and font of the bar in `main.go`.


## CREATING A BLOCK FUNCTIONS

On top of each block function you should run `bar.initBlock()`, this
is where most of the configuration happens. Here is a short
explanation of each parameter from left to right:

* The name of the block, this is and gets used as the name
  of the block map key. (`string`)
* The initial string the block should display. (`string`)
* The width of the block. (`int`)
* The aligment of the text, this can be `'l'` for left aligment, `'c'`
  for center aligment and `'r'` for right aligment. (`rune`)
* Additional x offset to further tweak the location of the text.
  (`int`)
* The foreground color of the block in hexadecimal. (`string`)
* The background color of the block in hexadecimal. (`string`)


---

Everything that should not be ran in a loop should of course be
specified before the `for` loop. For example setting up a connection
to mpd.

If you want something to only be done *after* the very first loop - an
example of this would be not waiting for a workspace chance event, but
immediately checking the current workspace. - use:

```go
init := true
for {
	if !init {
		// Things you only want to do after the first loop.
	}
	init = false
	...
```

This can be helpful because else the bar would display `"?"` before
the user changes his workspace for the first time.


---

When you've gathered all needed information you can:

* Update the foreground color of the block using
  `bar.updateBlockFg()`. This function takes two parameters, the first
  one being the block map key, and the second one the new string the
  new foreground color of the block, in hexadecimal.
* Update the background color of the block using
  `bar.updateBlockBg()`. This function takes the same parameters as
  `bar.updateBlockFg()`.
* Update the text of the block using `bar.updateBlockTxt()`. Again,
  this function takes two parameters, the second one being the new
  string the block should display.

Agter having ran the required `updateBlock*` functions you must run
`bar.redraw <- "blockname"`, where blockname is the name of the block
map key.


## AUTHORS

Camille Scholtz
