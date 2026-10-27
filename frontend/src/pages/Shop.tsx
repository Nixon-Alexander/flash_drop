import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { flashPublicApi, productApi } from "../api/endpoints";
import type { FlashSale, Product } from "../api/types";
import { Countdown, useNow } from "../components/Countdown";
import { ProductCard } from "../components/ProductCard";
import { Empty, Loading } from "../components/ui";
import { useAuth } from "../lib/auth";
import { useCart } from "../lib/cart";
import { dateTime, parseDbTime } from "../lib/format";
import { useToast } from "../lib/toast";

export function Shop() {
  const { user } = useAuth();
  const cart = useCart();
  const toast = useToast();
  const nav = useNavigate();
  const now = useNow(15000);

  const [products, setProducts] = useState<Product[] | null>(null);
  const [error, setError] = useState("");
  const [flash, setFlash] = useState<FlashSale[]>([]);
  const [query, setQuery] = useState("");
  const [busyId, setBusyId] = useState<number | null>(null);

  useEffect(() => {
    productApi.list().then(setProducts).catch((e: Error) => setError(e.message));
    flashPublicApi.list().then(setFlash);
  }, []);
  
  const byId = useMemo(() => new Map((products ?? []).map((p) => [p.id, p])), [products]);
  
  const live = flash.find((f) => f.status === "active" && parseDbTime(f.start_time).getTime() <= now && parseDbTime(f.end_time).getTime() > now);
  const upcoming = flash
    .filter((f) => f.status === "scheduled" && parseDbTime(f.start_time).getTime() > now)
    .sort((a, b) => +parseDbTime(a.start_time) - +parseDbTime(b.start_time))[0];

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase();
    return (products ?? []).filter((p) => p.is_active && (!q || p.name.toLowerCase().includes(q)));
  }, [products, query]);

  async function addToCart(p: Product) {
    if (!user) return nav("/masuk", { state: { from: "/" } });
    if (user.role === "admin") return toast("Akun admin tidak punya keranjang. Masuk sebagai pelanggan untuk berbelanja.", "error");
    setBusyId(p.id);
    try {
      await cart.add(p.id);
      toast(`${p.name} masuk keranjang`);
    } catch (e) {
      toast((e as Error).message, "error");
    } finally {
      setBusyId(null);
    }
  }

  function buyFlash(sale: FlashSale, p: Product, price: number) {
    if (!user) return nav("/masuk", { state: { from: "/" } });
    nav("/checkout", {
      state: {
        flash: { id: sale.id, name: sale.name },
        lines: [{ product_id: p.id, name: p.name, price, image_url: p.image_url, quantity: 1 }],
      },
    });
  }

  return (
    <>
      <section className="hero">
        <div className="hero-copy">
          <h1>Stok terbatas, harga hanya selama hitung mundur.</h1>
          <p>
            Setiap drop membuka beberapa produk dengan harga khusus. Saat stok drop habis atau waktu selesai, harga kembali normal.
          </p>
        </div>

        <aside className="board" aria-live="polite">
          {live ? (
            <>
              <p className="board-title">{live.name} sedang berlangsung</p>
              <Countdown to={live.end_time} />
              <p className="board-foot">Berakhir {dateTime(live.end_time)}</p>
            </>
          ) : upcoming ? (
            <>
              <p className="board-title">{upcoming.name} dimulai sebentar lagi</p>
              <Countdown to={upcoming.start_time} />
              <p className="board-foot">Mulai {dateTime(upcoming.start_time)}</p>
            </>
          ) : (
            <>
              <p className="board-title">Belum ada drop yang berjalan</p>
              <p className="board-idle">Produk reguler di bawah tetap bisa dibeli. Kembali lagi nanti untuk drop berikutnya.</p>
            </>
          )}
        </aside>
      </section>

      {live && (
        <section className="section">
          <div className="section-head">
            <h2>{live.name}</h2>
            {live.description && <p className="muted">{live.description}</p>}
          </div>
          <div className="grid">
            {live.items.map((it) => {
              const p = byId.get(it.product_id);
              if (!p) return null;
              return (
                <ProductCard
                  key={it.product_id}
                  product={p}
                  flashPrice={it.flash_price}
                  flashStock={it.stock}
                  onAdd={() => undefined}
                  onBuyFlash={() => buyFlash(live, p, it.flash_price)}
                />
              );
            })}
          </div>
        </section>
      )}

      <section className="section">
        <div className="section-head row">
          <h2>Semua produk</h2>
          <input
            className="search"
            type="search"
            placeholder="Cari produk"
            aria-label="Cari produk"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        {error ? (
          <Empty title="Produk tidak bisa dimuat">{error}</Empty>
        ) : !products ? (
          <Loading text="Memuat produk…" />
        ) : visible.length === 0 ? (
          <Empty title={query ? "Tidak ada produk yang cocok" : "Belum ada produk"}>
            {query ? "Coba kata kunci lain." : "Admin belum menambahkan produk."}
          </Empty>
        ) : (
          <div className="grid">
            {visible.map((p) => (
              <ProductCard key={p.id} product={p} busy={busyId === p.id} onAdd={() => addToCart(p)} />
            ))}
          </div>
        )}
      </section>
    </>
  );
}
