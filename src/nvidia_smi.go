package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
)

/*
SKIPPED TAGS:
	<mig_mode>
	<mig_devices>
	<ecc_mode>
	<ecc_errors>
	<retired_pages>
	<remapped_rows>
	<applications_clocks>
	<default_applications_clocks>
	<max_customer_boost_clocks>
	<supported_clocks>
	<accounted_processes>
*/

type NvidiaSmiOutput struct {
	DriverVersion string `xml:"driver_version"`
	CudaVersion   string `xml:"cuda_version"`
	AttachedGPUs  string `xml:"attached_gpus"`
	GPU           []struct {
		Id                       string `xml:"id,attr"`
		ProductName              string `xml:"product_name"`
		ProductBrand             string `xml:"product_brand"`
		ProductArchitecture      string `xml:"product_architecture"`
		DisplayMode              string `xml:"display_mode"`
		DisplayActive            string `xml:"display_active"`
		PersistenceMode          string `xml:"persistence_mode"`
		AccountingMode           string `xml:"accounting_mode"`
		AccountingModeBufferSize string `xml:"accounting_mode_buffer_size"`
		DriverModel              struct {
			CurrentDM string `xml:"current_dm"`
			PendingDM string `xml:"pending_dm"`
		} `xml:"driver_model"`
		Serial         string `xml:"serial"`
		UUID           string `xml:"uuid"`
		MinorNumber    string `xml:"minor_number"`
		VbiosVersion   string `xml:"vbios_version"`
		MultiGPUBoard  string `xml:"multigpu_board"`
		BoardId        string `xml:"board_id"`
		GPUPartNumber  string `xml:"gpu_part_number"`
		InfoRomVersion struct {
			ImgVersion string `xml:"img_version"`
			OemObject  string `xml:"oem_object"`
			EccObject  string `xml:"ecc_object"`
			PwrObject  string `xml:"pwr_object"`
		} `xml:"inforom_version"`
		GPUOperationMode struct {
			Current string `xml:"current_gom"`
			Pending string `xml:"pending_gom"`
		} `xml:"gpu_operation_mode"`
		GPUVirtualizationMode struct {
			VirtualizationMode string `xml:"virtualization_mode"`
			HostVGPUMode       string `xml:"host_vgpu_mode"`
		} `xml:"gpu_virtualization_mode"`
		IBMNPU struct {
			RelaxedOrderingMode string `xml:"relaxed_ordering_mode"`
		} `xml:"ibmnpu"`
		PCI struct {
			Bus         string `xml:"pci_bus"`
			Device      string `xml:"pci_device"`
			Domain      string `xml:"pci_domain"`
			DeviceId    string `xml:"pci_device_id"`
			BusId       string `xml:"pci_bus_id"`
			SubSystemId string `xml:"pci_sub_system_id"`
			GPULinkInfo struct {
				PCIeGen struct {
					Max     string `xml:"max_link_gen"`
					Current string `xml:"current_link_gen"`
				} `xml:"pcie_gen"`
				LinkWidth struct {
					Max     string `xml:"max_link_width"`
					Current string `xml:"current_link_width"`
				} `xml:"link_widths"`
			} `xml:"pci_gpu_link_info"`
			BridgeChip struct {
				Type string `xml:"bridge_chip_type"`
				Fw   string `xml:"bridge_chip_fw"`
			} `xml:"pci_bridge_chip"`
			ReplayCounter         string `xml:"replay_counter"`
			ReplayRolloverCounter string `xml:"replay_rollover_counter"`
			TxUtil                string `xml:"tx_util"`
			RxUtil                string `xml:"rx_util"`
		} `xml:"pci"`
		FanSpeed             string `xml:"fan_speed"`
		PerformanceState     string `xml:"performance_state"`
		ClockThrottleReasons struct {
			ClockThrottleReasonGPUIdle                   string `xml:"clocks_throttle_reason_gpu_idle"`
			ClockThrottleReasonApplicationsClocksSetting string `xml:"clocks_throttle_reason_applications_clocks_setting"`
			ClockThrottleReasonSWPowerCap                string `xml:"clocks_throttle_reason_sw_power_cap"`
			ClockThrottleReasonHWSlowdown                string `xml:"clocks_throttle_reason_hw_slowdown"`
			ClockThrottleReasonHWThermalSlowdown         string `xml:"clocks_throttle_reason_hw_thermal_slowdown"`
			ClockThrottleReasonHWPowerBrakeSlowdown      string `xml:"clocks_throttle_reason_hw_power_brake_slowdown"`
			ClockThrottleReasonSyncBoost                 string `xml:"clocks_throttle_reason_sync_boost"`
			ClockThrottleReasonSWThermalSlowdown         string `xml:"clocks_throttle_reason_sw_thermal_slowdown"`
			ClockThrottleReasonDisplayClocksSetting      string `xml:"clocks_throttle_reason_display_clocks_setting"`
		} `xml:"clocks_throttle_reasons"`
		FbMemoryUsage struct {
			Total string `xml:"total"`
			Used  string `xml:"used"`
			Free  string `xml:"free"`
		} `xml:"fb_memory_usage"`
		Bar1MemoryUsage struct {
			Total string `xml:"total"`
			Used  string `xml:"used"`
			Free  string `xml:"free"`
		} `xml:"bar1_memory_usage"`
		ComputeMode string `xml:"compute_mode"`
		Utilization struct {
			GPUUtil     string `xml:"gpu_util"`
			MemoryUtil  string `xml:"memory_util"`
			EncoderUtil string `xml:"encoder_util"`
			DecoderUtil string `xml:"decoder_util"`
		} `xml:"utilization"`
		EncoderStats struct {
			SessionCount   string `xml:"session_count"`
			AverageFPS     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
		} `xml:"encoder_stats"`
		FBCStats struct {
			SessionCount   string `xml:"session_count"`
			AverageFPS     string `xml:"average_fps"`
			AverageLatency string `xml:"average_latency"`
		} `xml:"fbc_stats"`
		Temperature struct {
			GPUTemp                string `xml:"gpu_temp"`
			GPUTempMaxThreshold    string `xml:"gpu_temp_max_threshold"`
			GPUTempSlowThreshold   string `xml:"gpu_temp_slow_threshold"`
			GPUTempMaxGpuThreshold string `xml:"gpu_temp_max_gpu_threshold"`
			GPUTargetTemperature   string `xml:"gpu_target_temperature"`
			MemoryTemp             string `xml:"memory_temp"`
			GPUTempMaxMemThreshold string `xml:"gpu_temp_max_mem_threshold"`
		} `xml:"temperature"`
		SupportedGPUTargetTemp struct {
			GPUTargetTempMin string `xml:"gpu_target_temp_min"`
			GPUTargetTempMax string `xml:"gpu_target_temp_max"`
		}
		PowerReadings struct { // backwards compatibility
			PowerState         string `xml:"power_state"`
			PowerManagement    string `xml:"power_management"`
			PowerDraw          string `xml:"power_draw"`
			PowerLimit         string `xml:"power_limit"`
			DefaultPowerLimit  string `xml:"default_power_limit"`
			EnforcedPowerLimit string `xml:"enforced_power_limit"`
			MinPowerLimit      string `xml:"min_power_limit"`
			MaxPowerLimit      string `xml:"max_power_limit"`
		} `xml:"power_readings"`
		GPUPowerReadings struct {
			PowerState          string `xml:"power_state"`
			PowerDraw           string `xml:"power_draw"`
			CurrentPowerLimit   string `xml:"current_power_limit"`
			RequestedPowerLimit string `xml:"requested_power_limit"`
			DefaultPowerLimit   string `xml:"default_power_limit"`
			MinPowerLimit       string `xml:"min_power_limit"`
			MaxPowerLimit       string `xml:"max_power_limit"`
		} `xml:"gpu_power_readings"`
		Clocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"clocks"`
		MaxClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"max_clocks"`
		ClockPolicy struct {
			AutoBoost        string `xml:"auto_boost"`
			AutoBoostDefault string `xml:"auto_boost_default"`
		} `xml:"clock_policy"`
		Processes struct {
			ProcessInfo []struct {
				Pid         int64  `xml:"pid"`
				Type        string `xml:"type"`
				ProcessName string `xml:"process_name"`
				UsedMemory  string `xml:"used_memory"`
			} `xml:"process_info"`
		} `xml:"processes"`
	} `xml:"gpu"`
}

func readNvidiaSmiOutput() (NvidiaSmiOutput, error) {
	var t NvidiaSmiOutput
	var stdout []byte
	var err error

	if *testFile != "" {
		// read test file
		stdout, err = ioutil.ReadFile(*testFile)
	} else {
		// execute system command
		cmd := exec.Command(*nvidiaSmiPath, "-q", "-x")
		stdout, err = cmd.Output()
	}
	if err != nil {
		return t, err
	}

	// parse XML
	if err := xml.Unmarshal(stdout, &t); err != nil {
		return t, fmt.Errorf("error parsing nvidia-smi output: %v", err)
	}

	return t, nil
}

func filterVersion(value string) string {
	r := regexp.MustCompile(`(?P<version>\d+\.\d+).*`)
	match := r.FindStringSubmatch(value)
	version := "0"
	if len(match) > 0 {
		version = match[1]
	}
	return version
}

func filterUnit(s string) string {
	r := regexp.MustCompile(`(?P<value>[\d\.]+) (?P<power>[KMGT]?[i]?)(?P<unit>.*)`)
	match := r.FindStringSubmatch(s)
	if len(match) == 0 {
		return "0"
	}

	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	power := result["power"]
	if value, err := strconv.ParseFloat(result["value"], 32); err == nil {
		switch power {
		case "K":
			value *= 1000
		case "M":
			value *= 1000 * 1000
		case "G":
			value *= 1000 * 1000 * 1000
		case "T":
			value *= 1000 * 1000 * 1000 * 1000
		case "Ki":
			value *= 1024
		case "Mi":
			value *= 1024 * 1024
		case "Gi":
			value *= 1024 * 1024 * 1024
		case "Ti":
			value *= 1024 * 1024 * 1024 * 1024
		}
		return fmt.Sprintf("%g", value)
	}
	return "0"
}

func filterNumber(value string) string {
	r := regexp.MustCompile("[^0-9.]")
	return r.ReplaceAllString(value, "")
}

func filterActive(value string) string {
	if value == "Active" {
		return "1"
	}
	return "0"
}
