package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"plugin"

	"github.com/BurntSushi/toml"
	"github.com/testbook/app-search-sync/client"
	. "github.com/testbook/app-search-sync/plugin"
)

type engineConfig struct {
	Name           string
	Namespace      string // mongo namespace
	ChangeStreamNS string
	DirectReadNS   string
	FunctionName   string // function name within plugins
	Plugin         MapperPlugin
}

type configOptions struct {
	MongoURL                 string      `toml:"mongo-url"`
	MongoOpLogDatabaseName   string      `toml:"mongo-oplog-database-name"`
	MongoOpLogCollectionName string      `toml:"mongo-oplog-collection-name"`
	GtmSettings              gtmSettings `toml:"gtm-settings"`
	ResumeName               string      `toml:"resume-name"`
	Version                  bool
	Verbose                  bool
	Resume                   bool           `toml:"resume"`
	ResumeStrategy           resumeStrategy `toml:"resume-strategy"`
	ResumeWriteUnsafe        bool           `toml:"resume-write-unsafe"`
	ResumeFromTimestamp      int64          `toml:"resume-from-timestamp"`
	Replay                   bool
	ConfigFile               string
	AppSearchURL             string `toml:"app-search-url"`
	AppSearchAPIKey          string `toml:"app-search-api-key"`
	AppSearchClients         int    `toml:"app-search-clients"`
	DirectReads              bool   `toml:"direct-reads"`
	ChangeStreams            bool   `toml:"change-streams"`
	ExitAfterDirectReads     bool   `toml:"exit-after-direct-reads"`
	PluginPath               string `toml:"plugin-path"`
	FlushBufferSize          int    `toml:"flush-buffer-size"`
	EngineConfig             []*engineConfig

	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}

func (config *configOptions) ParseCommandLineFlags() *configOptions {
	flag.StringVar(&config.AppSearchURL, "app-search-url", "", "App search connection URL")
	flag.StringVar(&config.AppSearchAPIKey, "app-search-api-key", "", "App search api key")
	flag.IntVar(&config.AppSearchClients, "app-search-clients", 0, "The number of concurrent app search clients")
	flag.StringVar(&config.MongoURL, "mongo-url", "", "MongoDB connection URL")
	flag.StringVar(&config.MongoOpLogDatabaseName, "mongo-oplog-database-name", "", "Override the database name which contains the mongodb oplog")
	flag.StringVar(&config.MongoOpLogCollectionName, "mongo-oplog-collection-name", "", "Override the collection name which contains the mongodb oplog")
	flag.StringVar(&config.ConfigFile, "f", "", "Location of configuration file")
	flag.BoolVar(&config.Version, "v", false, "True to print the version number")
	flag.BoolVar(&config.Verbose, "verbose", false, "True to output verbose messages")
	flag.BoolVar(&config.Resume, "resume", false, "True to capture the last timestamp of this run and resume on a subsequent run")
	flag.Var(&config.ResumeStrategy, "resume-strategy", "Strategy to use for resuming. 0=timestamp,1=token")
	flag.Int64Var(&config.ResumeFromTimestamp, "resume-from-timestamp", 0, "Timestamp to resume syncing from")
	flag.BoolVar(&config.ResumeWriteUnsafe, "resume-write-unsafe", false, "True to speedup writes of the last timestamp synched for resuming at the cost of error checking")
	flag.BoolVar(&config.Replay, "replay", false, "True to replay all events from the oplog and index them in elasticsearch")
	flag.StringVar(&config.ResumeName, "resume-name", "", "Name under which to load/store the resume state. Defaults to 'default'")
	flag.StringVar(&config.PluginPath, "plugin-path", "", "The file path to a .so file plugin")
	flag.BoolVar(&config.DirectReads, "direct-reads", false, "Set to true to read directly from MongoDB collections")
	flag.BoolVar(&config.ChangeStreams, "change-streams", false, "Set to true to enable change streams for MongoDB 3.6+")
	flag.BoolVar(&config.ExitAfterDirectReads, "exit-after-direct-reads", false, "Set to true to exit after direct reads are complete")
	flag.IntVar(&config.FlushBufferSize, "flush-buffer-size", 10, "After this number of docs the batch is flushed to appsearch")
	flag.Parse()
	return config
}

func (config *configOptions) LoadConfigFile() *configOptions {
	if config.ConfigFile != "" {
		var tomlConfig configOptions = configOptions{
			GtmSettings: GtmDefaultSettings(),
		}
		if _, err := toml.DecodeFile(config.ConfigFile, &tomlConfig); err != nil {
			panic(err)
		}
		if config.AppSearchURL == "" {
			config.AppSearchURL = tomlConfig.AppSearchURL
		}
		if config.AppSearchAPIKey == "" {
			config.AppSearchAPIKey = tomlConfig.AppSearchAPIKey
		}
		if config.MongoURL == "" {
			config.MongoURL = tomlConfig.MongoURL
		}
		if config.MongoOpLogDatabaseName == "" {
			config.MongoOpLogDatabaseName = tomlConfig.MongoOpLogDatabaseName
		}
		if config.MongoOpLogCollectionName == "" {
			config.MongoOpLogCollectionName = tomlConfig.MongoOpLogCollectionName
		}
		if !config.Verbose && tomlConfig.Verbose {
			config.Verbose = true
		}
		if !config.Replay && tomlConfig.Replay {
			config.Replay = true
		}
		if !config.DirectReads && tomlConfig.DirectReads {
			config.DirectReads = true
		}
		if !config.ChangeStreams && tomlConfig.ChangeStreams {
			config.ChangeStreams = true
		}
		if !config.ExitAfterDirectReads && tomlConfig.ExitAfterDirectReads {
			config.ExitAfterDirectReads = true
		}
		if !config.Resume && tomlConfig.Resume {
			config.Resume = true
		}
		if config.ResumeStrategy == 0 {
			config.ResumeStrategy = tomlConfig.ResumeStrategy
		}
		if !config.ResumeWriteUnsafe && tomlConfig.ResumeWriteUnsafe {
			config.ResumeWriteUnsafe = true
		}
		if config.ResumeFromTimestamp == 0 {
			config.ResumeFromTimestamp = tomlConfig.ResumeFromTimestamp
		}
		if config.Resume && config.ResumeName == "" {
			config.ResumeName = tomlConfig.ResumeName
		}
		if config.PluginPath == "" {
			config.PluginPath = tomlConfig.PluginPath
		}
		if config.FlushBufferSize == 0 {
			config.FlushBufferSize = tomlConfig.FlushBufferSize
		}
		config.GtmSettings = tomlConfig.GtmSettings
		config.EngineConfig = tomlConfig.EngineConfig
	}
	return config
}

func (config *configOptions) SetDefaults() *configOptions {
	if config.MongoURL == "" {
		config.MongoURL = mongoUrlDefault
	}
	if config.ResumeName == "" {
		config.ResumeName = resumeNameDefault
	}
	if config.FlushBufferSize == 0 {
		config.FlushBufferSize = indexClientBufferDefault
	}
	if config.InfoLogger == nil {
		config.InfoLogger = log.New(os.Stdout, "INFO ", log.Flags())
	}
	if config.ErrorLogger == nil {
		config.ErrorLogger = log.New(os.Stdout, "ERROR ", log.Flags())
	}
	return config
}

func (config *configOptions) LoadPlugin() *configOptions {
	if config.PluginPath == "" {
		if config.Verbose {
			config.InfoLogger.Println("no plugins detected")
		}
		return config
	}
	p, err := plugin.Open(config.PluginPath)
	if err != nil {
		config.ErrorLogger.Fatalf("Unable to load plugin <%s>: %s", config.PluginPath, err)
	}

	for _, m := range config.EngineConfig {
		if m.FunctionName == "" {
			continue
		}
		f, err := p.Lookup(m.FunctionName)
		if err != nil {
			config.ErrorLogger.Fatalf("Unable to lookup symbol <%s> for plugin <%s>: %s", m.FunctionName, config.PluginPath, err)
		}

		switch f := f.(type) {
		case func(*MapperPluginInput) (*MapperPluginOutput, error):
			m.Plugin = f
		default:
			config.ErrorLogger.Fatalf("Plugin symbol <%s> must be typed %T", m.FunctionName, m.Plugin)
		}
	}
	if config.Verbose {
		config.InfoLogger.Printf("plugin <%s> loaded succesfully\n", config.PluginPath)
	}
	return config
}

func (config *configOptions) GetHTTPConfig() client.HTTPConfig {
	httpConfig := client.HTTPConfig{
		Addr:      config.AppSearchURL,
		UserAgent: fmt.Sprintf("%s v%s", Name, Version),
		APIKey:    config.AppSearchAPIKey,
	}
	return httpConfig
}
