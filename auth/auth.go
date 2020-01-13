package auth

import (
	"crypto/rsa"
	"encoding/json"
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
func Logout(ctx *goweb.Context, postLogout func(id_token string)) {
	expire := time.Now().Add(-7 * 24 * time.Hour)
	newCookie := http.Cookie{
		Name:    access_token_cookie_name,
		Value:   "",
		Expires: expire,
	}
	http.SetCookie(ctx.Writer, &newCookie)
	s, err := GetSessionByToken(ctx)
	if err != nil {
		panic(err)
	}
	postLogout(s.token.Extra("id_token").(string))
}

func HasLoggedIn(ctx *goweb.Context) bool {
	_, err := GetSessionByToken(ctx)
	return err == nil
}
func CheckToken(ctx *goweb.Context, introspectTokenURL string) (ok bool, err error) {
	accessToken, err := GetBearerToken(ctx)
	if err != nil {
		session, err := GetSessionByToken(ctx)
		if err != nil {
			return false, err
		}
		accessToken = session.token.AccessToken
	}
	b := common.SendRestApiRequest("GET", accessToken, introspectTokenURL, nil, true)
	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return false, err
	}
	if m["error"] != nil {
		return false, errors.New(m["error"].(string))
	}
	isActive := m["data"].(bool)
	if !isActive {
		return false, errors.New("the token is not valid")
	}
	return true, nil
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
func GetSessionByToken(ctx *goweb.Context) (*session, error) {
	cookie, err := ctx.Request.Cookie(access_token_cookie_name)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(sessions); i++ {
		if sessions[i].token.AccessToken == cookie.Value {
			return &sessions[i], nil
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
