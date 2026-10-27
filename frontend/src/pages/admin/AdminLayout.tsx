import { NavLink, Outlet } from "react-router-dom";

export function AdminLayout() {
  return (
    <div className="admin">
      <nav className="admin-nav" aria-label="Menu admin">
        <NavLink to="/admin" end>Produk</NavLink>
        <NavLink to="/admin/kategori">Kategori</NavLink>
        <NavLink to="/admin/flash-sale">Flash sale</NavLink>
      </nav>
      <div className="admin-body">
        <Outlet />
      </div>
    </div>
  );
}
