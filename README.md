# E-Voting System

Aplikasi e-voting berbasis web yang dibangun dengan Go dan HTML templates. Sistem ini mendukung dua role utama: **Super Admin** dan **Admin** dengan fitur voting berbasis token yang aman.

## Fitur Utama

### Super Admin
- ✅ Mengelola semua pemilihan (elections)
- ✅ Membuat, mengedit, dan menghapus pemilihan
- ✅ Mengelola pengguna admin
- ✅ Mengassign admin ke pemilihan tertentu
- ✅ Dashboard dengan statistik lengkap

### Admin
- ✅ Mengelola pemilihan yang di-assign
- ✅ Mengelola kandidat dalam pemilihan
- ✅ Generate dan mengelola token voting
- ✅ Memonitor votes yang masuk
- ✅ Melihat laporan dan statistik pemilihan

### Sistem Voting
- ✅ Voting berbasis token unik
- ✅ Satu token hanya bisa digunakan sekali
- ✅ Interface voting yang user-friendly
- ✅ Konfirmasi sebelum submit vote
- ✅ Hasil voting real-time

## Teknologi yang Digunakan

- **Backend**: Go (Golang)
- **Database**: SQLite
- **Frontend**: HTML, CSS, JavaScript, Bootstrap 5
- **Router**: Gorilla Mux
- **Session**: Gorilla Sessions
- **Password Hashing**: bcrypt

## Instalasi dan Setup

### Prerequisites
- Go 1.19 atau lebih baru
- Git

### Langkah Instalasi

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd evoting-app
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Jalankan aplikasi**
   ```bash
   go run main.go
   ```

4. **Akses aplikasi**
   - Buka browser dan kunjungi: `http://localhost:8080`

## Konfigurasi

Aplikasi menggunakan environment variables untuk konfigurasi:

```bash
# Database file path (default: evoting.db)
DATABASE_URL=evoting.db

# Server port (default: 8080)
PORT=8080

# Session secret key (ganti di production)
SESSION_SECRET=your-secret-key-change-this-in-production
```

## Login Default

### Super Admin
- **Username**: `superadmin`
- **Password**: `password`

> ⚠️ **Penting**: Ganti password default setelah login pertama!

## Struktur Database

Aplikasi menggunakan SQLite dengan tabel-tabel berikut:

- `users` - Data pengguna (super admin dan admin)
- `elections` - Data pemilihan
- `candidates` - Data kandidat dalam pemilihan
- `voting_tokens` - Token untuk voting
- `votes` - Data vote yang masuk
- `election_admins` - Relasi admin dengan pemilihan

## Cara Penggunaan

### 1. Setup Pemilihan (Super Admin)

1. Login sebagai super admin
2. Buat pemilihan baru di menu "Manage Elections"
3. Buat user admin baru di menu "Manage Users"
4. Assign admin ke pemilihan di menu "Assign Admin"

### 2. Setup Kandidat dan Token (Admin)

1. Login sebagai admin
2. Pilih pemilihan yang di-assign
3. Tambahkan kandidat di menu "Candidates"
4. Generate token voting di menu "Tokens"
5. Bagikan token ke pemilih

### 3. Voting (Pemilih)

1. Kunjungi halaman utama aplikasi
2. Klik "Vote Now"
3. Masukkan token yang diberikan
4. Pilih kandidat
5. Konfirmasi dan submit vote

### 4. Monitoring (Admin)

1. Monitor vote masuk di menu "Votes"
2. Lihat laporan dan statistik di menu "Reports"
3. Export data jika diperlukan

## Keamanan

- ✅ Password di-hash menggunakan bcrypt
- ✅ Session-based authentication
- ✅ Role-based access control
- ✅ Token voting unik dan sekali pakai
- ✅ Input validation dan sanitization
- ✅ CSRF protection melalui session

## API Endpoints

### Public Routes
- `GET /` - Halaman utama
- `GET /login` - Halaman login
- `POST /login` - Proses login
- `GET /vote` - Form voting
- `POST /vote` - Submit vote
- `POST /logout` - Logout

### Super Admin Routes
- `GET /admin/superadmin/dashboard` - Dashboard super admin
- `GET /admin/superadmin/elections` - Kelola pemilihan
- `GET /admin/superadmin/users` - Kelola pengguna
- Dan lainnya...

### Admin Routes
- `GET /admin/admin/dashboard` - Dashboard admin
- `GET /admin/admin/elections/{id}/candidates` - Kelola kandidat
- `GET /admin/admin/elections/{id}/tokens` - Kelola token
- Dan lainnya...

## Development

### Struktur Folder
```
evoting-app/
├── main.go                 # Entry point aplikasi
├── internal/
│   ├── config/            # Konfigurasi aplikasi
│   ├── database/          # Database setup dan migrasi
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # Middleware (auth, etc)
│   └── models/           # Data models
├── web/
│   ├── templates/        # HTML templates
│   └── static/          # CSS, JS, images
└── README.md
```

### Menambah Fitur Baru

1. Tambahkan model di `internal/models/`
2. Update database schema di `internal/database/`
3. Buat handler di `internal/handlers/`
4. Tambahkan route di `main.go`
5. Buat template HTML di `web/templates/`

## Troubleshooting

### Database Issues
```bash
# Reset database (hapus file database)
rm evoting.db

# Restart aplikasi untuk recreate database
go run main.go
```

### Permission Issues
```bash
# Pastikan file database writable
chmod 644 evoting.db
```

### Port Already in Use
```bash
# Gunakan port lain
PORT=8081 go run main.go
```

## Contributing

1. Fork repository
2. Buat feature branch
3. Commit changes
4. Push ke branch
5. Buat Pull Request

## License

MIT License - lihat file LICENSE untuk detail.

## Support

Untuk pertanyaan atau issue, silakan buat issue di repository ini.
# evoting-go
