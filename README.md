# Go Fiber Real-Time Chat Application

![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)
![Framework](https://img.shields.io/badge/Fiber-v2-00ADD8?style=for-the-badge&logo=go)
![Database](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker)
![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)

Aplikasi chat real-time yang kaya fitur dan berperforma tinggi, dibangun menggunakan Go (Fiber) untuk backend dan Vanilla JavaScript untuk frontend. Proyek ini mendemonstrasikan arsitektur backend yang bersih, penggunaan WebSocket untuk komunikasi dua arah, dan otentikasi berbasis JWT yang aman.

---

### Ganti Screenshot ini!
_Ganti URL di bawah dengan screenshot aplikasi Anda._

![App Screenshot](https://via.placeholder.com/800x450.png?text=Screenshot+Aplikasi+Anda)

---

## ✨ Fitur Utama

- **Otentikasi Aman:** Sistem registrasi, login, dan logout lengkap dengan JWT (Access & Refresh Tokens).
- **Chat Multi-Ruangan:**
  - **Direct Messages (DM):** Ruang chat privat 1-lawan-1 yang dibuat secara otomatis.
  - **Grup Chat:** Buat dan bergabung ke dalam grup dengan banyak anggota.
- **Komunikasi Real-Time:** Pengiriman dan penerimaan pesan instan via WebSocket.
- **Lobi & Riwayat Chat:** Lihat daftar room dan akses riwayat percakapan sebelumnya saat memasuki room.
- **Database Persistence:** Semua data (pengguna, room, pesan) disimpan secara permanen di database PostgreSQL.
- **UI Intuitif:** Antarmuka frontend yang dinamis, menampilkan nama lawan bicara untuk DM dan avatar yang digenerate otomatis.

## 🚀 Tumpukan Teknologi (Tech Stack)

| Komponen | Teknologi                                                              |
| :------- | :--------------------------------------------------------------------- |
| **Backend**      | Go, [Fiber v2](https://gofiber.io/), [GORM](https://gorm.io/), [Google Wire](https://github.com/google/wire), PostgreSQL, WebSocket |
| **Frontend**     | Vanilla JavaScript (ES6+), HTML5, CSS3                         |
| **Deployment**   | Docker, Docker Compose                                         |

## 🏗️ Arsitektur

Aplikasi ini menggunakan arsitektur berlapis (Layered Architecture) untuk memastikan kode yang bersih, modular, dan mudah dipelihara.

```
+----------------+      +----------------+      +------------------+      +----------+
|     Client     |----->|     Handler    |----->|      Service     |----->| Repository |
| (JS/WebSocket) |      | (Validation)   |      | (Business Logic) |      | (Database) |
+----------------+      +----------------+      +------------------+      +----------+
```

- **Handler**: Menerima permintaan HTTP & WebSocket, melakukan validasi input.
- **Service**: Mengandung semua logika bisnis aplikasi.
- **Repository**: Bertanggung jawab atas semua komunikasi dengan database.
- **Dependency Injection**: [Google Wire](https://github.com/google/wire) digunakan untuk mengelola dependensi secara otomatis.

## 🏁 Memulai Proyek (Getting Started)

### Prasyarat

- [Go](https://golang.org/dl/) (versi 1.18+)
- [Docker](https://www.docker.com/get-started) & [Docker Compose](https://docs.docker.com/compose/install/)

---

### Metode 1: Menggunakan Docker (Direkomendasikan)

Cara termudah untuk menjalankan aplikasi secara lokal.

1.  **Clone repositori:**
    ```bash
    git clone https://github.com/your-username/go-fiber-chatapp.git
    cd go-fiber-chatapp
    ```

2.  **Buat file `.env`:**
    Buat file `.env` di direktori root proyek. Anda bisa menyalin dari contoh di bawah. Password `POSTGRES_PASSWORD` di sini harus sama dengan yang ada di `docker-compose.yml`.

    ```env
    # App Config
    APP_PORT=8080

    # Database Config (used by Go app to connect to Docker container)
    DB_HOST=db
    DB_USER=postgres
    DB_PASSWORD=your_super_secret_password # Ganti dengan password yang aman
    DB_NAME=chat_app_db
    DB_PORT=5432
    DB_SSLMODE=disable

    # JWT Config
    JWT_SECRET=your_jwt_secret_key # Ganti dengan secret yang kuat
    ACCESS_TOKEN_EXP_DAYS=1
    REFRESH_TOKEN_EXP_DAYS=7
    ```

3.  **Jalankan dengan Docker Compose:**
    Perintah ini akan membangun image dan menjalankan container untuk aplikasi Go dan database PostgreSQL.
    ```bash
    docker-compose up --build
    ```
    Aplikasi akan dapat diakses di `http://localhost:8080`.

---

### Metode 2: Setup Manual

Gunakan metode ini jika Anda tidak ingin menggunakan Docker.

1.  **Clone repositori** (lihat langkah di atas).

2.  **Pastikan PostgreSQL berjalan** di sistem Anda dan Anda memiliki database yang telah dibuat.

3.  **Buat file `.env`:**
    Gunakan contoh di atas, tetapi sesuaikan `DB_HOST`, `DB_PORT`, `DB_USER`, dan `DB_PASSWORD` dengan konfigurasi PostgreSQL lokal Anda. Umumnya `DB_HOST` akan menjadi `localhost`.

4.  **Install dependensi Go:**
    ```bash
    go mod tidy
    ```

5.  **Generate kode Dependency Injection:**
    Perintah ini wajib dijalankan untuk membuat file `wire_gen.go`.
    ```bash
    go generate ./...
    ```

6.  **Jalankan aplikasi:**
    ```bash
    go run .
    ```
    Server akan berjalan di `http://localhost:8080`. GORM akan otomatis membuat skema tabel saat aplikasi pertama kali dijalankan.

## ↔️ Endpoint API

| Method | Endpoint             | Deskripsi                                                |
| ------ | -------------------- | -------------------------------------------------------- |
| `POST` | `/api/auth/register` | Mendaftarkan pengguna baru.                              |
| `POST` | `/api/auth/login`    | Login dan mendapatkan token JWT.                         |
| `POST` | `/api/auth/logout`   | Logout (menambahkan token ke daftar hitam).              |
| `POST` | `/api/auth/refresh`  | Mendapatkan access token baru.                           |
| `GET`  | `/api/profile`       | Mendapatkan profil pengguna yang sedang login.           |
| `PUT`  | `/api/profile`       | Memperbarui profil pengguna.                             |
| `GET`  | `/api/users`         | Mendapatkan daftar semua pengguna.                       |
| `GET`  | `/api/rooms`         | Mendapatkan daftar room yang diikuti pengguna.           |
| `POST` | `/api/rooms`         | Membuat room baru (DM atau Grup).                        |
| `GET`  | `/chat/ws/:roomId`   | Endpoint untuk koneksi WebSocket ke sebuah room.         |

## 🗺️ Rencana Pengembangan (Roadmap)

- [ ] Status Online/Offline pengguna.
- [ ] Indikator "Sedang Mengetik..."
- [ ] Penambahan Unit Test & Integration Test.
- [ ] Fitur manajemen grup (undang/keluarkan anggota).
- [ ] Kemampuan mengunggah file atau gambar.

## 📄 Lisensi

Proyek ini dilisensikan di bawah [Lisensi MIT](LICENSE).
