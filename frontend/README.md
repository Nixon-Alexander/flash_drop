# Flash Drop Frontend

React 18 + TypeScript + Vite untuk backend `flash_drop` (branch `dev`). Tanpa library UI/state tambahan, hanya `react-router-dom`.

## Menjalankan

```bash
npm install
cp .env.example .env      # sesuaikan VITE_API_TARGET dengan APP_HOST:APP_PORT backend
npm run dev               # http://localhost:5173
```

Saat development, Vite mem-proxy `/api` dan `/uploads` ke backend, jadi CORS tidak diperlukan.
Build produksi: `npm run build` (hasil di `dist/`).

## Fitur

| Area | Isi |
| --- | --- |
| Toko | Hero dengan hitung mundur drop, grid produk, pencarian, tambah ke keranjang |
| Akun | Daftar, masuk, token JWT disimpan di localStorage, logout otomatis saat 401 |
| Keranjang | Ubah jumlah (maks 10 per produk), hapus item |
| Checkout | Alamat, 6 metode pembayaran yang aktif di backend, lalu redirect ke Midtrans |
| Beli drop | Checkout langsung dengan `flash_sale_id` lewat `POST /flash_sales/checkout` |
| Admin | CRUD produk (upload gambar), CRUD kategori, buat/ubah status/hapus flash sale |

Admin: masuk dengan akun berole `admin` (di backend, role admin harus diatur langsung di database; pendaftaran publik selalu `customer`).

## Hal di backend yang perlu Anda tambahkan atau perhatikan

1. **Gambar tidak disajikan.** Backend menyimpan file di `./uploads` tetapi tidak ada route statisnya. Tambahkan di `main.go`:
   ```go
   router.Static("/uploads", "./uploads")
   ```
2. **Customer tidak bisa melihat flash sale.** Semua `GET /flash_sales` ada di grup admin. Hero dan bagian "drop" di toko memanggil `GET /api/v1/flash_sales/` dan otomatis kosong bila gagal. Cukup satu baris di `routes/user_routes.go`:
   ```go
   userFlashSales.GET("/", ch.flashSale_c.GetAllEvent)
   ```
   Frontend sudah menyaring hanya yang berstatus `active` dan dalam rentang waktunya.
3. **Status flash sale tidak berubah otomatis.** Admin harus mengubahnya ke `active` / `ended` secara manual (ada dropdown di panel).
4. **`/flash_sales/checkout` memakai `BuyItems`, bukan `BuyFlashSale`.** Tetap berfungsi karena frontend selalu mengirim `flash_sale_id`, tetapi `flash_sale_id` tidak akan diwajibkan oleh server.
5. **Edit produk selalu butuh gambar.** `EditProduct` mengambil `ImageUrl[0]`, jadi form edit meminta gambar ulang.
6. **Keranjang tidak dikosongkan oleh checkout.** Frontend mengosongkannya setelah pesanan berhasil dibuat.
7. **Belum ada endpoint riwayat pesanan, daftar kategori untuk customer, atau detail produk.** Karena itu tidak ada halaman riwayat/filter kategori di sisi customer.
8. **Deploy terpisah** (frontend dan backend beda origin) butuh middleware CORS di Gin, lalu isi `VITE_API_BASE`.
9. Login dengan password salah mengembalikan HTTP 500 (bukan 401); pesan errornya tetap ditampilkan apa adanya.
10. Waktu flash sale dikirim sebagai `YYYY-MM-DD HH:MM:SS` tanpa zona waktu; samakan zona waktu database dengan zona Anda.

## Struktur

```
src/
  api/        client (fetch + envelope), endpoints, types (cermin struct Go)
  lib/        auth, cart, toast (context), format (rupiah, tanggal)
  components/ Layout, Countdown, ProductCard, ui (Modal, guard, dll)
  pages/      Shop, Auth, Cart, Checkout, admin/*
```
