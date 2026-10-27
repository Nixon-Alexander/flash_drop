import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Empty, Loading, ProductImage } from "../components/ui";
import { MAX_QTY, useCart } from "../lib/cart";
import { rupiah } from "../lib/format";
import { useToast } from "../lib/toast";

export function Cart() {
  const { items, subtotal, loading, setQty } = useCart();
  const toast = useToast();
  const nav = useNavigate();
  const [busy, setBusy] = useState<number | null>(null);

  async function change(productId: number, qty: number) {
    setBusy(productId);
    try {
      await setQty(productId, qty);
    } catch (e) {
      toast((e as Error).message, "error");
    } finally {
      setBusy(null);
    }
  }

  if (loading && items.length === 0) return <Loading text="Memuat keranjang…" />;

  if (items.length === 0) {
    return (
      <Empty title="Keranjang masih kosong">
        Pilih produk di toko, lalu kembali ke sini untuk membayar. <Link to="/">Ke toko</Link>
      </Empty>
    );
  }

  return (
    <div className="split">
      <section>
        <h1>Keranjang</h1>
        <ul className="lines">
          {items.map((it) => (
            <li key={it.cart_items_id} className="line">
              <div className="line-img"><ProductImage name={it.image_url} alt={it.name} /></div>
              <div className="line-info">
                <h3>{it.name}</h3>
                <p className="muted">{it.category_name}</p>
                <button className="link-btn" disabled={busy === it.product_id} onClick={() => change(it.product_id, 0)}>
                  Hapus
                </button>
              </div>
              <div className="stepper" role="group" aria-label={`Jumlah ${it.name}`}>
                <button disabled={busy === it.product_id} onClick={() => change(it.product_id, it.quantity - 1)} aria-label="Kurangi">−</button>
                <span>{it.quantity}</span>
                <button disabled={busy === it.product_id || it.quantity >= MAX_QTY} onClick={() => change(it.product_id, it.quantity + 1)} aria-label="Tambah">+</button>
              </div>
              <div className="line-price">{rupiah(it.price * it.quantity)}</div>
            </li>
          ))}
        </ul>
      </section>
      <aside className="summary">
        <h2>Ringkasan</h2>
        <dl>
          <div><dt>Subtotal</dt><dd>{rupiah(subtotal)}</dd></div>
          <div><dt>Ongkos kirim</dt><dd>Dihitung saat checkout</dd></div>
        </dl>
        <button className="btn btn-primary btn-wide" onClick={() => nav("/checkout")}>Lanjut ke pembayaran</button>
        <p className="muted small">Maksimal {MAX_QTY} per produk dalam satu pesanan.</p>
      </aside>
    </div>
  );
}
