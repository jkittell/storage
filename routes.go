package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jkittell/data/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var volume = "/data/files"

type FileHandler struct {
	files database.MongoDB[FileEntry]
}

func NewFileHandler(env string) FileHandler {
	db, err := database.NewMongoDB[FileEntry](env, "files")
	if err != nil {
		log.Fatal(err)
	}
	return FileHandler{files: db}
}

func (h FileHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var file FileEntry
	val, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error(), "id": val})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "file": file})
}

func (h FileHandler) List(c *gin.Context) {
	files, err := h.files.All(c, nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "files": files})
}

func (h FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// use uuid folder to prevent file name collisions
	id := uuid.New()
	filePath := filepath.Join(volume, id.String(), file.Filename)

	if err = c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	fileEntry := FileEntry{
		Id:        id,
		Name:      file.Filename,
		Size:      file.Size,
		CreatedAt: time.Now(),
	}
	if err = h.files.Insert(c, fileEntry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Return a success message and the file metadata
	c.JSON(http.StatusOK, gin.H{"success": true, "file": fileEntry})
}

func (h FileHandler) Download(c *gin.Context) {
	id := c.Param("id")
	var file FileEntry
	val, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error(), "id": val})
		return
	}

	filePath := filepath.Join(volume, id, file.Name)

	fileData, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer fileData.Close()

	fileHeader := make([]byte, 512)
	_, err = fileData.Read(fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	fileContentType := http.DetectContentType(fileHeader)
	fileInfo, err := fileData.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	c.Header("Content-Type", fileContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.File(filePath)
}

func (h FileHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var file FileEntry

	val, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error(), "id": val})
		return
	}

	filePath := filepath.Join(volume, id)

	err = os.RemoveAll(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	err = h.files.Delete(c, bson.D{{Key: "_id", Value: val}}, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true, "file": file,
	})
}
