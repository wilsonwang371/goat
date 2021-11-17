package barfeed

import (
	"goalgotrade/common"
	"goalgotrade/feed"

	"github.com/go-gota/gota/series"
)

type baseBarFeed struct {
	*feed.BaseFeed
}

func NewBaseBarFeed(stype series.Type, maxlen int) common.BarFeed {
	return &baseBarFeed{
		BaseFeed: feed.NewBaseFeed(stype, maxlen),
	}
}

func (b *baseBarFeed) GetCurrentBars() []common.Bar {
	// TODO: Implement me
	return nil
}
