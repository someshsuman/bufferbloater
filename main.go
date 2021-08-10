package main

import (
	"flag"

	"go.uber.org/zap"
)

func main() {

	workload := flag.String("workload", "90:10s,5:20s,4:100s,1:150s", "RPS1:duration1,RPS2:duration...")
	rqTimeout := flag.String("request-timeout", "5s", "Request timeout duration in sec")
	targetServer := flag.String("target", "0.0.0.0:19002", "ServerAddress:Port")
	flag.Parse()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("couldn't initialize logging")
	}
	sugar := logger.Sugar()

	bb, err := NewBufferbloater(*workload, *rqTimeout, *targetServer, sugar)
	if err != nil {
		sugar.Fatalw("failed to create bufferbloater",
			"error", err)
	}

	bb.Run()

	sugar.Infof("ok %+v", &bb)

}
