package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Gddr6Output []struct {
	GpuId string `json:"pci_id"`
	Temp  int    `json:"temp"`
}

func getGddr6Temperatures() (map[string]int, error) {
	if _, err := os.Stat(*gddr6Path); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	var t Gddr6Output
	var stdout []byte
	var err error

	cmd := exec.Command(*gddr6Path, "-j")
	stdout, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(stdout, &t); err != nil {
		return nil, fmt.Errorf("error parsing gddr6 output: %v", err)
	}

	result := make(map[string]int)
	for _, item := range t {
		result[item.GpuId] = item.Temp
	}
	return result, nil
}
