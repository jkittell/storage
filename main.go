package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	router := gin.New()
	fileHandler := NewFileHandler(".env")

	router.POST("/file", fileHandler.Upload)
	router.GET("/file/:id", fileHandler.Get)
	router.GET("/file", fileHandler.List)
	router.GET("/file/download/:id", fileHandler.Download)
	router.DELETE("/file/:id", fileHandler.Delete)

	err := router.Run(":80")
	if err != nil {
		log.Fatal(err)
	}
}
