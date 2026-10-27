import type { ApiEnvelope } from "./types";

const BASE = (import.meta.env.VITE_API_BASE ?? "") + "/api/v1";
const TOKEN_KEY = "fd_token";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

export const tokenStore = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

type Listener = () => void;
let onUnauthorized: Listener | null = null;
export const setUnauthorizedHandler = (fn: Listener | null) => {
  onUnauthorized = fn;
};

interface Options {
  method?: string;
  json?: unknown;
  form?: FormData;
  auth?: boolean;
}

export async function request<T>(path: string, opts: Options = {}): Promise<T> {
  const headers: Record<string, string> = {};
  const token = tokenStore.get();
  if (opts.auth !== false && token) headers.Authorization = `Bearer ${token}`;

  let body: BodyInit | undefined;
  if (opts.form) {
    body = opts.form; // browser mengatur boundary multipart sendiri
  } else if (opts.json !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(opts.json);
  }

  let res: Response;
  try {
    res = await fetch(BASE + path, { method: opts.method ?? "GET", headers, body });
  } catch {
    throw new ApiError("Tidak bisa terhubung ke server. Pastikan backend sedang berjalan.", 0);
  }

  let payload: Partial<ApiEnvelope<T>> | null = null;
  try {
    payload = await res.json();
  } catch {
    /* body kosong / bukan JSON */
  }

  if (!res.ok || payload?.status === "error") {
    if (res.status === 401 && opts.auth !== false) onUnauthorized?.();
    throw new ApiError(payload?.message || `Permintaan gagal (${res.status})`, res.status);
  }
  return payload?.data as T;
}

// Gambar produk disimpan backend sebagai nama file; disajikan lewat /uploads/<nama>
export const imageSrc = (name?: string) => (name ? `${import.meta.env.VITE_API_BASE ?? ""}/uploads/${name}` : "");
