package barfeed

import (
	"goalgotrade/common"
	"goalgotrade/core"
)

type BarFeed interface {
	common.Subject
	GetNewValueEvent() common.Event
	GetCurrentBars() []common.Bar
}

type barFeed struct {
	*core.DefaultSubject
	event common.Event
}

func (b *barFeed) NewBarFeed() *barFeed {
	return &barFeed{
		DefaultSubject: core.NewDefaultSubject(),
		event:          core.NewEvent(),
	}
}

func (b *barFeed) GetNewValueEvent() common.Event {
	return b.event
}

func (b *barFeed) GetCurrentBars() []common.Bar {
	return nil
}
