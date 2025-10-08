package test

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
	Age  int
}

func (s User) TableName() string {
	return "test_user"
}

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 🧪 1. Simpan data dummy
	if err := db.Create(&User{Name: "Amal", Age: 27}).Error; err != nil {
		t.Fatalf("failed to insert record: %v", err)
	}

	// 🔍 2. Ambil data dan cek hasilnya
	var u User
	if err := db.First(&u, "name = ?", "Amal").Error; err != nil {
		t.Fatalf("failed to fetch record: %v", err)
	}

	return db.Model(&User{})
}
