package Controllers

import (
	storage "cloud.google.com/go/storage"
	video "cloud.google.com/go/videointelligence/apiv1"
	videopb "cloud.google.com/go/videointelligence/apiv1/videointelligencepb"
	"context"
	"github.com/gin-gonic/gin"
	pag "github.com/gobeam/mongo-go-pagination"
	"github.com/google/uuid"
	"github.com/tavsec/devto-mongodb-hackathon/Services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func VideoStore(c *gin.Context) {
	ctx := context.Background()
	client, err := video.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func(client *video.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	fileUuid := uuid.New()

	storageClient, _ := storage.NewClient(ctx)

	f, uploadedFile, _ := c.Request.FormFile("video")

	if filepath.Ext(uploadedFile.Filename) != ".mp4" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Uploaded file must be mp4 video",
		})
		return
	}

	sw := storageClient.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(fileUuid.String()).NewWriter(ctx)

	_, err = io.Copy(sw, f)
	err = sw.Close()
	if err != nil {
		return
	}

	coll := Services.MongoClient.Database(os.Getenv("MONGO_DB_DATABASE")).Collection("videos")
	doc := bson.D{{"filename", uploadedFile.Filename}, {"size", uploadedFile.Size}, {"uuid", fileUuid.String()}}
	mongoRecord, err := coll.InsertOne(context.TODO(), doc)

	if err != nil {
		panic(err)
	}

	op, err := client.AnnotateVideo(ctx, &videopb.AnnotateVideoRequest{
		InputUri: "gs://" + os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET") + "/" + fileUuid.String(),
		Features: []videopb.Feature{
			videopb.Feature_LABEL_DETECTION,
		},
	})
	if err != nil {
		log.Fatalf("Failed to start annotation job: %v", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to annotate: %v", err)
	}

	result := resp.GetAnnotationResults()[0]
	var featuresList []Feature
	for _, annotation := range result.SegmentLabelAnnotations {

		var timestamps []Timestamp

		for _, segment := range annotation.Segments {
			start := segment.Segment.StartTimeOffset.AsDuration()
			end := segment.Segment.EndTimeOffset.AsDuration()
			timestamps = append(timestamps, Timestamp{
				Start: start.Milliseconds(),
				End:   end.Milliseconds(),
			})

		}

		f := Feature{Description: annotation.Entity.Description, Timestamps: timestamps}
		featuresList = append(featuresList, f)

	}

	filter := bson.D{{"_id", mongoRecord.InsertedID}}
	update := bson.D{{"$set", bson.D{{"features", featuresList}}}}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"features": featuresList,
	})
}

type SearchBody struct {
	Keyword string `json:"keyword"`
	Page    int64  `json:"page"`
	PerPage int64  `json:"perPage"`
}

type Video struct {
	Id        primitive.ObjectID `bson:"_id"`
	Filename  string             `bson:"filename"`
	Size      int                `bson:"size"`
	UUID      string             `bson:"uuid"`
	Features  []Feature          `bson:"features"`
	SignedURL string
}

type Timestamp struct {
	Start int64 `bson:"start"`
	End   int64 `bson:"end"`
}

type Feature struct {
	Description string      `bson:"description"`
	Timestamps  []Timestamp `bson:"timestamps"`
}

type PaginatedResult struct {
	Videos     []Video
	Pagination pag.PaginationData
}

func VideoSearch(c *gin.Context) {
	var searchBody SearchBody
	err := c.ShouldBindJSON(&searchBody)
	if err != nil {
		err = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	collection := Services.MongoClient.Database(os.Getenv("MONGO_DB_DATABASE")).Collection("videos")

	searchStage := bson.M{"$search": bson.D{{"index", os.Getenv("MONGO_DB_SEARCH_INDEX")}, {"text", bson.D{{"path", bson.D{{"wildcard", "*"}}}, {"query", searchBody.Keyword}}}}}

	aggPaginatedData, err := pag.New(collection).Context(context.TODO()).Limit(searchBody.PerPage).Page(searchBody.Page).Aggregate(searchStage)
	if err != nil {
		panic(err)
	}

	storageClient, _ := storage.NewClient(context.Background())

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}

	var results []Video
	for _, raw := range aggPaginatedData.Data {
		var v *Video
		if marshallErr := bson.Unmarshal(raw, &v); marshallErr == nil {
			v.SignedURL, _ = storageClient.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).SignedURL(v.UUID, opts)
			results = append(results, *v)
		}

	}

	c.JSON(http.StatusOK, PaginatedResult{Videos: results, Pagination: aggPaginatedData.Pagination})

}
