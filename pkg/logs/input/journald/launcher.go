// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package journald

import (
	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/restart"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Launcher is in charge of starting and stopping new journald tailers
type Launcher struct {
	sources          *config.LogSources
	pipelineProvider pipeline.Provider
	auditor          *auditor.Auditor
	tailers          map[string]*Tailer
}

// New returns a new Launcher.
func New(sources *config.LogSources, pipelineProvider pipeline.Provider, auditor *auditor.Auditor) *Launcher {
	return &Launcher{
		sources:          sources,
		pipelineProvider: pipelineProvider,
		auditor:          auditor,
		tailers:          make(map[string]*Tailer),
	}
}

// Start starts new tailers.
func (l *Launcher) Start() {
	for _, source := range l.sources.GetValidSourcesWithType(config.JournaldType) {
		identifier := source.Config.Path
		if _, exists := l.tailers[identifier]; exists {
			// set up only one tailer per journal
			continue
		}
		tailer, err := l.setupTailer(source)
		if err != nil {
			log.Warn("Could not set up journald tailer: ", err)
		} else {
			l.tailers[identifier] = tailer
		}
	}
}

// Stop stops all active tailers
func (l *Launcher) Stop() {
	stopper := restart.NewParallelStopper()
	for identifier, tailer := range l.tailers {
		stopper.Add(tailer)
		delete(l.tailers, identifier)
	}
	stopper.Stop()
}

// setupTailer configures and starts a new tailer,
// returns the tailer or an error.
func (l *Launcher) setupTailer(source *config.LogSource) (*Tailer, error) {
	tailer := NewTailer(source, l.pipelineProvider.NextPipelineChan())
	err := tailer.Start(l.auditor.GetLastCommittedOffset(tailer.Identifier()))
	if err != nil {
		return nil, err
	}
	return tailer, nil
}
