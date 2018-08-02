// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package config

import (
	"github.com/DataDog/datadog-agent/pkg/config"
)

// LogsAgent is the global configuration object
var LogsAgent = config.Datadog

// Build returns logs-agent sources
func Build() *LogSources {
	var sources []*LogSource

	if LogsAgent.GetBool("logs_config.container_collect_all") {
		// append source to collect all logs from all containers.
		containersSource := NewLogSource("container_collect_all", &LogsConfig{
			Type:    DockerType,
			Service: "docker",
			Source:  "docker",
		})
		sources = append(sources, containersSource)
	}

	tcpForwardPort := LogsAgent.GetInt("logs_config.tcp_forward_port")
	if tcpForwardPort > 0 {
		// append source to collect all logs forwarded by TCP on a given port.
		tcpForwardSource := NewLogSource("tcp_forward", &LogsConfig{
			Type: TCPType,
			Port: tcpForwardPort,
		})
		sources = append(sources, tcpForwardSource)
	}

	return NewLogSources(sources)
}
