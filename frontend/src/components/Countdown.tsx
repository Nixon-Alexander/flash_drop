import { useEffect, useState } from "react";
import { parseDbTime } from "../lib/format";

export function useNow(intervalMs = 1000) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return now;
}

const pad = (n: number) => String(n).padStart(2, "0");

export function Countdown({ to }: { to: string }) {
  const now = useNow();
  const left = Math.max(0, parseDbTime(to).getTime() - now);
  const s = Math.floor(left / 1000);
  const parts = [
    { v: Math.floor(s / 86400), label: "hari" },
    { v: Math.floor((s % 86400) / 3600), label: "jam" },
    { v: Math.floor((s % 3600) / 60), label: "menit" },
    { v: s % 60, label: "detik" },
  ];
  return (
    <div className="clock" role="timer" aria-label={`Sisa waktu ${parts.map((p) => `${p.v} ${p.label}`).join(" ")}`}>
      {parts.map((p) => (
        <div key={p.label} className="clock-seg">
          <span className="clock-num">{pad(p.v)}</span>
          <span className="clock-label">{p.label}</span>
        </div>
      ))}
    </div>
  );
}
