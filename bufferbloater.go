package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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
func clientConfigParse(workload string, rqTimeout string, target string) (client.Config, error) {
	// TODO: validate config

	serverAddress := strings.Split(target, ":")
	serverIp := serverAddress[0]
	port, err := strconv.Atoi(serverAddress[1])
	if err != nil {
		return client.Config{}, err
	}
	clientConfig := client.Config{
		TargetServer: client.Target{
			Address: serverIp,
			Port:    uint(port),
		},
	}

	d, err := time.ParseDuration(rqTimeout)
	if err != nil {
		return client.Config{}, err
	}
	clientConfig.RequestTimeout = d
	workloadInput := strings.Split(workload, ",")
	for _, stage := range workloadInput {
		rpsRate := strings.Split(stage, ":")

		rps, err := strconv.Atoi(rpsRate[0])
		if err != nil {
			return client.Config{}, err
		}
		d, err := time.ParseDuration(rpsRate[1])
		if err != nil {
			return client.Config{}, err
		}

		workloadStage := client.WorkloadStage{
			RPS:      uint(rps),
			Duration: d,
		}
		clientConfig.Workload = append(clientConfig.Workload, workloadStage)
	}
	fmt.Print(clientConfig)

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

func NewBufferbloater(workload string, rqTimeout string, target string, logger *zap.SugaredLogger) (*Bufferbloater, error) {
	bb := Bufferbloater{
		log:      logger,
		statsMgr: stats.NewStatsMgrImpl(logger),
	}

	clientConfig, err := clientConfigParse(workload, rqTimeout, target)
	if err != nil {
		bb.log.Fatalw("failed to create client config",
			"error", err)
	}
	bb.c = client.NewClient(clientConfig, logger, bb.statsMgr)
	return &bb, nil
}

func (bb *Bufferbloater) Run() {
	// TODO: make folder configurable.
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Creating status in path=" + path)
	defer bb.statsMgr.DumpStatsToFolder(path + "/data")

	stopStats := make(chan struct{}, 1)
	var statsWg sync.WaitGroup
	statsWg.Add(1)
	go bb.statsMgr.PeriodicStatsCollection(100*time.Millisecond, stopStats, &statsWg)

	var wg sync.WaitGroup
	wg.Add(1)
	go bb.c.Start(&wg)
	wg.Wait()

	stopStats <- struct{}{}
	statsWg.Wait()
}
