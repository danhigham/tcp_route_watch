package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"./route_listener"
)

// TokenResponse
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string `json:"scope"`
	Jti         string `json:"jti"`
}

func main() {
	clientID := "tcp_router"
	clientSecret := "tcp-router-secret"
	uaaEndpoint := "uaa.local.pcfdev.io"
	apiEndpoint := "api.local.pcfdev.io"

	token, err := getToken(uaaEndpoint, clientID, clientSecret)
	if err != nil {
		panic(err)
	}

	eventResponse, err := getRouteEvents(apiEndpoint, token)
	if err != nil {
		fmt.Print("ERROR!")
		panic(err)
	}

	ch := make(chan routelistener.RouteUpdate)
	rl := routelistener.New(&eventResponse)
	go rl.Listen(ch)

	for {
		ru := <-ch
		fmt.Printf("\n%+v\n", ru)
	}
}

func getToken(uaa string, clientID string, clientSecret string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	body := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials&response_type=token", clientID, clientSecret)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/oauth/token", uaa), strings.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(bytes.NewReader(responseBody))
	var tokenResponse TokenResponse
	dec.Decode(&tokenResponse)

	return tokenResponse.AccessToken, nil
}

func getRouteEvents(api string, token string) (http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		// IdleConnTimeout: 90 * time.Second,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/routing/v1/tcp_routes/events", api), nil)
	if err != nil {
		return http.Response{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)

	if err != nil {
		return http.Response{}, err
	}

	return *resp, nil
}
