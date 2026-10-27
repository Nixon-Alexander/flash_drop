import { request } from "./client";
import type {
  AuthPayload, CartItem, Category, CheckoutInput, CheckoutResult,
  FlashSale, FlashSaleInput, FlashSaleStatus, ItemInput, Product,
} from "./types";

// Catatan: route Gin didefinisikan dengan "/" di akhir, jadi path dibuat persis begitu
// agar tidak terkena redirect 301/307.

export const authApi = {
  login: (email: string, password: string) =>
    request<AuthPayload>("/login", { method: "POST", json: { email, password }, auth: false }),
  register: (name: string, email: string, password: string) =>
    request<AuthPayload>("/signup", { method: "POST", json: { name, email, password }, auth: false }),
};

export const productApi = {
  list: async () => (await request<Product[] | null>("/product/", { auth: false })) ?? [],
};

export const cartApi = {
  list: async () => (await request<CartItem[] | null>("/cart/")) ?? [],
  add: (item: ItemInput[]) => request<null>("/cart/", { method: "POST", json: { item } }),
  set: (item: ItemInput[]) => request<null>("/cart/", { method: "PUT", json: { item } }), // quantity 0 = hapus
};

export const checkoutApi = {
  buy: (input: CheckoutInput) => request<CheckoutResult>("/checkout/", { method: "POST", json: input }),
  buyFlash: (input: CheckoutInput) =>
    request<CheckoutResult>("/flash_sales/checkout", { method: "POST", json: input }),
};

export const flashPublicApi = {
  // Backend belum punya route ini untuk customer (lihat README). Gagal = daftar kosong.
  list: async () => {
    try {
      return (await request<FlashSale[] | null>("/flash_sales/")) ?? [];
    } catch {
      return [] as FlashSale[];
    }
  },
};

export const adminApi = {
  categories: async () => (await request<Category[] | null>("/admin/category/")) ?? [],
  createCategory: (name: string) => request<{ id: number }>("/admin/category/", { method: "POST", json: { name } }),
  editCategory: (id: number, name: string) => request<unknown>(`/admin/category/${id}`, { method: "PUT", json: { name } }),
  deleteCategory: (id: number) => request<unknown>(`/admin/category/${id}`, { method: "DELETE" }),

  createProduct: (form: FormData) => request<{ id: number }>("/admin/product/", { method: "POST", form }),
  editProduct: (id: number, form: FormData) => request<unknown>(`/admin/product/${id}`, { method: "PUT", form }),
  deleteProduct: (id: number) => request<unknown>(`/admin/product/${id}`, { method: "DELETE" }),

  flashSales: async () => (await request<FlashSale[] | null>("/admin/flash_sales/")) ?? [],
  createFlashSale: (input: FlashSaleInput) => request<null>("/admin/flash_sales/", { method: "POST", json: input }),
  setFlashStatus: (id: number, status: FlashSaleStatus) =>
    request<null>(`/admin/flash_sales/${id}/status`, { method: "PATCH", json: { status } }),
  deleteFlashSale: (id: number) => request<null>(`/admin/flash_sales/${id}`, { method: "DELETE" }),
};
