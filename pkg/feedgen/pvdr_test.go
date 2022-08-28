package feedgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/js"
	"goat/pkg/logger"

	"go.uber.org/zap"
)

func TestTradingViewSimple(t *testing.T) {
	user := os.Getenv("TRADINGVIEW_USER")
	pass := os.Getenv("TRADINGVIEW_PASS")
	if user == "" || pass == "" {
		t.Skip("TRADINGVIEW_USER and TRADINGVIEW_PASS must be set")
	}
	gen := NewLiveBarFeedGenerator(
		NewTradingViewDataProvider(user, pass),
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(&config.Config{}, gen, 100, "")
	disp.AddSubject(feed)

	go gen.Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}

func TestFakeSimple(t *testing.T) {
	gen := NewLiveBarFeedGenerator(
		NewFakeDataProvider(),
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(&config.Config{}, gen, 100, "")
	disp.AddSubject(feed)

	go gen.Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}

func Test2FakeSimple(t *testing.T) {
	gen := NewMultiLiveBarFeedGenerator(
		[]BarDataProvider{NewFakeDataProvider(), NewFakeDataProvider()},
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(&config.Config{}, gen, 100, "")
	disp.AddSubject(feed)

	go gen.Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}

func TestFx678DataGen(t *testing.T) {
	count := 0

	f := NewFx678DataProvider()

	for {
		if bar, err := f.(*fx678DataProvider).getOneBar("XAUUSD"); err != nil {
			logger.Logger.Info("error getting a bar", zap.Error(err))
			count++
			time.Sleep(time.Second * 5)
		} else {
			t.Log(bar)
			return
		}

		if count > 10 {
			t.Error("failed to get bar")
			return
		}
	}
}

func TestGoldPriceOrgDataGen(t *testing.T) {
	count := 0

	f := NewGoldPriceOrgDataProvider()

	for {
		if bar, err := f.(*goldPriceOrgDataProvider).getOneBar("XAUUSD"); err != nil {
			logger.Logger.Info("error getting a bar", zap.Error(err))
			count++
			time.Sleep(time.Second * 5)
		} else {
			t.Log(bar)
			return
		}

		if count > 10 {
			t.Error("failed to get bar")
			return
		}
	}
}

var runWg *sync.WaitGroup

func startLive() error {
	logger.Logger.Info("start live strategy data feed")
	if runWg == nil {
		panic("runWg is nil")
	}
	runWg.Done()
	return nil
}

func TestMultiProviders(t *testing.T) {
	cfg := config.Config{}

	pArr := []BarDataProvider{
		NewFakeDataProvider(),
		NewFakeDataProvider(),
	}
	gen := NewMultiLiveBarFeedGenerator(
		pArr,
		cfg.Symbol,
		[]core.Frequency{core.REALTIME},
		100)

	if gen == nil {
		logger.Logger.Error("failed to create feed generator")
		os.Exit(1)
	}
	runWg = &sync.WaitGroup{}
	runWg.Add(1)

	go gen.WaitAndRun(runWg)

	feed := core.NewGenericDataFeed(&config.Config{}, gen, 100, "")

	rt := js.NewStrategyRuntime(&cfg, feed, startLive)
	script, err := ioutil.ReadFile("../../samples/strategies/simple.js")
	if err != nil {
		logger.Logger.Error("failed to read script file", zap.Error(err))
		os.Exit(1)
	}
	if compiledScript, err := rt.Compile(string(script)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		sel := js.NewJSStrategyEventListener(rt)
		broker := core.NewDummyBroker(feed)
		strategy := core.NewStrategyController(&cfg, sel, broker, feed)

		if val, err := rt.Execute(compiledScript); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println(val)
		}

		go func() {
			time.Sleep(time.Second * 15)
			gen.Finish()
		}()

		strategy.Run()
	}
}
