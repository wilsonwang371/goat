package barfeed

import (
	"goalgotrade/common"
	"goalgotrade/core"
)

type barFeed struct {
	*core.DefaultSubject
	event common.Event
}

func (b *barFeed) NewBarFeed() common.BarFeed {
	return &barFeed{
		DefaultSubject: core.NewDefaultSubject(),
		event:          core.NewEvent(),
	}
}

func (b *barFeed) GetNewValueEvent() common.Event {
	return b.event
}

func (b *barFeed) GetCurrentBars() []common.Bar {
	// TODO: Implement me
	return nil
}
