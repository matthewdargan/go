// #include <u.h>
// #include <libc.h>
// #include <draw.h>
// #include <thread.h>
// #include <cursor.h>
// #include <mouse.h>
// #include <keyboard.h>
// #include <frame.h>
// #include <fcall.h>
// #include <plumb.h>
// #include <libsec.h>
// #include "dat.h"
// #include "fns.h"

package main

import (
	"time"

	"9fans.net/go/cmd/acme/internal/util"
	"9fans.net/go/draw"
	"9fans.net/go/draw/frame"
)

var scrtmp *draw.Image

func scrpos(r draw.Rectangle, p0 int, p1 int, tot int) draw.Rectangle {
	q := r
	h := q.Max.Y - q.Min.Y
	if tot == 0 {
		return q
	}
	if tot > 1024*1024 {
		tot >>= 10
		p0 >>= 10
		p1 >>= 10
	}
	if p0 > 0 {
		q.Min.Y += h * p0 / tot
	}
	if p1 < tot {
		q.Max.Y -= h * (tot - p1) / tot
	}
	if q.Max.Y < q.Min.Y+2 {
		if q.Min.Y+2 <= r.Max.Y {
			q.Max.Y = q.Min.Y + 2
		} else {
			q.Min.Y = q.Max.Y - 2
		}
	}
	return q
}

func scrlresize() {
	scrtmp.Free()
	var err error
	scrtmp, err = display.AllocImage(draw.Rect(0, 0, 32, display.ScreenImage.R.Max.Y), display.ScreenImage.Pix, false, draw.NoFill)
	if err != nil {
		util.Fatal("scroll alloc")
	}
}

func textscrdraw(t *Text) {
	if t.w == nil || t != &t.w.body {
		return
	}
	if scrtmp == nil {
		scrlresize()
	}
	r := t.scrollr
	b := scrtmp
	r1 := r
	r1.Min.X = 0
	r1.Max.X = r.Dx()
	r2 := scrpos(r1, t.org, t.org+t.fr.NumChars, t.file.b.nc)
	if !(r2 == t.lastsr) {
		t.lastsr = r2
		b.Draw(r1, t.fr.Cols[frame.BORD], nil, draw.ZP)
		b.Draw(r2, t.fr.Cols[frame.BACK], nil, draw.ZP)
		r2.Min.X = r2.Max.X - 1
		b.Draw(r2, t.fr.Cols[frame.BORD], nil, draw.ZP)
		t.fr.B.Draw(r, b, nil, draw.Pt(0, r1.Min.Y))
		/*flushimage(display, 1); // BUG? */
	}
}

func scrsleep(dt time.Duration) {
	timer := time.NewTimer(dt)
	defer timer.Stop()
	select {
	case <-timer.C:
	case mousectl.Mouse = <-mousectl.C:
	}
}

func textscroll(t *Text, but int) {
	s := t.scrollr.Inset(1)
	h := s.Max.Y - s.Min.Y
	x := (s.Min.X + s.Max.X) / 2
	oldp0 := ^0
	first := true
	for {
		display.Flush()
		my := mouse.Point.Y
		if my < s.Min.Y {
			my = s.Min.Y
		}
		if my >= s.Max.Y {
			my = s.Max.Y
		}
		if !(mouse.Point == draw.Pt(x, my)) {
			display.MoveCursor(draw.Pt(x, my))
			mousectl.Read() /* absorb event generated by moveto() */
		}
		var p0 int
		if but == 2 {
			y := my
			p0 = int(int64(t.file.b.nc) * int64(y-s.Min.Y) / int64(h))
			if p0 >= t.q1 {
				p0 = textbacknl(t, p0, 2)
			}
			if oldp0 != p0 {
				textsetorigin(t, p0, false)
			}
			oldp0 = p0
			mousectl.Read()
			goto Continue
		}
		if but == 1 {
			p0 = textbacknl(t, t.org, (my-s.Min.Y)/t.fr.Font.Height)
		} else {
			p0 = t.org + t.fr.CharOf(draw.Pt(s.Max.X, my))
		}
		if oldp0 != p0 {
			textsetorigin(t, p0, true)
		}
		oldp0 = p0
		/* debounce */
		if first {
			display.Flush()
			time.Sleep(200 * time.Millisecond)
			select {
			default:
				// non-blocking
			case mousectl.Mouse = <-mousectl.C:
				// ok
			}
			first = false
		}
		scrsleep(80 * time.Millisecond)
	Continue:
		if mouse.Buttons&(1<<(but-1)) == 0 {
			break
		}
	}
	for mouse.Buttons != 0 {
		mousectl.Read()
	}
}
