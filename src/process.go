package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/client"
)

var (
	regexDockerCgroup = regexp.MustCompile(`/docker/[0-9a-f]+`)
	regexContainerId  = regexp.MustCompile(`[0-9a-f]+$`)
	regexBootTime     = regexp.MustCompile(`btime\s+(\d+)`)
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
			log.Println("Docker inspect error:", err)
		}
	}
	return info
}

func containerIdForProcess(pid int64) string {
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid)); err == nil {
		cgroupId := string(regexDockerCgroup.Find(data))
		if cgroupId != "" {
			containerId := regexContainerId.FindString(cgroupId)
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
	pinfo.containerName = strings.TrimPrefix(ctJson.Name, "/")
	pinfo.dockerImage = ctJson.Config.Image
	t, err := time.Parse(time.RFC3339Nano, ctJson.State.StartedAt)
	if err == nil {
		pinfo.containerStartTs = float64(t.UnixNano()) / 1e9
	}
	return nil
}

func sysBootTime() int64 {
	if data, err := os.ReadFile("/proc/stat"); err == nil {
		if matches := regexBootTime.FindSubmatch(data); len(matches) > 1 {
			ts, _ := strconv.ParseInt(string(matches[1]), 10, 64)
			return ts
		}
	}
	return 0
}

var bootTime int64

func processStartTimestamp(pid int64) float64 {
	if bootTime == 0 {
		bootTime = sysBootTime()
	}
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
		fields := strings.Split(string(data), " ")
		if len(fields) > 21 {
			ts, _ := strconv.ParseInt(fields[21], 10, 64)
			return float64(bootTime) + float64(ts)/100
		}
	}
	return 0
}
