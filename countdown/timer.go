package countdown

import (
	"fmt"
	"time"

	"github.com/aelnahas/pomo/sessions"
	"github.com/aelnahas/pomo/task"
	"github.com/nsf/termbox-go"
)

type Countdown struct {
	timer        *time.Timer
	ticker       *time.Ticker
	queues       chan termbox.Event
	startDone    bool
	startX       int
	startY       int
	duration     time.Duration
	task         *task.Task
	sessiontType sessions.Type
}

const tick = time.Second

var controls = []string{
	"CTRL-C | ESC -> Quit",
	"p      | P   -> Pause",
	"c      | C   -> Continue",
}

var (
	Green  = termbox.RGBToAttribute(154, 255, 0)
	Red    = termbox.RGBToAttribute(255, 0, 68)
	Yellow = termbox.RGBToAttribute(255, 196, 0)
	Cyan   = termbox.RGBToAttribute(0, 239, 255)
)

var colorTypeMap map[sessions.Type]termbox.Attribute = map[sessions.Type]termbox.Attribute{
	sessions.Focus: termbox.ColorDefault,
	sessions.Short: Green,
	sessions.Long:  Cyan,
}

func New(d time.Duration, task *task.Task, sessionType sessions.Type) *Countdown {
	return &Countdown{
		duration:     d,
		task:         task,
		sessiontType: sessionType,
	}
}

func (c *Countdown) Draw(d time.Duration) {
	w, h := termbox.Size()
	clear()

	str := format(d)
	text := toText(str)

	fg := colorTypeMap[c.sessiontType]
	if c.sessiontType == sessions.Focus {
		remaining := int(100 * float64(d) / float64(c.duration))
		switch {
		case remaining > 25 && remaining < 50:
			fg = Yellow
		case remaining < 25:
			fg = Red
		}
	}

	if !c.startDone {
		c.startDone = true
		c.startX = w/2 - text.width()/2
		c.startY = h/2 - text.height()/2
	}

	x, y := c.startX, c.startY
	for _, s := range text {
		echo(s, x, y, fg)
		x += s.width()
	}

	description := Symbol([]string{string(c.sessiontType), c.task.Title})
	y += text.height()
	x = w/2 - description.width()/2
	echo(description, x, y, termbox.ColorDefault)
	showControls(w, h)
	showNumSessions(w, h, c.task.Sessions)
	flush()
}

func showControls(w, h int) {
	escape := Symbol(controls)
	echo(escape, 0, h-escape.height(), termbox.ColorDefault)
}

func showNumSessions(w, h, sessions int) {
	symbol := Symbol([]string{fmt.Sprintf("sessions : %d", sessions)})
	echo(symbol, w-symbol.width(), h-symbol.height(), termbox.ColorDefault)
}

func (c *Countdown) start(d time.Duration) {
	c.timer = time.NewTimer(d)
	c.ticker = time.NewTicker(tick)
}

func (c *Countdown) stop() {
	c.timer.Stop()
	c.ticker.Stop()
}

func (c *Countdown) countdown(timeLeft time.Duration, countUp bool) error {
	c.start(timeLeft)
	if countUp {
		timeLeft = 0
	}
	c.Draw(timeLeft)
	defer termbox.Close()

	for {
		select {
		case ev := <-c.queues:
			if ev.Type == termbox.EventKey && (ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC) {
				return fmt.Errorf("Error: timer interrupted")
			}
			if ev.Ch == 'p' || ev.Ch == 'P' {
				c.stop()
			}
			if ev.Ch == 'c' || ev.Ch == 'C' {
				c.start(timeLeft)
			}
		case <-c.ticker.C:
			if countUp {
				timeLeft += tick
			} else {
				timeLeft -= tick
			}
			c.Draw(timeLeft)
		case <-c.timer.C:
			return nil
		}
	}
}

func (c *Countdown) Run() error {
	termbox.SetOutputMode(termbox.OutputRGB)
	if err := termbox.Init(); err != nil {
		return err
	}

	c.queues = make(chan termbox.Event)
	go func() {
		for {
			c.queues <- termbox.PollEvent()
		}
	}()
	return c.countdown(c.duration, false)
}

func format(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h < 1 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
