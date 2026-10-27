import { useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { ErrorNote } from "../components/ui";
import { useAuth } from "../lib/auth";

export function Auth({ mode }: { mode: "login" | "register" }) {
  const isLogin = mode === "login";
  const { login, register } = useAuth();
  const nav = useNavigate();
  const from = (useLocation().state as { from?: string } | null)?.from;

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    if (!isLogin && password.length < 8) return setError("Kata sandi minimal 8 karakter.");
    setBusy(true);
    try {
      const u = isLogin ? await login(email, password) : await register(name, email, password);
      nav(u.role === "admin" ? "/admin" : from || "/", { replace: true });
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="auth">
      <h1>{isLogin ? "Masuk" : "Buat akun"}</h1>
      <p className="muted">{isLogin ? "Masuk untuk mengisi keranjang dan ikut drop." : "Satu akun untuk semua drop."}</p>
      <form className="form" onSubmit={submit}>
        {!isLogin && (
          <label>
            Nama
            <input value={name} onChange={(e) => setName(e.target.value)} required minLength={2} maxLength={100} autoComplete="name" />
          </label>
        )}
        <label>
          Email
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoComplete="email" />
        </label>
        <label>
          Kata sandi
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={isLogin ? undefined : 8}
            maxLength={72}
            autoComplete={isLogin ? "current-password" : "new-password"}
          />
        </label>
        <ErrorNote message={error} />
        <button className="btn btn-primary btn-wide" disabled={busy}>
          {busy ? "Memproses…" : isLogin ? "Masuk" : "Daftar"}
        </button>
      </form>
      <p className="muted">
        {isLogin ? (
          <>Belum punya akun? <Link to="/daftar" state={{ from }}>Daftar</Link></>
        ) : (
          <>Sudah punya akun? <Link to="/masuk" state={{ from }}>Masuk</Link></>
        )}
      </p>
    </div>
  );
}
