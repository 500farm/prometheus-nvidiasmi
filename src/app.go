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
	"strings"
)

const listenAddress = ":9202"
const nvidiaSmiPath = "/usr/bin/nvidia-smi"

var testMode string

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
		cmd = exec.Command(nvidiaSmiPath, "-q", "-x")
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
			"id":       GPU.Id,
			"short_id": regexp.MustCompile(`^\d{8}:`).ReplaceAllString(GPU.Id, ""),
			"uuid":     GPU.UUID,
			"name":     GPU.ProductName,
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
			labelValues["process_pid"] = fmt.Sprintf("%d", Process.Pid)
			labelValues["process_type"] = Process.Type
			labelValues["process_name"] = Process.ProcessName
			var ts int64
			labelValues["container_id"], labelValues["container_name"], labelValues["docker_image"], ts = containerInfo(Process.Pid)
			labelValues["container_create_timestamp"] = fmt.Sprintf("%d", ts)

			writeMetric(w, "process_start_timestamp", labelValues, fmt.Sprintf("%f", processStartTimestamp(Process.Pid)))
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

	log.Print("Nvidia SMI exporter listening on " + listenAddress)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(listenAddress, nil)
}
