package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rwynn/gtm"
	"github.com/testbook/app-search-sync/client"
	"github.com/testbook/app-search-sync/plugin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type indexEngineCtx struct {
	name      string
	namespace string
	docs      []interface{}
	plugin    plugin.MapperPlugin
}

type indexClient struct {
	gtmCtx       *gtm.OpCtxMulti
	config       *configOptions
	mongo        *mongo.Client
	client       client.Client
	indexWg      *sync.WaitGroup
	indexMutex   *sync.Mutex
	indexC       chan *gtm.Op
	lastTs       primitive.Timestamp
	tokens       bson.M
	lastUpdateTs time.Time
	engines      map[string]*indexEngineCtx
}

type dbcol struct {
	db  string
	col string
}

func (ic *indexClient) parseNamespace(namespace string) (*dbcol, error) {
	dbCol := strings.SplitN(namespace, ".", 2)
	if len(dbCol) != 2 {
		return nil, fmt.Errorf("view namespace is invalid: %s", namespace)
	}
	d := &dbcol{
		db:  dbCol[0],
		col: dbCol[1],
	}
	return d, nil
}

func (ic *indexClient) setupEngines() error {
	if len(ic.config.EngineConfig) == 0 {
		return fmt.Errorf("no engine config found")
	}

	ic.engines = make(map[string]*indexEngineCtx)
	for _, engine := range ic.config.EngineConfig {
		ic.engines[engine.Namespace] = &indexEngineCtx{
			name:      engine.Name,
			namespace: engine.Namespace,
			plugin:    engine.Plugin,
		}
	}
	return nil
}

func (ic *indexClient) batchIndex() (err error) {
	ic.indexMutex.Lock()
	defer ic.indexMutex.Unlock()

	docs := 0
	for idx, e := range ic.engines {
		if len(e.docs) == 0 {
			continue
		}

		docs += len(e.docs)
		if err = ic.client.Index(e.name, e.docs); err != nil {
			break
		}
		ic.engines[idx].docs = []interface{}{}
	}

	if ic.config.Verbose {
		if docs > 0 {
			ic.config.InfoLogger.Printf("%d docs flushed\n", docs)
		}
	}
	ic.lastUpdateTs = time.Now()
	return
}

func (ic *indexClient) saveTs() (err error) {
	if !(ic.config.Resume && ic.lastTs.T > 0) {
		return err
	}

	if err = ic.batchIndex(); err != nil {
		return err
	}
	if ic.config.ResumeStrategy == tokenResumeStrategy {
		err = saveTokens(ic.mongo, ic.tokens, ic.config)
		if err == nil {
			ic.tokens = bson.M{}
		}
	} else {
		err = saveTimestamp(ic.mongo, ic.lastTs, ic.config)
	}

	ic.lastTs = primitive.Timestamp{}
	return err
}

func (ic *indexClient) lookupInView(orig *gtm.Op, namespace string) (op *gtm.Op, err error) {
	v, err := ic.parseNamespace(namespace)
	if err != nil {
		return nil, err
	}

	col := ic.mongo.Database(v.db).Collection(v.col)
	result := col.FindOne(context.Background(), bson.M{
		"_id": orig.Id,
	})
	if err = result.Err(); err == nil {
		doc := make(map[string]interface{})
		if err = result.Decode(&doc); err == nil {
			op = &gtm.Op{
				Id:        orig.Id,
				Data:      doc,
				Operation: orig.Operation,
				Namespace: namespace,
				Source:    gtm.DirectQuerySource,
				Timestamp: orig.Timestamp,
			}
		}
	}
	return
}

func (ic *indexClient) addDocument(op *gtm.Op) error {
	engine := ic.engines[op.Namespace]
	if engine == nil {
		return nil
	}
	if engine.namespace != "" && op.IsSourceOplog() {
		var err error
		op, err = ic.lookupInView(op, engine.namespace) // fetch from mongo
		if err != nil {
			return err
		}
	}

	if engine.plugin != nil {
		inp := &plugin.MapperPluginInput{
			Document:    op.Doc,
			Data:        op.Data,
			Database:    op.GetDatabase(),
			Collection:  op.GetCollection(),
			Operation:   op.Operation,
			Namespace:   op.Namespace,
			MongoClient: ic.mongo,
		}
		upd, err := engine.plugin(inp)
		if err != nil {
			return err
		}
		if upd.Skip {
			return nil
		}
		op.Doc = upd.Document
	}
	engine.docs = append(engine.docs, op.Doc)

	if op.IsSourceOplog() {
		ic.lastTs = op.Timestamp
		if ic.config.ResumeStrategy == tokenResumeStrategy {
			ic.tokens[op.ResumeToken.StreamID] = op.ResumeToken.ResumeToken
		}
	}

	if len(engine.docs) >= ic.config.FlushBufferSize {
		if err := ic.batchIndex(); err != nil {
			return err
		}
	}
	return nil
}

func (ic *indexClient) index() {
	for {
		select {
		case err := <-ic.gtmCtx.ErrC:
			if err == nil {
				break
			}
			exitStatus = 1
			ic.config.ErrorLogger.Println(err)

		case op, open := <-ic.gtmCtx.OpC:
			if op == nil {
				if !open {
					if err := ic.saveTs(); err != nil {
						exitStatus = 1
						ic.config.ErrorLogger.Println(err)
					}
					return
				}
				break
			}
			if err := ic.addDocument(op); err != nil {
				exitStatus = 1
				ic.config.ErrorLogger.Println(err)
			}
		}
	}
}

func (ic *indexClient) directReads() {
	directReadsFunc := func() {
		ic.gtmCtx.DirectReadWg.Wait()
		ic.config.InfoLogger.Println("Direct reads completed")

		if ic.config.Resume && ic.config.ResumeStrategy == timestampResumeStrategy {
			saveTimestampFromReplStatus(ic.mongo, ic.config)
		}
		if ic.config.ExitAfterDirectReads {
			ic.gtmCtx.Stop()
		}
	}
	if ic.config.DirectReads {
		go directReadsFunc()
	}
}

func (ic *indexClient) startIndex() {
	for i := 0; i < ic.config.AppSearchClients; i += 1 {
		go ic.index()
	}
}

func (ic *indexClient) flusher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C // Periodic flush

		ic.config.InfoLogger.Println("batchingIndex from ticker")
		if err := ic.batchIndex(); err != nil {
			ic.config.ErrorLogger.Println("err in flusher", err)
		}
	}
}

func (ic *indexClient) startFlusher() {
	if int64(ic.config.FlushInterval) > 0 {
		go ic.flusher(time.Second * time.Duration(ic.config.FlushInterval))
	}
}

func (ic *indexClient) start() {
	ic.startIndex()
	ic.startFlusher()
	ic.directReads()
}
