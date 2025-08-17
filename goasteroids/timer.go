package goasteroids

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Timer struct {
	currencyTicks int
	targetTicks   int
}

func NewTimer(d time.Duration) *Timer {
	return &Timer{
		currencyTicks: 0,
		targetTicks:   int(d.Milliseconds()) * ebiten.TPS() / 1000,
	}
}

func (t *Timer) Update() {
	if t.currencyTicks < t.targetTicks {
		t.currencyTicks++
	}
}

func (t *Timer) IsReady() bool {
	return t.currencyTicks >= t.targetTicks
}

func (t *Timer) Reset() {
	t.currencyTicks = 0
}
