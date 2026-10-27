import { useEffect, useState, type ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { imageSrc } from "../api/client";
import type { Role } from "../api/types";
import { useAuth } from "../lib/auth";

const PLACEHOLDER =
  "data:image/svg+xml;utf8," +
  encodeURIComponent(
    `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 80 100'><rect width='80' height='100' fill='#DDE1F2'/><path d='M44 26 28 54h12l-4 20 18-30H42z' fill='#B6BDDA'/></svg>`,
  );

export function ProductImage({ name, alt }: { name?: string; alt: string }) {
  const [src, setSrc] = useState(imageSrc(name) || PLACEHOLDER);
  useEffect(() => setSrc(imageSrc(name) || PLACEHOLDER), [name]);
  return <img src={src} alt={alt} loading="lazy" onError={() => setSrc(PLACEHOLDER)} />;
}

export function RequireRole({ role, children }: { role: Role; children: ReactNode }) {
  const { user } = useAuth();
  const loc = useLocation();
  if (!user) return <Navigate to="/masuk" replace state={{ from: loc.pathname + loc.search }} />;
  if (user.role !== role) return <Navigate to={user.role === "admin" ? "/admin" : "/"} replace />;
  return <>{children}</>;
}

export function Empty({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="empty">
      <h3>{title}</h3>
      {children && <p>{children}</p>}
    </div>
  );
}

export function Loading({ text = "Memuat…" }: { text?: string }) {
  return <p className="loading">{text}</p>;
}

export function ErrorNote({ message }: { message: string }) {
  return message ? (
    <p className="form-error" role="alert">
      {message}
    </p>
  ) : null;
}

export function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);
  return (
    <div className="modal-backdrop" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal" role="dialog" aria-modal="true" aria-label={title}>
        <div className="modal-head">
          <h2>{title}</h2>
          <button className="btn btn-ghost" onClick={onClose} aria-label="Tutup">
            Tutup
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}
