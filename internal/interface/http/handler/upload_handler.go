package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arabella/ai-studio-backend/internal/interface/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler handles file uploads
type UploadHandler struct{}

// NewUploadHandler creates a new UploadHandler
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadImage handles image uploads
// @Summary Upload image
// @Description Upload an image file (for thumbnails, etc.)
// @Tags admin
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /admin/upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Verify authentication
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Get the file from form data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "No file provided",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/octet-stream": true, // Some browsers send this
	}

	contentType := file.Header.Get("Content-Type")
	if !allowedTypes[contentType] && !strings.HasSuffix(strings.ToLower(file.Filename), ".jpg") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".jpeg") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".png") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".gif") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".webp") {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid file type. Only images are allowed (jpg, jpeg, png, gif, webp)",
			Code:    "INVALID_FILE_TYPE",
			Details: fmt.Sprintf("Content-Type: %s", contentType),
		})
		return
	}

	// Validate file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "File too large. Maximum size is 10MB",
			Code:    "FILE_TOO_LARGE",
			Details: fmt.Sprintf("File size: %d bytes", file.Size),
		})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create upload directory",
			Code:    "INTERNAL_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		// Default to .jpg if no extension
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to save file",
			Code:    "INTERNAL_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Verify the file was saved and is readable
	fileInfo, err := os.Stat(filePath)
	if err != nil || fileInfo.Size() == 0 {
		os.Remove(filePath) // Clean up
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to verify uploaded file",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Return the URL to access the uploaded file
	// Use the server base URL from config or construct from request
	baseURL := c.GetHeader("Origin")
	if baseURL == "" {
		// Fallback to constructing from request
		scheme := "https"
		if c.GetHeader("X-Forwarded-Proto") == "http" || c.Request.TLS == nil {
			scheme = "http"
		}
		host := c.GetHeader("Host")
		if host == "" {
			host = c.Request.Host
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, host)
	}

	// Remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	imageURL := fmt.Sprintf("%s/uploads/%s", baseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      imageURL,
		"filename": filename,
		"size":     fileInfo.Size(),
		"uploaded_at": time.Now().UTC().Format(time.RFC3339),
	})
}
