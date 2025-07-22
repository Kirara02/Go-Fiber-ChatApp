# Go Fiber Real-Time Chat Application

![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)
![Framework](https://img.shields.io/badge/Fiber-v2-00ADD8?style=for-the-badge&logo=go)
![Database](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql)

Aplikasi chat real-time yang kaya fitur, dibangun dengan Go (Fiber) untuk backend dan Vanilla JavaScript untuk frontend. Proyek ini mendemonstrasikan arsitektur backend yang bersih dan berlapis (Handler, Service, Repository) dengan Dependency Injection, WebSocket untuk komunikasi dua arah, dan otentikasi berbasis JWT.

<!-- Ganti dengan path screenshot aplikasi Anda -->

![App Screenshot](https://via.placeholder.com/800x450.png?text=App+Screenshot+Here)

## Fitur Utama

- **Otentikasi Pengguna:** Registrasi, Login, dan Logout yang aman menggunakan JWT (Access & Refresh Token).
- **Chat Multi-Ruangan:** Pengguna dapat membuat dan bergabung ke banyak ruangan.
  - **Direct Messages (DM):** Chat privat otomatis antara dua pengguna, dengan validasi untuk mencegah duplikasi.
  - **Grup Chat:** Room dengan 3+ anggota, lengkap dengan nama grup dan pemilik (owner).
- **Pesan Real-Time:** Komunikasi instan menggunakan WebSocket.
- **Lobi Chat:** Setelah login, pengguna dapat melihat daftar room yang mereka ikuti atau membuat room baru dengan memilih pengguna lain.
- **Persistensi Data:** Seluruh pengguna, room, keanggotaan, dan riwayat chat disimpan di database PostgreSQL.
- **Riwayat Chat:** Riwayat pesan dimuat secara otomatis saat memasuki sebuah room.
- **UI Dinamis & Cerdas:**
  - Frontend menampilkan nama DM sebagai nama lawan bicara.
  - Form pembuatan grup bersifat dinamis (input nama hanya muncul jika diperlukan).
  - Avatar profil dihasilkan secara otomatis berdasarkan nama pengguna.

## Tumpukan Teknologi (Tech Stack)

### Backend

- **Bahasa:** Go
- **Framework Web:** [Fiber v2](https://gofiber.io/)
- **ORM:** [GORM](https://gorm.io/)
- **Database:** PostgreSQL
- **Komunikasi Real-Time:** Fiber WebSocket
- **Otentikasi:** JSON Web Tokens (JWT)
- **Dependency Injection:** [Google Wire](https://github.com/google/wire)

### Frontend

- **HTML5**
- **CSS3** (tanpa framework)
- **Vanilla JavaScript** (ES6+)

## Arsitektur

Proyek ini mengikuti arsitektur berlapis (Layered Architecture) yang bersih untuk memisahkan tanggung jawab dan meningkatkan kemudahan pemeliharaan serta pengujian.

**Client** → **Handler** → **Service** → **Repository** → **Database**

- **Handler:** Menerima permintaan HTTP/WebSocket, memvalidasi input dasar, dan memanggil Service.
- **Service:** Berisi logika bisnis inti aplikasi (misalnya, memeriksa apakah DM sudah ada, membuat nama grup).
- **Repository:** Bertanggung jawab untuk semua interaksi dengan database (query, transaksi).
- **Dependency Injection:** [Google Wire](https://github.com/google/wire) digunakan untuk mengelola dan menyediakan dependensi secara otomatis di seluruh aplikasi, membuat kode menjadi lebih modular.

## Memulai Proyek (Getting Started)

### Prasyarat

- [Go](https://golang.org/dl/) versi 1.18 atau lebih baru.
- [PostgreSQL](https://www.postgresql.org/download/) yang sedang berjalan.
- [Wire CLI](https://github.com/google/wire#command-line-tool) (opsional, `go generate` juga bisa digunakan).

### Instalasi & Konfigurasi

1.  **Clone repositori ini:**

    ```bash
    git clone https://github.com/your-username/go-fiber-chatapp.git
    cd go-fiber-chatapp
    ```

2.  **Siapkan environment variables:**
    Buat file `.env` di direktori root dan salin konten dari `.env.example` (jika ada) atau gunakan contoh di bawah. Ganti nilainya sesuai dengan konfigurasi database dan rahasia JWT Anda.

    _Contoh `.env`_:

    ```env
    DB_HOST=localhost
    DB_USER=postgres
    DB_PASSWORD=your_password
    DB_NAME=chat_app_db
    DB_PORT=5432
    DB_SSLMODE=disable

    JWT_SECRET=your_super_secret_key
    ACCESS_TOKEN_EXP_DAYS=1
    REFRESH_TOKEN_EXP_DAYS=7
    ```

3.  **Install dependensi Go:**

    ```bash
    go mod tidy
    ```

4.  **Hasilkan kode Dependency Injection:**
    Langkah ini sangat penting. Wire akan membaca `wire.go` dan membuat `wire_gen.go`.

    ```bash
    go generate ./...
    ```

    Atau jika Anda menginstal Wire CLI:

    ```bash
    wire .
    ```

5.  **Jalankan aplikasi:**
    ```bash
    go run .
    ```
    Server akan berjalan di `http://localhost:8080`. GORM akan secara otomatis melakukan migrasi skema database saat pertama kali dijalankan.

## Endpoint API

Semua endpoint di bawah ini (kecuali `/auth/register` dan `/auth/login`) memerlukan token otentikasi.

| Method | Endpoint             | Deskripsi                                                |
| ------ | -------------------- | -------------------------------------------------------- |
| `POST` | `/api/auth/register` | Mendaftarkan pengguna baru.                              |
| `POST` | `/api/auth/login`    | Login dan mendapatkan token JWT.                         |
| `POST` | `/api/auth/logout`   | Menambahkan token saat ini ke denylist (logout).         |
| `POST` | `/api/auth/refresh`  | Mendapatkan access token baru menggunakan refresh token. |
| `GET`  | `/api/profile`       | Mendapatkan detail profil pengguna yang sedang login.    |
| `PUT`  | `/api/profile`       | Memperbarui profil pengguna yang sedang login.           |
| `GET`  | `/api/users`         | Mendapatkan daftar semua pengguna.                       |
| `GET`  | `/api/rooms`         | Mendapatkan daftar semua room yang diikuti pengguna.     |
| `POST` | `/api/rooms`         | Membuat room baru (DM atau Grup).                        |
| `GET`  | `/chat/ws/:roomId`   | Endpoint untuk memulai koneksi WebSocket ke room.        |

## Rencana Pengembangan (Future Improvements)

- [ ] **Status Online/Offline:** Menampilkan status pengguna secara real-time.
- [ ] **Indikator "Sedang Mengetik...":** Menampilkan saat pengguna lain sedang mengetik.
- [ ] **Unit & Integration Test:** Menambahkan pengujian untuk setiap lapisan.
- [ ] **Manajemen Grup:** Menambahkan fitur "undang/keluarkan anggota" dan "ubah nama grup" untuk pemilik.
- [ ] **Unggah File/Gambar:** Kemampuan untuk berbagi gambar atau file di dalam chat.

## Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).

```

```
