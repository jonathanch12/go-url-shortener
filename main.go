package main

import (
	"fmt"
	"time"
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

func main() {
	fmt.Println("Test")
}
