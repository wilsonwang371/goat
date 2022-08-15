package feedgen

import (
	"encoding/json"
	"fmt"
	"goalgotrade/pkg/core"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	lg "goalgotrade/pkg/logger"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"

	"github.com/imroc/req"
	"github.com/recws-org/recws"
)

// TradingViewSessionStringLength ...
const (
	TradingViewSessionStringLength = 12
	TradingViewSignInUrl           = "https://www.tradingview.com/accounts/signin/"
	TradingViewWebSocketUrl        = "wss://data.tradingview.com/socket.io/websocket"
)

var frequencyTable map[core.Frequency]string

func init() {
	frequencyTable = map[core.Frequency]string{
		core.MINUTE: "1",
		core.HOUR:   "60",
		core.DAY:    "1D",
	}
}

// GetAuthToken ...
func GetAuthToken(username, password string) (string, error) {
	headers := req.Header{
		"authority":  "www.tradingview.com",
		"user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
		"origin":     "https://www.tradingview.com",
		"referer":    "https://www.tradingview.com/",
	}
	param := req.Param{
		"username": username,
		"password": password,
		"remember": "on",
	}

	resp, err := req.Post(TradingViewSignInUrl, headers, param)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{}
	err = resp.ToJSON(&result)
	if err != nil {
		return "", fmt.Errorf("convert to json format failed. %v. response: %v", err, resp)
	}
	if userRaw, ok := result["user"]; ok {
		if user, ok := userRaw.(map[string]interface{}); !ok {
			return "", fmt.Errorf("invalid user argument. result: %v", result)
		} else {
			if auth, ok := user["auth_token"]; ok {
				if auth, ok := auth.(string); ok {
					return auth, nil
				}
				return "", fmt.Errorf("invalid auth argument. result: %v", result)
			}
		}
	}
	return "", fmt.Errorf("invalid response data. result: %v", result)
}

type tradingViewWSDataProvider struct {
	ws               recws.RecConn
	username         string
	password         string
	querySessionName string
	chatSessionName  string
	authToken        string
	instrument       string
	freqList         []core.Frequency

	barC chan core.Bar
}

// NewTradingViewDataProvider ...
func NewTradingViewDataProvider(username, password string) BarDataProvider {
	res := &tradingViewWSDataProvider{
		querySessionName: TVGenQuerySession(),
		chatSessionName:  TVGenChatSession(),
		username:         username,
		password:         password,
		authToken:        "",
		barC:             make(chan core.Bar, 1024),
	}
	return res
}

// TVGenSession ...
func TVGenSession() string {
	res := [TradingViewSessionStringLength]byte{}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < TradingViewSessionStringLength; i++ {
		res[i] = byte('a' + rand.Intn(26))
	}
	return string(res[:])
}

// TVGenQuerySession ...
func TVGenQuerySession() string {
	res := TVGenSession()
	return "qs_" + res
}

// TVGenChatSession ...
func TVGenChatSession() string {
	res := TVGenSession()
	return "cs_" + res
}

// TVBuildMsgHdr ...
func TVBuildMsgHdr(msg []byte) []byte {
	return []byte(fmt.Sprintf("~m~%d~m~", len(msg)))
}

// TVBuildMsgBody ...
func TVBuildMsgBody(methodName string, paramList []interface{}) ([]byte, error) {
	if params, err := json.Marshal(map[string]interface{}{
		"m": methodName,
		"p": paramList,
	}); err == nil {
		return params, nil
	}
	return nil, fmt.Errorf("failed to marshal message body")
}

// TVBuildMsg ...
func TVBuildMsg(methodName string, paramList []interface{}) ([]byte, error) {
	if body, err := TVBuildMsgBody(methodName, paramList); err == nil {
		header := TVBuildMsgHdr(body)
		return append(header, body...), nil
	} else {
		return nil, err
	}
}

type dTypeDef struct {
	M string        `json:"m"`
	P []interface{} `json:"p"`
}

type pTypeDef struct {
	S1 struct {
		Lbs struct {
			BarCloseTime int `json:"bar_close_time"`
		} `json:"lbs"`
		Ns struct {
			D       string        `json:"d"`
			Indexes []interface{} `json:"indexes"`
		} `json:"ns"`
		S []struct {
			I int       `json:"i"`
			V []float64 `json:"v"`
		} `json:"s"`
		T string `json:"t"`
	} `json:"s1"`
}

func (t *tradingViewWSDataProvider) tvDataParse(data []byte) ([]core.Bar, error) {
	var res []core.Bar
	parsedData := dTypeDef{}
	freq := core.INVALID
	for _, v := range t.freqList {
		if v != core.REALTIME {
			freq = v
			break
		}
	}
	if freq == core.INVALID {
		lg.Logger.Fatal("invalid frequency")
	}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return nil, err
	}
	if parsedData.M == "du" {
		for _, pvalue := range parsedData.P {
			data2, err := json.Marshal(pvalue)
			if err != nil {
				return nil, fmt.Errorf("invalid data 0")
			}

			parsedInnerData := pTypeDef{}
			if err := json.Unmarshal(data2, &parsedInnerData); err != nil {
				continue
			}
			if len(parsedInnerData.S1.S) == 0 {
				return nil, fmt.Errorf("no data")
			}
			for _, svalue := range parsedInnerData.S1.S {
				// lg.Logger.Debug("parsed data", zap.Any("quote", svalue))
				bar := core.NewBasicBar(time.Unix(int64(svalue.V[0]), 0),
					svalue.V[1], svalue.V[2], svalue.V[3], svalue.V[4], svalue.V[4],
					int64(svalue.V[5]), core.REALTIME)
				res = append(res, bar)
			}
			return res, nil
		}
		lg.Logger.Info("no new data")
	} else if parsedData.M == "timescale_update" {
		for _, pvalue := range parsedData.P {
			data2, err := json.Marshal(pvalue)
			if err != nil {
				return nil, fmt.Errorf("invalid data 0")
			}

			parsedInnerData := pTypeDef{}
			if err := json.Unmarshal(data2, &parsedInnerData); err != nil {
				continue
			}
			if len(parsedInnerData.S1.S) == 0 {
				return nil, fmt.Errorf("no data")
			}
			for _, svalue := range parsedInnerData.S1.S {
				bar := core.NewBasicBar(time.Unix(int64(svalue.V[0]), 0),
					svalue.V[1], svalue.V[2], svalue.V[3], svalue.V[4], svalue.V[4],
					int64(svalue.V[5]), core.REALTIME)
				res = append(res, bar)
			}
			return res, nil
		}
		// TODO: implement me
	} else {
		// lg.Logger.Debug("skip the data we dont care", zap.String("method", parsedData.M), zap.String("data", string(data)))
		return nil, fmt.Errorf("skipped data")
	}
	return nil, fmt.Errorf("invalid data 3")
}

func (t *tradingViewWSDataProvider) sendRawMessage(message []byte) error {
	t.ws.WriteMessage(websocket.TextMessage, message)
	return nil
}

func (t *tradingViewWSDataProvider) sendMessage(methodName string, paramList []interface{}) error {
	if data, err := TVBuildMsg(methodName, paramList); err == nil {
		return t.sendRawMessage(data)
	} else {
		return err
	}
}

func (t *tradingViewWSDataProvider) init(instrument string, freqList []core.Frequency) error {
	count := 0
	for _, freq := range freqList {
		if freq == core.REALTIME {
			continue
		}
		if _, ok := frequencyTable[freq]; !ok {
			return fmt.Errorf("frequency not supported")
		}
		count++
	}
	if count > 1 {
		return fmt.Errorf("too many frequencies")
	}
	t.instrument = instrument
	t.freqList = freqList
	t.reset()
	lg.Logger.Info("tradingview fetcher init", zap.String("instrument", instrument), zap.Any("frequencies", freqList))
	return nil
}

func (t *tradingViewWSDataProvider) setupConnection() error {
	freq := core.INVALID
	for _, v := range t.freqList {
		if v != core.REALTIME {
			freq = v
			break
		}
	}
	if freq == core.INVALID {
		lg.Logger.Fatal("invalid frequency")
	}
	lg.Logger.Info("initialize new connection")
	t.sendMessage("set_auth_token",
		[]interface{}{"unauthorized_user_token"})
	t.sendMessage("chart_create_session",
		[]interface{}{t.chatSessionName, ""})
	t.sendMessage("quote_create_session",
		[]interface{}{t.querySessionName})
	t.sendMessage("quote_set_fields",
		[]interface{}{
			t.querySessionName, "ch", "chp", "current_session", "description",
			"local_description", "language", "exchange", "fractional", "is_tradable", "lp", "lp_time",
			"minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type",
			"update_mode", "volume", "currency_code", "rchp", "rtc",
		})
	t.sendMessage("quote_add_symbols",
		[]interface{}{
			t.querySessionName,
			t.instrument,
			map[string]interface{}{"flags": []string{"force_permission"}},
		})

	t.sendMessage("resolve_symbol",
		[]interface{}{t.chatSessionName, "symbol_1", "={\"symbol\":\"" + t.instrument + "\",\"adjustment\":\"splits\"}"})
	t.sendMessage("create_series",
		[]interface{}{t.chatSessionName, "s1", "s1", "symbol_1", frequencyTable[freq], 300})

	t.sendMessage("quote_fast_symbols",
		[]interface{}{t.querySessionName, t.instrument})

	t.sendMessage("create_study",
		[]interface{}{
			t.chatSessionName, "st1", "st1", "s1", "Volume@tv-basicstudies-118",
			map[string]interface{}{
				"length":         20,
				"col_prev_close": "false",
			},
		})
	t.sendMessage("quote_hibernate_all",
		[]interface{}{t.querySessionName})

	go t.fetchBarsLoop()
	return nil
}

func (t *tradingViewWSDataProvider) connect() error {
	lg.Logger.Info("tradingview fetcher connecting")
	authToken, err := GetAuthToken(t.username, t.password)
	if err != nil {
		return err
	}
	t.authToken = authToken

	headers := http.Header{
		"authority":  []string{"www.tradingview.com"},
		"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"},
		"origin":     []string{"https://data.tradingview.com"},
	}
	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
		SubscribeHandler: t.setupConnection,
	}
	t.ws = ws
	t.ws.Dial(TradingViewWebSocketUrl, headers)

	return nil
}

func (t *tradingViewWSDataProvider) stop() error {
	return t.reset()
}

func (t *tradingViewWSDataProvider) reset() error {
	t.querySessionName = TVGenQuerySession()
	t.chatSessionName = TVGenChatSession()
	if t.ws.IsConnected() {
		t.ws.Close()
	}
	return nil
}

func (t *tradingViewWSDataProvider) datatype() series.Type {
	return series.Float
}

func (t *tradingViewWSDataProvider) nextBars() (map[string]core.Bar, error) {
	tmp := <-t.barC

	res := make(map[string]core.Bar)
	res[t.instrument] = tmp
	return res, nil
}

func (t *tradingViewWSDataProvider) fetchBarsLoop() error {
	r := regexp.MustCompile("~m~\\d+~m~~h~\\d+$")
	r2 := regexp.MustCompile("~m~\\d+~m~")
	for {
		if !t.ws.IsConnected() {
			return fmt.Errorf("got disconnected")
		}
		if msgType, data, err := t.ws.ReadMessage(); err != nil {
			return err
		} else {
			if msgType != websocket.TextMessage {
				return fmt.Errorf("reply data is not text message")
			}
			if r.MatchString(string(data)) {
				// we got a message that we need to echo back
				t.ws.ReadMessage()
				t.ws.WriteMessage(websocket.TextMessage, data)
			} else {
				split := r2.Split(string(data), -1)
				for _, v := range split {
					if len(v) == 0 {
						continue
					}
					barList, _ := t.tvDataParse([]byte(v))
					if len(barList) > 0 {
						for _, bar := range barList {
							t.barC <- bar
						}
					}
				}
			}
		}
	}
}
