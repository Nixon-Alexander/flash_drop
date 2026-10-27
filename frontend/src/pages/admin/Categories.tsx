import { useCallback, useEffect, useState, type FormEvent } from "react";
import { adminApi } from "../../api/endpoints";
import type { Category } from "../../api/types";
import { Empty, Loading } from "../../components/ui";
import { useToast } from "../../lib/toast";

export function AdminCategories() {
  const toast = useToast();
  const [cats, setCats] = useState<Category[] | null>(null);
  const [error, setError] = useState("");
  const [name, setName] = useState("");
  const [editId, setEditId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");

  const load = useCallback(() => {
    adminApi.categories().then(setCats).catch((e: Error) => setError(e.message));
  }, []);
  useEffect(load, [load]);

  async function run(fn: () => Promise<unknown>, ok: string) {
    try {
      await fn();
      toast(ok);
      load();
    } catch (e) {
      toast((e as Error).message, "error");
    }
  }

  async function add(e: FormEvent) {
    e.preventDefault();
    await run(() => adminApi.createCategory(name.trim()), "Kategori ditambahkan");
    setName("");
  }

  return (
    <>
      <div className="section-head"><h1>Kategori</h1></div>
      <form className="inline-form" onSubmit={add}>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Nama kategori baru" aria-label="Nama kategori baru" required />
        <button className="btn btn-primary">Tambah kategori</button>
      </form>
      {error ? <Empty title="Data tidak bisa dimuat">{error}</Empty> : !cats ? <Loading /> : cats.length === 0 ? (
        <Empty title="Belum ada kategori">Kategori dipakai untuk mengelompokkan produk.</Empty>
      ) : (
        <ul className="rows">
          {cats.map((c) => (
            <li key={c.id}>
              {editId === c.id ? (
                <form
                  className="inline-form grow"
                  onSubmit={async (e) => {
                    e.preventDefault();
                    await run(() => adminApi.editCategory(c.id, editName.trim()), "Kategori diperbarui");
                    setEditId(null);
                  }}
                >
                  <input value={editName} onChange={(e) => setEditName(e.target.value)} aria-label="Nama kategori" required autoFocus />
                  <button className="btn btn-primary">Simpan</button>
                  <button type="button" className="btn btn-ghost" onClick={() => setEditId(null)}>Batal</button>
                </form>
              ) : (
                <>
                  <span className="grow"><strong>{c.name}</strong> <span className="muted small">{c.slug}</span></span>
                  <button className="link-btn" onClick={() => { setEditId(c.id); setEditName(c.name); }}>Ubah nama</button>
                  <button
                    className="link-btn danger"
                    onClick={() => confirm(`Hapus kategori "${c.name}"?`) && run(() => adminApi.deleteCategory(c.id), "Kategori dihapus")}
                  >
                    Hapus
                  </button>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </>
  );
}
