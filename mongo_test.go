package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongo(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)

	mongoLogger := NewMongoLogger(testLogger, 10240)

	// Create a client with our logger options.

	clientOptions := options.
		Client().
		ApplyURI("mongodb://rocky:27017").
		SetLoggerOptions(mongoLogger.Options())

	client, err := mongo.Connect(context.Background(), clientOptions)

	require.NoError(t, err)

	defer func() {
		_ = client.Disconnect(context.TODO())
	}()

	// Make a database request to test our logging solution.
	coll := client.Database("test").Collection("test")

	_, err = coll.InsertOne(context.TODO(), bson.D{{"Alice", "123"}})
	require.NoError(t, err)
}

func TestMongoWithMonitor(t *testing.T) {
	testLogger, _ := NewEasyLogger(true, false, ``, `test`)

	mongoLogger := NewMongoLogger(testLogger, 10240)

	// Create a client with our logger options.

	clientOptions := options.
		Client().
		ApplyURI("mongodb://rocky:27017").
		SetMonitor(mongoLogger.CommandMonitor())
	client, err := mongo.Connect(context.Background(), clientOptions)

	require.NoError(t, err)

	defer func() {
		_ = client.Disconnect(context.TODO())
	}()

	// Make a database request to test our logging solution.
	coll := client.Database("test").Collection("test")

	_, err = coll.InsertOne(context.TODO(), bson.D{{"Alice", "123"}})
	require.NoError(t, err)
}
