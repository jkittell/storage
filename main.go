package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	router := gin.New()
	fileHandler := NewFileHandler()

	router.POST("/files", fileHandler.Upload)
	router.GET("/files/:id", fileHandler.Get)
	router.GET("/files", fileHandler.List)
	router.GET("/files/:id/download", fileHandler.Download)
	router.DELETE("/files/:id", fileHandler.Delete)

	err := router.Run(":80")
	if err != nil {
		log.Fatal(err)
	}
}
