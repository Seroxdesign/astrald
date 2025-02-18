package admin

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cryptopunkscc/astrald/log"
	"io"
	"strings"
	"time"
)

const useColorTerminal = true

type Terminal struct {
	Color  bool
	log    *log.Logger
	output log.Output
	io.ReadWriter
}

type Header string
type Keyword string
type Faded string
type Important string

func NewTerminal(rw io.ReadWriter, logger *log.Logger) *Terminal {
	l := logger.Tag("")
	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		return []log.Op{log.OpText{Text: s}}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(Header)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.White},
			log.OpText{Text: strings.ToUpper(string(s))},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(Keyword)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Yellow},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(Faded)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.BrightBlack},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(Important)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.BrightRed},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		t, ok := v.(time.Time)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpText{Text: t.Format(timestampFormat)},
		}, true
	})
	return &Terminal{
		Color:      useColorTerminal,
		ReadWriter: rw,
		output:     log.NewColorOutput(rw),
		log:        l,
	}
}

func (t *Terminal) Sprintf(f string, v ...any) string {
	var buf = &bytes.Buffer{}
	var out log.Output

	if t.Color {
		out = log.NewColorOutput(buf)
	} else {
		out = log.NewMonoOutput(buf)
	}

	out.Do(t.log.Renderf(f, v...)...)

	return buf.String()
}

func (t *Terminal) Printf(f string, v ...any) {
	if t.Color {
		t.output.Do(t.log.Renderf(f, v...)...)
		return
	}
	fmt.Fprintf(t, f, v...)
}

func (t *Terminal) Println(v ...any) {
	if t.Color {
		for _, i := range v {
			ops, b := t.log.Render(i)
			if b {
				t.output.Do(ops...)
			} else {
				fmt.Fprintln(t, i)
			}
		}
		t.output.Do(log.OpText{Text: "\n"})
		return
	}
	fmt.Fprintln(t, v...)
}

func (t *Terminal) Scanf(f string, v ...any) {
	fmt.Fscanf(t, f, v...)
}

func (t *Terminal) ScanLine() (string, error) {
	var scanner = bufio.NewScanner(t)
	if !scanner.Scan() {
		return "", io.EOF
	}
	return scanner.Text(), nil
}
