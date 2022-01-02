package main

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type AerInfo struct {
	AerFatalCount       int
	AerNonFatalCount    int
	AerCorrectableCount int
}

type VendorInfo struct {
	Vendor       string
	Device       string
	SubsysVendor string
	SubsysDevice string
}

func aerInfo(id string) AerInfo {
	result := AerInfo{-1, -1, -1}

	path := "/sys/bus/pci/devices/" +
		strings.ToLower(regexp.MustCompile(`^0000(\d{4})`).ReplaceAllString(id, "$1")) + "/"

	t, err := ioutil.ReadFile(path + "aer_dev_fatal")
	if err == nil {
		result.AerFatalCount, _ = strconv.Atoi(string(regexp.MustCompile(`TOTAL_ERR_FATAL (\d+)`).FindSubmatch(t)[1]))
	}

	t, err = ioutil.ReadFile(path + "aer_dev_nonfatal")
	if err == nil {
		result.AerNonFatalCount, _ = strconv.Atoi(string(regexp.MustCompile(`TOTAL_ERR_NONFATAL (\d+)`).FindSubmatch(t)[1]))
	}

	t, err = ioutil.ReadFile(path + "aer_dev_correctable")
	if err == nil {
		result.AerCorrectableCount, _ = strconv.Atoi(string(regexp.MustCompile(`TOTAL_ERR_COR (\d+)`).FindSubmatch(t)[1]))
	}

	return result
}

func vendorInfo(id string) VendorInfo {
	result := VendorInfo{}
	cmd := exec.Command("/usr/bin/lspci", "-vmm", "-s", id)
	out, err := cmd.Output()
	if err != nil {
		return result
	}
	re := regexp.MustCompile(`^([A-Za-z]+):\s+(.+)$`)
	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) >= 3 {
			k := m[1]
			v := m[2]
			if k == "Vendor" {
				result.Vendor = v
			} else if k == "Device" {
				result.Device = v
			} else if k == "SVendor" {
				result.SubsysVendor = v
			} else if k == "SDevice" {
				result.SubsysDevice = v
			}
		}
	}
	return result
}
