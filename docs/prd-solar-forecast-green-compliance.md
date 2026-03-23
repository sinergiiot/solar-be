# PRD — Solar Forecast: Green Compliance & Incentive Platform

**Version:** 2.0  
**Status:** Draft  
**Owner:** PT Sinergi IoT Nusantara  
**Last Updated:** Maret 2026

---

## 1. Problem Statement

Pemilik PLTS di Indonesia — dari skala rumah tangga hingga industri — tidak memiliki
sistem yang membantu mereka **membuktikan dan mengklaim** keuntungan finansial, fiskal,
dan reputasi dari investasi energi surya mereka. Data produksi ada, tapi tidak tersusun
dalam format yang bisa dilampirkan ke instansi pemerintah, lembaga REC, atau laporan ESG.

Platform Solar Forecast sudah menghasilkan data harian (kWh, hemat biaya, CO2 avoided).
Dua gap yang perlu ditutup secara bersamaan:

1. **Lapisan pelaporan resmi** — format yang bisa dilampirkan ke instansi, kumulatif jangka panjang
2. **Model bisnis berkelanjutan** — tier system yang mendorong konversi sambil tetap memberi nilai di free tier

---

## 2. Target Pengguna & Tier Mapping

| Segmen                 | Kapasitas  | Tier             | Kebutuhan Utama                               |
| ---------------------- | ---------- | ---------------- | --------------------------------------------- |
| Rumah tangga           | < 10 kWp   | Free → Pro       | Forecast harian, laporan hemat tagihan        |
| UKM / Pemilik bangunan | 10–50 kWp  | Pro              | Laporan bulanan PDF, insentif PBB, CSV export |
| Industri menengah      | 50–100 kWp | Pro → Enterprise | REC readiness, CSR reporting, multi-site      |
| Industri besar / EPC   | > 100 kWp  | Enterprise       | Carbon credit MRV, ESG dashboard, white-label |

---

## 3. Tier Structure

| Fitur                         |     🆓 Free     | ⭐ Pro (Rp 99rb/bln)  | 🏢 Enterprise (Rp 499rb/bln) |
| ----------------------------- | :-------------: | :-------------------: | :--------------------------: |
| Solar Profile                 |     1 site      |        5 site         |          Unlimited           |
| Device IoT                    |    1 device     |       10 device       |          Unlimited           |
| Forecast Harian               |       ✅        |          ✅           |              ✅              |
| History Forecast & Actual     |     7 hari      |        90 hari        |    Unlimited + CSV Export    |
| Notifikasi Email              |  ✅ (1x/hari)   |          ✅           |              ✅              |
| Notifikasi Telegram/WA        |       ❌        |          ✅           |              ✅              |
| Weather Risk Status           |   Label saja    | ✅ + Rekomendasi aksi |    ✅ + Rekomendasi aksi     |
| Estimasi Produksi per Periode | Pagi/Siang/Sore |  ✅ + Trend mingguan  |      ✅ + Trend bulanan      |
| **Laporan Bulanan PDF**       |       ❌        |          ✅           |    ✅ + White-label logo     |
| **Surat Keterangan PBB**      |       ❌        |          ✅           |              ✅              |
| **MWh Accumulator & REC**     |   Angka saja    |  ✅ + Notifikasi REC  |       ✅ + Laporan PDF       |
| **CO2 Tracker & MRV Report**  |   Angka saja    |      PDF bulanan      |      PDF + White-label       |
| **ESG Dashboard Multi-site**  |       ❌        |          ❌           |              ✅              |
| **ESG Report PDF**            |       ❌        |          ❌           |       ✅ + White-label       |
| Fleet Dashboard (Map)         |       ❌        |          ❌           |              ✅              |
| Soiling Detection             |       ❌        |          ❌           |              ✅              |
| API Access Eksternal          |       ❌        |     Rate-limited      |         Full access          |
| Priority Support              |    Community    |     Email 48 jam      |      Dedicated WA 4 jam      |

**Pricing psychology:**

- Free → Rp 0 — akuisisi user, flywheel data aktual
- Pro → Rp 99.000/bln atau Rp 899.000/thn (hemat 25%) — di bawah barrier psikologis Rp 100rb
- Enterprise → Rp 499.000/bln atau custom quote untuk fleet besar

---

## 4. Regulatory Landscape (Konteks Indonesia 2026)

### 4a. Permen ESDM No. 2 Tahun 2024

Kapasitas PLTS Atap diizinkan hingga 100% daya terpasang PLN, dengan skema net metering.
**Data yang dibutuhkan:** kWh produksi harian, kapasitas terpasang, periode.
**Tier yang mendukung:** Free (data), Pro (laporan resmi).

### 4b. Insentif PBB dari Pemda

Beberapa pemda memberikan diskon PBB bagi pemilik PLTS.
**Data yang dibutuhkan:** Surat keterangan produksi energi tahunan resmi.
**Tier yang mendukung:** Pro (Surat Keterangan PDF).

### 4c. REC — Renewable Energy Certificate

1 REC = 1 MWh dari sumber EBT. Diperdagangkan di ICDX dan via PLN GEAS.
**Data yang dibutuhkan:** Akumulasi MWh per site, format terverifikasi.
**Tier yang mendukung:** Free (angka), Pro (notifikasi + laporan PDF).

### 4d. Carbon Credit — IDX Carbon / SRN-PPI

Diperdagangkan ~Rp 96.000–144.000/ton CO2. Butuh proses MRV.
**Data yang dibutuhkan:** CO2 avoided per bulan/tahun, metodologi transparan.
**Tier yang mendukung:** Free (angka di dashboard), Pro (MRV PDF bulanan), Enterprise (white-label).

### 4e. ESG Reporting & Tender Keunggulan

Perusahaan dengan bukti EBT lebih unggul di tender BUMN/pemerintah.
**Data yang dibutuhkan:** Dashboard multi-site, % energi hijau, CO2 avoided.
**Tier yang mendukung:** Enterprise (ESG Dashboard + ESG Report PDF).

---

## 5. Goals & Success Metrics

| Goal                                   | Metrik                                       |
| -------------------------------------- | -------------------------------------------- |
| Free tier cukup berguna untuk akuisisi | MAU Free tier, data actual yang masuk        |
| Konversi Free → Pro                    | Konversi rate %, waktu median hingga upgrade |
| User Pro aktif pakai laporan           | % user Pro yang download laporan per bulan   |
| Enterprise terbukti untuk tender/ESG   | Jumlah ESG Report PDF yang digenerate        |
| Platform sustainable secara bisnis     | MRR (Monthly Recurring Revenue)              |

---

## 6. Trigger Konversi (Upgrade Nudge)

| Momen                           | Pesan                                                          |
| ------------------------------- | -------------------------------------------------------------- |
| User coba tambah profile ke-2   | "Upgrade ke Pro untuk kelola hingga 5 site."                   |
| User buka history > 7 hari      | "Lihat data historis 90 hari dengan Pro."                      |
| User aktifkan Telegram/WA notif | "Notifikasi multi-channel tersedia di Pro."                    |
| User klik tombol Export/PDF     | "Laporan PDF tersedia di paket Pro."                           |
| 30 hari setelah registrasi      | Email: "Anda sudah hemat Rp X. Tingkatkan monitoring Anda."    |
| REC pertama tercapai (1 MWh)    | "Site Anda siap klaim REC! Upgrade Pro untuk laporan PDF-nya." |

---

## 7. Features Overview (per Epic)

### Epic 0 — Tier System & Billing (Fondasi)

Middleware enforcement tier, billing integration, upgrade/downgrade flow.
**Harus selesai sebelum Epic 1–5 dirilis ke publik.**

### Epic 1 — Monthly Energy Report PDF (Pro)

Laporan bulanan produksi energi per site, siap cetak dan lampirkan ke instansi.

### Epic 2 — Annual Summary & PBB Letter (Pro)

Rekapitulasi tahunan + surat keterangan produksi energi format resmi.

### Epic 3 — MWh Accumulator & REC Readiness (Free angka / Pro laporan)

Tracking akumulasi MWh kumulatif, notifikasi REC, laporan PDF REC-ready.

### Epic 4 — CO2 Avoided Tracker & MRV Report (Free angka / Pro PDF)

Laporan CO2 avoided dengan metodologi MRV, estimasi nilai carbon credit.

### Epic 5 — ESG Dashboard & Report (Enterprise)

Dashboard multi-site + ESG Report PDF white-label untuk tender dan CSR.

---

## 8. Technical Foundation (Existing)

### Sudah Ada di Codebase

- `plan_tier` field di `notification_preferences` (`free` / `paid`) — perlu extend ke `free|pro|enterprise`
- Multi-profile & multi-device — skema DB sudah mendukung
- Multi-channel notification — Email, Telegram, WhatsApp terintegrasi native
- CO2 calculation per region — sudah berjalan di scheduler
- Weather Risk Status — sudah konsisten di dashboard, forecast, history

### Perlu Dibangun

| Komponen                      | Estimasi | Prioritas | Epic   |
| ----------------------------- | -------- | --------- | ------ |
| Extend `plan_tier` validation | 0.5 hari | 🔴 Tinggi | Epic 0 |
| Tier Enforcement Middleware   | 2-3 hari | 🔴 Tinggi | Epic 0 |
| Profile/Device Count Limit    | 1 hari   | 🔴 Tinggi | Epic 0 |
| History Day-Limit per Tier    | 1 hari   | 🔴 Tinggi | Epic 0 |
| Billing/Payment Integration   | 5-7 hari | 🔴 Tinggi | Epic 0 |
| Upgrade/Downgrade UI          | 2-3 hari | 🔴 Tinggi | Epic 0 |
| Frontend Tier Gate + Lock UI  | 2 hari   | 🔴 Tinggi | Epic 0 |
| PDF Generation (gofpdf)       | 3-5 hari | 🟡 Sedang | Epic 1 |
| CSV Export                    | 2 hari   | 🟡 Sedang | Epic 1 |
| MWh Accumulator DB + Job      | 2 hari   | 🟡 Sedang | Epic 3 |
| ESG Dashboard Multi-site      | 5-7 hari | 🟢 Rendah | Epic 5 |
| Fleet Dashboard / Map View    | 5-7 hari | 🟢 Rendah | Epic 5 |

---

## 9. Non-Goals (Out of Scope v1)

- Integrasi langsung ke sistem PLN / ICDX / SRN-PPI
- Verifikasi oleh third-party auditor
- Marketplace jual-beli REC atau carbon credit
- Soiling detection algorithm (Enterprise fase 2)

---

## 10. Dependencies & Risks

| Risiko                                | Mitigasi                                                           |
| ------------------------------------- | ------------------------------------------------------------------ |
| Payment gateway delay                 | Manual upgrade via DB untuk early adopters sambil billing dibangun |
| Format laporan PBB berbeda tiap pemda | Template generik + field customizable                              |
| Regulasi REC berubah                  | Format laporan sebagai template yang bisa diupdate tanpa deploy    |
| Data actual tidak konsisten           | Tampilkan coverage % dan warning jika data < 80%                   |
| User free tidak convert               | A/B test pesan nudge, audit trigger points tiap bulan              |
