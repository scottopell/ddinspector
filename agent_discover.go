package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/yaml.v2"
)

type RunningAgentInfo struct {
	cmdPort   int
	authToken string
	pid       int32
}

var AgentModRegex = regexp.MustCompile(`mod\s+github.com\/DataDog\/datadog-agent`)

var StaticAgentConfLocations = []string{
	"/etc/datadog-agent",
	"/opt/datadog-agent/etc",
	"/usr/local/etc/datadog-agent",
	"c:\\programdata\\datadog",
}

func DiscoverRunningAgent() (*RunningAgentInfo, error) {
	// Idea is to run `go version -m /proc/*/exe` and look for
	// `mod     github.com/DataDog/datadog-agent` in the output
	procs, err := process.Processes()
	if err != nil {
		log.Print("Could not fetch list of running processes: ", err)
		return nil, err
	}
	for _, proc := range procs {
		exe, err := proc.Exe()
		if err != nil {
			continue
		}
		// TODO ideally we don't need to shell out to the `go` binary
		// See https://pkg.go.dev/runtime/debug#ReadBuildInfo
		// See https://cs.opensource.google/go/go/+/refs/tags/go1.20.1:src/cmd/go/internal/version/version.go
		goVersionCmd := exec.Command("go", "version", "-m", exe)
		versionOutput, err := goVersionCmd.Output()
		if err != nil {
			// Probably not a go binary if `go version -m ` is failing
			continue
		}
		if AgentModRegex.Match(versionOutput) {
			// Woohoo! This correctly finds the running agent process
			// Next step is to find the config for the process and then the
			// cmd-port and auth-token
			// To find cmd-port and auth-token, we need to define the possible ways to configure them:
			// 1. cmd-line args: `-c config-file.yaml`
			// 2. env var options `DD_CMD_PORT` and `DD_CMD_HOST`
			// 3. Config file at some path
			//     - config file names: "datadog.yaml"
			//     - config file search path: [ '.', DD_CONF_PATH, StaticAgentConfigLocations ]

			// really this is all duplicating existing functionality in the agent
			// and its probably pretty hard to get all the edge cases right
			// (ie what if they specify a command line config, but override CMD_PORT in an env variable)
			// so best-effort is perfectly fine here

			// 1. Search through cmdline
			if info, err := getConfigFromCmdline(proc); err == nil {
				return info, nil
			}
			// 2. Env Vars
			// TODO

			// 3. Config file at known path
			// TODO

			return nil, fmt.Errorf("found running agent (pid %d), but couldn't get config values for it", proc.Pid)
		}
	}

	return nil, fmt.Errorf("no running agent found, maybe lacking permission to inspect")
}

func configPathToRunningAgentInfo(configPath string, pid int32) (*RunningAgentInfo, error) {
	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	configData := make(map[interface{}]interface{})

	err = yaml.Unmarshal(dat, &configData)

	if err != nil {
		return nil, err
	}

	var cmdPortInt int
	var ok bool
	cmdPort := configData["cmd_port"]
	if cmdPort == nil {
		cmdPortInt = 5001
	} else if cmdPortInt, ok = cmdPort.(int); !ok {
		return nil, fmt.Errorf("cmd_port was not empty and not an int")
	}

	authTokenLocation := configData["auth_token_file_path"]
	var authTokenLocationString string
	if authTokenLocation == nil {
		// default to current dir relative to configfile used
		authTokenLocationString = filepath.Join(filepath.Dir(configPath), "auth_token")
	} else if authTokenLocationString, ok = authTokenLocation.(string); !ok {
		return nil, fmt.Errorf("auth_token_file_path was not a string and was not empty")
	}
	authTokenContentsDat, err := ioutil.ReadFile(authTokenLocationString)
	if err != nil {
		return nil, err
	}

	return &RunningAgentInfo{
		cmdPort:   cmdPortInt,
		authToken: string(authTokenContentsDat),
		pid:       pid,
	}, nil
}

func getConfigFromCmdline(proc *process.Process) (*RunningAgentInfo, error) {
	nextArgIsConfigPath := false
	cmdLineSlice, err := proc.CmdlineSlice()
	if err != nil {
		return nil, err
	}

	processCwd, err := proc.Cwd()
	if err != nil {
		return nil, err
	}
	for _, arg := range cmdLineSlice {
		if nextArgIsConfigPath {
			configPath := filepath.Join(processCwd, arg)
			return configPathToRunningAgentInfo(configPath, proc.Pid)
		}
		if arg == "-c" {
			nextArgIsConfigPath = true
		}
	}
	return nil, fmt.Errorf("no config found in command line")
}
