package feedgen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"goat/pkg/core"
	"goat/pkg/logger"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

var goldPriceOrgSupportedSymbols []string = []string{"XAU", "XAG"}

type GoldPriceOrgBar struct {
	Timestamp  uint64 `json:"ts"`
	Timestamp2 uint64 `json:"tsj"`
	Date       string `json:"date"`
	Items      []struct {
		Currency string  `json:"curr"`
		XauPrice float64 `json:"xauPrice"`
		XagPrice float64 `json:"xagPrice"`
		ChgXau   float64 `json:"chgXau"`
		ChgXag   float64 `json:"chgXag"`
		PcXau    float64 `json:"pcXau"`
		PcXag    float64 `json:"pcXag"`
		XauClose float64 `json:"xauClose"`
		XagClose float64 `json:"xagClose"`
	} `json:"items"`
}

type goldPriceOrgDataProvider struct {
	instrument string
	freqList   []core.Frequency
	stopped    bool
}

func (f *goldPriceOrgDataProvider) getOneBar(instrument string) (core.Bar, error) {
	barRaw := GoldPriceOrgBar{}
	reqUrl := "https://data-asg.goldprice.org/dbXRates/USD"
	if resp, err := f.sendRequest(reqUrl); err != nil {
		log.Printf("error sending request: %v", err)
		return nil, err
	} else {
		if err := json.Unmarshal([]byte(resp), &barRaw); err != nil {
			log.Printf("error unmarshalling response: %v", err)
			return nil, err
		}

		if len(barRaw.Items) != 1 {
			return nil, fmt.Errorf("unexpected number of items in response: %d", len(barRaw.Items))
		}

		if barRaw.Items[0].Currency != "USD" {
			return nil, fmt.Errorf("unexpected currency in response: %s", barRaw.Items[0].Currency)
		}

		t := time.Unix(int64(barRaw.Timestamp2/1000), 0)

		if instrument == "XAU" {
			return core.NewBasicBar(t, barRaw.Items[0].XauPrice, barRaw.Items[0].XauPrice, barRaw.Items[0].XauPrice, barRaw.Items[0].XauPrice, barRaw.Items[0].XauPrice, 0, core.REALTIME), nil
		} else if instrument == "XAG" {
			return core.NewBasicBar(t, barRaw.Items[0].XagPrice, barRaw.Items[0].XagPrice, barRaw.Items[0].XagPrice, barRaw.Items[0].XagPrice, barRaw.Items[0].XagPrice, 0, core.REALTIME), nil
		} else {
			return nil, fmt.Errorf("unexpected instrument: %s", instrument)
		}
	}
}

func (f *goldPriceOrgDataProvider) sendRequest(reqUrl string) (string, error) {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: RequestTimeoutDuration,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return string(body), nil
}

func (f *goldPriceOrgDataProvider) init(instrument string, freqList []core.Frequency) error {
	found := false

	for _, sym := range goldPriceOrgSupportedSymbols {
		if sym == instrument {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("instrument %s not supported", instrument)
	}

	if len(freqList) == 0 {
		return fmt.Errorf("freqList is empty")
	}
	for _, freq := range freqList {
		if freq != core.REALTIME {
			return fmt.Errorf("freq %v not supported", freq)
		}
	}
	f.instrument = instrument
	f.freqList = freqList
	return nil
}

func (f *goldPriceOrgDataProvider) connect() error {
	return nil
}

func (f *goldPriceOrgDataProvider) nextBars() (core.Bars, error) {
	// this can return nothing but with no error, you should not block this forever
	if f.stopped {
		return nil, fmt.Errorf("goldprice.org data provider is stopped")
	}
	time.Sleep(SleepDuration)
	basicBar, err := f.getOneBar(f.instrument)
	if err != nil {
		logger.Logger.Warn("error getting a bar: %v", zap.Error(err))
		return nil, err
	}
	return core.Bars{f.instrument: basicBar}, nil
}

func (f *goldPriceOrgDataProvider) reset() error {
	return nil
}

func (f *goldPriceOrgDataProvider) stop() error {
	f.stopped = true
	return nil
}

func (f *goldPriceOrgDataProvider) datatype() series.Type {
	return series.Float
}

func NewGoldPriceOrgDataProvider() BarDataProvider {
	return &goldPriceOrgDataProvider{
		stopped: false,
	}
}
