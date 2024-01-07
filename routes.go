package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jkittell/data/database"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var volume = "/data/files"

type FileHandler struct {
	files database.MongoDB[FileEntry]
}

func NewFileHandler() FileHandler {
	db, err := database.NewMongoDB[FileEntry]("files")
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "id": val})
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h FileHandler) List(c *gin.Context) {
	files, err := h.files.All(c, nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

type Form struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (h FileHandler) Upload(c *gin.Context) {
	var form Form
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	id := uuid.New()
	filePath := filepath.Join(volume, id.String(), form.File.Filename)
	if err = c.SaveUploadedFile(form.File, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileEntry := FileEntry{
		Id:        id,
		Name:      form.File.Filename,
		Size:      form.File.Size,
		CreatedAt: time.Now(),
	}
	if err = h.files.Insert(c, fileEntry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a success message and the file metadata
	c.JSON(http.StatusOK, fileEntry)
}

func (h FileHandler) Download(c *gin.Context) {
	id := c.Param("id")
	var file FileEntry
	val, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "id": val})
		return
	}

	filePath := filepath.Join(volume, id, file.Name)

	fileData, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer fileData.Close()

	fileHeader := make([]byte, 512)
	_, err = fileData.Read(fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fileContentType := http.DetectContentType(fileHeader)
	fileInfo, err := fileData.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	file, err = h.files.FindByID(c, val)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "id": val})
		return
	}

	filePath := filepath.Join(volume, id)

	err = os.RemoveAll(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.files.Delete(c, bson.D{{Key: "_id", Value: val}}, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file": file,
	})
}
