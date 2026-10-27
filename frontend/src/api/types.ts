// Bentuk data mengikuti struct JSON di backend Go (backend/types/*.go)

export type Role = "customer" | "admin";

export interface ApiEnvelope<T> {
  status: "success" | "error";
  code: number;
  data: T;
  message: string;
}

export interface AuthPayload {
  id?: number;
  name: string;
  email: string;
  role: Role;
  token: string;
}

export interface Product {
  id: number;
  category_id: number;
  name: string;
  slug: string;
  description: string;
  price: number;
  stock: number;
  image_url: string; // nama file di folder uploads backend
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: number;
  name: string;
  slug: string;
}

export interface CartItem {
  cart_items_id: number;
  cart_id: number;
  product_id: number;
  quantity: number;
  category_name: string;
  slug: string;
  name: string;
  price: number;
  image_url: string;
}

export interface ItemInput {
  product_id: number;
  quantity: number;
}

export type FlashSaleStatus = "scheduled" | "active" | "ended" | "cancelled";

export interface FlashSaleItem {
  product_id: number;
  flash_price: number;
  stock: number;
}

export interface FlashSale {
  id: number;
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  status: FlashSaleStatus;
  items: FlashSaleItem[];
}

export interface FlashSaleInput {
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  items: FlashSaleItem[];
}

export type PaymentMethod =
  | "gopay"
  | "shopeepay"
  | "permata_va"
  | "bca_va"
  | "bni_va"
  | "cimb_va";

export interface ShippingDetail {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  address: string;
  city: string;
  postal_code: string;
  country_code: string;
  shipping_cost: number;
  shipping_name: string;
}

export interface CheckoutInput {
  shipping_detail: ShippingDetail;
  payment_type: PaymentMethod;
  flash_sale_id?: number;
  items: ItemInput[];
}

export interface CheckoutResult {
  token: string;
  redirect_url: string;
}
