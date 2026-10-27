import { rupiah } from "../lib/format";
import type { Product } from "../api/types";
import { ProductImage } from "./ui";

interface Props {
  product: Product;
  flashPrice?: number;
  flashStock?: number;
  busy?: boolean;
  onAdd: () => void;
  onBuyFlash?: () => void;
}

export function ProductCard({ product, flashPrice, flashStock, busy, onAdd, onBuyFlash }: Props) {
  const isFlash = flashPrice !== undefined;
  const stock = isFlash ? Math.min(product.stock, flashStock ?? product.stock) : product.stock;
  const soldOut = stock <= 0;
  const price = isFlash ? flashPrice : product.price;
  console.log(isFlash)

  return (
    <article className={`tile${isFlash ? " tile-flash" : ""}`}>
      <div className="tile-media">
        <ProductImage name={product.image_url} alt={product.name} />
        <div className="sticker" aria-label={`Harga ${rupiah(price)}`}>
          {isFlash && <s>{rupiah(product.price)}</s>}
          <strong>{rupiah(price)}</strong>
        </div>
        {soldOut && <div className="soldout">Habis</div>}
      </div>
      <div className="tile-body">
        <h3>{product.name}</h3>
        <p className="muted">{soldOut ? "Stok habis" : stock <= 5 ? `Sisa ${stock}` : "Stok tersedia"}</p>
        {isFlash && onBuyFlash ? (
          <button className="btn btn-hot" disabled={soldOut} onClick={onBuyFlash}>
            Beli sekarang
          </button>
        ) : (
          <button className="btn btn-primary" disabled={soldOut || busy} onClick={onAdd}>
            {busy ? "Menambahkan…" : "Tambah ke keranjang"}
          </button>
        )}
      </div>
    </article>
  );
}
