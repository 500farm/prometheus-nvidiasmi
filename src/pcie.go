package main

import (
	"io/ioutil"
	"regexp"
	"strconv"
)

type PcieInfo struct {
	AerFatalCount       int
	AerNonFatalCount    int
	AerCorrectableCount int
}

func pcieInfo(id string) PcieInfo {
	result := PcieInfo{-1, -1, -1}

	path := "/sys/bus/pci/devices/" +
		regexp.MustCompile(`^0000(\d{4}):`).ReplaceAllString(id, "$1") + "/"

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
