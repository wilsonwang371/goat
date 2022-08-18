package feedgen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"goat/pkg/core"
	"goat/pkg/logger"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

type Fx678Bar struct {
	Status string   `json:"s"`
	Time   []string `json:"t"`
	Close  []string `json:"c"`
	Open   []string `json:"o"`
	High   []string `json:"h"`
	Low    []string `json:"l"`
	Price  []string `json:"p"`
	Volume []string `json:"v"`
	Bid    []string `json:"b"`
	Se     []string `json:"se"`
}

type fx678DataProvider struct {
	instrument string
	freqList   []core.Frequency
	stopped    bool
}

const (
	SleepDuration          = 10 * time.Second
	RequestTimeoutDuration = 10 * time.Second
)

var symbolMap map[string]string = map[string]string{
	"XAU": "WGJS",
}

func getABar(instrument string) (core.Bar, error) {
	barRaw := Fx678Bar{}
	reqUrl := fmt.Sprintf("https://api-q.fx678img.com/getQuote.php?exchName=%s&symbol=%s&st=%.16f", symbolMap[instrument], instrument, rand.Float64())
	if resp, err := sendRequest(reqUrl); err != nil {
		log.Printf("error sending request: %v", err)
		return nil, err
	} else {
		if err := json.Unmarshal([]byte(resp), &barRaw); err != nil {
			log.Printf("error unmarshalling response: %v", err)
			return nil, err
		}
		o, err := strconv.ParseFloat(barRaw.Open[0], 64)
		if err != nil {
			log.Printf("error parsing open: %v", err)
			return nil, err
		}
		h, err := strconv.ParseFloat(barRaw.High[0], 64)
		if err != nil {
			log.Printf("error parsing high: %v", err)
			return nil, err
		}
		l, err := strconv.ParseFloat(barRaw.Low[0], 64)
		if err != nil {
			log.Printf("error parsing low: %v", err)
			return nil, err
		}
		c, err := strconv.ParseFloat(barRaw.Close[0], 64)
		if err != nil {
			log.Printf("error parsing close: %v", err)
			return nil, err
		}
		v, err := strconv.ParseFloat(barRaw.Volume[0], 64)
		if err != nil {
			log.Printf("error parsing volume: %v", err)
			return nil, err
		}
		i, err := strconv.ParseInt(barRaw.Time[0], 10, 64)
		if err != nil {
			log.Printf("error parsing time: %v", err)
			return nil, err
		}
		t := time.Unix(i, 0)
		return core.NewBasicBar(t, o, h, l, c, c, int64(v), core.REALTIME), nil
	}
}

func sendRequest(reqUrl string) (string, error) {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://quote.fx678.com")
	req.Header.Set("Referer", "https://quote.fx678.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.134 Safari/537.36 Edg/103.0.1264.71")
	req.Header.Set("Sec-Ch-Ua", "\".Not/A)Brand\";v=\"99\", \"Microsoft Edge\";v=\"103\", \"Chromium\";v=\"103\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")

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

func (f *fx678DataProvider) init(instrument string, freqList []core.Frequency) error {
	if _, ok := symbolMap[instrument]; !ok {
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

func (f *fx678DataProvider) connect() error {
	return nil
}

func (f *fx678DataProvider) nextBars() (map[string]core.Bar, error) {
	if f.stopped {
		return nil, fmt.Errorf("fx678 data provider is stopped")
	}
	time.Sleep(SleepDuration)
	basicBar, err := getABar(f.instrument)
	if err != nil {
		logger.Logger.Warn("error getting a bar: %v", zap.Error(err))
		return nil, err
	}
	return map[string]core.Bar{f.instrument: basicBar}, nil
}

func (f *fx678DataProvider) reset() error {
	return nil
}

func (f *fx678DataProvider) stop() error {
	f.stopped = true
	return nil
}

func (f *fx678DataProvider) datatype() series.Type {
	return series.Float
}

func NewFx678DataProvider() BarDataProvider {
	return &fx678DataProvider{
		stopped: false,
	}
}
