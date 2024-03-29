package core

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"goat/pkg/common"
	"goat/pkg/config"
	"goat/pkg/db"
	"goat/pkg/logger"
	"goat/pkg/metrics"

	"go.uber.org/zap"
)

type Order interface{}

type OrderEvent interface {
	GetOrder() Order
	GetEventType()
	GetEventInfo()
}

type StrategyEventListener interface {
	OnStart(args ...interface{}) error
	OnIdle() error
	OnFinish(args ...interface{}) error
	OnBars(bars Bars) error
	OnOrderUpdated(order Order) error
	OnOrderEvent(orderEvent OrderEvent) error
}

type StrategyController interface {
	Run()
	Stop()
}

type strategyEventListener struct{}

// OnBars implements StrategyEventListener
func (s *strategyEventListener) OnBars(bars Bars) error {
	logger.Logger.Info("onBars", zap.Any("bars", bars))
	return nil
}

// OnFinish implements StrategyEventListener
func (s *strategyEventListener) OnFinish(args ...interface{}) error {
	logger.Logger.Info("onFinish")
	return nil
}

// OnIdle implements StrategyEventListener
func (s *strategyEventListener) OnIdle() error {
	// logger.Logger.Info("onIdle")
	return nil
}

// OnOrderEvent implements StrategyEventListener
func (s *strategyEventListener) OnOrderEvent(orderEvent OrderEvent) error {
	logger.Logger.Info("onOrderEvent", zap.Any("orderEvent", orderEvent))
	return nil
}

// OnOrderUpdated implements StrategyEventListener
func (s *strategyEventListener) OnOrderUpdated(order Order) error {
	logger.Logger.Info("onOrderUpdated", zap.Any("order", order))
	return nil
}

// OnStart implements StrategyEventListener
func (s *strategyEventListener) OnStart(args ...interface{}) error {
	// logger.Logger.Info("onStart")
	return nil
}

func NewSimpleStrategyEventListener() StrategyEventListener {
	return &strategyEventListener{}
}

type strategyController struct {
	ctx      context.Context
	cfg      *config.Config
	dumpDB   *db.DB
	listener StrategyEventListener
	broker   Broker
	dataFeed DataFeed

	dispatcher Dispatcher

	barProcessedEvent Event

	barDataDumpC chan *db.BarData
	closeC       chan struct{}
	dumpWg       sync.WaitGroup
}

func (s *strategyController) onStart(args ...interface{}) error {
	logger.Logger.Debug("onStart")
	return s.listener.OnStart()
}

func (s *strategyController) onIdle(args ...interface{}) error {
	// logger.Logger.Info("onIdle")
	time.Sleep(common.IdleSleepDuration)
	return s.listener.OnIdle()
}

func (s *strategyController) barDumpWorkerLoop() {
	defer s.dumpWg.Done()
	barDataList := []*db.BarData{}
	for {
		select {
		case barData := <-s.barDataDumpC:
			barDataList = append(barDataList, barData)
			if len(barDataList) >= 1000 {
				s.dumpDB.CreateInBatches(barDataList, len(barDataList)).Commit()
				barDataList = nil
			}
		case <-s.ctx.Done():
			for {
				select {
				case barData := <-s.barDataDumpC:
					barDataList = append(barDataList, barData)
				default:
					if s.dumpDB != nil && len(barDataList) > 0 {
						s.dumpDB.CreateInBatches(barDataList, len(barDataList)).Commit()
						barDataList = nil
						s.dumpDB.Commit()
					}
					s.dumpDB = nil
					return
				}
			}
		case <-s.closeC:
			for {
				select {
				case barData := <-s.barDataDumpC:
					barDataList = append(barDataList, barData)
				default:
					if s.dumpDB != nil && len(barDataList) > 0 {
						s.dumpDB.CreateInBatches(barDataList, len(barDataList)).Commit()
						barDataList = nil
						s.dumpDB.Commit()
					}
					s.dumpDB = nil
					return
				}
			}
		default:
			if len(barDataList) > 100 {
				s.dumpDB.CreateInBatches(barDataList, len(barDataList)).Commit()
				barDataList = nil
			} else {
				time.Sleep(time.Microsecond * 100)
			}
		}
	}
}

func (s *strategyController) onBars(args ...interface{}) error {
	// logger.Logger.Debug("StrategyController onBars is called.")
	if len(args) != 2 {
		return fmt.Errorf("onBars args length should be 2")
	}

	// currentTime := args[0].(time.Time)
	data := args[1].(map[string]interface{})
	bars := make(Bars, len(data))
	for k, v := range data {
		bars[k] = v.(Bar)
	}

	// logger.Logger.Debug("onBars",
	// 	zap.Time("time", currentTime),
	// 	zap.Any("bars", bars))
	if s.dumpDB != nil {
		for symbol, bar := range bars {
			if r := bar.GetMeta(BarMetaIsRecovery); r != nil && r.(bool) {
				// we need to skip saving recovery bars
				metrics.BarsNotSaved.Inc()
				continue
			}
			data := &db.BarData{
				Symbol:    symbol,
				DateTime:  bar.DateTime().Unix(),
				Open:      bar.Open(),
				High:      bar.High(),
				Low:       bar.Low(),
				Close:     bar.Close(),
				Volume:    bar.Volume(),
				AdjClose:  bar.AdjClose(),
				Frequency: int64(bar.Frequency()),
			}
			// dump the new data to dump channel
			s.barDataDumpC <- data
			// logger.Logger.Debug("barDataDumpC", zap.Any("data", data))
		}
	}

	s.listener.OnBars(bars)
	s.barProcessedEvent.Emit(bars)

	return nil
}

func (s *strategyController) onOrderEvent(args ...interface{}) error {
	logger.Logger.Info("onOrderEvent")
	if len(args) != 2 {
		return fmt.Errorf("onOrderEvent args length should be 2")
	}

	broker := args[0].(Broker)
	orderEvent := args[1].(OrderEvent)

	logger.Logger.Info("onOrderEvent",
		zap.Any("broker", broker),
		zap.Any("orderEvent", orderEvent))

	s.listener.OnOrderUpdated(orderEvent.GetOrder())

	//TODO: handle order event
	/*
			 # Notify the position about the order event.
		        pos = self.__orderToPosition.get(order.getId(), None)
		        if pos is not None:
		            # Unlink the order from the position if its not active anymore.
		            if not order.isActive():
		                self.unregisterPositionOrder(pos, order)

		            pos.onOrderEvent(orderEvent)
	*/

	s.listener.OnOrderEvent(orderEvent)

	return nil
}

func (s *strategyController) Run() {
	s.dispatcher.Run()
	s.listener.OnFinish()
	s.closeC <- struct{}{}
	close(s.closeC)
	s.dumpWg.Wait()
}

func (s *strategyController) Stop() {
	s.dispatcher.Stop()
}

func NewStrategyController(ctx context.Context, cfg *config.Config, strategyEventListener StrategyEventListener,
	broker Broker, dataFeed DataFeed,
) StrategyController {
	controller := &strategyController{
		ctx:               ctx,
		cfg:               cfg,
		dumpDB:            nil,
		listener:          strategyEventListener,
		broker:            broker,
		dataFeed:          dataFeed,
		dispatcher:        NewDispatcher(ctx),
		barProcessedEvent: NewEvent(),
		barDataDumpC:      make(chan *db.BarData, 1000),
		closeC:            make(chan struct{}, 1),
		dumpWg:            sync.WaitGroup{},
	}

	var err error
	if cfg.Dump.BarDumpDB != "" {
		controller.dumpDB, err = db.NewSQLiteDataBase(cfg.Dump.BarDumpDB,
			cfg.Dump.RemoveOldBars)
		if err != nil {
			logger.Logger.Fatal("failed to create dump db", zap.Error(err))
			os.Exit(1)
		}
	}

	controller.dispatcher.AddSubject(controller.broker)
	controller.dispatcher.AddSubject(controller.dataFeed)

	controller.dispatcher.GetStartEvent().Subscribe(controller.onStart)
	controller.dispatcher.GetIdleEvent().Subscribe(controller.onIdle)

	controller.dataFeed.GetNewValueEvent().Subscribe(controller.onBars)
	controller.broker.GetOrderUpdatedEvent().Subscribe(controller.onOrderEvent)

	controller.dumpWg.Add(1)
	go controller.barDumpWorkerLoop()

	return controller
}
