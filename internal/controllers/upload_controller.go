package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strings"

	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// uploadDir is where uploaded images are stored and served from (/uploads/...).
const uploadDir = "uploads"

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}

type UploadController struct{}

func NewUploadController() *UploadController { return &UploadController{} }

// Upload godoc
// @Summary  Upload an image, returns its URL
// @Tags     Uploads
// @Accept   multipart/form-data
// @Produce  json
// @Security BearerAuth
// @Param    file  formData  file  true  "Image file (jpg/png/webp/gif)"
// @Success  200   {object}  map[string]interface{}
// @Router   /uploads [post]
func (ctrl *UploadController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		response.BadRequest(c, "Only image files are allowed (jpg, png, webp, gif)", nil)
		return
	}
	if file.Size > 5<<20 { // 5 MB
		response.BadRequest(c, "File too large (max 5MB)", nil)
		return
	}

	// Random, collision-free filename.
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	name := hex.EncodeToString(b) + ext

	if err := c.SaveUploadedFile(file, filepath.Join(uploadDir, name)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save file", nil)
		return
	}

	response.Success(c, "Uploaded", gin.H{"url": "/" + uploadDir + "/" + name})
}
