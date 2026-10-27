import { useCallback, useEffect, useState, type FormEvent } from "react";
import { adminApi, productApi } from "../../api/endpoints";
import type { FlashSale, FlashSaleItem, FlashSaleStatus, Product } from "../../api/types";
import { Empty, ErrorNote, Loading, Modal } from "../../components/ui";
import { dateTime, rupiah, toOffsetDateTime } from "../../lib/format";
import { useToast } from "../../lib/toast";

const STATUS: Record<FlashSaleStatus, string> = { scheduled: "Terjadwal", active: "Aktif", ended: "Selesai", cancelled: "Dibatalkan" };

export function AdminFlashSales() {
  const toast = useToast();
  const [sales, setSales] = useState<FlashSale[] | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState("");
  const [creating, setCreating] = useState(false);

  const load = useCallback(() => {
    Promise.all([adminApi.flashSales(), productApi.list()])
      .then(([s, p]) => {
        setSales(s);
        setProducts(p);
      })
      .catch((e: Error) => setError(e.message));
  }, []);
  useEffect(load, [load]);

  const pName = (id: number) => products.find((p) => p.id === id)?.name ?? `Produk #${id}`;

  async function run(fn: () => Promise<unknown>, ok: string) {
    try {
      await fn();
      toast(ok);
      load();
    } catch (e) {
      toast((e as Error).message, "error");
    }
  }

  return (
    <>
      <div className="section-head row">
        <h1>Flash sale</h1>
        <button className="btn btn-primary" onClick={() => setCreating(true)} disabled={products.length === 0}>Buat drop</button>
      </div>
      <p className="note">Status tidak berubah otomatis di backend. Ubah ke "Aktif" saat drop dimulai dan ke "Selesai" saat berakhir.</p>

      {error ? <Empty title="Data tidak bisa dimuat">{error}</Empty> : !sales ? <Loading /> : sales.length === 0 ? (
        <Empty title="Belum ada drop">Buat drop pertama: pilih produk, harga flash, dan jumlah stoknya.</Empty>
      ) : (
        <div className="stack">
          {sales.map((s) => (
            <article key={s.id} className="sale">
              <header>
                <div>
                  <h3>{s.name}</h3>
                  <p className="muted small">{dateTime(s.start_time)} sampai {dateTime(s.end_time)}</p>
                </div>
                <span className={`pill pill-${s.status}`}>{STATUS[s.status]}</span>
              </header>
              {s.description && <p>{s.description}</p>}
              <ul className="sale-items">
                {s.items.map((it) => (
                  <li key={it.product_id}><span>{pName(it.product_id)}</span><span>{rupiah(it.flash_price)}, kuota {it.stock}</span></li>
                ))}
              </ul>
              <footer>
                <label className="inline-select">Status
                  <select value={s.status} onChange={(e) => run(() => adminApi.setFlashStatus(s.id, e.target.value as FlashSaleStatus), "Status diperbarui")}>
                    {(Object.keys(STATUS) as FlashSaleStatus[]).map((k) => <option key={k} value={k}>{STATUS[k]}</option>)}
                  </select>
                </label>
                <button className="link-btn danger" onClick={() => confirm(`Hapus drop "${s.name}"?`) && run(() => adminApi.deleteFlashSale(s.id), "Drop dihapus")}>Hapus</button>
              </footer>
            </article>
          ))}
        </div>
      )}
      {creating && <CreateForm products={products} onClose={() => setCreating(false)} onSaved={() => { setCreating(false); load(); }} />}
    </>
  );
}

function CreateForm({ products, onClose, onSaved }: { products: Product[]; onClose: () => void; onSaved: () => void }) {
  const toast = useToast();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [start, setStart] = useState("");
  const [end, setEnd] = useState("");
  const [items, setItems] = useState<{ product_id: string; flash_price: string; stock: string }[]>([
    { product_id: String(products[0]?.id ?? ""), flash_price: "", stock: "" },
  ]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const patch = (i: number, k: string, v: string) => setItems((l) => l.map((x, idx) => (idx === i ? { ...x, [k]: v } : x)));

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    if (new Date(end) <= new Date(start)) return setError("Waktu berakhir harus setelah waktu mulai.");
    const ids = items.map((i) => i.product_id);
    if (new Set(ids).size !== ids.length) return setError("Satu produk hanya boleh muncul sekali dalam satu drop.");
    const payloadItems: FlashSaleItem[] = items.map((i) => ({ product_id: Number(i.product_id), flash_price: Number(i.flash_price), stock: Number(i.stock) }));
    setBusy(true);
    try {
      await adminApi.createFlashSale({
        name,
        description,
        start_time: toOffsetDateTime(start),
        end_time: toOffsetDateTime(end),
        items: payloadItems,
      });
      toast("Drop dibuat dengan status terjadwal");
      onSaved();
    } catch (err) {
      setError((err as Error).message);
      setBusy(false);
    }
  }

  return (
    <Modal title="Buat drop baru" onClose={onClose}>
      <form className="form" onSubmit={submit}>
        <label>Nama drop<input value={name} onChange={(e) => setName(e.target.value)} required /></label>
        <label>Deskripsi<input value={description} onChange={(e) => setDescription(e.target.value)} /></label>
        <div className="two">
          <label>Mulai<input type="datetime-local" value={start} onChange={(e) => setStart(e.target.value)} required /></label>
          <label>Berakhir<input type="datetime-local" value={end} onChange={(e) => setEnd(e.target.value)} required /></label>
        </div>
        <fieldset className="fields">
          <legend>Produk dalam drop</legend>
          {items.map((it, i) => (
            <div key={i} className="item-row">
              <select value={it.product_id} onChange={(e) => patch(i, "product_id", e.target.value)} aria-label="Produk">
                {products.map((p) => <option key={p.id} value={p.id}>{p.name} ({rupiah(p.price)})</option>)}
              </select>
              <input type="number" min={1} placeholder="Harga flash" aria-label="Harga flash" value={it.flash_price} onChange={(e) => patch(i, "flash_price", e.target.value)} required />
              <input type="number" min={1} placeholder="Kuota" aria-label="Kuota" value={it.stock} onChange={(e) => patch(i, "stock", e.target.value)} required />
              <button type="button" className="link-btn danger" disabled={items.length === 1} onClick={() => setItems((l) => l.filter((_, idx) => idx !== i))}>Hapus</button>
            </div>
          ))}
          <button type="button" className="btn btn-ghost" onClick={() => setItems((l) => [...l, { product_id: String(products[0]?.id ?? ""), flash_price: "", stock: "" }])}>Tambah produk</button>
        </fieldset>
        <p className="muted small">
          {start && end
            ? <>Dikirim sebagai {toOffsetDateTime(start)} sampai {toOffsetDateTime(end)}.</>
            : "Waktu dikirim lengkap dengan offset zona waktu perangkat Anda."}
        </p>
        <ErrorNote message={error} />
        <button className="btn btn-primary btn-wide" disabled={busy}>{busy ? "Menyimpan…" : "Buat drop"}</button>
      </form>
    </Modal>
  );
}
