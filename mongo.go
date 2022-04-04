package main

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/rwynn/gtm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func saveTokens(client *mongo.Client, tokens bson.M, config *configOptions) error {
	var err error
	if len(tokens) == 0 {
		return err
	}
	col := client.Database(Name).Collection("tokens")
	bwo := options.BulkWrite().SetOrdered(false)
	var models []mongo.WriteModel
	for streamID, token := range tokens {
		filter := bson.M{
			"resumeName": config.ResumeName,
			"streamID":   streamID,
		}
		update := bson.M{"$set": bson.M{
			"resumeName": config.ResumeName,
			"streamID":   streamID,
			"token":      token,
		}}
		model := mongo.NewUpdateManyModel()
		model.SetUpsert(true)
		model.SetFilter(filter)
		model.SetUpdate(update)
		models = append(models, model)
	}
	_, err = col.BulkWrite(context.Background(), models, bwo)
	return err
}

func cleanMongoURL(URL string) string {
	const (
		redact    = "REDACTED"
		scheme    = "mongodb://"
		schemeSrv = "mongodb+srv://"
	)
	url := URL
	hasScheme := strings.HasPrefix(url, scheme)
	hasSchemeSrv := strings.HasPrefix(url, schemeSrv)
	url = strings.TrimPrefix(url, scheme)
	url = strings.TrimPrefix(url, schemeSrv)
	userEnd := strings.IndexAny(url, "@")
	if userEnd != -1 {
		url = redact + "@" + url[userEnd+1:]
	}
	if hasScheme {
		url = scheme + url
	} else if hasSchemeSrv {
		url = schemeSrv + url
	}
	return url
}

func (config *configOptions) DialMongo() (*mongo.Client, error) {
	rb := bson.NewRegistryBuilder()
	rb.RegisterTypeMapEntry(bsontype.DateTime, reflect.TypeOf(time.Time{}))
	reg := rb.Build()
	clientOptions := options.Client()
	clientOptions.ApplyURI(config.MongoURL)
	clientOptions.SetAppName(Name)
	clientOptions.SetRegistry(reg)
	if config.Resume && config.ResumeWriteUnsafe {
		clientOptions.SetWriteConcern(writeconcern.New(writeconcern.W(0), writeconcern.J(false)))
	}
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	mongoOk := make(chan bool)
	go config.cancelConnection(mongoOk)
	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	close(mongoOk)
	return client, nil
}

func (config *configOptions) cancelConnection(mongoOk chan bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer signal.Stop(sigs)
	select {
	case <-mongoOk:
		return
	case <-sigs:
		os.Exit(exitStatus)
	}
}

func saveTimestamp(client *mongo.Client, ts primitive.Timestamp, config *configOptions) error {
	col := client.Database(Name).Collection("resume")
	doc := map[string]interface{}{
		"ts": ts,
	}
	opts := options.Update()
	opts.SetUpsert(true)
	_, err := col.UpdateOne(context.Background(), bson.M{
		"_id": config.ResumeName,
	}, bson.M{
		"$set": doc,
	}, opts)
	return err
}

func saveTimestampFromReplStatus(client *mongo.Client, config *configOptions) {
	if rs, err := gtm.GetReplStatus(client); err == nil {
		var ts primitive.Timestamp
		if ts, err = rs.GetLastCommitted(); err == nil {
			saveTimestamp(client, ts, config)
		}
	}
}
