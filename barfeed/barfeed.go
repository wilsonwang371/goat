package barfeed

import (
	"goalgotrade/common"
	"goalgotrade/feed"
)

type baseBarFeed struct {
	*feed.BaseFeed
}

func NewBaseBarFeed(maxlen int) common.BarFeed {
	return &baseBarFeed{
		BaseFeed: feed.NewBaseFeed(maxlen),
	}
}

func (b *baseBarFeed) GetCurrentBars() []common.Bar {
	// TODO: Implement me
	return nil
}
