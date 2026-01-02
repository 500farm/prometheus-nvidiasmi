package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regex patterns for better performance
var (
	regexPciIdPrefix    = regexp.MustCompile(`^0000(\d{4})`)
	regexAerFatal       = regexp.MustCompile(`TOTAL_ERR_FATAL (\d+)`)
	regexAerNonFatal    = regexp.MustCompile(`TOTAL_ERR_NONFATAL (\d+)`)
	regexAerCorrectable = regexp.MustCompile(`TOTAL_ERR_COR (\d+)`)
	regexLspciField     = regexp.MustCompile(`^([A-Za-z]+):\s+(.+)$`)
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
		strings.ToLower(regexPciIdPrefix.ReplaceAllString(id, "$1")) + "/"

	t, err := os.ReadFile(path + "aer_dev_fatal")
	if err == nil {
		if matches := regexAerFatal.FindSubmatch(t); len(matches) > 1 {
			result.AerFatalCount, _ = strconv.Atoi(string(matches[1]))
		}
	}

	t, err = os.ReadFile(path + "aer_dev_nonfatal")
	if err == nil {
		if matches := regexAerNonFatal.FindSubmatch(t); len(matches) > 1 {
			result.AerNonFatalCount, _ = strconv.Atoi(string(matches[1]))
		}
	}

	t, err = os.ReadFile(path + "aer_dev_correctable")
	if err == nil {
		if matches := regexAerCorrectable.FindSubmatch(t); len(matches) > 1 {
			result.AerCorrectableCount, _ = strconv.Atoi(string(matches[1]))
		}
	}

	return result
}

func initVendorInfo() {
	cmd := exec.Command("/usr/sbin/update-pciids")
	_, err := cmd.Output()
	if err != nil {
		log.Println("Error updating PCI IDs:", err)
	}
}

func vendorInfo(id string) VendorInfo {
	result := VendorInfo{}
	cmd := exec.Command("/usr/bin/lspci", "-vmm", "-s", id)
	out, err := cmd.Output()
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(out), "\n") {
		m := regexLspciField.FindStringSubmatch(line)
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
