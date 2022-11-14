package Services

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var MongoClient *mongo.Client

func MongoDBInitialize() {
	uri := os.Getenv("MONGODB_URI")

	var err error

	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

}
