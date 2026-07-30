package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedImageExts restricts uploads to common image formats.
var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

const maxUploadSize = 5 << 20 // 5 MB

// UploadImage godoc
// POST /api/v1/upload (protected)
// Accepts a multipart/form-data "image" field, saves it under ./uploads,
// and returns a URL that can be stored on a Product/Category's image_url field.
//
// NOTE: This is Day 1/2 "structure ready" — local disk storage is fine for
// development. Before production, swap SaveUploadedFile for an S3/GCS/Cloud
// Storage client so images survive redeploys and scale past a single server.
func UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided (expected form field 'image')"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type. Allowed: jpg, jpeg, png, webp"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"image_url": "/uploads/" + filename,
	})
}
