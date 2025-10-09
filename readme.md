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
db := // inisialisasi GORM DB
sx := searchx.SetDB(*db)
```

---

## 🔍 1. Get Data

Ambil semua data dari tabel aktif:

```go
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

## 📊 2. Summary / Aggregate Query

Gunakan `.Summary()` untuk membuat query agregasi seperti `SUM()`, `COUNT()`, `AVG()`, `MIN()`, `MAX()`.

```go
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

## 🔄 3. Union Query + Sort

`searchx` mendukung **UNION query** antar tabel atau model berbeda, lengkap dengan filter dan sorting setelah digabung.

```go
result := []map[string]any{}

qStaff := db.Session(&gorm.Session{}).Model(&Staff{}).Select("id, name, age, sales")

search_result := searchx.SetDB(*db).
    Union(*searchx.SetDB(*qStaff)).
    Search([]map[string]string{
        {
            "search_column":    "name",
            "search_condition": "is not null",
        },
    }).
    Sort([]map[string]string{
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

## 📄 4. Pagination

Gunakan `.Paginate(page, limit, &result)` untuk melakukan paginasi otomatis, lengkap dengan total count.

```go
result := map[string]any{}

search_result := searchx.SetDB(*db).
    Paginate(1, 10, &result)

if search_result.Err != nil {
    log.Fatal(search_result.Err)
}

fmt.Println("Total:", result["total"])
fmt.Println("Data:", result["data"])
fmt.Println("Total Pages:", result["total_pages"])
```

---

## 🧱 Struktur Fungsi Utama

| Fungsi                           | Deskripsi                            |
| -------------------------------- | ------------------------------------ |
| `SetDB(db)`                      | Inisialisasi wrapper searchx         |
| `Get(&result)`                   | Menjalankan query utama              |
| `Summary(map[string]string)`     | Menambahkan kolom agregasi           |
| `GetSummary(&result)`            | Eksekusi query summary               |
| `Union(query)`                   | Menggabungkan dua query (UNION)      |
| `Sort([]map[string]string)`      | Menambahkan ORDER BY dinamis         |
| `Paginate(page, limit, &result)` | Paginate otomatis dengan total count |
| `Err`                            | Properti error jika query gagal      |
| `Raw`, `RawSummary`, `RawUnion`  | Query SQL terakhir yang dijalankan   |

---

## 🧩 Contoh Query Chaining (Custom)

`searchx` bisa dikombinasikan dengan query builder GORM biasa:

```go
result := []map[string]any{}

db := DB.Model(&User{}).
    Where("status = ?", "active")
sx := searchx.SetDB(*db)

search_result := sx.Summary(map[string]string{
    "total_amount": "SUM(amount)",
}).GetSummary(&result)
```

---

## test
```bash
go test -v ./test
```

## 🧠 Lisensi

MIT License © 2025 [Amal Sholihan](https://github.com/amalsholihan)
