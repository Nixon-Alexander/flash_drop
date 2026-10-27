import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { cartApi } from "../api/endpoints";
import type { CartItem } from "../api/types";
import { useAuth } from "./auth";

export const MAX_QTY = 10; // batas per produk di backend

interface CartCtx {
  items: CartItem[];
  count: number;
  subtotal: number;
  loading: boolean;
  add: (productId: number, qty?: number) => Promise<void>;
  setQty: (productId: number, qty: number) => Promise<void>;
  clear: () => Promise<void>;
  refresh: () => Promise<void>;
}

const Ctx = createContext<CartCtx | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const isCustomer = user?.role === "customer";
  const [items, setItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    if (!isCustomer) {
      setItems([]);
      return;
    }
    setLoading(true);
    try {
      setItems(await cartApi.list());
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [isCustomer]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const add = useCallback(
    async (productId: number, qty = 1) => {
      await cartApi.add([{ product_id: productId, quantity: qty }]);
      await refresh();
    },
    [refresh],
  );

  const setQty = useCallback(
    async (productId: number, qty: number) => {
      await cartApi.set([{ product_id: productId, quantity: qty }]);
      await refresh();
    },
    [refresh],
  );

  const clear = useCallback(async () => {
    if (items.length === 0) return;
    await cartApi.set(items.map((i) => ({ product_id: i.product_id, quantity: 0 })));
    await refresh();
  }, [items, refresh]);

  const value = useMemo<CartCtx>(
    () => ({
      items,
      loading,
      count: items.reduce((n, i) => n + i.quantity, 0),
      subtotal: items.reduce((n, i) => n + i.quantity * i.price, 0),
      add, setQty, clear, refresh,
    }),
    [items, loading, add, setQty, clear, refresh],
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useCart() {
  const c = useContext(Ctx);
  if (!c) throw new Error("useCart harus di dalam CartProvider");
  return c;
}
