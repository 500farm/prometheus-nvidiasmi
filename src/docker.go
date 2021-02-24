package main

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"github.com/prometheus/common/log"
)

type DockerInspectOutput []struct {
	Name   string `json:"Name"`
	Config struct {
		Image string `json:"Image"`
	} `json:"Config"`
}

func containerInfo(pid string) (string, string, string) {
	containerId := ""
	containerName := ""
	dockerImage := ""

	if data, err := ioutil.ReadFile("/proc/" + pid + "/cgroup"); err == nil {
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
				}
			}
		}
	}

	return containerId, containerName, dockerImage
}
