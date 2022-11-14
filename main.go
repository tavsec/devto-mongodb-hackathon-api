package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tavsec/devto-mongodb-hackathon/Controllers"
	"github.com/tavsec/devto-mongodb-hackathon/Services"
	"log"
)

func init() {
	Services.MongoDBInitialize()
}

func main() {
	r := gin.Default()
	r.POST("/videos", Controllers.VideoStore)
	r.GET("/videos", Controllers.VideoSearch)

	err := r.Run()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	fmt.Println("Server running on :8080")
}
