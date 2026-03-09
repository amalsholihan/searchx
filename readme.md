# 🧠 searchx — Query Builder untuk GORM yang Fleksibel

`searchx` adalah helper package untuk memperluas kemampuan **GORM** dalam melakukan pencarian dinamis, agregasi, union, sorting, dan paginasi dengan sintaks yang sederhana dan konsisten.
Package ini cocok untuk kebutuhan **API filtering**, **report generator**, atau **dynamic SQL builder** tanpa menulis query mentah.

---

## 🚀 Instalasi

```bash
go get github.com/amalsholihan/searchx
```

Import di kode kamu:

```go
import "github.com/amalsholihan/searchx"
```

---

## ⚙️ Setup Awal

Bungkus objek `*gorm.DB` kamu menggunakan `searchx.SetDB()` agar bisa menggunakan seluruh fitur `searchx`.

```go
// inisialisasi GORM DB
db := DB.Model(&User{}).
	Select(`id, name`)
sx := searchx.SetDB(*db)
```

---

## 🔍 1. Get Data

Ambil semua data dari tabel aktif:

```go
db := DB.Model(&User{}).
	Select(`id, name`)
var result []map[string]any

sx := searchx.SetDB(*db)
search_result := sx.Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output

```sql
SELECT * FROM `test_user`
```

---

## 🎯 2. Search dengan Nested Structure

`searchx` mendukung **nested search** dengan struktur AND/OR yang kompleks menggunakan format JSON atau map.

### 2.1 Search Sederhana

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    Search([]map[string]any{
        {
            "search_column":    "name",
            "search_condition": "=",
            "search_text":      "Annissa",
        },
    }).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` WHERE name = 'Annissa'
```

### 2.2 Search dengan Format JSON

Gunakan format JSON untuk search yang lebih kompleks:

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    SearchWithJSON(`{
        "and": [
            {
                "field": "name",
                "operator": "=",
                "value": "Annissa"
            },
            {
                "field": "age",
                "operator": ">",
                "value": 20
            }
        ]
    }`).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` WHERE (name = 'Annissa' AND age > 20)
```

### 2.3 Search dengan Nested OR dalam AND

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    SearchWithJSON(`{
        "and": [
            {
                "field": "name",
                "operator": "=",
                "value": "Annissa"
            },
            {
                "or": [
                    {
                        "field": "age",
                        "operator": ">",
                        "value": 20
                    },
                    {
                        "field": "age",
                        "operator": "<",
                        "value": 30
                    }
                ]
            }
        ]
    }`).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` WHERE (name = 'Annissa' AND (age > 20 OR age < 30))
```

### 2.4 Search dengan Key "search" untuk Nested Structure

Gunakan key "search" untuk menghindari duplicate key issue:

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    Search([]map[string]any{
        {
            "search_column":    "name",
            "search_condition": "=",
            "search_text":      "Annissa",
        },
        {
            "search": []map[string]any{
                {
                    "search_column":    "name",
                    "search_condition": "=",
                    "search_text":      "Annissa",
                },
            },
        },
    }).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` WHERE name = 'Annissa' AND (name = 'Annissa')
```

### 2.5 Operator yang Didukung

| Operator | Deskripsi |
|----------|-----------|
| `=` | Sama dengan |
| `!=` | Tidak sama dengan |
| `>` | Lebih besar dari |
| `>=` | Lebih besar atau sama dengan |
| `<` | Lebih kecil dari |
| `<=` | Lebih kecil atau sama dengan |
| `like` | Pattern matching |
| `is null` | Nilai null |
| `is not null` | Nilai tidak null |

---

## 📊 3. Summary / Aggregate Query

Gunakan `.Summary()` untuk membuat query agregasi seperti `SUM()`, `COUNT()`, `AVG()`, `MIN()`, `MAX()`.

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := map[string]any{}

search_result := searchx.SetDB(*db).
    Summary(map[string]string{
        "total_sales": "sum(sales)",
        "max_sales":   "max(sales)",
        "min_sales":   "min(sales)",
    }).
    GetSummary(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw summary:", search_result.RawSummary)
fmt.Println("Total sales:", result["total_sales"])
fmt.Println("Max sales:", result["max_sales"])
fmt.Println("Min sales:", result["min_sales"])
```

### Output Query

```sql
SELECT sum(sales) as total_sales, max(sales) as max_sales, min(sales) as min_sales
FROM (select * from test_user) my_table_summary
```

---

## 🔄 4. Union Query + Sort

`searchx` mendukung **UNION query** antar tabel atau model berbeda, lengkap dengan filter dan sorting setelah digabung.

```go
dbUser := DB.Model(&User{}).
	Select(`id, name`)
dbStaff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")
result := []map[string]any{}

search_result := searchx.SetDB(*dbUser).
    Union(*searchx.SetDB(*dbStaff)).
    Search([]map[string]any{
        {
            "search_column":    "name",
            "search_condition": "is not null",
        },
    }).
    Sort([]map[string]any{
        {
            "sort_column": "name",
            "sort_type":   "asc",
        },
        {
            "sort_column": "id",
            "sort_type":   "desc",
        },
    }).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw Union Query:", search_result.RawUnion)
fmt.Println("Data:", result)
```

### Output Query

```sql
select * from (
  select id, name, age, sales from test_user where name is not null
  union
  select id, name, age, sales from test_staff where name is not null
) as my_table
order by name ASC, id DESC
```

---

## 📊 5. Sort dengan Nested Structure

`searchx` mendukung **nested sort** dengan format JSON atau map untuk sorting yang lebih kompleks.

### 5.1 Sort Sederhana

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    Sort([]map[string]any{
        {
            "sort_column": "name",
            "sort_type":   "asc",
        },
        {
            "sort_column": "id",
            "sort_type":   "desc",
        },
    }).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` ORDER BY name ASC, id DESC
```

### 5.2 Sort dengan Format JSON

Gunakan format JSON untuk sort yang lebih kompleks:

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    SortWithJSON(`{
        "sort": [
            {
                "field": "name",
                "direction": "asc"
            },
            {
                "field": "id",
                "direction": "desc"
            }
        ]
    }`).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` ORDER BY name ASC, id DESC
```

### 5.3 Sort dengan Key "sort" untuk Nested Structure

Gunakan key "sort" untuk menghindari duplicate key issue:

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    Sort([]map[string]any{
        {
            "sort_column": "name",
            "sort_type":   "asc",
        },
        {
            "sort": []map[string]any{
                {
                    "sort_column": "id",
                    "sort_type":   "desc",
                },
            },
        },
    }).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` ORDER BY name ASC, id DESC
```

### 5.4 Direction yang Didukung

| Direction | Deskripsi |
|-----------|-----------|
| `asc` | Ascending (naik) |
| `desc` | Descending (turun) |

---

## � 6. Pagination

Gunakan `.Paginate(page, limit, &result)` untuk melakukan paginasi otomatis, lengkap dengan total count.

```go
db := DB.Model(&User{}).
	Select(`id, name`)
result := searchx.Paginated{}

search_result := searchx.SetDB(*db).
    Paginate(1, 10, &result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Total:", result.Total)
fmt.Println("Data:", result.Data)
fmt.Println("Total Pages:", result.TotalPages)
```

---

## 🧱 Struktur Fungsi Utama

| Fungsi                           | Deskripsi                            |
| -------------------------------- | ------------------------------------ |
| `SetDB(db)`                      | Inisialisasi wrapper searchx         |
| `Get(&result)`                   | Menjalankan query utama              |
| `Search([]map[string]any)`       | Menambahkan WHERE dinamis            |
| `SearchWithJSON(json)`           | Menambahkan WHERE dengan format JSON |
| `Sort([]map[string]any)`        | Menambahkan ORDER BY dinamis         |
| `SortWithJSON(json)`             | Menambahkan ORDER BY dengan format JSON |
| `Summary(map[string]string)`     | Menambahkan kolom agregasi           |
| `GetSummary(&result)`            | Eksekusi query summary               |
| `Union(query)`                   | Menggabungkan dua query (UNION)      |
| `Paginate(page, limit, &result)` | Paginate otomatis dengan total count |
| `Err`                            | Properti error jika query gagal      |
| `Raw`, `RawSummary`, `RawUnion`  | Query SQL terakhir yang dijalankan   |

---

## ✨ Fitur Baru

### Nested Search dengan AND/OR Logic
- Mendukung struktur nested dengan key "and" dan "or"
- Format JSON untuk query yang lebih kompleks
- Key "search" untuk menghindari duplicate key issue
- Operator baru: `>` dan `<`

### Nested Sort
- Mendukung struktur nested dengan key "sort"
- Format JSON untuk sorting yang lebih kompleks
- Backward compatibility dengan format lama

### Keuntungan Nested Structure
- Lebih mudah dibaca dan dipahami
- Lebih fleksibel untuk query kompleks
- Menghindari duplicate key issue
- Konsisten dengan format JSON yang umum digunakan

---

## 🧩 Contoh Query Chaining (Custom)

`searchx` bisa dikombinasikan dengan query builder GORM biasa:

```go
result := []map[string]any{}

db := DB.Model(&User{}).
    Where("status = ?", "active")
search_result := searchx.SetDB(*db).
    Summary(map[string]string{
        "total_amount": "SUM(amount)",
    }).
    GetSummary(&result)
```

### Contoh Penggunaan Lengkap dengan Nested Search dan Sort

```go
db := DB.Model(&User{}).
	Select(`id, name, age, sales`)
result := []map[string]any{}

search_result := searchx.SetDB(*db).
    SearchWithJSON(`{
        "and": [
            {
                "field": "name",
                "operator": "=",
                "value": "Annissa"
            },
            {
                "or": [
                    {
                        "field": "age",
                        "operator": ">",
                        "value": 20
                    },
                    {
                        "field": "age",
                        "operator": "<",
                        "value": 30
                    }
                ]
            }
        ]
    }`).
    SortWithJSON(`{
        "sort": [
            {
                "field": "name",
                "direction": "asc"
            },
            {
                "field": "age",
                "direction": "desc"
            }
        ]
    }`).
    Get(&result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Raw query:", search_result.Raw)
fmt.Println("Data:", result)
```

### Output Query

```sql
SELECT * FROM `test_user` 
WHERE (name = 'Annissa' AND (age > 20 OR age < 30)) 
ORDER BY name ASC, age DESC
```

---

## 💡 Tips dan Best Practices

### 1. Backward Compatibility
Format lama masih didukung dan dapat digunakan bersamaan dengan format baru:

```go
// Format lama (masih berfungsi)
Search([]map[string]any{
    {
        "search_column":    "name",
        "search_condition": "=",
        "search_text":      "Annissa",
    },
})

// Format baru dengan JSON
SearchWithJSON(`{
    "and": [
        {
            "field": "name",
            "operator": "=",
            "value": "Annissa"
        }
    ]
}`)

// Format baru dengan key "search"
Search([]map[string]any{
    {
        "search": []map[string]any{
            {
                "search_column":    "name",
                "search_condition": "=",
                "search_text":      "Annissa",
            },
        },
    },
})
```

### 2. Menghindari Duplicate Key Issue
Gunakan key "search" atau "sort" untuk nested structure:

```go
// ❌ Akan menyebabkan duplicate key issue
Search([]map[string]any{
    {
        "search_column":    "name",
        "search_condition": "=",
        "search_text":      "Annissa",
    },
    {
        "search_column":    "age",
        "search_condition": ">",
        "search_text":      "20",
    },
})

// ✅ Gunakan key "search" untuk nested structure
Search([]map[string]any{
    {
        "search_column":    "name",
        "search_condition": "=",
        "search_text":      "Annissa",
    },
    {
        "search": []map[string]any{
            {
                "search_column":    "age",
                "search_condition": ">",
                "search_text":      "20",
            },
        },
    },
})
```

### 3. Debugging
Gunakan properti `Raw` untuk melihat query yang dihasilkan:

```go
search_result := searchx.SetDB(*db).
    SearchWithJSON(`{
        "and": [
            {
                "field": "name",
                "operator": "=",
                "value": "Annissa"
            }
        ]
    }`).
    Get(&result)

fmt.Println("Raw query:", search_result.Raw)
```

### 4. Performance
Nested structure tidak berat karena:
- Rekursif hanya dilakukan sekali saat parsing
- Query yang dihasilkan adalah SQL flat biasa
- Tidak ada overhead saat eksekusi database

---

## test

### Menjalankan Semua Test
```bash
go test -v ./test
```

### Struktur Test
Test telah dipisahkan menjadi beberapa file berdasarkan fungsionalitas:

| File Test | Deskripsi |
|-----------|-----------|
| `Get_test.go` | Test fungsi Get dasar |
| `Search_test.go` | Test search dengan berbagai format (JSON, nested, dll) |
| `Sort_test.go` | Test sort dengan berbagai format (JSON, nested, dll) |
| `Union_test.go` | Test union query dengan search dan sort |
| `Paginate_test.go` | Test fungsi paginasi |
| `PaginateUnion_test.go` | Test paginasi dengan union |
| `Summary_test.go` | Test fungsi summary/agregasi |

### Menjalankan Test Spesifik
```bash
# Test search saja
go test -v ./test -run TestSearch

# Test sort saja
go test -v ./test -run TestSort

# Test union saja
go test -v ./test -run TestUnion
```

## 🧠 Lisensi

MIT License © 2025 [Amal Sholihan](https://github.com/amalsholihan)
