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


## CREATING BLOCK FUNCTIONS

On top of each block function you should run `bar.initBlock()`, this
is where most of the configuration happens. Here is a short
explanation of each parameter from left to right:

* The name of the block, this is gets used as the name of the block
  map key. (`string`)
* The initial string the block should display. (`string`)
* The width of the block. (`int`)
* The aligment of the text, this can be `'l'` for left aligment, `'c'`
  for center aligment and `'r'` for right aligment. (`rune`)
* Additional x offset to further tweak the location of the text.
  (`int`)
* The foreground color of the block in hexadecimal. (`string`)
* The background color of the block in hexadecimal. (`string`)

You can also additionally specify mousebindings using:

```go
block.actions["buttonN"] = func() {
	// Do stuff.
}
```


---

Everything that should not be ran in a loop should of course be
specified before the `for` loop. For example setting up a connection
to `mpd`.

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

When you've gathered all needed information you can update the block
values using for example `block.bg = value` and running
`bar.redraw <- block`.


## TODO

* Create some kind of easy to use init function for blocks (instead of
  the `if !init` stuff I use at the moment).
* Add popups.
* Drop support for `ttf` fonts and use `pcf` fonts instead if
  possible.
* or maybe some kind of different format altogether that's easily
  hackable, such as suckless farbfeld?


## AUTHORS

Camille Scholtz
