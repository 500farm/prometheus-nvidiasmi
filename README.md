# Docker Prometheus Nvidia SMI Exporter

Dockerized Prometheus exporter for GPU statistics from [nvidia-smi](https://developer.nvidia.com/nvidia-system-management-interface), written in Go.
Supports multiple GPUs.

# How-To

`docker run --runtime nvidia -p 9202:9202/tcp sergeycheperis/prometheus-nvidiasmi`

Check result at: [http://localhost:9202/metrics](http://localhost:9202/metrics)

# Grafana dashboard

[Nvidia SMI Metrics dashboard](https://grafana.com/grafana/dashboards/12357) on Grafana Labs
