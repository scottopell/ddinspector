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
	"time"

	"github.com/rivo/tview"
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

func (df *DataFetcher) statusJson() map[string]interface{} {
	var result map[string]interface{}

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

type IStatusConfig struct {
	statusUri     string
	authToken     string
	statusPort    int
	telemetryPort int
}

type IStatusApp struct {
	app    *tview.Application
	client *http.Client
	df     *DataFetcher
}

func (is *IStatusApp) getStatusText() *tview.TextView {
	tv := tview.NewTextView().SetChangedFunc(func() {
		is.app.Draw()
	})

	statusObj := is.df.statusJson()
	fmt.Fprintf(tv, statusObj["version"].(string))

	return tv
}

func (is *IStatusApp) getRoot() *tview.Flex {
	flex := tview.NewFlex().
		AddItem(is.getStatusText(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Top"), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right (20 cols)"), 20, 1, false)
	return flex
}

func (is *IStatusApp) Run() {
	root := is.getRoot()
	if err := is.app.SetRoot(root, true).SetFocus(root).Run(); err != nil {
		panic(err)
	}
}

func getAuthToken(tester func(string) bool) (string, error) {
	// Start with list of common locations
	// TODO - would be cool to have some "autodiscovery" here based on the currently running agent
	locations := [...]string{
		"/Users/scott.opell/go/src/github.com/DataDog/datadog-agent/bin/agent/dist/auth_token",
		"/etc/datadog-agent",
		"/opt/datadog-agent/etc/auth_token",
	}

	for _, loc := range locations {
		auth_token, err := os.ReadFile(loc)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		}
		s := string(auth_token)
		if tester(s) {
			return s, nil
		} else {
			continue
		}
	}

	return "", errors.New("No auth locations passed")
}

func main() {
	df := NewDataFetcher("localhost", 5001)

	// TODO allow specifying auth_token via cli/env-var
	authToken, err := getAuthToken(df.testAuthToken)
	if err != nil {
		panic(err)
	}

	df.AuthToken = authToken

	app := IStatusApp{
		app: tview.NewApplication(),
		df:  df,
	}

	app.Run()
}
