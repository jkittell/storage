package main

import (
	"github.com/google/uuid"
	"time"
)

type FileEntry struct {
	Id        uuid.UUID `bson:"_id"`
	Name      string    `bson:"name"`
	Size      int64     `bson:"size"`
	CreatedAt time.Time `bson:"created_at"`
}
