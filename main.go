package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type URLMapping struct {
	ID        uint   `gorm:"primaryKey"`
	ShortCode string `gorm:"uniqueIndex;size:16"`
	LongURL   string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	ExpiresAt *time.Time // pointer so it can be nil (no expiry)
}

type URLClick struct {
	ID        uint `gorm:"primaryKey"`
	MappingID uint `gorm:"index"` // which URLMapping was clicked
	IPAddress string
	UserAgent string
	ClickedAt time.Time
}

type ShortenRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

var db *gorm.DB

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateShortCode(length int) (string, error) {
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		result[i] = base62Chars[num.Int64()]
	}

	return string(result), nil
}

func shortenHandler(c *gin.Context) {
	var req ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var existing URLMapping
	result := db.Where("long_url = ?", req.LongURL).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"short_url":  os.Getenv("BASE_URL") + "/" + existing.ShortCode,
			"short_code": existing.ShortCode,
		})
		return
	}

	var code string
	for {
		generated, err := generateShortCode(6)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate code"})
			return
		}

		var clash URLMapping
		if err := db.Where("short_code = ?", generated).First(&clash).Error; err != nil {
			code = generated
			break
		}
	}

	mapping := URLMapping{
		ShortCode: code,
		LongURL:   req.LongURL,
	}
	if err := db.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save mapping"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_url":  os.Getenv("BASE_URL") + "/" + code,
		"short_code": code,
	})
}

func redirectHandler(c *gin.Context) {
	code := c.Param("code")

	var mapping URLMapping
	err := db.Where("short_code = ?", code).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&mapping).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	click := URLClick{
		MappingID: mapping.ID,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		ClickedAt: time.Now(),
	}
	db.Create(&click)

	c.Redirect(http.StatusFound, mapping.LongURL)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	fmt.Println("Connected to the database")

	if err := db.AutoMigrate(&URLMapping{}, &URLClick{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	fmt.Println("Database migrated")

	router := gin.Default()

	router.POST("/shorten", shortenHandler)
	router.GET("/:code", redirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server listening on http://localhost:" + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed: ", err)
	}

}
