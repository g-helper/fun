package sso

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	fb "github.com/huandu/facebook/v2"
)

// FacebookConfig ...
type FacebookConfig struct {
	ClientID     string
	ClientSecret string
}

// LongLiveAccessToken ...
type LongLiveAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// CommonFacebookGraphMeData ...
type CommonFacebookGraphMeData struct {
	ID        string    `json:"id" bson:"id"`
	Email     string    `json:"email" bson:"email"`
	Name      string    `json:"name" bson:"name"`
	Photo     string    `json:"photo" bson:"photo"`
	Token     string    `json:"token" bson:"token"`
	ExpiredAt time.Time `json:"expiredAt" bson:"expiredAt"`
	Gender    string    `json:"gender" bson:"gender"`
	Picture   struct {
		Data struct {
			URL    string `json:"url" bson:"url"`
			Width  int    `json:"width" bson:"width"`
			Height int    `json:"height" bson:"height"`
		} `json:"data" bson:"data"`
	} `json:"picture" bson:"picture"`
}

// GetUserInfoFromFacebook ...
func GetUserInfoFromFacebook(config FacebookConfig, token string) (result CommonFacebookGraphMeData, err error) {
	var fbParam = fb.Params{
		"access_token": token,
		"fields":       "id,name,email,picture.type(large),gender",
	}

	// Request login facebook
	interfaceResponse, err := fb.Get("/me", fbParam)
	if err != nil {
		return
	}

	jsonString, _ := json.Marshal(interfaceResponse)
	json.Unmarshal(jsonString, &result)

	result.Photo = result.Picture.Data.URL
	result.Token = token

	longLiveToken, expireTime := GetLongLiveAccessTokenFacebook(config, result.Token)
	if longLiveToken == "" {
		return result, errors.New(fmt.Sprintf("Chúng tôi không thể đọc thông tin từ tài khoản của bạn"))
	}
	result.Token = longLiveToken
	result.ExpiredAt = expireTime
	return
}

// GetLongLiveAccessTokenFacebook ...
func GetLongLiveAccessTokenFacebook(config FacebookConfig, token string) (string, time.Time) {
	var (
		res LongLiveAccessToken
	)
	uri := fmt.Sprintf("https://graph.facebook.com/v20.0/oauth/access_token?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
		config.ClientID,
		config.ClientSecret,
		token)
	client := resty.New()
	resp, err := client.R().
		SetDebug(true).
		Get(uri)

	if err != nil || resp.StatusCode() != http.StatusOK {
		return "", time.Time{}
	}
	_ = json.Unmarshal(resp.Body(), &res)
	return res.AccessToken, time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
}
