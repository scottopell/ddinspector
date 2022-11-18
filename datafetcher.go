package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"
)

const (
	endpointStatus  = "/agent/status"
	endpointVersion = "/agent/version"
)

type DataFetcher struct {
	client    *http.Client
	AuthToken string
	host      string
	port      int
}

func NewDataFetcher(host string, port int) *DataFetcher {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// TODO shorten this, take into account the refresh cycle
	c := http.Client{Timeout: time.Duration(1) * time.Second, Transport: tr}

	df := DataFetcher{
		host:   host,
		port:   port,
		client: &c,
	}

	df.setAuthTokenFromEnv()

	return &df
}

func (df *DataFetcher) setAuthTokenFromEnv() {
	// TODO allow specifying auth_token via cli/env-var

	// Start with list of common locations
	// TODO - would be cool to have some "autodiscovery" here based on the currently running agent
	// https://github.com/mitchellh/go-ps
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	authTokenLoc1 := fmt.Sprintf("/Users/%s/go/src/github.com/DataDog/datadog-agent/bin/agent/dist/auth_token", currUser.Username)
	authTokenLoc2 := fmt.Sprintf("/Users/%s/code/datadog-agent/bin/agent/dist/auth_token", currUser.Username)
	locations := [...]string{
		authTokenLoc1,
		authTokenLoc2,
		"/etc/datadog-agent",
		"/opt/datadog-agent/etc/auth_token",
	}

	authToken := ""
	for _, loc := range locations {
		auth_token, err := os.ReadFile(loc)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			} else {
				break
			}
		}
		s := string(auth_token)
		if df.testAuthToken(s) {
			authToken = s
			break
		} else {
			continue
		}
	}

	if authToken == "" {
		panic(errors.New("No valid auth_tokens found!"))
	} else {
		df.AuthToken = authToken
	}
}

func (df *DataFetcher) constructRequest(endpoint string) *http.Request {
	uri := fmt.Sprintf("https://%s:%d/%s", df.host, df.port, endpoint)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", df.AuthToken))

	return req
}

func (df *DataFetcher) testAuthToken(token string) bool {
	uri := fmt.Sprintf("https://%s:%d/%s", df.host, df.port, endpointVersion)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := df.client.Do(req)
	if err != nil {
		log.Printf("Token %q failed with http err: %v", token, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (df *DataFetcher) statusJson() map[string]any {
	var result map[string]any

	req := df.constructRequest(endpointStatus)
	resp, err := df.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Body was: ", string(body))
		panic(err)
	}

	return result
}
