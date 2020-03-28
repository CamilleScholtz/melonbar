package main

// TODO: Instead of doing this using rune-count, do this using pixel-count.
func trim(txt string, l int) string {
	if len(txt) > l {
		return txt[0:l] + "..."
	}
	return txt
}
