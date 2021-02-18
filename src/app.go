package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const LISTEN_ADDRESS = ":9202"
const NVIDIA_SMI_PATH = "/usr/bin/nvidia-smi"

var testMode string

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

type NvidiaSmiLog struct {
	DriverVersion string `xml:"driver_version"`
	CudaVersion   string `xml:"cuda_version"`
	AttachedGPUs  string `xml:"attached_gpus"`
	GPU           []struct {
		Id                       string `xml:"id,attr"`
		ProductName              string `xml:"product_name"`
		ProductBrand             string `xml:"product_brand"`
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
		PowerReadings struct {
			PowerState         string `xml:"power_state"`
			PowerManagement    string `xml:"power_management"`
			PowerDraw          string `xml:"power_draw"`
			PowerLimit         string `xml:"power_limit"`
			DefaultPowerLimit  string `xml:"default_power_limit"`
			EnforcedPowerLimit string `xml:"enforced_power_limit"`
			MinPowerLimit      string `xml:"min_power_limit"`
			MaxPowerLimit      string `xml:"max_power_limit"`
		} `xml:"power_readings"`
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
				Pid         string `xml:"pid"`
				Type        string `xml:"type"`
				ProcessName string `xml:"process_name"`
				UsedMemory  string `xml:"used_memory"`
			} `xml:"process_info"`
		} `xml:"processes"`
	} `xml:"gpu"`
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

func promEscape(value string) string {
	var re = regexp.MustCompile(`[\\"]`)
	return `"` + strings.ReplaceAll(re.ReplaceAllString(value, `\$0`), "\n", `\n`) + `"`
}

func writeMetric(w http.ResponseWriter, name string, labelValues map[string]string, value string) {
	var meta string
	for k, v := range labelValues {
		if meta != "" {
			meta += ","
		}
		meta += k + "=" + promEscape(v)
	}
	if meta != "" {
		meta = "{" + meta + "}"
	}
	io.WriteString(w, "nvidiasmi_"+name+meta+" "+value+"\n")
}

func metrics(w http.ResponseWriter, r *http.Request) {
	var cmd *exec.Cmd
	if testMode == "1" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("/bin/cat", dir+"/sample-xmls/geforce-gtx-980.xml")
	} else {
		cmd = exec.Command(NVIDIA_SMI_PATH, "-q", "-x")
	}

	// Execute system command
	stdout, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error executing nvidia-smi: "+err.Error(), 500)
		return
	}

	// Parse XML
	var xmlData NvidiaSmiLog
	if err := xml.Unmarshal(stdout, &xmlData); err != nil {
		http.Error(w, "Error parsing nvidia-smi output: "+err.Error(), 500)
		return
	}

	// Output
	writeMetric(w, "driver_version", nil, filterVersion(xmlData.DriverVersion))
	writeMetric(w, "cuda_version", nil, xmlData.CudaVersion)
	writeMetric(w, "cuda_version", nil, xmlData.CudaVersion)
	writeMetric(w, "attached_gpus", nil, xmlData.AttachedGPUs)

	for _, GPU := range xmlData.GPU {
		labelValues := map[string]string{
			"id":   GPU.Id,
			"uuid": GPU.UUID,
			"name": GPU.ProductName,
		}

		writeMetric(w, "pci_pcie_gen_max", labelValues, GPU.PCI.GPULinkInfo.PCIeGen.Max)
		writeMetric(w, "pci_pcie_gen_current", labelValues, GPU.PCI.GPULinkInfo.PCIeGen.Current)
		writeMetric(w, "pci_link_width_max_multiplicator", labelValues, filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Max))
		writeMetric(w, "pci_link_width_current_multiplicator", labelValues, filterNumber(GPU.PCI.GPULinkInfo.LinkWidth.Current))
		writeMetric(w, "pci_replay_counter", labelValues, GPU.PCI.ReplayRolloverCounter)
		writeMetric(w, "pci_replay_rollover_counter", labelValues, GPU.PCI.ReplayRolloverCounter)
		writeMetric(w, "pci_tx_util_bytes_per_second", labelValues, filterUnit(GPU.PCI.TxUtil))
		writeMetric(w, "pci_rx_util_bytes_per_second", labelValues, filterUnit(GPU.PCI.RxUtil))
		writeMetric(w, "fan_speed_percent", labelValues, filterUnit(GPU.FanSpeed))
		writeMetric(w, "performance_state_int", labelValues, filterNumber(GPU.PerformanceState))
		writeMetric(w, "fb_memory_usage_total_bytes", labelValues, filterUnit(GPU.FbMemoryUsage.Total))
		writeMetric(w, "fb_memory_usage_used_bytes", labelValues, filterUnit(GPU.FbMemoryUsage.Used))
		writeMetric(w, "fb_memory_usage_free_bytes", labelValues, filterUnit(GPU.FbMemoryUsage.Free))
		writeMetric(w, "bar1_memory_usage_total_bytes", labelValues, filterUnit(GPU.Bar1MemoryUsage.Total))
		writeMetric(w, "bar1_memory_usage_used_bytes", labelValues, filterUnit(GPU.Bar1MemoryUsage.Used))
		writeMetric(w, "bar1_memory_usage_free_bytes", labelValues, filterUnit(GPU.Bar1MemoryUsage.Free))
		writeMetric(w, "utilization_gpu_percent", labelValues, filterUnit(GPU.Utilization.GPUUtil))
		writeMetric(w, "utilization_memory_percent", labelValues, filterUnit(GPU.Utilization.MemoryUtil))
		writeMetric(w, "utilization_encoder_percent", labelValues, filterUnit(GPU.Utilization.EncoderUtil))
		writeMetric(w, "utilization_decoder_percent", labelValues, filterUnit(GPU.Utilization.DecoderUtil))
		writeMetric(w, "encoder_session_count", labelValues, GPU.EncoderStats.SessionCount)
		writeMetric(w, "encoder_average_fps", labelValues, GPU.EncoderStats.AverageFPS)
		writeMetric(w, "encoder_average_latency", labelValues, GPU.EncoderStats.AverageLatency)
		writeMetric(w, "fbc_session_count", labelValues, GPU.FBCStats.SessionCount)
		writeMetric(w, "fbc_average_fps", labelValues, GPU.FBCStats.AverageFPS)
		writeMetric(w, "fbc_average_latency", labelValues, GPU.FBCStats.AverageLatency)
		writeMetric(w, "gpu_temp_celsius", labelValues, filterUnit(GPU.Temperature.GPUTemp))
		writeMetric(w, "gpu_temp_max_threshold_celsius", labelValues, filterUnit(GPU.Temperature.GPUTempMaxThreshold))
		writeMetric(w, "gpu_temp_slow_threshold_celsius", labelValues, filterUnit(GPU.Temperature.GPUTempSlowThreshold))
		writeMetric(w, "gpu_temp_max_gpu_threshold_celsius", labelValues, filterUnit(GPU.Temperature.GPUTempMaxGpuThreshold))
		writeMetric(w, "gpu_target_temp_celsius", labelValues, filterUnit(GPU.Temperature.GPUTargetTemperature))
		writeMetric(w, "gpu_target_temp_min_celsius", labelValues, filterUnit(GPU.SupportedGPUTargetTemp.GPUTargetTempMin))
		writeMetric(w, "gpu_target_temp_max_celsius", labelValues, filterUnit(GPU.SupportedGPUTargetTemp.GPUTargetTempMax))
		writeMetric(w, "memory_temp_celsius", labelValues, filterUnit(GPU.Temperature.MemoryTemp))
		writeMetric(w, "gpu_temp_max_mem_threshold_celsius", labelValues, filterUnit(GPU.Temperature.GPUTempMaxMemThreshold))
		writeMetric(w, "power_state_int", labelValues, filterNumber(GPU.PowerReadings.PowerState))
		writeMetric(w, "power_draw_watts", labelValues, filterUnit(GPU.PowerReadings.PowerDraw))
		writeMetric(w, "power_limit_watts", labelValues, filterUnit(GPU.PowerReadings.PowerLimit))
		writeMetric(w, "default_power_limit_watts", labelValues, filterUnit(GPU.PowerReadings.DefaultPowerLimit))
		writeMetric(w, "enforced_power_limit_watts", labelValues, filterUnit(GPU.PowerReadings.EnforcedPowerLimit))
		writeMetric(w, "min_power_limit_watts", labelValues, filterUnit(GPU.PowerReadings.MinPowerLimit))
		writeMetric(w, "max_power_limit_watts", labelValues, filterUnit(GPU.PowerReadings.MaxPowerLimit))
		writeMetric(w, "clock_graphics_hertz", labelValues, filterUnit(GPU.Clocks.GraphicsClock))
		writeMetric(w, "clock_graphics_max_hertz", labelValues, filterUnit(GPU.MaxClocks.GraphicsClock))
		writeMetric(w, "clock_sm_hertz", labelValues, filterUnit(GPU.Clocks.SmClock))
		writeMetric(w, "clock_sm_max_hertz", labelValues, filterUnit(GPU.MaxClocks.SmClock))
		writeMetric(w, "clock_mem_hertz", labelValues, filterUnit(GPU.Clocks.MemClock))
		writeMetric(w, "clock_mem_max_hertz", labelValues, filterUnit(GPU.MaxClocks.MemClock))
		writeMetric(w, "clock_video_hertz", labelValues, filterUnit(GPU.Clocks.VideoClock))
		writeMetric(w, "clock_video_max_hertz", labelValues, filterUnit(GPU.MaxClocks.VideoClock))
		writeMetric(w, "clock_policy_auto_boost", labelValues, filterUnit(GPU.ClockPolicy.AutoBoost))
		writeMetric(w, "clock_policy_auto_boost_default", labelValues, filterUnit(GPU.ClockPolicy.AutoBoostDefault))
		writeMetric(w, "clocks_throttle_reason_gpu_idle", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonGPUIdle))
		writeMetric(w, "clocks_throttle_reason_applications_clocks_setting", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonApplicationsClocksSetting))
		writeMetric(w, "clocks_throttle_reason_sw_power_cap", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonSWPowerCap))
		writeMetric(w, "clocks_throttle_reason_hw_slowdown", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonHWSlowdown))
		writeMetric(w, "clocks_throttle_reason_hw_thermal_slowdown", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonHWThermalSlowdown))
		writeMetric(w, "clocks_throttle_reason_hw_power_brake_slowdown", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonHWPowerBrakeSlowdown))
		writeMetric(w, "clocks_throttle_reason_sync_boost", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonSyncBoost))
		writeMetric(w, "clocks_throttle_reason_sw_thermal_slowdown", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonSWThermalSlowdown))
		writeMetric(w, "clocks_throttle_reason_display_clocks_setting", labelValues, filterActive(GPU.ClockThrottleReasons.ClockThrottleReasonDisplayClocksSetting))

		for _, Process := range GPU.Processes.ProcessInfo {
			labelValues["process_pid"] = Process.Pid
			labelValues["process_type"] = Process.Type
			labelValues["process_name"] = Process.ProcessName

			writeMetric(w, "process_used_memory_bytes", labelValues, filterUnit(Process.UsedMemory))
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Nvidia SMI Exporter</title>
    </head>
    <body>
        <h1>Nvidia SMI Exporter</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
	io.WriteString(w, html)
}

func main() {
	testMode = os.Getenv("TEST_MODE")
	if testMode == "1" {
		log.Print("Test mode is enabled")
	}

	log.Print("Nvidia SMI exporter listening on " + LISTEN_ADDRESS)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(LISTEN_ADDRESS, nil)
}
