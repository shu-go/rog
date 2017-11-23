package rog

import (
	"bytes"
	"fmt"
	"io"
	//"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	//"strconv"
	"sync"
	"time"
)

var (
	d2tbl [100][2]byte
)

const (
	Ldate         = log.Ldate
	Ltime         = log.Ltime
	Lmicroseconds = log.Lmicroseconds
	Llongfile     = log.Llongfile
	Lshortfile    = log.Lshortfile
	LUTC          = log.LUTC
	LstdFlags     = log.LstdFlags

	Lcompat = 1 << 8
)

func init() {
	for i := 0; i < len(d2tbl); i++ {
		if i < 10 {
			d2tbl[i] = [2]byte{'0', '0' + byte(i)}
		} else {
			d2tbl[i] = [2]byte{'0' + byte(i/10), '0' + byte(i%10)}
		}
	}
}

type Hook interface {
	Run(v ...interface{}) bool
}

type logger struct {
	mu sync.Mutex

	out    io.Writer
	prefix string
	flag   int

	hbuf *bytes.Buffer // prefix ~ file line
	bbuf *bytes.Buffer // format and v...

	cache struct {
		year, month, day int
		dateCache        []byte
		hour, minute     int
		timeCache        []byte
	}

	viaExposed bool

	hooks []Hook

	bounds []interface{}
}

func New(out io.Writer, prefix string, flag int) *logger {
	return &logger{
		out:    out,
		prefix: prefix,
		flag:   flag,
	}
}

func (l *logger) Bind(values ...interface{}) *logger {
	newLogger := *l // dup
	newLogger.bounds = make([]interface{}, 0, len(l.bounds)+len(values))
	newLogger.bounds = append(newLogger.bounds, l.bounds...)
	newLogger.bounds = append(newLogger.bounds, values...)
	return &newLogger
}

func (l *logger) Hook(h Hook) {
	l.mu.Lock()
	l.hooks = append(l.hooks, h)
	l.mu.Unlock()
}

func (l *logger) ResetHooks() {
	l.mu.Lock()
	l.hooks = nil
	l.mu.Unlock()
}

func (l *logger) Print(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var values []interface{}
	if len(l.bounds) > 0 {
		values = make([]interface{}, 0, len(v)+len(l.bounds))
		values = append(values, l.bounds...)
		values = append(values, v...)
	} else {
		values = v
	}

	for _, h := range l.hooks {
		if h.Run(values...) {
			return
		}
	}

	if l.out == nil {
		return
	}

	if l.viaExposed {
		l.outputHeader(3)
	} else {
		l.outputHeader(2)
	}

	if l.flag&Lcompat == 0 {
		fmt.Fprintln(l.out, values...)
		return
	}

	if l.bbuf == nil {
		l.bbuf = new(bytes.Buffer)
	} else {
		l.bbuf.Reset()
	}
	prevIsString := false
	isString := false
	for i, vv := range values {
		_, ok := vv.(string)
		isString = ok

		if i == 0 || (isString && prevIsString) {
			//nop
		} else {
			l.bbuf.WriteByte(' ')
		}
		fmt.Fprint(l.bbuf, vv)

		prevIsString = isString
	}

	chk := l.bbuf.Bytes()
	if len(chk) > 0 && chk[len(chk)-1] == '\n' {
	} else {
		l.bbuf.WriteByte('\n')
	}

	l.bbuf.WriteTo(l.out)

}

func (l *logger) Printf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var values []interface{}
	if len(l.bounds) > 0 {
		values = make([]interface{}, 0, len(v)+len(l.bounds))
		values = append(values, l.bounds...)
		values = append(values, v...)
	} else {
		values = v
	}

	// NO HOOK

	if l.out == nil {
		return
	}

	if l.viaExposed {
		l.outputHeader(3)
	} else {
		l.outputHeader(2)
	}

	if l.flag&Lcompat == 0 {
		fmt.Fprintln(l.out, values...)
		return
	}

	if l.bbuf == nil {
		l.bbuf = new(bytes.Buffer)
	} else {
		l.bbuf.Reset()
	}
	fmt.Fprintf(l.bbuf, format, values...)

	chk := l.bbuf.Bytes()
	if len(chk) > 0 && chk[len(chk)-1] == '\n' {
	} else {
		l.bbuf.WriteByte('\n')
	}

	l.bbuf.WriteTo(l.out)
}

func (l *logger) outputHeader(calldepth int) {
	if l.out == nil {
		return
	}

	if l.hbuf == nil {
		l.hbuf = new(bytes.Buffer)
	} else {
		l.hbuf.Reset()
	}

	if l.prefix != "" {
		l.hbuf.Write([]byte(l.prefix))
	}

	var now time.Time
	if l.flag&(Ldate|Ltime) != 0 {
		if l.flag&LUTC != 0 {
			now = time.Now().UTC()
		} else {
			now = time.Now()
		}
	}
	if l.flag&Ldate != 0 {
		year, month, day := now.Date()
		if !(year == l.cache.year && int(month) == l.cache.month && day == l.cache.day) {
			if l.cache.dateCache == nil {
				//l.cache.dateCache = make([]byte, 0, 10)
				l.cache.dateCache = make([]byte, 11)
			} else {
				//l.cache.dateCache = l.cache.dateCache[:0]
			}
			/*
				l.cache.dateCache = append(l.cache.dateCache, d2tbl[year/100][:]...)
				l.cache.dateCache = append(l.cache.dateCache, d2tbl[year%100][:]...)
				l.cache.dateCache = append(l.cache.dateCache, '/')
				l.cache.dateCache = append(l.cache.dateCache, d2tbl[int(month)][:]...)
				l.cache.dateCache = append(l.cache.dateCache, '/')
				l.cache.dateCache = append(l.cache.dateCache, d2tbl[day][:]...)
				l.cache.dateCache = append(l.cache.dateCache, ' ')
			*/
			l.cache.dateCache[0] = d2tbl[year/100][0]
			l.cache.dateCache[1] = d2tbl[year/100][1]
			l.cache.dateCache[2] = d2tbl[year%100][0]
			l.cache.dateCache[3] = d2tbl[year%100][1]
			l.cache.dateCache[4] = '/'
			l.cache.dateCache[5] = d2tbl[int(month)][0]
			l.cache.dateCache[6] = d2tbl[int(month)][1]
			l.cache.dateCache[7] = '/'
			l.cache.dateCache[8] = d2tbl[day][0]
			l.cache.dateCache[9] = d2tbl[day][1]
			l.cache.dateCache[10] = ' '

			l.cache.year, l.cache.month, l.cache.day = year, int(month), day
		}
		l.hbuf.Write(l.cache.dateCache)
	}
	if l.flag&Ltime != 0 {
		hour, minute, second := now.Clock()
		if !(hour == l.cache.hour && minute == l.cache.minute) {
			if l.cache.timeCache == nil {
				//l.cache.timeCache = make([]byte, 0, 6)
				l.cache.timeCache = make([]byte, 6)
			} else {
				//l.cache.timeCache = l.cache.timeCache[:0]
			}
			/*
				l.cache.timeCache = append(l.cache.timeCache, d2tbl[hour][:]...)
				l.cache.timeCache = append(l.cache.timeCache, ':')
				l.cache.timeCache = append(l.cache.timeCache, d2tbl[minute][:]...)
				l.cache.timeCache = append(l.cache.timeCache, ':')
			*/
			l.cache.timeCache[0] = d2tbl[hour][0]
			l.cache.timeCache[1] = d2tbl[hour][1]
			l.cache.timeCache[2] = ':'
			l.cache.timeCache[3] = d2tbl[minute][0]
			l.cache.timeCache[4] = d2tbl[minute][1]
			l.cache.timeCache[5] = ':'
		}
		l.hbuf.Write(l.cache.timeCache)
		l.hbuf.Write(d2tbl[second][:])

		if l.flag&Lmicroseconds != 0 {
			micro := (now.Nanosecond() / 1000) % 1000000
			l.hbuf.WriteByte('.')
			//fmt.Fprintf(l.hbuf, "%06d", micro)
			a := d2tbl[int(micro/10000)]
			b := d2tbl[int(micro/100)%100]
			c := d2tbl[int(micro)%100]
			l.hbuf.Write(a[:])
			l.hbuf.Write(b[:])
			l.hbuf.Write(c[:])
		}

		l.hbuf.WriteByte(' ')
	}

	if l.flag&(Llongfile|Lshortfile) != 0 {
		// https://golang.org/src/log/log.go?s=10258:10300#L153
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}

		if l.flag&Lshortfile != 0 {
			file = filepath.Base(file)
		}
		fmt.Fprintf(l.hbuf, "%s:%d: ", file, line)
		//l.hbuf.Write([]byte(file))
		//l.hbuf.WriteByte(':')
		//writePositiveInt(line, l.hbuf)
		//l.hbuf.WriteByte(' ')
	}

	l.hbuf.WriteTo(l.out)
}

func (l *logger) SetPrefix(prefix string) {
	l.mu.Lock()
	l.prefix = prefix
	l.mu.Unlock()
}

func (l *logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

func (l *logger) SetFlags(flag int) {
	l.mu.Lock()
	l.flag = flag
	l.mu.Unlock()
}

func (l *logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

func (l *logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	l.out = out
	l.mu.Unlock()
}

// from log.itoa
func writePositiveInt(v int, out io.Writer) {
	if v <= 0 {
		out.Write([]byte{'0'})
		return
	}

	var buf [20]byte
	for i := len(buf) - 1; i >= 0; i -= 2 {
		/*
			buf[i] = '0' + byte(v%10)
			v /= 10
			if v == 0 {
				out.Write(buf[i:])
				return
			}
		*/
		d2 := d2tbl[v%100]
		buf[i] = d2[1]
		buf[i-1] = d2[0]
		v /= 100
		if v == 0 {
			if d2[0] == '0' {
				out.Write(buf[i:])
			} else {
				out.Write(buf[i-1:])
			}
			return
		}
	}
	out.Write(buf[:])
}
