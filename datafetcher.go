package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	endpointStatus         = "/agent/status"
	endpointVersion        = "/agent/version"
	endpointDogstatsdStats = "/agent/dogstatsd-stats"
	endpointConfig         = "/agent/config"
	settingDogstatsStats   = "dogstatsd_stats"
)

// metricStat holds how many times a metric has been
// processed and when was the last time.
// COPY_PASTA from pkg/server/dogstatsd
type MetricStat struct {
	Name     string    `json:"name"`
	Count    uint64    `json:"count"`
	LastSeen time.Time `json:"last_seen"`
	Tags     string    `json:"tags"`
}

type AgentDataFetcher struct {
	client    *http.Client
	AuthToken string
	host      string
	port      int
	agentPid  int32
}

func NewAgentDataFetcher() *AgentDataFetcher {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// TODO Enable timeout here taking into account the refresh interval
	// but disabled for now because when running inside a debugger this will cause a
	// panic when the timout is exceeded
	//c := http.Client{Timeout: time.Duration(1) * time.Second, Transport: tr}
	c := http.Client{Transport: tr}

	df := AgentDataFetcher{
		host:   "localhost",
		port:   5001,
		client: &c,
	}

	runningAgent, err := DiscoverRunningAgent()
	if err != nil {
		log.Print("Failed to find running agent with valid config. Err: ", err)
	} else {
		if !df.testAuthToken(runningAgent.authToken) {
			log.Printf("Running agent's auth-token fetch failed! Something is wrong... falling back to legacy codepath")
			return nil
		} else {
			log.Print("Found Valid Running Agent:", runningAgent)
			df.port = runningAgent.cmdPort
			df.AuthToken = runningAgent.authToken
			df.agentPid = runningAgent.pid
		}
	}

	return &df
}

func (df *AgentDataFetcher) constructGetRequest(endpoint string) *http.Request {
	uri := fmt.Sprintf("https://%s:%d/%s", df.host, df.port, endpoint)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", df.AuthToken))

	return req
}

func (df *AgentDataFetcher) constructPostRequest(endpoint string, contentType string, body io.Reader) *http.Request {
	uri := fmt.Sprintf("https://%s:%d/%s", df.host, df.port, endpoint)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", df.AuthToken))

	return req
}

func (df *AgentDataFetcher) testAuthToken(token string) bool {
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

func (df *AgentDataFetcher) statusJson() map[string]any {
	var result map[string]any

	req := df.constructGetRequest(endpointStatus)
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

func (df *AgentDataFetcher) GetDogstatsdCaptureEnablementValue() bool {
	req := df.constructGetRequest(fmt.Sprintf("%s/%s", endpointConfig, settingDogstatsStats))
	resp, err := df.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Unmarshal failed. Body was: ", string(body))
		panic(err)
	}

	return result["value"].(bool)
}

func (df *AgentDataFetcher) EnableDogstatsdCapture() error {
	reqBody := fmt.Sprintf("value=%s", html.EscapeString("true"))
	req := df.constructPostRequest(fmt.Sprintf("%s/%s", endpointConfig, settingDogstatsStats), "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(reqBody)))
	resp, err := df.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Unmarshal failed. Body was: ", string(body))
		panic(err)
	}

	if newValue, ok := result["value"]; ok {
		if newValue != true {
			return errors.New("unable to enable dogstatsd capture, after sending 'enable' it remained false")
		}
	}

	return nil
}

func (df *AgentDataFetcher) fetchDogstatsdCaptureData() (map[uint64]MetricStat, error) {
	req := df.constructGetRequest(endpointDogstatsdStats)
	resp, err := df.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dogStats map[uint64]MetricStat
	if err := json.Unmarshal(body, &dogStats); err != nil {
		fmt.Println("Body was: ", string(body))
		return nil, err
	}

	return dogStats, nil

}
