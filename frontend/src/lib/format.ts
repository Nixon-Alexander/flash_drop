const idr = new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 });
export const rupiah = (n: number) => idr.format(Math.round(n));

export const parseDbTime = (s: string) =>
  new Date(/[zZ]$|[+-]\d{2}:?\d{2}$/.test(s) ? s : s.replace(" ", "T") + "Z");

export const dateTime = (iso: string) =>
  parseDbTime(iso).toLocaleString("id-ID", {
    day: "numeric", month: "short", year: "numeric",
    hour: "2-digit", minute: "2-digit",
    timeZone: "Asia/Jakarta",
  }) + " WIB";

// <input type="datetime-local"> -> "YYYY-MM-DD HH:MM:SS" (format yang aman untuk kolom DATETIME MySQL)
export const toOffsetDateTime = (local: string) => {
  const off = -new Date(local).getTimezoneOffset(); // menit di timur UTC
  const sign = off >= 0 ? "+" : "-";
  const abs = Math.abs(off);
  const p = (n: number) => String(n).padStart(2, "0");
  return `${local}:00${sign}${p(Math.floor(abs / 60))}:${p(abs % 60)}`;
};

export const splitName = (full: string) => {
  const [first, ...rest] = full.trim().split(/\s+/);
  return { first: first ?? "", last: rest.join(" ") };
};
