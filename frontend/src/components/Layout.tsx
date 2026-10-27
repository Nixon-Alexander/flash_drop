import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../lib/auth";
import { useCart } from "../lib/cart";

function Logo() {
  return (
    <Link to="/" className="logo" aria-label="Flash Drop, beranda">
      <svg width="26" height="26" viewBox="0 0 32 32" aria-hidden="true">
        <rect width="32" height="32" fill="var(--cobalt)" />
        <path d="M18 4 8 18h7l-2 10 11-15h-7z" fill="var(--pink)" />
      </svg>
      <span>Flash Drop</span>
    </Link>
  );
}

export function Layout() {
  const { user, logout } = useAuth();
  const { count } = useCart();
  const nav = useNavigate();

  return (
    <>
      <header className="topbar">
        <div className="wrap topbar-in">
          <Logo />
          <nav className="topnav" aria-label="Navigasi utama">
            <NavLink to="/" end>Toko</NavLink>
            {user?.role === "admin" && <NavLink to="/admin">Panel admin</NavLink>}
          </nav>
          <div className="topbar-end">
            {user?.role !== "admin" && (
              <NavLink to="/keranjang" className="cart-link">
                Keranjang
                {count > 0 && <span className="badge">{count}</span>}
              </NavLink>
            )}
            {user ? (
              <>
                <span className="who">{user.name}</span>
                <button
                  className="btn btn-ghost"
                  onClick={() => {
                    logout();
                    nav("/");
                  }}
                >
                  Keluar
                </button>
              </>
            ) : (
              <>
                <Link to="/masuk" className="btn btn-ghost">Masuk</Link>
                <Link to="/daftar" className="btn btn-primary">Daftar</Link>
              </>
            )}
          </div>
        </div>
      </header>
      <main className="wrap page">
        <Outlet />
      </main>
      <footer className="foot">
        <div className="wrap">Flash Drop. Pembayaran lewat Midtrans, pesanan yang belum dibayar otomatis dibatalkan setelah 15 menit.</div>
      </footer>
    </>
  );
}
