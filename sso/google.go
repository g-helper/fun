package sso

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// CommonGoogleOauth2Data ...
type CommonGoogleOauth2Data struct {
	ID    string `json:"sub"`
	AUD   string `json:"aud"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Photo string `json:"picture"`
	Token string `json:"token"`
}

const oauthGoogleURLAPI = "https://oauth2.googleapis.com/tokeninfo?id_token="

// GetUserInfoFromGoogle ...
func GetUserInfoFromGoogle(token string, googleClientId string) (result CommonGoogleOauth2Data, err error) {
	client := resty.New()
	resp, err := client.R().
		SetDebug(true).
		Get(oauthGoogleURLAPI)
	if err != nil {
		return
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		fmt.Println("GetUserInfoFromGoogle Error : ", err.Error())
		return result, err
	}
	result.Token = token

	if result.AUD != googleClientId {
		result = CommonGoogleOauth2Data{}
		return
	}

	return
}
