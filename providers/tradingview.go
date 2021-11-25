package providers

import (
	"fmt"

	"github.com/imroc/req"
)

const (
	TradingViewSignInUrl = "https://www.tradingview.com/accounts/signin/"
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
