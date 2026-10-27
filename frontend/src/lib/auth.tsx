import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { authApi } from "../api/endpoints";
import { setUnauthorizedHandler, tokenStore } from "../api/client";
import type { AuthPayload, Role } from "../api/types";

export interface User {
  name: string;
  email: string;
  role: Role;
}

interface AuthCtx {
  user: User | null;
  login: (email: string, password: string) => Promise<User>;
  register: (name: string, email: string, password: string) => Promise<User>;
  logout: () => void;
}

const USER_KEY = "fd_user";
const Ctx = createContext<AuthCtx | null>(null);

// Cek kedaluwarsa token JWT di sisi klien (hanya untuk UX; server tetap memverifikasi)
function tokenAlive(token: string | null) {
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")));
    return !payload.exp || payload.exp * 1000 > Date.now();
  } catch {
    return false;
  }
}

function loadUser(): User | null {
  if (!tokenAlive(tokenStore.get())) return null;
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) ?? "null");
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(loadUser);

  const logout = useCallback(() => {
    tokenStore.clear();
    localStorage.removeItem(USER_KEY);
    setUser(null);
  }, []);

  const persist = useCallback((p: AuthPayload) => {
    const u: User = { name: p.name, email: p.email, role: p.role };
    tokenStore.set(p.token);
    localStorage.setItem(USER_KEY, JSON.stringify(u));
    setUser(u);
    return u;
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(logout);
    return () => setUnauthorizedHandler(null);
  }, [logout]);

  const value = useMemo<AuthCtx>(
    () => ({
      user,
      logout,
      login: async (email, password) => persist(await authApi.login(email, password)),
      register: async (name, email, password) => persist(await authApi.register(name, email, password)),
    }),
    [user, logout, persist],
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth() {
  const c = useContext(Ctx);
  if (!c) throw new Error("useAuth harus di dalam AuthProvider");
  return c;
}
