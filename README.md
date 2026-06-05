# Report-B2B

Aplikasi backend service untuk sistem reporting B2B yang dibangun menggunakan Go (Golang) dengan framework Gin.

## 📋 Deskripsi

Report-B2B adalah layanan REST API untuk mengelola dan mengirimkan laporan bisnis B2B secara otomatis. Aplikasi ini dilengkapi dengan fitur scheduler untuk pengiriman laporan secara berkala melalui email.

## 🚀 Fitur Utama

* **Report Management** - Pengelolaan dan pembuatan laporan.
* **Email Integration** - Integrasi dengan layanan Como Email untuk pengiriman laporan.
* **Auto Scheduler** - Penjadwalan otomatis untuk pengiriman laporan (Auto UW Report).
* **Role-based Middleware** - Sistem middleware berbasis role untuk keamanan API.
* **Excel Export** - Fitur ekspor laporan ke format Excel.

## 🛠️ Tech Stack

* **Language:** Go 1.26.3
* **Framework:** Gin (v1.12.0)
* **Database:** MySQL
* **Authentication:** JWT (`dgrijalva/jwt-go`)
* **Logging:** Zerolog
* **Environment:** godotenv

## 📁 Struktur Project

```text
Report-B2B/
├── client/                # HTTP client untuk layanan eksternal (Como Email)
│   └── como.go
├── config/                # Konfigurasi database dan aplikasi
│   ├── como.go
│   └── config.go
├── helpers/               # Helper functions (Excel, Scheduler)
│   ├── exel.go
│   └── schaduler.go
├── middleware/            # HTTP middleware
│   └── middleware.go
├── model/                 # Data models
│   ├── report/            # Model untuk report
│   │   ├── como.go
│   │   └── report.go
│   └── role/              # Model untuk role
│       └── role.go
├── module/                # Business modules
│   ├── report/            # Module report
│   │   ├── handler/
│   │   │   └── handler.go
│   │   ├── repository/
│   │   │   └── reportRepo.go
│   │   ├── usecase/
│   │   │   └── reportUC.go
│   │   ├── reportRepoInterface.go
│   │   └── reportUCInterface.go
│   └── role/              # Module role dan middleware
│       ├── middlewareUsecase/
│       │   └── middlewareUC.go
│       └── miiddlewareRepo/
│           └── middlewareRepo.go
├── .env                   # Environment variables (tidak di-commit)
├── .gitignore             # Git ignore file
├── go.mod                 # Go modules
├── go.sum                 # Go dependencies checksum
├── main.go                # Entry point aplikasi
└── README.md              # Dokumentasi project
```

## ⚙️ Konfigurasi

Buat file `.env` di root project dengan konfigurasi berikut:

```env
# Server Configuration
PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password
DB_NAME=report_b2b

# Connection Pool
MAX_IDLE=10
MAX_CONN=100

# File Configuration
PATH_FILE=/path/to/files

# Auto UW Report
AUTO_UW_REPORT_RECIPIENTS=email1@example.com,email2@example.com

# Como Email Configuration
# BaseURL=
# ApiKey=
# FromEmail=
```

## 🚀 Cara Menjalankan

### Prerequisites

* Go 1.26.3 atau lebih baru
* MySQL Database
* Git

### Langkah-langkah

#### 1. Clone Repository

```bash
git clone https://github.com/YugaAdiIrawan/Report-B2B.git
cd Report-B2B
```

#### 2. Install Dependencies

```bash
go mod download
```

#### 3. Konfigurasi Environment

```bash
cp .env.example .env
```

Kemudian sesuaikan isi file `.env` dengan environment yang digunakan.

#### 4. Jalankan Aplikasi

```bash
go run main.go
```

Aplikasi akan berjalan di:

```text
http://localhost:{PORT}
```

## 📚 API Documentation

### Reports

Endpoint yang berkaitan dengan pembuatan, pengelolaan, dan pengiriman laporan.

> Detail endpoint mengikuti implementasi pada module report.

## 📝 Arsitektur

Project ini menerapkan pendekatan **Clean Architecture** dengan pemisahan tanggung jawab sebagai berikut:

### Handler

Bertanggung jawab menerima dan mengembalikan HTTP request/response.

### Usecase

Berisi business logic aplikasi.

### Repository

Berfungsi sebagai data access layer untuk berinteraksi dengan database.

### Model

Merepresentasikan struktur data yang digunakan aplikasi.

### Client

Berfungsi untuk integrasi dengan layanan eksternal seperti Como Email.

## 📄 License

Internal Project - PT Asuransi Ciputra Indonesia.
