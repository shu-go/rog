package test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/shu-go/gotwant"
	"github.com/shu-go/rog"
)

func init() {
	l := rog.New(os.Stderr, "", rog.LstdFlags|rog.Lmicroseconds)
	l.Print("Lmicroseconds")

	l = rog.New(os.Stderr, "", rog.LstdFlags|rog.Lshortfile)
	l.Print("Lshortfile")
}

func TestPrint(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", rog.LstdFlags|rog.Lcompat)

	stdl.Print("hello")
	l.Print("hello")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	stdl.Print("hello", "world")
	l.Print("hello", "world")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))
}

func TestPrintf(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", rog.LstdFlags|rog.Lcompat)

	stdl.Printf("hello")
	l.Printf("hello")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	stdl.Print("hello", "world")
	l.Print("hello", "world")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	stdl.Printf("aaa %v, %v", "hello", "world")
	l.Printf("aaa %v, %v", "hello", "world")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))
}

func TestBind(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", rog.LstdFlags|rog.Lcompat)
	ll := l.Bind("[L]")

	stdl.Print("[L]abc")
	ll.Print("abc")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	lll := ll.Bind(123, "{L}")
	stdl.Print("[L] 123 {L}def")
	lll.Print("def")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	stdl.Print("ghi")
	l.Print("ghi")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))
}

type counterHook struct {
	count int
}

func (h *counterHook) Run(v ...interface{}) bool {
	h.count++
	return true
}

type asyncCounterHook struct {
	count int
}

func (h *asyncCounterHook) Run(v ...interface{}) bool {
	go func() {
		h.count++
		time.Sleep(time.Millisecond)
	}()
	return true
}

func TestHook(t *testing.T) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", log.LstdFlags)

	cnth := &counterHook{}
	l.Hook(cnth)

	gotwant.Test(t, cnth.count, 0)

	l.Print("hello")
	gotwant.Test(t, cnth.count, 1)

	l.Print("hello")
	gotwant.Test(t, cnth.count, 2)
}

func TestPrefix(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "[DEBUG]", log.LstdFlags)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", rog.LstdFlags|rog.Lcompat)

	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))

	stdl.SetPrefix("[Info]")
	l.SetPrefix("[Info]")
	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")
	gotwant.Test(t, buf.String(), stdbuf.String(), gotwant.Format("%q"))
}

func TestNil(t *testing.T) {
	stdl := log.New(ioutil.Discard, "[DEBUG]", log.LstdFlags)
	l := rog.New(nil, "[DEBUG]", rog.LstdFlags|rog.Lcompat)

	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")

	stdl.SetPrefix("[Info]")
	l.SetPrefix("[Info]")
	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")
}

func TestFilename(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags|log.Lshortfile)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", log.LstdFlags|log.Lshortfile)

	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")
	//gotwant.Test(t, buf.String(), stdbuf.String())
}

func TestUTC(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags|log.LUTC)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", rog.LstdFlags|rog.LUTC|rog.Lcompat)

	stdl.Print("a", "b", "c")
	l.Print("a", "b", "c")
	gotwant.Test(t, buf.String(), stdbuf.String())
}

func TestDebug(t *testing.T) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "", log.LstdFlags)

	buf := bytes.NewBufferString("")
	l := rog.New(buf, "", log.LstdFlags)

	rog.Debug("Debug output disabled")
	rog.EnableDebug()
	rog.Debug("Debug output enabled")

	rog.EnableDebug(l)

	stdl.Print("Debug output")
	rog.Debug("Debug output")
	gotwant.Test(t, buf.String(), stdbuf.String())
}

func TestFileHook(t *testing.T) {
	l := rog.New(nil, "", 0)
	l.Hook(rog.FileHook("./hooked.txt", os.ModePerm))
	l.Print("hello, world")
}

func BenchmarkStdLog(b *testing.B) {
	stdbuf := bytes.NewBufferString("")
	stdl := log.New(stdbuf, "[DEBUG]", log.LstdFlags)

	for i := 0; i < b.N; i++ {
		stdl.Print("a", "b", "c", i)
	}
}

func BenchmarkRog(b *testing.B) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", rog.LstdFlags)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkRogCompat(b *testing.B) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", rog.LstdFlags|rog.Lcompat)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkMicroseconds(b *testing.B) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", rog.LstdFlags|rog.Lmicroseconds)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkRogWithHook(b *testing.B) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", log.LstdFlags)

	cnth := &counterHook{}
	l.Hook(cnth)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkRogWithAsyncHook(b *testing.B) {
	buf := bytes.NewBufferString("")
	l := rog.New(buf, "[DEBUG]", log.LstdFlags)

	cnth := &asyncCounterHook{}
	l.Hook(cnth)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkStdLogDiscard(b *testing.B) {
	stdl := log.New(ioutil.Discard, "[DEBUG]", log.LstdFlags)

	for i := 0; i < b.N; i++ {
		stdl.Print("a", "b", "c", i)
	}
}

func BenchmarkRogDiscard(b *testing.B) {
	l := rog.New(nil, "[DEBUG]", log.LstdFlags)

	for i := 0; i < b.N; i++ {
		l.Print("a", "b", "c", i)
	}
}

func BenchmarkDebugDisabled(b *testing.B) {
	rog.DisableDebug()
	for i := 0; i < b.N; i++ {
		rog.Debug("a", "b", "c", i)
	}
}
