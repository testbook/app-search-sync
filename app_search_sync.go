package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rwynn/gtm"
	client "github.com/testbook/app-search-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	Name                     = "app-search-sync"
	Version                  = "1.0.0"
	mongoUrlDefault          = "mongodb://localhost:27017"
	indexClientsDefault      = 10
	indexClientBufferDefault = 10
	resumeNameDefault        = "default"
	defaultHttpAddr          = ":8010"
	gtmChannelSizeDefault    = 512
	defaultConfigFile        = "config.go"
)

func main() {
	config := &configOptions{
		GtmSettings: GtmDefaultSettings(),
		InfoLogger:  log.New(os.Stdout, "INFO ", log.Flags()),
		ErrorLogger: log.New(os.Stdout, "ERROR ", log.Flags()),
	}
	config.ParseCommandLineFlags()
	config.LoadConfigFile().SetDefaults().LoadPlugin()

	if len(config.EngineConfig) == 0 {
		config.ErrorLogger.Fatalln("No engine configuration found")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	coreMongo, engagementMongo, testMongo, err := config.DialMongo()
	if err != nil {
		config.ErrorLogger.Fatalf("Unable to connect to mongodb: %s", err)
	}
	defer coreMongo.Disconnect(context.Background())
	defer engagementMongo.Disconnect(context.Background())
	defer testMongo.Disconnect(context.Background())

	client, err := client.NewHTTPClient(config.GetHTTPConfig())
	if err != nil {
		config.ErrorLogger.Fatalf("Unable to create client: %s", err)
	}
	defer client.Close()

	gtmCtx := gtm.StartMulti([]*mongo.Client{coreMongo, testMongo}, config.buildGtmOptions())
	defer gtmCtx.Stop()
	ic := &indexClient{
		indexMutex:      &sync.Mutex{},
		tokens:          bson.M{},
		client:          client,
		config:          config,
		gtmCtx:          gtmCtx,
		coreMongo:       coreMongo,
		engagementMongo: engagementMongo,
		testMongo:       testMongo,
		stats: &bulkProcessorStats{
			Enabled: config.Stats,
		},
	}
	if err = ic.setupEngines(); err != nil {
		config.ErrorLogger.Fatalf("Error to setup engines: %s", err)
	}
	ic.start()
	go startHTTPServer(&httpServerCtx{indexConfig: ic})

	<-c
	ic.config.InfoLogger.Println("Stopping all workers and shutting down")
	ic.batchIndex()
}
