// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.

// +build docker

package collectors

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-agent/pkg/errors"
	taggerutil "github.com/DataDog/datadog-agent/pkg/tagger/utils"
	"github.com/DataDog/datadog-agent/pkg/util/docker"
	ecsutil "github.com/DataDog/datadog-agent/pkg/util/ecs"
)

const (
	ecsFargateCollectorName = "ecs_fargate"
	ecsFargateExpireFreq    = 5 * time.Minute
)

// ECSFargateCollector polls the ecs metadata api.
type ECSFargateCollector struct {
	infoOut      chan<- []*TagInfo
	expire       *taggerutil.Expire
	lastExpire   time.Time
	lastSeen     map[string]interface{}
	expireFreq   time.Duration
	labelsAsTags map[string]string
}

// Detect tries to connect to the ECS metadata API
func (c *ECSFargateCollector) Detect(out chan<- []*TagInfo) (CollectionMode, error) {
	var err error

	if ecsutil.IsFargateInstance() {
		c.infoOut = out
		c.lastExpire = time.Now()
		c.expireFreq = ecsFargateExpireFreq
		c.expire, err = taggerutil.NewExpire(ecsFargateExpireFreq)
		c.labelsAsTags = retrieveMappingFromConfig("docker_labels_as_tags")

		if err != nil {
			return FetchOnlyCollection, fmt.Errorf("Failed to instantiate the container expiring process")
		}
		return FetchOnlyCollection, nil
	}

	return NoCollection, fmt.Errorf("Failed to connect to task metadata API, ECS tagging will not work")
}

// Fetch fetches ECS tags for a container on demand
func (c *ECSFargateCollector) Fetch(container string) ([]string, []string, error) {
	meta, err := ecsutil.GetTaskMetadata()
	if err != nil {
		return []string{}, []string{}, err
	}
	updates, err := c.parseMetadata(meta)
	if err != nil {
		return []string{}, []string{}, err
	}
	c.infoOut <- updates

	// Throttle deletion computations
	if time.Now().After(c.lastExpire.Add(c.expireFreq)) {
		expireList, err := c.expire.ComputeExpires()
		if err != nil {
			return []string{}, []string{}, err
		}
		expiries, err := c.parseExpires(expireList)
		if err != nil {
			return []string{}, []string{}, err
		}
		c.infoOut <- expiries
		c.lastExpire = time.Now()
	}

	for _, info := range updates {
		if info.Entity == container {
			return info.LowCardTags, info.HighCardTags, nil
		}
	}
	// container not found in updates
	return []string{}, []string{}, errors.NewNotFound(container)
}

// parseExpires transforms event from the PodWatcher to TagInfo objects
func (c *ECSFargateCollector) parseExpires(idList []string) ([]*TagInfo, error) {
	var output []*TagInfo
	for _, id := range idList {
		info := &TagInfo{
			Source:       ecsFargateCollectorName,
			Entity:       docker.ContainerIDToEntityName(id),
			DeleteEntity: true,
		}
		output = append(output, info)
	}
	return output, nil
}

func ecsFargateFactory() Collector {
	return &ECSFargateCollector{}
}

func init() {
	registerCollector(ecsFargateCollectorName, ecsFargateFactory, NodeOrchestrator)
}
