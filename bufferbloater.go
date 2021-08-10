package main

import (
	"io/ioutil"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/tonya11en/bufferbloater/client"
	"github.com/tonya11en/bufferbloater/stats"
)

type Bufferbloater struct {
	log      *zap.SugaredLogger
	c        *client.Client
	statsMgr *stats.StatsMgr
}

// Basic representation of the parsed yaml file before the durations are parsed.
type parsedYamlConfig struct {
	Client struct {
		Workload []struct {
			Rps      uint
			Duration string
		} `yaml:"workload"`
		RqTimeout    string `yaml:"rq_timeout"`
		TargetServer struct {
			Address string
			Port    uint
		} `yaml:"target_server"`
	} `yaml:"client"`
}

// Creates a properly typed client config.
func clientConfigParse(parsedConfig parsedYamlConfig) (client.Config, error) {
	// TODO: validate config

	clientConfig := client.Config{
		TargetServer: client.Target{
			Address: parsedConfig.Client.TargetServer.Address,
			Port:    parsedConfig.Client.TargetServer.Port,
		},
	}

	d, err := time.ParseDuration(parsedConfig.Client.RqTimeout)
	if err != nil {
		return client.Config{}, err
	}
	clientConfig.RequestTimeout = d

	for _, stage := range parsedConfig.Client.Workload {
		d, err := time.ParseDuration(stage.Duration)
		if err != nil {
			return client.Config{}, err
		}

		workloadStage := client.WorkloadStage{
			RPS:      stage.Rps,
			Duration: d,
		}
		clientConfig.Workload = append(clientConfig.Workload, workloadStage)
	}

	return clientConfig, nil
}

func parseConfigFromFile(configFilename string) (parsedYamlConfig, error) {
	// Read the config file.
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return parsedYamlConfig{}, err
	}

	// Parse the config file yaml.
	var parsedConfig parsedYamlConfig
	err = yaml.UnmarshalStrict([]byte(data), &parsedConfig)
	if err != nil {
		return parsedYamlConfig{}, err
	}

	return parsedConfig, nil
}

func NewBufferbloater(configFilename string, logger *zap.SugaredLogger) (*Bufferbloater, error) {
	bb := Bufferbloater{
		log:      logger,
		statsMgr: stats.NewStatsMgrImpl(logger),
	}

	parsedConfig, err := parseConfigFromFile(configFilename)
	if err != nil {
		bb.log.Fatalw("failed to parse yaml file",
			"error", err)
	}

	clientConfig, err := clientConfigParse(parsedConfig)
	if err != nil {
		bb.log.Fatalw("failed to create client config",
			"error", err)
	}
	bb.c = client.NewClient(clientConfig, logger, bb.statsMgr)
	return &bb, nil
}

func (bb *Bufferbloater) Run() {
	// TODO: make folder configurable.
	defer bb.statsMgr.DumpStatsToFolder("data")

	stopStats := make(chan struct{}, 1)
	var statsWg sync.WaitGroup
	statsWg.Add(1)
	go bb.statsMgr.PeriodicStatsCollection(100*time.Millisecond, stopStats, &statsWg)

	var wg sync.WaitGroup
	wg.Add(1)
	//go bb.s.Start(&wg)
	go bb.c.Start(&wg)
	wg.Wait()

	stopStats <- struct{}{}
	statsWg.Wait()
}
