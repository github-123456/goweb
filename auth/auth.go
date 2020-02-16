package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	"github.com/swishcloud/gostudy/common"
	"github.com/swishcloud/gostudy/keygenerator"

	"github.com/swishcloud/goweb"
	"golang.org/x/oauth2"
)

var access_token_cookie_name string

var sessions []session

type session struct {
	token  *oauth2.Token
	Claims map[string]interface{}
	Data   map[string]interface{}
}

func (s *session) GetToken() *oauth2.Token {
	return s.token
}
func Login(ctx *goweb.Context, token *oauth2.Token, jwk_json_url string) *session {
	//todo:mutex.Lock()
	//todo:defer mutex.Unlock()
	session := session{}
	session.token = token
	session.Claims = extractIdTokenCliams(token.Extra("id_token").(string), jwk_json_url)
	session.Data = map[string]interface{}{}
	cookie := http.Cookie{Name: access_token_cookie_name, Value: session.token.AccessToken, Path: "/", Expires: time.Now().Add(7 * 24 * time.Hour)}
	sessions = append(sessions, session)
	http.SetCookie(ctx.Writer, &cookie)
	return &session
}
func Logout(rac *common.RestApiClient, ctx *goweb.Context, conf *oauth2.Config, introspectTokenURL string, skip_tls_verify bool, postLogout func(id_token string)) {
	expire := time.Now().Add(-7 * 24 * time.Hour)
	newCookie := http.Cookie{
		Name:    access_token_cookie_name,
		Value:   "",
		Expires: expire,
	}
	http.SetCookie(ctx.Writer, &newCookie)
	s, err := GetSessionByToken(rac, ctx, conf, introspectTokenURL, skip_tls_verify)
	if err != nil {
		panic(err)
	}
	postLogout(s.token.Extra("id_token").(string))
}

func HasLoggedIn(rac *common.RestApiClient, ctx *goweb.Context, conf *oauth2.Config, introspectTokenURL string, skip_tls_verify bool) bool {
	_, err := GetSessionByToken(rac, ctx, conf, introspectTokenURL, skip_tls_verify)
	return err == nil
}
func CheckToken(rac *common.RestApiClient, conf *oauth2.Config, token *oauth2.Token, introspectTokenURL string, skip_tls_verify bool) (sub string, err error) {
	rar := common.NewRestApiRequest("GET", introspectTokenURL, nil).UseToken(conf, token)
	resp, err := rac.Do(rar)
	if err != nil {
		return "", err
	}
	m, err := common.ReadAsMap(resp.Body)
	if err != nil {
		return "", err
	}
	if m["error"] != nil {
		return "", errors.New(m["error"].(string))
	}
	data := m["data"].(map[string]interface{})
	isActive := data["active"].(bool)
	if !isActive {
		return "", errors.New("the token is not valid")
	}
	return data["sub"].(string), nil
}
func GetBearerToken(ctx *goweb.Context) (string, error) {
	authorization := ctx.Request.Header["Authorization"]
	if len(authorization) == 0 {
		return "", errors.New("not found bearer token")
	}
	if match, _ := regexp.MatchString("Bearer .+", authorization[0]); !match {
		return "", errors.New("not found bearer token")
	}
	token := []rune(authorization[0])
	token = token[7:]
	return string(token), nil
}
func removeSessionAt(index int) {
	sessions = append(sessions[:index], sessions[index+1:]...)
}
func GetSessionByToken(rac *common.RestApiClient, ctx *goweb.Context, conf *oauth2.Config, introspectTokenURL string, skip_tls_verify bool) (*session, error) {
	cookie, err := ctx.Request.Cookie(access_token_cookie_name)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(sessions); i++ {
		s := sessions[i]
		if s.token.AccessToken == cookie.Value {
			_, err := CheckToken(rac, conf, s.token, introspectTokenURL, skip_tls_verify)
			if err != nil {
				removeSessionAt(i)
				return nil, err
			}
			return &s, nil
		}
	}
	return nil, errors.New("not found session")
}
func init() {
	k, err := keygenerator.NewKey(4, false, false, false, true)
	if err != nil {
		panic(err)
	}
	access_token_cookie_name = "access_token_" + k
}

func extractIdTokenCliams(tokenString string, jwk_json_url string) map[string]interface{} {
	jwk, err := jwk.Fetch(jwk_json_url)
	if err != nil {
		panic(err)
	}
	k, err := jwk.Keys[0].Materialize()
	pk := k.(*rsa.PublicKey)
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return pk, nil
	})
	if err != nil {
		panic(err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims
	} else {
		return nil
	}
}
