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

## Struktur

```
src/
  api/        client (fetch + envelope), endpoints, types (cermin struct Go)
  lib/        auth, cart, toast (context), format (rupiah, tanggal)
  components/ Layout, Countdown, ProductCard, ui (Modal, guard, dll)
  pages/      Shop, Auth, Cart, Checkout, admin/*
```
