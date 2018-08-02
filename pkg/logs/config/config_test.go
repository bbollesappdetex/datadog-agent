// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDatadogConfig(t *testing.T) {
	assert.Equal(t, false, LogsAgent.GetBool("log_enabled"))
	assert.Equal(t, false, LogsAgent.GetBool("logs_enabled"))
	assert.Equal(t, "", LogsAgent.GetString("logset"))
	assert.Equal(t, "agent-intake.logs.datadoghq.com", LogsAgent.GetString("logs_config.dd_url"))
	assert.Equal(t, 10516, LogsAgent.GetInt("logs_config.dd_port"))
	assert.Equal(t, false, LogsAgent.GetBool("logs_config.dev_mode_no_ssl"))
	assert.Equal(t, true, LogsAgent.GetBool("logs_config.dev_mode_use_proto"))
	assert.Equal(t, 100, LogsAgent.GetInt("logs_config.open_files_limit"))
	assert.Equal(t, 9000, LogsAgent.GetInt("logs_config.frame_size"))
	assert.Equal(t, -1, LogsAgent.GetInt("logs_config.tcp_forward_port"))
}

func TestBuild(t *testing.T) {
	var sources *LogSources
	var source *LogSource

	// should return an error
	sources = Build()
	assert.Equal(t, 0, len(sources.GetValidSources()))

	// should return the default tail all containers source
	LogsAgent.Set("logs_config.container_collect_all", true)
	LogsAgent.Set("logs_config.tcp_forward_port", -1)
	sources = Build()
	assert.Equal(t, 1, len(sources.GetValidSources()))
	source = sources.GetValidSources()[0]
	assert.Equal(t, "container_collect_all", source.Name)
	assert.Equal(t, DockerType, source.Config.Type)
	assert.Equal(t, "docker", source.Config.Service)
	assert.Equal(t, "docker", source.Config.Source)

	// should return the tcp forward source
	LogsAgent.Set("logs_config.container_collect_all", false)
	LogsAgent.Set("logs_config.tcp_forward_port", 1234)
	sources = Build()
	assert.Equal(t, 1, len(sources.GetValidSources()))
	source = sources.GetValidSources()[0]
	assert.Equal(t, "tcp_forward", source.Name)
	assert.Equal(t, TCPType, source.Config.Type)
	assert.Equal(t, 1234, source.Config.Port)

	// should return the container collect all and tcp forward sources
	LogsAgent.Set("logs_config.container_collect_all", true)
	LogsAgent.Set("logs_config.tcp_forward_port", 1234)
	sources = Build()
	assert.Equal(t, 2, len(sources.GetValidSources()))

	source = sources.GetValidSources()[0]
	assert.Equal(t, "container_collect_all", source.Name)
	assert.Equal(t, DockerType, source.Config.Type)
	assert.Equal(t, "docker", source.Config.Service)
	assert.Equal(t, "docker", source.Config.Source)

	source = sources.GetValidSources()[1]
	assert.Equal(t, "tcp_forward", source.Name)
	assert.Equal(t, TCPType, source.Config.Type)
	assert.Equal(t, 1234, source.Config.Port)
}
