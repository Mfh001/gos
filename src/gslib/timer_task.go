package gslib

import (
	"github.com/ryszard/goskiplist/skiplist"
)

type TimerTask struct {
}

func (t *TimerTask) Add(key string, delay int, cb func()) {
}

func (t *TimerTask) Update(key string, delay int) {
}

func (t *TimerTask) Finish(key string) {
}

func (t *TimerTask) Del(key string) {
}
