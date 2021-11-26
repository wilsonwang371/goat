package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/imroc/req"
	"github.com/recws-org/recws"
)

const (
	TradingViewSignInUrl    = "https://www.tradingview.com/accounts/signin/"
	TradingViewWebSocketUrl = "wss://data.tradingview.com/socket.io/websocket"
)

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

func TradingViewConnect() {
	headers := http.Header{
		"authority":  []string{"www.tradingview.com"},
		"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"},
		"origin":     []string{"https://data.tradingview.com"},
	}
	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
	}
	ws.Dial(TradingViewWebSocketUrl, headers)

	for {
		if !ws.IsConnected() {
			log.Printf("Websocket disconnected %s", ws.GetURL())
			continue
		}
		log.Printf("connected")
		ws.Close()
		break
	}
}

type TradingViewWSFetcherAdapter struct {
	ws               recws.RecConn
	username         string
	password         string
	querySessionName string
	chatSessionName  string
	authToken        string
}

const TradingViewSessionStringLength = 12

func NewTradingViewWSFetcherAdapter(username, password string) *TradingViewWSFetcherAdapter {
	res := TradingViewWSFetcherAdapter{
		querySessionName: generateQuerySession(),
		chatSessionName:  generateChatSession(),
		username:         username,
		password:         password,
		authToken:        "",
	}
	return &res
}

func generateSession() string {
	res := [TradingViewSessionStringLength]byte{}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < TradingViewSessionStringLength; i++ {
		res[i] = byte('a' + rand.Intn(26))
	}
	return string(res[:])
}

func generateQuerySession() string {
	res := generateSession()
	return "qs_" + res
}

func generateChatSession() string {
	res := generateSession()
	return "cs_" + res
}

func buildMessageHeader(msg []byte) []byte {
	return []byte(fmt.Sprintf("~m~%d~m~", len(msg)))
}

func buildMessageBody(methodName string, paramList []interface{}) ([]byte, error) {
	if params, err := json.Marshal(map[string]interface{}{
		"m": methodName,
		"p": paramList,
	}); err == nil {
		return params, nil
	}
	return nil, fmt.Errorf("failed to marshal message body")
}

func buildMessage(methodName string, paramList []interface{}) ([]byte, error) {
	if body, err := buildMessageBody(methodName, paramList); err == nil {
		header := buildMessageHeader(body)
		return append(header, body...), nil
	} else {
		return nil, err
	}
}

func (t *TradingViewWSFetcherAdapter) sendRawMessage(message []byte) error {
	t.ws.WriteMessage(websocket.TextMessage, message)
	return nil
}

func (t *TradingViewWSFetcherAdapter) sendMessage(methodName string, paramList []interface{}) error {
	if data, err := buildMessage(methodName, paramList); err == nil {
		return t.sendRawMessage(data)
	} else {
		return err
	}
}

func (t *TradingViewWSFetcherAdapter) Connect() error {
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
	}
	ws.Dial(TradingViewWebSocketUrl, headers)
	t.ws = ws

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
			"BINANCE:BTCUSDT",
			map[string]interface{}{"flags": []string{"force_permission"}},
		})

	t.sendMessage("resolve_symbol",
		[]interface{}{t.chatSessionName, "symbol_1", "={\"symbol\":\"BINANCE:BTCUSDT\",\"adjustment\":\"splits\"}"})
	t.sendMessage("create_series",
		[]interface{}{t.chatSessionName, "s1", "s1", "symbol_1", "1", 300})

	t.sendMessage("quote_fast_symbols",
		[]interface{}{t.querySessionName, "BINANCE:BTCUSDT"})

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

	// TODO: implement me
	return nil
}

func (t *TradingViewWSFetcherAdapter) Reset() error {
	t.querySessionName = generateQuerySession()
	t.chatSessionName = generateChatSession()
	return nil
}
