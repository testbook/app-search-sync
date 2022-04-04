package main

import (
	"context"
	"time"

	"github.com/rwynn/gtm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type gtmSettings struct {
	ChannelSize    int    `toml:"channel-size"`
	BufferSize     int    `toml:"buffer-size"`
	BufferDuration string `toml:"buffer-duration"`
}

func GtmDefaultSettings() gtmSettings {
	return gtmSettings{
		ChannelSize:    gtmChannelSizeDefault,
		BufferSize:     32,
		BufferDuration: "75ms",
	}
}

func notAppSearchSync() gtm.OpFilter {
	return func(op *gtm.Op) bool {
		return op.GetDatabase() != Name
	}
}

func isInsertOrUpdate(op *gtm.Op) bool {
	return op.IsInsert() || op.IsUpdate()
}

func (config *configOptions) onlyMeasured() gtm.OpFilter {
	if config.ChangeStreams {
		return func(op *gtm.Op) bool {
			return true
		}
	}
	measured := make(map[string]bool)
	for _, m := range config.EngineConfig {
		measured[m.Namespace] = true
		if m.Namespace != "" {
			measured[m.Namespace] = true
		}
	}
	return func(op *gtm.Op) bool {
		return measured[op.Namespace]
	}
}

func (config *configOptions) getDirectReadNSList() []string {
	if !config.DirectReads {
		return nil
	}

	directReadNSList := make([]string, 0)

	for _, m := range config.EngineConfig {
		if m.DirectReadNS != "" {
			directReadNSList = append(directReadNSList, m.Namespace)
		}
	}
	return directReadNSList
}

func (config *configOptions) getChangeStreamNSList() []string {
	changeStreamNSList := make([]string, 0)

	for _, m := range config.EngineConfig {
		if m.ChangeStreamNS != "" {
			changeStreamNSList = append(changeStreamNSList, m.Namespace)
		}
	}
	return changeStreamNSList
}

func (config *configOptions) getTimestampGen() (after gtm.TimestampGenerator) {
	if config.ResumeStrategy != timestampResumeStrategy {
		return after
	}

	if config.Replay {
		after = func(client *mongo.Client, options *gtm.Options) (primitive.Timestamp, error) {
			return primitive.Timestamp{}, nil
		}
	} else if config.ResumeFromTimestamp != 0 {
		after = func(client *mongo.Client, options *gtm.Options) (primitive.Timestamp, error) {
			return primitive.Timestamp{
				T: uint32(config.ResumeFromTimestamp),
				I: 1,
			}, nil
		}
	} else if config.Resume {
		after = func(client *mongo.Client, options *gtm.Options) (primitive.Timestamp, error) {
			var ts primitive.Timestamp
			col := client.Database(Name).Collection("resume")
			result := col.FindOne(context.Background(), bson.M{
				"_id": config.ResumeName,
			})
			if err := result.Err(); err == nil {
				doc := make(map[string]interface{})
				if err = result.Decode(&doc); err == nil {
					if doc["ts"] != nil {
						ts = doc["ts"].(primitive.Timestamp)
						ts.I += 1
					}
				}
			}
			if ts.T == 0 {
				ts, _ = gtm.LastOpTimestamp(client, options)
			}
			config.ErrorLogger.Printf("Resuming from timestamp %+v", ts)
			return ts, nil
		}
	}
	return
}

func (config *configOptions) buildTokenGen() gtm.ResumeTokenGenenerator {
	var token gtm.ResumeTokenGenenerator
	if !config.Resume || (config.ResumeStrategy != tokenResumeStrategy) {
		return token
	}

	token = func(client *mongo.Client, streamID string, options *gtm.Options) (interface{}, error) {
		var t interface{} = nil
		var err error
		col := client.Database(Name).Collection("tokens")
		result := col.FindOne(context.Background(), bson.M{
			"resumeName": config.ResumeName,
			"streamID":   streamID,
		})
		if err = result.Err(); err == nil {
			doc := make(map[string]interface{})
			if err = result.Decode(&doc); err == nil {
				t = doc["token"]
				if t != nil {
					config.InfoLogger.Printf("Resuming stream '%s' from collection %s.tokens using resume name '%s'",
						streamID, Name, config.ResumeName)
				}
			}
		}
		return t, err
	}
	return token
}

func (config *configOptions) buildGtmOptions() *gtm.Options {
	var nsFilter, filter, directReadFilter gtm.OpFilter

	filterChain := []gtm.OpFilter{notAppSearchSync(), config.onlyMeasured(), isInsertOrUpdate}
	filter = gtm.ChainOpFilters(filterChain...)
	bufferDuration, err := time.ParseDuration(config.GtmSettings.BufferDuration)
	if err != nil {
		config.ErrorLogger.Fatalf("Unable to parse gtm buffer duration %s: %s", config.GtmSettings.BufferDuration, err)
	}

	after := config.getTimestampGen()
	token := config.buildTokenGen()

	gtmOpts := &gtm.Options{
		After:               after,
		Token:               token,
		Filter:              filter,
		NamespaceFilter:     nsFilter,
		OpLogDisabled:       len(config.getDirectReadNSList()) > 0,
		OpLogDatabaseName:   config.MongoOpLogDatabaseName,
		OpLogCollectionName: config.MongoOpLogCollectionName, // oplog.rs
		ChannelSize:         config.GtmSettings.ChannelSize,
		Ordering:            gtm.AnyOrder,
		WorkerCount:         4,
		BufferDuration:      bufferDuration,
		BufferSize:          config.GtmSettings.BufferSize,
		DirectReadNs:        config.getDirectReadNSList(),
		DirectReadFilter:    directReadFilter,
		Log:                 config.InfoLogger,
		ChangeStreamNs:      config.getChangeStreamNSList(),
	}
	return gtmOpts
}
