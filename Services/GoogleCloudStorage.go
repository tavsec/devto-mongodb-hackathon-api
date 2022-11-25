package Services

import (
	"cloud.google.com/go/storage"
	"context"
	"os"
)

var StorageBucket *storage.BucketHandle
var StorageClient *storage.Client

func GoogleCloudStorageInitialize() {
	StorageClient, _ = storage.NewClient(context.TODO())
	StorageBucket = StorageClient.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET"))
}
