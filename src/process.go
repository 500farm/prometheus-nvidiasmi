package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

type DockerInspectOutput []struct {
	Name    string `json:"Name"`
	Created string `json:"Created"`
	Config  struct {
		Image string `json:"Image"`
	} `json:"Config"`
}

type ProcessInfo struct {
	processName      string
	processStartTs   float64
	containerId      string
	containerName    string
	dockerImage      string
	containerStartTs float64
}

func processInfo(pid int64) ProcessInfo {
	var info ProcessInfo

	if t, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		info.processName = t
	}
	info.processStartTs = processStartTimestamp(pid)

	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid)); err == nil {
		info.containerId = string(regexp.MustCompile(`/docker/[0-9a-f]+`).Find(data))
		if info.containerId != "" {
			dockerId := regexp.MustCompile(`[0-9a-f]+$`).FindString(info.containerId)
			cmd := exec.Command("docker", "inspect", dockerId)
			output, err := cmd.Output()
			if err != nil {
				log.Errorln("Command execution error:", err)
			} else {
				var result DockerInspectOutput
				err := json.Unmarshal(output, &result)
				if err != nil {
					log.Errorln("JSON parse error:", err)
				} else if len(output) > 0 {
					info.containerName = strings.TrimLeft(result[0].Name, "/")
					info.dockerImage = result[0].Config.Image
					t, err := time.Parse(time.RFC3339Nano, result[0].Created)
					if err == nil {
						info.containerStartTs = float64(t.UnixNano()) / 1e9
					}
				}
			}
		}
	}

	return info
}

func sysBootTime() int64 {
	if data, err := ioutil.ReadFile("/proc/stat"); err == nil {
		ts, _ := strconv.ParseInt(string(regexp.MustCompile(`btime\s+(\d+)`).FindSubmatch(data)[1]), 10, 64)
		return ts
	}
	return 0
}

var bootTime int64

func processStartTimestamp(pid int64) float64 {
	if bootTime == 0 {
		bootTime = sysBootTime()
	}
	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		ts, _ := strconv.ParseInt(strings.Split(string(data), " ")[21], 10, 64)
		return float64(bootTime) + float64(ts)/100
	}
	return 0
}
