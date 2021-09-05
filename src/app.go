package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag(
		"listen",
		"Address to listen on.",
	).Default(":9202").String()
	nvidiaSmiPath = kingpin.Flag(
		"nvidia-smi-path",
		"Path to nvidia-smi",
	).Default("/usr/bin/nvidia-smi").String()
	updateInterval = kingpin.Flag(
		"update-interval",
		"How often to run nvidia-smi",
	).Default("5s").Duration()
	testFile = kingpin.Flag(
		"test-file",
		"Run in test mode (read nvidia-smi xml output from specified file)",
	).String()
)

// read and store

type OutputData struct {
	nvidiaSmiOutput NvidiaSmiOutput
	pcieInfo        map[string]PcieInfo   // by GPU Id
	processInfo     map[int64]ProcessInfo // by PID
}

var storedOutput OutputData

func readData() error {
	var data OutputData

	nvSmi, err := readNvidiaSmiOutput()
	if err != nil {
		return err
	}
	data.nvidiaSmiOutput = nvSmi

	data.pcieInfo = make(map[string]PcieInfo)
	data.processInfo = make(map[int64]ProcessInfo)

	for _, gpu := range nvSmi.GPU {
		data.pcieInfo[gpu.Id] = pcieInfo(gpu.Id)
		for _, process := range gpu.Processes.ProcessInfo {
			if _, ok := data.processInfo[process.Pid]; !ok {
				data.processInfo[process.Pid] = processInfo(process.Pid)
			}
		}
	}

	storedOutput = data
	return nil
}

// output

func promEscape(value string) string {
	var re = regexp.MustCompile(`[\\"]`)
	return `"` + strings.ReplaceAll(re.ReplaceAllString(value, `\$0`), "\n", `\n`) + `"`
}

func writeMetric(w http.ResponseWriter, name string, labelValues map[string]string, value string) {
	// make sorted array of keys to achieve a fixed order (otherwise the map iteration order is random each time)
	labelKeys := make([]string, 0, len(labelValues))
	for k := range labelValues {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)

	var meta string
	for _, k := range labelKeys {
		if meta != "" {
			meta += ","
		}
		meta += k + "=" + promEscape(labelValues[k])
	}
	if meta != "" {
		meta = "{" + meta + "}"
	}

	io.WriteString(w, "nvidiasmi_"+name+meta+" "+value+"\n")
}

func metrics(w http.ResponseWriter, r *http.Request) {
	output := storedOutput.nvidiaSmiOutput

	// Output
	writeMetric(w, "driver_version", nil, filterVersion(output.DriverVersion))
	writeMetric(w, "cuda_version", nil, output.CudaVersion)
	writeMetric(w, "cuda_version", nil, output.CudaVersion)
	writeMetric(w, "attached_gpus", nil, output.AttachedGPUs)

	for _, GPU := range output.GPU {
		labelValues := map[string]string{
			"id":       GPU.Id,
			"short_id": regexp.MustCompile(`^\d{8}:`).ReplaceAllString(GPU.Id, ""),
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

		pcie := storedOutput.pcieInfo[GPU.Id]
		labelValues["aer_type"] = "fatal"
		writeMetric(w, "aer_counter", labelValues, strconv.Itoa(pcie.AerFatalCount))
		labelValues["aer_type"] = "non-fatal"
		writeMetric(w, "aer_counter", labelValues, strconv.Itoa(pcie.AerNonFatalCount))
		labelValues["aer_type"] = "correctable"
		writeMetric(w, "aer_counter", labelValues, strconv.Itoa(pcie.AerCorrectableCount))
		delete(labelValues, "aer_type")

		labelValues["uuid"] = GPU.UUID
		labelValues["name"] = GPU.ProductName
		writeMetric(w, "gpu_info", labelValues, "1.0")

		for _, Process := range GPU.Processes.ProcessInfo {
			labelValues2 := map[string]string{
				"id":            labelValues["id"],
				"short_id":      labelValues["short_id"],
				"pid":           fmt.Sprintf("%d", Process.Pid),
				"proocess_type": Process.Type,
			}
			writeMetric(w, "process_used_memory_bytes", labelValues2, filterUnit(Process.UsedMemory))
		}
	}

	for pid, pInfo := range storedOutput.processInfo {
		labelValues := map[string]string{
			"pid": fmt.Sprintf("%d", pid),
		}

		writeMetric(w, "process_start_timestamp", labelValues, fmt.Sprintf("%f", pInfo.processStartTs))
		if pInfo.containerStartTs > 0 {
			writeMetric(w, "process_container_start_timestamp", labelValues, fmt.Sprintf("%f", pInfo.containerStartTs))
		}

		labelValues["process_name"] = pInfo.processName
		if pInfo.containerId != "" {
			labelValues["container_id"] = pInfo.containerId
			labelValues["container_name"] = pInfo.containerName
			labelValues["docker_image"] = pInfo.dockerImage
		}

		writeMetric(w, "process_info", labelValues, "1.0")
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
	kingpin.Version(version.Print("nvidiasmi_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *testFile != "" {
		log.Infoln("Test mode is enabled")
	}

	err := readData()
	if err != nil {
		// initial update must succeed, otherwise exit
		log.Fatalln(err)
	}

	go func() {
		for {
			time.Sleep(*updateInterval)
			err := readData()
			if err != nil {
				log.Errorln(err)
			}
		}
	}()

	log.Infoln("Nvidia SMI exporter listening on " + *listenAddress)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(*listenAddress, nil)
}
