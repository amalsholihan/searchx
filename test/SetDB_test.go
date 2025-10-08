package test

import (
	"testing"

	"github.com/amalsholihan/searchx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Contoh model
type User struct {
	ID   uint
	Name string
	Age  int
}

func TestSetDB(t *testing.T) {
	// 🧠 1. Buat database SQLite in-memory
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	// 🧱 2. Auto-migrate model
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate model: %v", err)
	}

	// 🔧 3. Gunakan Searchx.SetDB()
	s := &searchx.Searchx{}
	s.SetDB(*db)

	// 🧪 4. Simpan data dummy
	if err := s.DB.Create(&User{Name: "Amal", Age: 27}).Error; err != nil {
		t.Fatalf("failed to insert record: %v", err)
	}

	// 🔍 5. Ambil data dan cek hasilnya
	var u User
	if err := s.DB.First(&u, "name = ?", "Amal").Error; err != nil {
		t.Fatalf("failed to fetch record: %v", err)
	}

	if u.Age != 27 {
		t.Errorf("expected Age=27, got %d", u.Age)
	}

	// ✅ 6. Pastikan DB masih valid
	sqlDB, err := s.DB.DB()
	if err != nil {
		t.Fatalf("expected DB connection valid, got error: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("expected DB alive, got ping error: %v", err)
	}
}
