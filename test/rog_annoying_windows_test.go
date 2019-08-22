package test

import (
	"testing"

	"github.com/shu-go/rog"
)

func DoNotTestAnnoyingHook(t *testing.T) {
	l := rog.Discard.Bind()
	l.Hook(rog.AnnoyingHook())

	l.Print("Hello?")
	l.Print("Are you there?")

	// not hooked
	rog.Discard.Print("I'm not here.")
}
