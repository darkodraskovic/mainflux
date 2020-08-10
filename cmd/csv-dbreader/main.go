// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	r "github.com/go-redis/redis"
	"github.com/mainflux/mainflux/readers/dbreader"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers/dbreader/api"
	csv "github.com/mainflux/mainflux/readers/dbreader/csv"
	"github.com/mainflux/mainflux/readers/dbreader/redis"

	natspub "github.com/mainflux/mainflux/pkg/messaging/nats"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	nats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName     = "csv-dbreader"
	defLogLevel = "debug"
	defHTTPPort = "9205"
	defNatsURL  = nats.DefaultURL

	defESURL          = "localhost:6379"
	defESPass         = ""
	defESDB           = "0"
	defESConsumerName = "dbreader"
	defFSUser         = ""
	defFSPass         = ""

	defCSVConfig = ""

	envHTTPPort = "MF_CSV_HTTP_PORT"
	envLogLevel = "MF_CSV_ADAPTER_LOG_LEVEL"

	envFSUser = "MF_CSV_FS_USER"
	envFSPass = "MF_CSV_FS_PASS"

	envNatsURL = "MF_NATS_URL"

	envESURL          = "MF_THINGS_ES_URL"
	envESPass         = "MF_THINGS_ES_PASS"
	envESDB           = "MF_THINGS_ES_DB"
	envESConsumerName = "MF_CSV_ADAPTER_EVENT_CONSUMER"

	envCSVConfig = "MF_CSV_ADAPTER_CONFIG_FILE"

	readerType = "csv"
)

type config struct {
	httpPort       string
	natsURL        string
	esURL          string
	esPass         string
	esDB           string
	esConsumerName string
	logLevel       string
	cfgFile        string
	FSUser         string
	FSPass         string
}

func main() {

	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	publisher, err := natspub.NewPublisher(cfg.natsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer publisher.Close()

	esConn := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	defer esConn.Close()

	readerManager := dbreader.NewReaderManager(cfg.cfgFile, readerType, csv.New, logger)
	params := map[string]string{csv.FSUser: cfg.FSUser, csv.FSPass: cfg.FSPass}
	readerManager.SetParams(params)

	svc := dbreader.New(publisher, readerManager, logger)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "csv_dbreader",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "csv_dbreader",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	go subscribeToThingsES(svc, esConn, cfg.esConsumerName, logger)

	errs := make(chan error, 2)

	go startHTTPServer(cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	readerManager.LoadAll()
	readerManager.StartAll(svc)

	err = <-errs

	readerManager.SaveAll()

	logger.Error(fmt.Sprintf("CSV reader service terminated: %s", err))
}

func loadConfig() config {
	return config{
		httpPort:       mainflux.Env(envHTTPPort, defHTTPPort),
		natsURL:        mainflux.Env(envNatsURL, defNatsURL),
		logLevel:       mainflux.Env(envLogLevel, defLogLevel),
		esURL:          mainflux.Env(envESURL, defESURL),
		esPass:         mainflux.Env(envESPass, defESPass),
		esDB:           mainflux.Env(envESDB, defESDB),
		esConsumerName: mainflux.Env(envESConsumerName, defESConsumerName),
		cfgFile:        mainflux.Env(envCSVConfig, defCSVConfig),
		FSUser:         mainflux.Env(envFSUser, defFSUser),
		FSPass:         mainflux.Env(envFSPass, defFSPass),
	}
}

func connectToRedis(url, pass, DB string, logger logger.Logger) *r.Client {
	db, err := strconv.Atoi(DB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Connected to redis %s", url))
	return r.NewClient(&r.Options{
		Addr:     url,
		Password: pass,
		DB:       db,
	})
}

func subscribeToThingsES(svc dbreader.Service, client *r.Client, prefix string, logger logger.Logger) {
	eventStore := redis.NewEventStore(svc, client, prefix, logger)
	if err := eventStore.Subscribe("mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Failed to subscribe to Redis event source: %s", err))
	}
}

func startHTTPServer(cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	logger.Info(fmt.Sprintf("CSV reader service started, exposed port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, api.MakeHandler())
}
