package sso

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/thoas/go-funk"
)

const (
	oauthURL       = "https://open.tiktokapis.com/v2/oauth/token/"
	userInfoURL    = "https://open.tiktokapis.com/v2/user/info/"
	oauthGrantType = "authorization_code"
	userInfoField  = "open_id,union_id,avatar_url,display_name,username"
)

// TiktokConfig ...
type TiktokConfig struct {
	ClientSecret string
	ClientKey    string
	RedirectURI  string
}

// ResponseTiktokGetInfo ...
type ResponseTiktokGetInfo struct {
	Data struct {
		User struct {
			ID          string `json:"open_id"`
			Avatar      string `json:"avatar_url"`
			DisplayName string `json:"display_name"`
			Username    string `json:"username"`
		} `json:"user"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		LogId   string `json:"log_id"`
	} `json:"error"`
}

// CommonTiktokOauth2Data ...
type CommonTiktokOauth2Data struct {
	ID                    string    `json:"sub"`
	Name                  string    `json:"name"`
	Scope                 string    `json:"scope"`
	Photo                 string    `json:"photo"`
	Username              string    `json:"username"`
	Token                 string    `json:"token"`
	ExpiredAt             time.Time `json:"expiredAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiredAt time.Time `json:"refreshTokenExpiredAt"`
}

// TiktokGetUserInfoByCode ...
func TiktokGetUserInfoByCode(config TiktokConfig, code, redirectURI string) (*CommonTiktokOauth2Data, error) {
	accessToken, err := tiktokGetAccessToken(config, code, redirectURI)
	if err != nil {
		fmt.Println("Error GetAccessToken: ", err.Error())
		return nil, err
	}
	scopes := strings.Split(accessToken.Scope, ",")
	if !funk.Contains(scopes, "user.info.profile") {
		return nil, errors.New(fmt.Sprintf("Bạn cần cung cấp đủ các quyền mà ứng dụng yêu cầu"))
	}

	client := resty.New()

	resp, err := client.R().
		SetDebug(true).
		Get(userInfoURL + "?fields=" + userInfoField)
	if err != nil {
		return nil, err
	}
	var (
		resInfo ResponseTiktokGetInfo
	)
	_ = json.Unmarshal(resp.Body(), &resInfo)
	if resInfo.Error.Code != "ok" {
		return nil, errors.New(resInfo.Error.Message)
	}
	return &CommonTiktokOauth2Data{
		ID:                    resInfo.Data.User.ID,
		Scope:                 accessToken.Scope,
		Name:                  resInfo.Data.User.DisplayName,
		Photo:                 resInfo.Data.User.Avatar,
		Username:              resInfo.Data.User.Username,
		Token:                 accessToken.AccessToken,
		ExpiredAt:             time.Now().Add(time.Duration(accessToken.ExpiresIn) * time.Second),
		RefreshToken:          accessToken.RefreshToken,
		RefreshTokenExpiredAt: time.Now().Add(time.Duration(accessToken.RefreshExpiresIn) * time.Second),
	}, nil
}

type TiktokAccessToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	OpenID           string `json:"open_id"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	LogId            string `json:"log_id"`
}

func tiktokGetAccessToken(config TiktokConfig, code, redirectURI string) (TiktokAccessToken, error) {
	var (
		res TiktokAccessToken
	)
	client := resty.New()
	body := map[string]string{
		"client_key":    config.ClientKey,
		"client_secret": config.ClientSecret,
		"grant_type":    oauthGrantType,
		"redirect_uri":  redirectURI,
		"code":          code,
	}
	resp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}).
		SetFormData(body).
		SetDebug(true).
		Post(oauthURL)
	if err != nil {
		return TiktokAccessToken{}, err
	}
	err = json.Unmarshal(resp.Body(), &res)
	if err != nil {
		return TiktokAccessToken{}, err
	}
	return res, nil
}
