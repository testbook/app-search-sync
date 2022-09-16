package plugin

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type MapperPlugin func(*MapperPluginInput) (*MapperPluginOutput, error)

type MapperPluginInput struct {
	Id                interface{}            // original document id
	Data              map[string]interface{} // parsed map from data
	Document          interface{}            // the original document from MongoDB
	Database          string                 // the origin database in MongoDB
	Collection        string                 // the origin collection in MongoDB
	Namespace         string                 // the entire namespace for the original document
	Operation         string                 // "i" for a insert or "u" for update
	CoreMongo         *mongo.Client          // Core MongoDB driver client
	LearnMongo        *mongo.Client          // Learn MongoDB driver client
	EngagementMongo   *mongo.Client          // Engagement MongoDB driver client
	TestMongo         *mongo.Client          // Test MongoDB driver client
	UpdateDescription map[string]interface{} // map describing changes to the document
}

type MapperPluginOutput struct {
	Document        interface{} // an updated document to index into Elasticsearch
	Index           string      // the name of the index to use
	Type            string      // the document type
	Routing         string      // the routing value to use
	Drop            bool        // set to true to indicate that the document should not be indexed but removed
	Passthrough     bool        // set to true to indicate the original document should be indexed unchanged
	Parent          string      // the parent id to use
	Version         int64       // the version of the document
	VersionType     string      // the version type of the document (internal, external, external_gte)
	Pipeline        string      // the pipeline to index with
	RetryOnConflict int         // how many times to retry updates before failing
	Skip            bool        // set to true to indicate the the document should be ignored
	ID              string      // override the _id of the indexed document; not recommended
}
