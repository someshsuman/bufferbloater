package main

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/tonya11en/bufferbloater/server"
	"github.com/tonya11en/bufferbloater/stats"
)

type Bufferbloater struct {
	log      *zap.SugaredLogger
	s        *server.Server
	statsMgr *stats.StatsMgr
}

func serverConfigParse(latency string, port int, thread int, duration string) (server.Config, error) {
	// TODO: validate config

	serverConfig := server.Config{
		ListenPort: uint(port),
		Threads:    uint(thread),
	}

	s := server.LatencySegment{}
	// Calculate the latency distribution.
	s.WeightSum = 0
	s.LatencyDistribution = []server.WeightedLatency{}

	latencyInput := strings.Split(latency, ",")
	for _, stage := range latencyInput {
		weights := strings.Split(stage, ":")

		weight, err := strconv.Atoi(weights[0])
		if err != nil {
			return server.Config{}, err
		}
		latencyDuration, err := time.ParseDuration(weights[1])
		if err != nil {
			return server.Config{}, err
		}

		weightedLatency := server.WeightedLatency{
			Weight:  uint(weight),
			Latency: latencyDuration,
		}
		s.WeightSum += uint(weight)
		s.LatencyDistribution = append(s.LatencyDistribution, weightedLatency)
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return server.Config{}, err
	}
	s.SegmentDuration = d

	serverConfig.Profile = append(serverConfig.Profile, s)

	return serverConfig, nil
}

func NewBufferbloater(latency string, port int, thread int, duration string, logger *zap.SugaredLogger) (*Bufferbloater, error) {
	bb := Bufferbloater{
		log:      logger,
		statsMgr: stats.NewStatsMgrImpl(logger),
	}

	serverConfig, err := serverConfigParse(latency, port, thread, duration)
	if err != nil {
		bb.log.Fatalw("failed to create server config",
			"error", err)
	}
	bb.s = server.NewServer(serverConfig, logger, bb.statsMgr)

	return &bb, nil
}

func (bb *Bufferbloater) Run() {

	var wg sync.WaitGroup
	wg.Add(1)
	go bb.s.Start(&wg)
	wg.Wait()
}
