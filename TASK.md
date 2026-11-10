# Implement Product CRUD API – Tepat Supplier Tools (Backend Only)

Bertindak sebagai **Senior Go Backend Engineer**. Implementasikan fitur **Product CRUD API** pada repository backend ini dengan mengikuti kontrak API dari frontend:

`https://github.com/nizarlabib/tepat-supplier-tools/blob/main/docs/FRONTEND_DATA_CONTRACT.md`

**PENTING: Scope pekerjaan hanya repository backend Go. Jangan mengubah, membuat, menghapus, atau melakukan commit perubahan apa pun pada repository frontend.** File `FRONTEND_DATA_CONTRACT.md` hanya menjadi referensi untuk memahami struktur data dan kebutuhan API.

Pastikan implementasi mengikuti arsitektur, standar coding, response format, dan pola yang sudah diterapkan di backend.

## 1. Analisis codebase sebelum implementasi

Sebelum menulis kode, periksa terlebih dahulu:

- Struktur folder dan arsitektur project.
- Definisi model dan base model yang sudah tersedia.
- Implementasi `BaseUUIDModel` dan `BaseModelTimestamp`.
- Pola migration database.
- Pola repository, use case/service, handler, dan dependency injection.
- Standar routing dan middleware.
- Implementasi authentication dan authorization/RBAC.
- Standar request validation.
- Format response success dan error.
- Standar logging dan error handling.
- Implementasi Swagger/OpenAPI.

Jangan membuat pola atau abstraksi baru jika project sudah memiliki standar yang bisa digunakan. Ikuti pola yang paling konsisten dengan codebase saat ini.

## 2. Referensi API contract

Baca dan pahami `FRONTEND_DATA_CONTRACT.md` secara menyeluruh.

Gunakan dokumen tersebut sebagai referensi untuk menentukan:

- Struktur data Product.
- Struktur data Category.
- Field yang dibutuhkan frontend.
- Endpoint public.
- Query parameters.
- Pagination.
- Struktur response.

Product membutuhkan informasi:

- `id`
- `name`
- `slug`
- `brand`
- `category_id`
- `category`
- `description`
- `image_url`
- `image_alt`
- `price`
- `is_active`
- `created_at`
- `updated_at`

**Field `price` merupakan tambahan kebutuhan backend pada implementasi ini.** Tambahkan dukungannya pada model, database, DTO, validation, repository, use case, handler, dan dokumentasi API backend.

Jangan melakukan perubahan terhadap file frontend contract. Jika ditemukan ketidaksesuaian antara kebutuhan API dan dokumen referensi, jelaskan dalam final report sebagai catatan untuk tim frontend.

## 3. Standard model ID, timestamp, dan soft delete

WAJIB menggunakan `BaseUUIDModel` yang sudah tersedia di project sebagai standar model.

Struktur acuan:

```go
type BaseUUIDModel struct {
    ID uuid.UUID `json:"id" gorm:"primarykey;default:gen_random_uuid();"`
    *BaseModelTimestamp
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

Ketentuan:

1. Gunakan `BaseUUIDModel` sebagai embedded struct pada model Product dan model lain yang relevan.
2. Gunakan `uuid.UUID` sebagai primary key. Jangan menggunakan `uint`, `int`, atau string untuk ID.
3. ID dibuat otomatis melalui PostgreSQL `gen_random_uuid()`, mengikuti mekanisme yang sudah digunakan project.
4. Gunakan `BaseModelTimestamp` melalui embedded `BaseUUIDModel`. Jangan mendefinisikan ulang `CreatedAt` dan `UpdatedAt`.
5. Gunakan `gorm.DeletedAt` untuk soft delete.
6. Operasi DELETE harus menggunakan mekanisme soft delete GORM, bukan permanent delete.
7. Query public listing dan detail harus mengecualikan record yang sudah di-soft-delete.
8. Foreign key yang mengarah ke model berbasis `BaseUUIDModel` harus menggunakan `uuid.UUID`.
9. Pastikan migration tidak membuat kolom ID atau timestamp duplikat.
10. Periksa strategi unique slug dan soft delete. Pastikan slug tidak menyebabkan konflik ketika record di-soft-delete. Gunakan partial unique index jika sesuai dengan standar database dan migration project.

Sebelum implementasi, periksa definisi aktual `BaseUUIDModel` dan `BaseModelTimestamp`. Jangan membuat base model baru atau mengubah standar tersebut.

## 4. Product model dan database schema

Implementasikan model Product dengan field berikut:

| Field | Type | Keterangan |
|---|---|---|
| ID | uuid.UUID | Menggunakan BaseUUIDModel |
| Name | string | Nama produk |
| Slug | string | Unique slug |
| Brand | nullable string | Merek produk |
| CategoryID | uuid.UUID | Foreign key ke Category |
| Description | text | Deskripsi produk |
| ImageURL | nullable string | URL gambar |
| ImageAlt | nullable string | Alt text gambar |
| Price | nullable decimal | Harga produk |
| IsActive | bool | Status aktif |
| CreatedAt | dari BaseUUIDModel | Timestamp otomatis |
| UpdatedAt | dari BaseUUIDModel | Timestamp otomatis |
| DeletedAt | dari BaseUUIDModel | Soft delete |

### Ketentuan field Price

- Gunakan tipe database `NUMERIC(15,2)`.
- Jangan menggunakan `float32` atau `float64` untuk menyimpan nilai uang.
- Gunakan decimal library atau tipe decimal yang sudah digunakan project.
- Nilai harga harus non-negatif.
- Field `price` boleh `null`.
- `null` berarti harga belum ditentukan atau produk perlu meminta penawaran harga.
- Representasi JSON menggunakan number atau null, bukan string berformat mata uang.
- Mata uang yang digunakan adalah IDR.

### Ketentuan database

- Gunakan UUID sebagai primary key.
- Gunakan foreign key untuk relasi Category.
- Slug harus unik sesuai strategi soft delete.
- `is_active` memiliki default `true`.
- Tambahkan index yang relevan untuk pencarian, filter kategori, status aktif, dan slug.
- Gunakan migration mengikuti mekanisme project.
- Jangan mengubah atau menghapus data existing secara tidak perlu.

Jika tabel Category belum tersedia, implementasikan model dan migration Category dengan struktur:

| Field | Type |
|---|---|
| ID | uuid.UUID |
| Name | string |
| Slug | string |
| Description | nullable string |
| IsActive | bool |
| SortOrder | int |
| CreatedAt / UpdatedAt / DeletedAt | dari BaseUUIDModel |

Kategori awal mengikuti referensi frontend contract:

- `mata-bor`
- `hand-tap`
- `insert-widia`
- `aksesoris-machining`

Pastikan seed kategori tidak menghasilkan data duplikat ketika dijalankan berulang kali.

## 5. API endpoints

Implementasikan endpoint berikut.

### Public endpoints

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | `/api/v1/products` | Mendapatkan daftar produk |
| GET | `/api/v1/products/{slug}` | Mendapatkan detail produk berdasarkan slug |
| GET | `/api/v1/categories` | Mendapatkan daftar kategori aktif |

### Management endpoints

| Method | Endpoint | Fungsi |
|---|---|---|
| POST | `/api/v1/products` | Membuat produk |
| PUT | `/api/v1/products/{id}` | Memperbarui seluruh data produk |
| PATCH | `/api/v1/products/{id}/status` | Mengubah status aktif produk |
| DELETE | `/api/v1/products/{id}` | Soft delete produk |

Management endpoints wajib menggunakan authentication dan authorization sesuai mekanisme RBAC yang sudah ada di project.

Jika permission untuk Product belum tersedia, tambahkan mengikuti pola permission existing. Jangan membuat sistem authorization baru.

## 6. Product listing dan pagination

Endpoint:

`GET /api/v1/products`

Mendukung query parameters:

| Parameter | Type | Default | Keterangan |
|---|---|---|---|
| search | string | kosong | Pencarian nama, brand, dan description |
| category | string | kosong | Filter berdasarkan category slug |
| page | integer | 1 | Halaman |
| per_page | integer | 9 | Jumlah data per halaman |

Ketentuan:

- Search menggunakan case-insensitive partial matching.
- Filter kategori menggunakan slug kategori.
- Pagination dilakukan di database.
- Default `per_page` adalah 9.
- Maksimum `per_page` adalah 50.
- Hanya produk aktif dan belum di-soft-delete yang ditampilkan pada public listing.
- Hitung total data dan total halaman secara akurat.
- Pastikan query efisien dan tidak menimbulkan N+1 query.

Format pagination mengikuti referensi frontend contract:

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 9,
    "total": 25,
    "total_pages": 3
  }
}
```

Sesuaikan struktur tersebut dengan response envelope standar backend tanpa menghilangkan informasi pagination.

## 7. Product detail

Endpoint:

`GET /api/v1/products/{slug}`

Ketentuan:

- Mengambil produk berdasarkan slug.
- Sertakan informasi kategori.
- Hanya produk aktif yang dapat diakses melalui public endpoint.
- Produk yang sudah di-soft-delete harus dianggap tidak ditemukan.
- Kembalikan error sesuai standar project jika produk tidak ditemukan.

## 8. Product create

Endpoint:

`POST /api/v1/products`

Contoh request:

```json
{
  "name": "Nachi Drill Bit",
  "brand": "Nachi",
  "category_id": "UUID_CATEGORY",
  "description": "High quality drill bit",
  "image_url": "https://example.com/image.jpg",
  "image_alt": "Nachi Drill Bit",
  "price": 150000,
  "is_active": true
}
```

Ketentuan:

- ID dibuat otomatis oleh database.
- Slug dibuat otomatis berdasarkan nama produk.
- Pastikan slug unik. Jika sudah digunakan, tambahkan suffix yang sesuai.
- Validasi `category_id` dan pastikan kategori tersedia.
- Validasi harga tidak boleh negatif.
- Field `price` boleh null.
- Jangan mempercayai ID, timestamp, atau field sistem dari client.
- Tangani duplicate slug dan database constraint dengan benar.

## 9. Product update

Endpoint:

`PUT /api/v1/products/{id}`

Ketentuan:

- Update data produk berdasarkan UUID.
- Terapkan validasi yang sama dengan proses create.
- Jika nama produk berubah, tentukan strategi update slug secara konsisten dan hindari konflik.
- Pastikan relasi kategori valid.
- Jangan mengubah ID dan timestamp secara manual.
- Produk yang sudah di-soft-delete tidak boleh diperbarui melalui endpoint normal.

## 10. Product status

Endpoint:

`PATCH /api/v1/products/{id}/status`

Contoh request:

```json
{
  "is_active": false
}
```

Ketentuan:

- Hanya mengubah status aktif produk.
- Jangan mengubah field lain.
- Pastikan record belum di-soft-delete.

## 11. Product delete

Endpoint:

`DELETE /api/v1/products/{id}`

Ketentuan:

- Gunakan GORM soft delete.
- Jangan melakukan permanent delete.
- Record yang sudah dihapus tidak muncul di public listing dan detail.
- Tangani request terhadap ID yang tidak ditemukan atau sudah dihapus.
- Pastikan perilaku unique slug setelah soft delete konsisten dengan strategi index yang digunakan.

## 12. Response format

Ikuti response envelope yang sudah digunakan project.

Jangan membuat format response baru hanya untuk Product.

Pastikan response success dan error konsisten dengan endpoint existing, termasuk penggunaan:

- `message`
- `data`
- `meta`
- `error_code`
- HTTP status code

Untuk pagination, gabungkan metadata pagination dengan mekanisme response standar backend tanpa mengubah kontrak frontend.

## 13. Architecture dan implementation quality

Implementasikan fitur menggunakan struktur arsitektur yang sudah diterapkan project.

Pisahkan tanggung jawab sesuai pola existing:

- Model / Entity
- Request dan Response DTO
- Repository
- Use Case / Service
- Handler
- Route
- Migration dan Seeder

Ketentuan:

- Handler tidak boleh berisi business logic.
- Repository bertanggung jawab terhadap query database.
- Use case/service bertanggung jawab terhadap business logic.
- Gunakan context untuk operasi database.
- Gunakan transaction jika proses membutuhkan konsistensi beberapa operasi database.
- Tangani error secara eksplisit.
- Hindari duplikasi kode.
- Hindari overengineering.
- Gunakan naming yang konsisten dan mudah dipahami.
- Jangan menambahkan dependency baru jika fitur dapat menggunakan dependency yang sudah tersedia.

## 14. Validation dan error handling

Implementasikan validasi untuk:

- Name wajib diisi.
- Category ID wajib valid.
- Price tidak boleh negatif.
- Slug harus unik.
- URL gambar harus valid jika diisi.
- Pagination harus berupa integer positif.
- `per_page` tidak boleh melebihi batas maksimum.
- ID harus berupa UUID valid.

Gunakan validator dan mekanisme error handling yang sudah tersedia di project.

Pastikan error memiliki HTTP status dan error code yang sesuai. Jangan mengembalikan detail internal database kepada client.

## 15. Testing

Buat unit test dan integration test mengikuti pola testing project.

Minimal mencakup:

**Product CRUD**
- Create product berhasil.
- Create product dengan price null.
- Create product dengan price valid.
- Create product dengan price negatif ditolak.
- Duplicate slug ditangani.
- Update product berhasil.
- Update category tidak valid ditolak.
- Update status berhasil.
- Soft delete berhasil.
- Product yang di-soft-delete tidak dapat diakses melalui public endpoint.

**Product listing**
- Listing berhasil.
- Search berdasarkan nama.
- Search berdasarkan brand.
- Search berdasarkan description.
- Filter berdasarkan category slug.
- Pagination berjalan benar.
- `per_page` default 9.
- `per_page` maksimum 50.
- Produk inactive tidak ditampilkan.
- Produk soft-deleted tidak ditampilkan.

**Model dan database**
- UUID terbentuk dengan benar.
- CreatedAt dan UpdatedAt menggunakan mekanisme BaseModelTimestamp.
- DeletedAt bekerja sesuai mekanisme GORM soft delete.
- Relasi Product dan Category menggunakan UUID.
- Migration tidak menghasilkan kolom duplikat.

## 16. Swagger dan dokumentasi backend

Perbarui Swagger/OpenAPI backend untuk seluruh endpoint Product dan Category yang diimplementasikan.

Dokumentasikan:

- HTTP method dan endpoint.
- Authentication requirement.
- Query parameters.
- Request body.
- Response success.
- Response error.
- Pagination metadata.
- Field `price` beserta kemungkinan nilai null.
- Contoh UUID.
- Soft delete behavior jika relevan untuk dokumentasi management API.

Pastikan dokumentasi dapat di-generate menggunakan command yang berlaku di project.

Perbarui README backend dengan informasi:

- Product API endpoints.
- Category API endpoints.
- Query parameters dan pagination.
- Contoh request dan response.
- Cara menjalankan migration.
- Cara menjalankan seeder.
- Cara menjalankan testing.
- Cara generate Swagger.

## 17. Final verification

Setelah implementasi selesai:

1. Jalankan formatter Go.
2. Jalankan lint jika tersedia.
3. Jalankan seluruh test.
4. Jalankan migration atau verifikasi migration.
5. Jalankan proses build.
6. Pastikan route terdaftar dengan benar.
7. Pastikan Swagger berhasil di-generate.
8. Periksa response API terhadap frontend contract sebagai referensi.
9. Periksa kembali penggunaan `BaseUUIDModel` pada model yang relevan.
10. Pastikan tidak ada implementasi sementara, TODO yang belum diselesaikan, atau kode duplikat yang tidak diperlukan.
11. Pastikan hanya file pada repository backend yang berubah.

Jangan mengklaim test, migration, atau build berhasil jika belum benar-benar dijalankan.

## 18. Final report

Setelah implementasi selesai, berikan ringkasan:

- Fitur yang berhasil diimplementasikan.
- Daftar endpoint.
- Perubahan database dan migration.
- Perubahan model dan penggunaan BaseUUIDModel.
- Perubahan Swagger dan README backend.
- Daftar test beserta hasilnya.
- Command yang perlu dijalankan untuk menjalankan fitur.
- Ketidaksesuaian atau catatan terkait frontend contract, jika ada.
- Pekerjaan yang belum selesai jika ada.

**Batasan mutlak:** Pekerjaan hanya untuk backend Go. Jangan melakukan perubahan pada repository frontend, jangan mengedit `FRONTEND_DATA_CONTRACT.md`, dan jangan membuat implementasi frontend.