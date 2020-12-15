/*
=======================================================================
Copyright (c) 2017-2020 Aarna Networks, Inc.
All rights reserved.
======================================================================
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
          http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================================================================
*/

package authproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type AuthProxy struct {
	AuthProxyConf AuthProxyConfig
}

// AuthProxyConfig holds inputs of authproxy
type AuthProxyConfig struct {
	Issuer      string `json:"issuer"`
	RedirectURI string `json:"redirect_uri"`
	ClientID    string `json:"client_id"`
}

// NewAppHandler interface implementing REST callhandler
func NewAppHandler() *AuthProxy {
	return &AuthProxy{}
}

// OpenIDConfiguration struct to map response from OIDC
type OpenIDConfiguration struct {
	Issuer                string `json:"issuer"`
	AuthzEndpoint         string `json:"authorization_endpoint"`
	TokenEndPoint         string `json:"token_endpoint"`
	IntrospectionEndpoint string `json:"introspection_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// TokenConfig struct holds tokens
type TokenConfig struct {
	ACCESSTOKEN string `json:"access_token"`
	IDTOKEN     string `json:"id_token"`
}

// RealmConfig struct holds public_key of issuer
type RealmConfig struct {
	PublicKey string `json:"public_key"`
}

var openIDConfig *OpenIDConfiguration
var realmConfig *RealmConfig

// Loads the openIDconfig from the OIDC only once
func getOpenIDConfig(issuer string) OpenIDConfiguration {
	if openIDConfig != nil {
		log.Println("openidconfig is not null and returning the cached value")
		return *openIDConfig
	}
	log.Println("openidconfig is null and loading the values")
	url := issuer + ".well-known/openid-configuration"
	response, err := http.Get(url)
	if err != nil {
		log.Printf("The openidconfig HTTP request failed with error %s", err)
		return *openIDConfig
	}

	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(bodyBytes, &openIDConfig)
	return *openIDConfig
}

// LoginHandler redirects to client login page and sets cookie with the original path
func (h AuthProxy) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler start")

	rd := r.FormValue("rd")
	scope := r.FormValue("scope")

	log.Printf("[LoginHandler] url Param 'rd' is: %s, 'scope' is: %s\n", string(rd), string(scope))
	redirect := r.Header.Get("X-Auth-Request-Redirect")
	log.Println("redirect url from HEADER is: " + redirect)
	if len(redirect) == 0 {
		redirect = rd
	}

	cookie := http.Cookie{
		Name:     "org",
		Value:    redirect,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: false,
	}
	// Set cookie with original URL
	http.SetCookie(w, &cookie)
	state := "1234" // Optional parameter included in all login redirects
	if len(scope) == 0 {
		// generate token with offline_access scope so that it can be stored in cookie and reused
		scope = "openid offline_access"
	}

	// get authorization endpoint from function openidconfig
	authzEndpoint := getOpenIDConfig(h.AuthProxyConf.Issuer).AuthzEndpoint

	// Construct redirect URL with params
	u, _ := url.Parse(authzEndpoint)
	q := u.Query()
	q.Add("client_id", h.AuthProxyConf.ClientID)
	// h.AuthProxyConf.RedirectURI is the callback endpoint of middleend.
	// after successful authentication, url will be redirected to this one
	q.Add("redirect_uri", h.AuthProxyConf.RedirectURI)
	q.Add("response_type", "code")
	q.Add("scope", scope)
	q.Add("state", state)
	u.RawQuery = q.Encode()

	log.Println("[LoginHandler] Redireced URL -> " + u.String())
	http.Redirect(w, r, u.String(), http.StatusFound)
}

/*
 * CallbackHandler reads the OIDC config
 * Gets token with API and sets id and access tokens in cookies
 * Redirects to original URL
 */
func (h AuthProxy) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	code := r.FormValue("code")
	tokenEndpoint := getOpenIDConfig(h.AuthProxyConf.Issuer).TokenEndPoint
	log.Printf("[CallbackHandler] state: %s , code: %s , tokenEndpoint: %s \n", state, code, tokenEndpoint)

	client := http.Client{}
	form := url.Values{}
	form.Add("client_id", h.AuthProxyConf.ClientID)
	form.Add("client_secret", "")
	form.Add("grant_type", "authorization_code")
	form.Add("code", code)
	form.Add("redirect_uri", h.AuthProxyConf.RedirectURI)
	request, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(request)
	if err != nil {
		log.Printf("[CallbackHandler] HTTP request to %s failed\n", tokenEndpoint)
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[CallbackHandler] Error while reading response from tokenEndpoint")
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var tokenConfig TokenConfig
	json.Unmarshal(body, &tokenConfig)
	log.Printf("[CallbackHandler] access_token: %s \n id_token: %s\n", tokenConfig.ACCESSTOKEN, tokenConfig.IDTOKEN)

	// Construct the original URL with cookie org
	var orginalURL string
	cookie, err := r.Cookie("org")
	if err == nil {
		orginalURL = cookie.Value
	}
	fmt.Println("[CallbackHandler] orginalURL from cookie: " + orginalURL)

	// Create cookies with id_token, access_token
	idTokenCookie := http.Cookie{
		Name:     "idtoken",
		Value:    tokenConfig.IDTOKEN,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: false,
	}
	accessTokencookie := http.Cookie{
		Name:     "accesstoken",
		Value:    tokenConfig.ACCESSTOKEN,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: false,
	}

	http.SetCookie(w, &idTokenCookie)
	http.SetCookie(w, &accessTokencookie)

	// Finally return the original URL with the cookies
	http.Redirect(w, r, orginalURL, http.StatusFound)
}

// AuthHandler verifies the token and returns response
func (h AuthProxy) AuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[AuthHandler] Authenticating the token")

	var idToken string
	cookie, err := r.Cookie("idtoken")
	if err == nil {
		cookieVal := cookie.Value
		idToken = cookieVal
	}

	if idToken == "" {
		log.Println("[AuthHandler] id token is nil ")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	error := validateToken(h.AuthProxyConf.Issuer, idToken)
	if error != nil {
		log.Println("[AuthHandler] Issue with token and returning failed response")
		w.WriteHeader(http.StatusUnauthorized)
	}
}

/*
* Validates JWT token
* verifies signature, token expiry and invalid check... etc
 */
func validateToken(issuer string, reqToken string) error {
	log.Printf("[AuthHandler] Validating JWT token: \n%s\n", reqToken)

	//load realm public key only once
	if realmConfig == nil {
		log.Println("[AuthHandler] realmconfig is null and loading the value")
		response, err := http.Get(issuer)
		if err != nil {
			log.Printf("[AuthHandler] Error while retreiving issuer details : %s\n", err)
			return err
		}
		defer response.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(bodyBytes, &realmConfig)
	}
	SecretKey := "-----BEGIN CERTIFICATE-----\n" + realmConfig.PublicKey + "\n-----END CERTIFICATE-----"
	key, er := jwt.ParseRSAPublicKeyFromPEM([]byte(SecretKey))
	if er != nil {
		log.Println("[AuthHandler] Error occured while parsing public key")
		log.Println(er)
		return er
	}

	token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("[AuthHandler] Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		log.Println("[AuthHandler] Error while parsing token")
		log.Println(err)
		return err
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("[AuthHandler] Token is valid")
	}
	return nil
}
