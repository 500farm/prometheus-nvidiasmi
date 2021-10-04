package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/prometheus/common/log"
)

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

	if cid := containerIdForProcess(pid); cid != "" {
		if err := dockerInspect(cid, &info); err != nil {
			log.Errorln("Docker inspect:", err)
		}
	}
	return info
}

func containerIdForProcess(pid int64) string {
	if data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid)); err == nil {
		cgroupId := string(regexp.MustCompile(`/docker/[0-9a-f]+`).Find(data))
		if cgroupId != "" {
			containerId := regexp.MustCompile(`[0-9a-f]+$`).FindString(cgroupId)
			return containerId
		}
	}
	return ""
}

var cli *client.Client

func dockerInspect(cid string, pinfo *ProcessInfo) error {
	if cli == nil {
		var err error
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}
	}
	ctJson, err := cli.ContainerInspect(context.Background(), cid)
	if err != nil {
		return err
	}
	pinfo.containerId = cid
	pinfo.containerName = strings.TrimLeft(ctJson.Name, "/")
	pinfo.dockerImage = ctJson.Config.Image
	t, err := time.Parse(time.RFC3339Nano, ctJson.State.StartedAt)
	if err == nil {
		pinfo.containerStartTs = float64(t.UnixNano()) / 1e9
	}
	return nil
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
