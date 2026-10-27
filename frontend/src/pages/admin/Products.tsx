import { useCallback, useEffect, useState, type FormEvent } from "react";
import { adminApi, productApi } from "../../api/endpoints";
import type { Category, Product } from "../../api/types";
import { Empty, ErrorNote, Loading, Modal, ProductImage } from "../../components/ui";
import { rupiah } from "../../lib/format";
import { useToast } from "../../lib/toast";

interface FormState {
  name: string;
  description: string;
  category_id: string;
  price: string;
  stock: string;
  is_active: boolean;
}
const blank: FormState = { name: "", description: "", category_id: "", price: "", stock: "", is_active: true };

export function AdminProducts() {
  const toast = useToast();
  const [products, setProducts] = useState<Product[] | null>(null);
  const [cats, setCats] = useState<Category[]>([]);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<Product | "new" | null>(null);

  const load = useCallback(() => {
    Promise.all([productApi.list(), adminApi.categories()])
      .then(([p, c]) => {
        setProducts(p);
        setCats(c);
      })
      .catch((e: Error) => setError(e.message));
  }, []);
  useEffect(load, [load]);

  const catName = (id: number) => cats.find((c) => c.id === id)?.name ?? "-";

  async function remove(p: Product) {
    if (!confirm(`Hapus "${p.name}"? Produk yang sudah pernah dipesan atau ada di flash sale mungkin tidak bisa dihapus.`)) return;
    try {
      await adminApi.deleteProduct(p.id);
      toast("Produk dihapus");
      load();
    } catch (e) {
      toast((e as Error).message, "error");
    }
  }

  return (
    <>
      <div className="section-head row">
        <h1>Produk</h1>
        <button className="btn btn-primary" onClick={() => setEditing("new")} disabled={cats.length === 0}>Tambah produk</button>
      </div>
      {cats.length === 0 && !error && <p className="note">Buat minimal satu kategori sebelum menambah produk.</p>}
      {error ? <Empty title="Data tidak bisa dimuat">{error}</Empty> : !products ? <Loading /> : products.length === 0 ? (
        <Empty title="Belum ada produk">Tambahkan produk pertama untuk mulai berjualan.</Empty>
      ) : (
        <div className="table-wrap">
          <table>
            <thead><tr><th>Produk</th><th>Kategori</th><th className="num">Harga</th><th className="num">Stok</th><th>Status</th><th /></tr></thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id}>
                  <td><div className="cell-prod"><div className="line-img sm"><ProductImage name={p.image_url} alt="" /></div>{p.name}</div></td>
                  <td>{catName(p.category_id)}</td>
                  <td className="num">{rupiah(p.price)}</td>
                  <td className="num">{p.stock}</td>
                  <td><span className={`pill ${p.is_active ? "pill-on" : ""}`}>{p.is_active ? "Tayang" : "Disembunyikan"}</span></td>
                  <td className="actions">
                    <button className="link-btn" onClick={() => setEditing(p)}>Edit</button>
                    <button className="link-btn danger" onClick={() => remove(p)}>Hapus</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {editing && (
        <ProductForm
          product={editing === "new" ? null : editing}
          cats={cats}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            load();
          }}
        />
      )}
    </>
  );
}

function ProductForm({ product, cats, onClose, onSaved }: { product: Product | null; cats: Category[]; onClose: () => void; onSaved: () => void }) {
  const toast = useToast();
  const [f, setF] = useState<FormState>(
    product
      ? { name: product.name, description: product.description, category_id: String(product.category_id), price: String(product.price), stock: String(product.stock), is_active: product.is_active }
      : { ...blank, category_id: cats[0] ? String(cats[0].id) : "" },
  );
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    // Backend mewajibkan gambar pada pembuatan maupun pengeditan produk
    if (!file) return setError(product ? "Pilih gambar lagi untuk menyimpan perubahan (backend selalu meminta gambar)." : "Gambar produk wajib diisi.");
    const fd = new FormData();
    fd.set("name", f.name);
    fd.set("description", f.description);
    fd.set("category_id", f.category_id);
    fd.set("price", f.price);
    fd.set("stock", f.stock);
    fd.set("is_active", String(f.is_active));
    fd.set("image_url", file);
    setBusy(true);
    try {
      if (product) await adminApi.editProduct(product.id, fd);
      else await adminApi.createProduct(fd);
      toast(product ? "Produk diperbarui" : "Produk ditambahkan");
      onSaved();
    } catch (err) {
      setError((err as Error).message);
      setBusy(false);
    }
  }

  const bind = (k: keyof FormState) => (e: { target: { value: string } }) => setF((s) => ({ ...s, [k]: e.target.value }));

  return (
    <Modal title={product ? "Edit produk" : "Tambah produk"} onClose={onClose}>
      <form className="form" onSubmit={submit}>
        <label>Nama<input value={f.name} onChange={bind("name")} required /></label>
        <label>Kategori
          <select value={f.category_id} onChange={bind("category_id")} required>
            {cats.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </label>
        <label>Deskripsi<textarea rows={3} value={f.description} onChange={bind("description")} /></label>
        <div className="two">
          <label>Harga (Rp)<input type="number" min={0} step="1" value={f.price} onChange={bind("price")} required /></label>
          <label>Stok<input type="number" min={0} step="1" value={f.stock} onChange={bind("stock")} required /></label>
        </div>
        <label>Gambar<input type="file" accept="image/*" onChange={(e) => setFile(e.target.files?.[0] ?? null)} /></label>
        <label className="check"><input type="checkbox" checked={f.is_active} onChange={(e) => setF((s) => ({ ...s, is_active: e.target.checked }))} /> Tayangkan di toko</label>
        <ErrorNote message={error} />
        <button className="btn btn-primary btn-wide" disabled={busy}>{busy ? "Menyimpan…" : "Simpan produk"}</button>
      </form>
    </Modal>
  );
}
