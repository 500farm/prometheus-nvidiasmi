package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func containerInfo(pid int64) (string, string, string, int64) {
	containerId := ""
	containerName := ""
	dockerImage := ""
	containerCreateTimestamp := int64(0)

	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid)); err == nil {
		containerId = string(regexp.MustCompile(`/docker/[0-9a-f]+`).Find(data))
		if containerId != "" {
			dockerId := regexp.MustCompile(`[0-9a-f]+$`).FindString(containerId)
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
					containerName = strings.TrimLeft(result[0].Name, "/")
					dockerImage = result[0].Config.Image
					t, err := time.Parse(time.RFC3339Nano, result[0].Created)
					if err == nil {
						containerCreateTimestamp = t.Unix()
					}
				}
			}
		}
	}

	return containerId, containerName, dockerImage, containerCreateTimestamp
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
