// Package docker comment :)
package docker

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/b92c/gowatch/pkg/metrics"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type ContainerLog struct {
	ID   string
	Logs []string
}

type FormattedLog struct {
	Service string
	Line    string
}

var (
	previousStats = make(map[string]container.StatsResponse)
	historyStore  = make(map[string]metrics.ContainerStats)
	statsMutex    sync.RWMutex
)

func getContainerStats(ctx context.Context, apiClient *client.Client, containerID string) metrics.ContainerStats {
	stats, err := apiClient.ContainerStats(ctx, containerID, client.ContainerStatsOptions{Stream: false})
	if err != nil {
		return metrics.ContainerStats{}
	}
	defer stats.Body.Close()

	var statsJSON container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return metrics.ContainerStats{}
	}

	statsMutex.Lock()
	defer statsMutex.Unlock()

	prevStats, exists := previousStats[containerID]
	if exists {
		statsJSON.PreCPUStats = prevStats.CPUStats
	}
	previousStats[containerID] = statsJSON

	stat := ParseStats(statsJSON)

	// Update history
	h, exists := historyStore[containerID]
	if !exists {
		h = metrics.ContainerStats{}
	}

	now := time.Now()
	h.CPUPercent = stat.CPUPercent
	h.MemUsage = stat.MemUsage
	h.NetRxBytes = stat.NetRxBytes
	h.NetTxBytes = stat.NetTxBytes
	h.NetRxPackets = stat.NetRxPackets
	h.NetTxPackets = stat.NetTxPackets
	h.DiskReadBytes = stat.DiskReadBytes
	h.DiskWriteBytes = stat.DiskWriteBytes
	h.DiskReadOps = stat.DiskReadOps
	h.DiskWriteOps = stat.DiskWriteOps
	h.PIDsCurrent = stat.PIDsCurrent
	h.OOMEvents = stat.OOMEvents

	h.CPUHistory = append(h.CPUHistory, metrics.MetricPoint{Timestamp: now, Value: stat.CPUPercent})
	h.MemHistory = append(h.MemHistory, metrics.MetricPoint{Timestamp: now, Value: float64(stat.MemUsage)})

	// Limit history
	if len(h.CPUHistory) > metrics.MaxHistoryPoints {
		h.CPUHistory = h.CPUHistory[1:]
	}
	if len(h.MemHistory) > metrics.MaxHistoryPoints {
		h.MemHistory = h.MemHistory[1:]
	}

	historyStore[containerID] = h
	return h
}

func getContainerLogs(ctx context.Context, apiClient *client.Client, containerID string) []string {
	logs, err := apiClient.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "50",
	})
	if err != nil {
		return []string{"Error fetching logs"}
	}
	defer logs.Close()

	return ParseLogs(logs)
}

func getHostInfo() metrics.HostInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return metrics.HostInfo{
		CPUCount: runtime.NumCPU(),
		MemTotal: memStats.Sys,
		MemFree:  memStats.Frees,
	}
}

type Containers struct {
	C        []Container
	Logs     []ContainerLog
	FlatLogs []FormattedLog
	Host     metrics.HostInfo
}

type Container struct {
	WorkingDir     string
	SOVersion      string
	Status         string
	State          string
	Command        string
	DependsOn      string
	ID             string
	Image          string
	ConfigFile     string
	Service        string
	CPUHistory     []metrics.MetricPoint
	Log            []string
	MemHistory     []metrics.MetricPoint
	CPUPercent     float64
	NetTxBytes     uint64
	NetRxPackets   uint64
	NetTxPackets   uint64
	DiskReadBytes  uint64
	DiskWriteBytes uint64
	DiskReadOps    uint64
	DiskWriteOps   uint64
	PIDsCurrent    uint64
	OOMEvents      uint64
	CreatedAt      int64
	NetRxBytes     uint64
	MemUsage       uint64
}

func WatchContainers(ctx context.Context, apiClient *client.Client) (Containers, error) {
	cntList, err := apiClient.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return Containers{}, err
	}

	var containers Containers
	for _, c := range cntList.Items {
		stat := getContainerStats(ctx, apiClient, c.ID)
		logs := getContainerLogs(ctx, apiClient, c.ID)
		containers.C = append(containers.C, Container{
			ID: c.ID, Image: c.Image, Status: c.Status, State: string(c.State),
			Command: c.Command, DependsOn: c.Labels["com.docker.compose.depends_on"],
			Service:    c.Labels["com.docker.compose.service"],
			SOVersion:  c.Labels["org.opencontainers.image.ref.name"] + " " + c.Labels["org.opencontainers.image.version"],
			WorkingDir: c.Labels["com.docker.compose.project.working_dir"], ConfigFile: c.Labels["com.docker.compose.project.config_files"],
			CreatedAt:      c.Created,
			CPUPercent:     stat.CPUPercent,
			MemUsage:       stat.MemUsage,
			NetRxBytes:     stat.NetRxBytes,
			NetTxBytes:     stat.NetTxBytes,
			NetRxPackets:   stat.NetRxPackets,
			NetTxPackets:   stat.NetTxPackets,
			DiskReadBytes:  stat.DiskReadBytes,
			DiskWriteBytes: stat.DiskWriteBytes,
			DiskReadOps:    stat.DiskReadOps,
			DiskWriteOps:   stat.DiskWriteOps,
			PIDsCurrent:    stat.PIDsCurrent,
			OOMEvents:      stat.OOMEvents,
			CPUHistory:     stat.CPUHistory,
			MemHistory:     stat.MemHistory,
			Log:            logs,
		})
	}
	containers.Host = getHostInfo()

	for _, c := range containers.C {
		serviceName := c.Service
		if serviceName == "" {
			if len(c.ID) >= 12 {
				serviceName = c.ID[:12]
			} else {
				serviceName = c.ID
			}
		}
		for _, line := range c.Log {
			containers.FlatLogs = append(containers.FlatLogs, FormattedLog{
				Service: serviceName,
				Line:    line,
			})
		}
	}

	return containers, nil
}
