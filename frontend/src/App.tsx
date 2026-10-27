import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { RequireRole } from "./components/ui";
import { Auth } from "./pages/Auth";
import { Cart } from "./pages/Cart";
import { Checkout } from "./pages/Checkout";
import { Shop } from "./pages/Shop";
import { AdminLayout } from "./pages/admin/AdminLayout";
import { AdminCategories } from "./pages/admin/Categories";
import { AdminFlashSales } from "./pages/admin/FlashSales";
import { AdminProducts } from "./pages/admin/Products";

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Shop />} />
        <Route path="masuk" element={<Auth mode="login" />} />
        <Route path="daftar" element={<Auth mode="register" />} />
        <Route path="keranjang" element={<RequireRole role="customer"><Cart /></RequireRole>} />
        <Route path="checkout" element={<RequireRole role="customer"><Checkout /></RequireRole>} />
        <Route path="admin" element={<RequireRole role="admin"><AdminLayout /></RequireRole>}>
          <Route index element={<AdminProducts />} />
          <Route path="kategori" element={<AdminCategories />} />
          <Route path="flash-sale" element={<AdminFlashSales />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
