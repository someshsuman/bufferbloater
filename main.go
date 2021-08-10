package main

import (
	"flag"

	"go.uber.org/zap"
)

func main() {
	latency := flag.String("latency", "90:50ms,5:60ms,4:100ms,1:150ms", "RPS1:duration1,RPS2:duration...")
	port := flag.Int("port", 9002, "Request timeout duration in sec")
	thread := flag.Int("thread", 8, "ServerAddress:Port")
	duration := flag.String("duration", "7200s", "duration for the server up")
	flag.Parse()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("couldn't initialize logging")
	}
	sugar := logger.Sugar()

	bb, err := NewBufferbloater(*latency, *port, *thread, *duration, sugar)
	if err != nil {
		sugar.Fatalw("failed to create bufferbloater",
			"error", err)
	}

	bb.Run()

	sugar.Infof("ok %+v", &bb)
}
