import { useState, type FormEvent } from "react";
import { Link, Navigate, useLocation } from "react-router-dom";
import { checkoutApi } from "../api/endpoints";
import type { PaymentMethod } from "../api/types";
import { ErrorNote, ProductImage } from "../components/ui";
import { useAuth } from "../lib/auth";
import { MAX_QTY, useCart } from "../lib/cart";
import { rupiah, splitName } from "../lib/format";

interface Line {
  product_id: number;
  name: string;
  price: number;
  image_url: string;
  quantity: number;
}
interface DirectState {
  flash?: { id: number; name: string };
  lines?: Line[];
}

const SHIPPING_COST = 10000; // backend memakai tarif tetap Rp10.000 (calculateShipping)

const METHODS: { id: PaymentMethod; label: string; hint: string }[] = [
  { id: "gopay", label: "GoPay", hint: "Dompet digital" },
  { id: "shopeepay", label: "ShopeePay", hint: "Dompet digital" },
  { id: "bca_va", label: "BCA Virtual Account", hint: "Transfer bank" },
  { id: "bni_va", label: "BNI Virtual Account", hint: "Transfer bank" },
  { id: "permata_va", label: "Permata Virtual Account", hint: "Transfer bank" },
  { id: "cimb_va", label: "CIMB Niaga Virtual Account", hint: "Transfer bank" },
];

export function Checkout() {
  const { user } = useAuth();
  const cart = useCart();
  const direct = (useLocation().state as DirectState | null) ?? {};
  const isFlash = !!direct.flash && !!direct.lines?.length;

  const [flashLines, setFlashLines] = useState<Line[]>(direct.lines ?? []);
  const lines: Line[] = isFlash
    ? flashLines
    : cart.items.map((i) => ({ product_id: i.product_id, name: i.name, price: i.price, image_url: i.image_url, quantity: i.quantity }));

  const init = splitName(user?.name ?? "");
  const [f, setF] = useState({ first_name: init.first, last_name: init.last, email: user?.email ?? "", phone: "", address: "", city: "", postal_code: "" });
  const [method, setMethod] = useState<PaymentMethod>("gopay");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  if (!isFlash && !cart.loading && cart.items.length === 0) return <Navigate to="/keranjang" replace />;

  const subtotal = lines.reduce((n, l) => n + l.price * l.quantity, 0);
  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF((s) => ({ ...s, [k]: e.target.value }));

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const payload = {
        shipping_detail: { ...f, country_code: "IDN", shipping_cost: SHIPPING_COST, shipping_name: "Reguler" },
        payment_type: method,
        items: lines.map((l) => ({ product_id: l.product_id, quantity: l.quantity })),
      };
      const res = isFlash
        ? await checkoutApi.buyFlash({ ...payload, flash_sale_id: direct.flash!.id })
        : await checkoutApi.buy(payload);

      // Backend tidak mengosongkan keranjang saat checkout, jadi dikosongkan di sini
      // setelah pesanan berhasil dibuat (stok sudah dikunci selama 15 menit).
      if (!isFlash) await cart.clear().catch(() => undefined);
      window.location.assign(res.redirect_url);
    } catch (err) {
      setError((err as Error).message);
      setBusy(false);
    }
  }

  return (
    <form className="split" onSubmit={submit}>
      <section>
        <h1>{isFlash ? `Checkout drop: ${direct.flash!.name}` : "Checkout"}</h1>

        <fieldset className="fields">
          <legend>Alamat pengiriman</legend>
          <div className="two">
            <label>Nama depan<input value={f.first_name} onChange={set("first_name")} required autoComplete="given-name" /></label>
            <label>Nama belakang<input value={f.last_name} onChange={set("last_name")} autoComplete="family-name" /></label>
          </div>
          <div className="two">
            <label>Email<input type="email" value={f.email} onChange={set("email")} required autoComplete="email" /></label>
            <label>Telepon<input type="tel" value={f.phone} onChange={set("phone")} required autoComplete="tel" placeholder="08xxxxxxxxxx" /></label>
          </div>
          <label>Alamat lengkap<input value={f.address} onChange={set("address")} required autoComplete="street-address" /></label>
          <div className="two">
            <label>Kota<input value={f.city} onChange={set("city")} required autoComplete="address-level2" /></label>
            <label>Kode pos<input value={f.postal_code} onChange={set("postal_code")} required inputMode="numeric" autoComplete="postal-code" /></label>
          </div>
        </fieldset>

        <fieldset className="fields">
          <legend>Metode pembayaran</legend>
          <div className="methods">
            {METHODS.map((m) => (
              <label key={m.id} className={`method${method === m.id ? " is-on" : ""}`}>
                <input type="radio" name="pay" value={m.id} checked={method === m.id} onChange={() => setMethod(m.id)} />
                <span>{m.label}</span>
                <small>{m.hint}</small>
              </label>
            ))}
          </div>
        </fieldset>
      </section>

      <aside className="summary">
        <h2>Pesanan</h2>
        <ul className="mini">
          {lines.map((l) => (
            <li key={l.product_id}>
              <div className="line-img sm"><ProductImage name={l.image_url} alt={l.name} /></div>
              <div>
                <strong>{l.name}</strong>
                {isFlash ? (
                  <div className="stepper sm">
                    <button type="button" aria-label="Kurangi" disabled={l.quantity <= 1} onClick={() => setFlashLines((s) => s.map((x) => (x.product_id === l.product_id ? { ...x, quantity: x.quantity - 1 } : x)))}>−</button>
                    <span>{l.quantity}</span>
                    <button type="button" aria-label="Tambah" disabled={l.quantity >= MAX_QTY} onClick={() => setFlashLines((s) => s.map((x) => (x.product_id === l.product_id ? { ...x, quantity: x.quantity + 1 } : x)))}>+</button>
                  </div>
                ) : (
                  <span className="muted small">{l.quantity} × {rupiah(l.price)}</span>
                )}
              </div>
              <span>{rupiah(l.price * l.quantity)}</span>
            </li>
          ))}
        </ul>
        <dl>
          <div><dt>Subtotal</dt><dd>{rupiah(subtotal)}</dd></div>
          <div><dt>Ongkos kirim</dt><dd>{rupiah(SHIPPING_COST)}</dd></div>
          <div className="total"><dt>Total</dt><dd>{rupiah(subtotal + SHIPPING_COST)}</dd></div>
        </dl>
        <ErrorNote message={error} />
        <button className="btn btn-hot btn-wide" disabled={busy}>{busy ? "Membuat pesanan…" : "Bayar sekarang"}</button>
        <p className="muted small">
          Stok dikunci 15 menit. Pesanan yang belum dibayar akan dibatalkan otomatis dan stok dikembalikan.
          {!isFlash && <> <Link to="/keranjang">Ubah keranjang</Link></>}
        </p>
      </aside>
    </form>
  );
}
