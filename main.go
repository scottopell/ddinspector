package main

import (
	"errors"
	"io/fs"
	"os"

	"github.com/rivo/tview"
)

const (
	endpointStatus  = "/agent/status"
	endpointVersion = "/agent/version"
)

func getAuthToken(tester func(string) bool) (string, error) {
	// Start with list of common locations
	// TODO - would be cool to have some "autodiscovery" here based on the currently running agent
	// https://github.com/mitchellh/go-ps
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
