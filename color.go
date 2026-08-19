// color.go
package main

import "os"

// colorEnabled returns true unless NO_COLOR env var is set.
func colorEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

func dim(s string) string {
	if !colorEnabled() {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

func green(s string) string {
	if !colorEnabled() {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

// checkMark returns a green ✓ or plain "x" when color is disabled.
func checkMark() string {
	if colorEnabled() {
		return green("✓")
	}
	return "x"
}

// dot returns a dimmed · or plain "-" when color is disabled.
func dot() string {
	if colorEnabled() {
		return dim("·")
	}
	return "-"
}
