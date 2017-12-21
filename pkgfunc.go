package rog

import (
	"os"
)

var (
	std     = New(os.Stderr, "", LstdFlags)
	Discard = New(nil, "", 0)
	debug   = New(os.Stderr, "", LstdFlags|Lshortfile)

	debugging = false
)

func init() {
	std.viaExposed = true
	debug.viaExposed = true
}

func EnableDebug(optLogger ...*logger) {
	debugging = true
	if len(optLogger) > 0 {
		debug = optLogger[0]
	}
}

func DisableDebug() {
	debugging = false
}

func Print(v ...interface{}) {
	std.Print(v...)
}

func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

func Debug(v ...interface{}) {
	if debugging {
		debug.Print(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if debugging {
		debug.Printf(format, v...)
	}
}
