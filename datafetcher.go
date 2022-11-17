package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

	return &DataFetcher{
		host:   host,
		port:   port,
		client: &c,
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
	req := df.constructRequest(endpointVersion)
	// Remove empty auth-token
	req.Header.Del("Authorization")

	// Add the one we want to test
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
