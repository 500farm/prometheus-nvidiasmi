# nvidiasmi_exporter
Prometheus exporter for GPU statistics from [nvidia-smi](https://developer.nvidia.com/nvidia-system-management-interface), written in Go. Supports multiple GPUs.

In addition to nvidia-smi info, also exports some info about processes using the GPU, docker containers processes belong to, and PCIe error stats.

This version is not dockerized because it requires access to docker daemon and /sys filesystem.

### Setup

```sh
$ make 
$ sudo make install
```

After that, the exporter will be started automatically by systemd on startup.

### Usage

By default, `nvidiasmi_exporter` listens on port 9202.

CLI args (specify in `/etc/systemd/system/nvidiasmi_exporter.service`):

```
--listen
    Address to listen on (default 9202).

--nvidia-smi-path
    Path to nvidia-smi (default /usr/bin/nvidia-smi).

--update-interval
    How often to run nvidia-smi (default 5s)

--test-file
    Run in test mode (read nvidia-smi xml output from specified file)
```

### Credits

Based on work of [Kristoph Junge](https://github.com/kristophjunge/docker-prometheus-nvidiasmi) and 
[Michaël Ferrand](https://github.com/e7d/docker-prometheus-nvidiasmi).


### Example output with annotations

```
### Driver info
nvidiasmi_info{attached_gpus="1",cuda_version="11.4",driver_version="470.63"} 1.0

### Nvidia_smi info
nvidiasmi_pci_pcie_gen_max{gpu_id="46:00.0"} 3
nvidiasmi_pci_pcie_gen_current{gpu_id="46:00.0"} 3
nvidiasmi_pci_link_width_max_multiplicator{gpu_id="46:00.0"} 16
nvidiasmi_pci_link_width_current_multiplicator{gpu_id="46:00.0"} 16
nvidiasmi_pci_replay_counter{gpu_id="46:00.0"} 0
nvidiasmi_pci_replay_rollover_counter{gpu_id="46:00.0"} 0
nvidiasmi_pci_tx_util_bytes_per_second{gpu_id="46:00.0"} 5.9e+07
nvidiasmi_pci_rx_util_bytes_per_second{gpu_id="46:00.0"} 4.67e+08
nvidiasmi_fan_speed_percent{gpu_id="46:00.0"} 30
nvidiasmi_performance_state_int{gpu_id="46:00.0"} 2
nvidiasmi_fb_memory_usage_total_bytes{gpu_id="46:00.0"} 2.5446842368e+10
nvidiasmi_fb_memory_usage_used_bytes{gpu_id="46:00.0"} 2.4920457216e+10
nvidiasmi_fb_memory_usage_free_bytes{gpu_id="46:00.0"} 5.26385152e+08
nvidiasmi_bar1_memory_usage_total_bytes{gpu_id="46:00.0"} 2.68435456e+08
nvidiasmi_bar1_memory_usage_used_bytes{gpu_id="46:00.0"} 7.340032e+06
nvidiasmi_bar1_memory_usage_free_bytes{gpu_id="46:00.0"} 2.61095424e+08
nvidiasmi_utilization_gpu_percent{gpu_id="46:00.0"} 23
nvidiasmi_utilization_memory_percent{gpu_id="46:00.0"} 1
nvidiasmi_utilization_encoder_percent{gpu_id="46:00.0"} 0
nvidiasmi_utilization_decoder_percent{gpu_id="46:00.0"} 0
nvidiasmi_encoder_session_count{gpu_id="46:00.0"} 0
nvidiasmi_encoder_average_fps{gpu_id="46:00.0"} 0
nvidiasmi_encoder_average_latency{gpu_id="46:00.0"} 0
nvidiasmi_fbc_session_count{gpu_id="46:00.0"} 0
nvidiasmi_fbc_average_fps{gpu_id="46:00.0"} 0
nvidiasmi_fbc_average_latency{gpu_id="46:00.0"} 0
nvidiasmi_gpu_temp_celsius{gpu_id="46:00.0"} 32
nvidiasmi_gpu_temp_max_threshold_celsius{gpu_id="46:00.0"} 98
nvidiasmi_gpu_temp_slow_threshold_celsius{gpu_id="46:00.0"} 95
nvidiasmi_gpu_temp_max_gpu_threshold_celsius{gpu_id="46:00.0"} 93
nvidiasmi_gpu_target_temp_celsius{gpu_id="46:00.0"} 75
nvidiasmi_gpu_target_temp_min_celsius{gpu_id="46:00.0"} 0
nvidiasmi_gpu_target_temp_max_celsius{gpu_id="46:00.0"} 0
nvidiasmi_memory_temp_celsius{gpu_id="46:00.0"} 0
nvidiasmi_gpu_temp_max_mem_threshold_celsius{gpu_id="46:00.0"} 0
nvidiasmi_power_state_int{gpu_id="46:00.0"} 2
nvidiasmi_power_draw_watts{gpu_id="46:00.0"} 112.56999969482422
nvidiasmi_power_limit_watts{gpu_id="46:00.0"} 375
nvidiasmi_default_power_limit_watts{gpu_id="46:00.0"} 350
nvidiasmi_enforced_power_limit_watts{gpu_id="46:00.0"} 375
nvidiasmi_min_power_limit_watts{gpu_id="46:00.0"} 100
nvidiasmi_max_power_limit_watts{gpu_id="46:00.0"} 400
nvidiasmi_clock_graphics_hertz{gpu_id="46:00.0"} 1.695e+09
nvidiasmi_clock_graphics_max_hertz{gpu_id="46:00.0"} 2.1e+09
nvidiasmi_clock_sm_hertz{gpu_id="46:00.0"} 1.695e+09
nvidiasmi_clock_sm_max_hertz{gpu_id="46:00.0"} 2.1e+09
nvidiasmi_clock_mem_hertz{gpu_id="46:00.0"} 9.501e+09
nvidiasmi_clock_mem_max_hertz{gpu_id="46:00.0"} 9.751e+09
nvidiasmi_clock_video_hertz{gpu_id="46:00.0"} 1.515e+09
nvidiasmi_clock_video_max_hertz{gpu_id="46:00.0"} 1.95e+09
nvidiasmi_clock_policy_auto_boost{gpu_id="46:00.0"} 0
nvidiasmi_clock_policy_auto_boost_default{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_gpu_idle{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_applications_clocks_setting{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_sw_power_cap{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_hw_slowdown{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_hw_thermal_slowdown{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_hw_power_brake_slowdown{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_sync_boost{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_sw_thermal_slowdown{gpu_id="46:00.0"} 0
nvidiasmi_clocks_throttle_reason_display_clocks_setting{gpu_id="46:00.0"} 0

### GPU name and UUID
nvidiasmi_gpu_info{device="GA102 [GeForce RTX 3090]",gpu_id="46:00.0",gpu_name="NVIDIA GeForce RTX 3090",gpu_uuid="GPU-325bf28b-e1e1-0628-3678-06673bdb76fd",subsys_device="Device 147d",subsys_vendor="NVIDIA Corporation",vendor="NVIDIA Corporation"} 1.0

### PCIe AER counters
nvidiasmi_aer_counter{aer_type="fatal",gpu_id="46:00.0"} 0
nvidiasmi_aer_counter{aer_type="non-fatal",gpu_id="46:00.0"} 0
nvidiasmi_aer_counter{aer_type="correctable",gpu_id="46:00.0"} 0

### Process/container info
nvidiasmi_process_up{gpu_id="46:00.0",pid="3890000",process_type="C"} 1.0
nvidiasmi_process_used_memory_bytes{gpu_id="46:00.0",pid="3890000",process_type="C"} 2.4917311488e+10
nvidiasmi_process_start_timestamp{pid="3890000"} 1633035666.200000
nvidiasmi_process_container_start_timestamp{pid="3890000"} 1632969100.384305
nvidiasmi_process_info{container_id="/docker/e3b272ff36976664996c4c537816ba712d63725e7b5f6d121e875a1d738c4f4a",container_name="C.1329067",docker_image="pytorch/pytorch",pid="3890000",process_name="/usr/bin/python3.6"} 1.0
