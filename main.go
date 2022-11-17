package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

const (
	endpointStatus = "/agent/status"
)

type IStatusConfig struct {
	statusUri     string
	authToken     string
	statusPort    int
	telemetryPort int
}

type IStatusApp struct {
	app    *tview.Application
	config *IStatusConfig
}

func (is *IStatusApp) fetchStatusJson() map[string]interface{} {
	var result map[string]interface{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// TODO shorten this, take into account the refresh cycle
	c := http.Client{Timeout: time.Duration(1) * time.Second, Transport: tr}

	endpoint := fmt.Sprintf("https://%s:%d/%s", is.config.statusUri, is.config.statusPort, endpointStatus)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", is.config.authToken))

	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}
	fmt.Println(result)

	return result

}

func (is *IStatusApp) getStatusText() *tview.TextView {

	// TODO
	tv := tview.NewTextView().SetChangedFunc(func() {
		is.app.Draw()
	})

	statusObj := is.fetchStatusJson()
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

func main() {
	authToken := "fadca502938f8f0c71a4bd20e84d99093f5227de1d87ee83a110297720f022c9" // READ FROM DISK
	config := IStatusConfig{
		statusUri:     "localhost",
		statusPort:    5001,
		authToken:     authToken,
		telemetryPort: 5000,
	}

	app := IStatusApp{
		app:    tview.NewApplication(),
		config: &config,
	}
	app.Run()
}
