# Perencanaan Fitur Freemium — Solar Forecast Platform

## Ringkasan Strategi

Prinsip utama: **Free tier harus cukup berguna** agar user mau mendaftar dan aktif (flywheel data), namun **cukup terbatas** di fitur operasional/enterprise sehingga user komersial terdorong upgrade.

---

## Matriks Fitur per Tier

| Fitur                                |       🆓 Free        | ⭐ Pro (Rp 99rb/bln)  | 🏢 Enterprise (Rp 499rb/bln) |
| ------------------------------------ | :------------------: | :-------------------: | :--------------------------: |
| **Solar Profile**                    |        1 site        |        5 site         |          Unlimited           |
| **Device IoT**                       |       1 device       |       10 device       |          Unlimited           |
| **Forecast Harian**                  |          ✅          |          ✅           |              ✅              |
| **History Forecast**                 |   7 hari terakhir    |        90 hari        |    Unlimited + Export CSV    |
| **History Actual**                   |   7 hari terakhir    |        90 hari        |    Unlimited + Export CSV    |
| **Notifikasi Email**                 |     ✅ (1x/hari)     |     ✅ (1x/hari)      |         ✅ (1x/hari)         |
| **Notifikasi Telegram**              |          ❌          |          ✅           |              ✅              |
| **Notifikasi WhatsApp**              |          ❌          |          ✅           |              ✅              |
| **Weather Risk Status**              |   ✅ (label saja)    | ✅ + Rekomendasi aksi |    ✅ + Rekomendasi aksi     |
| **Estimasi Produksi per Periode**    | ✅ (Pagi/Siang/Sore) |  ✅ + Trend mingguan  |      ✅ + Trend bulanan      |
| **Laporan ESG / CO₂**                |  Angka di dashboard  | PDF bulanan otomatis  |    PDF + White-label logo    |
| **Soiling Detection**                |          ❌          |          ❌           |     ✅ (alert otomatis)      |
| **Fleet Dashboard (Multi-site Map)** |          ❌          |          ❌           |              ✅              |
| **API Access (Eksternal)**           |          ❌          |     Rate-limited      |         Full access          |
| **Priority Support**                 |   Community/GitHub   |     Email 48 jam      |      Dedicated WA 4 jam      |

---

## Arsitektur Teknis yang Sudah Siap

Kabar baiknya, **fondasi kode sudah mendukung** skema tier ini:

### ✅ Sudah Ada di Codebase

1. **`plan_tier` field** — Tersimpan di tabel `notification_preferences` (`free` / `paid`). Validasi sudah ada di [notification/service.go](file:///Users/akbarsenawijaya/Akbar-WIT/forcast-solar-panel/internal/notification/service.go).
2. **Multi-profile & Multi-device** — Skema DB sudah mendukung banyak `solar_profiles` dan `devices` per user.
3. **Multi-channel notification** — Email, Telegram, WhatsApp sudah terintegrasi native.
4. **Emission factor dinamis** — CO₂ calculation per region sudah berjalan di scheduler.
5. **Weather Risk Status** — Label risiko cuaca sudah konsisten di dashboard, forecast, dan history.

### 🔧 Perlu Dibangun (Effort Estimation)

| Komponen                        | Estimasi | Prioritas           |
| ------------------------------- | -------- | ------------------- |
| **Tier Enforcement Middleware** | 2-3 hari | 🔴 Tinggi           |
| **Billing/Payment Integration** | 5-7 hari | 🔴 Tinggi           |
| **History Day-Limit per Tier**  | 1 hari   | 🟡 Sedang           |
| **Profile/Device Count Limit**  | 1 hari   | 🟡 Sedang           |
| **CSV Export**                  | 2 hari   | 🟡 Sedang           |
| **PDF ESG Report Generator**    | 3-5 hari | 🟢 Rendah (Phase 2) |
| **Fleet Dashboard / Map View**  | 5-7 hari | 🟢 Rendah (Phase 2) |
| **Upgrade/Downgrade UI**        | 2-3 hari | 🔴 Tinggi           |

---

## Rencana Implementasi (Sprint Plan)

### Sprint 1: Fondasi Tier System (Minggu 1-2)

**Goal:** Memisahkan `free` vs [pro](file:///Users/akbarsenawijaya/Akbar-WIT/forcast-solar-panel/internal/scheduler/scheduler.go#164-232) vs `enterprise` secara teknis.

1. **Extend `plan_tier`** — Ubah validasi dari hanya `free|paid` menjadi `free|pro|enterprise`.
2. **Tier Enforcement Middleware** — Buat Go middleware yang membaca `plan_tier` dari user context dan men-_gate_ akses ke fitur tertentu.
3. **Limit Solar Profile** — Tambahkan pengecekan jumlah profile saat `POST /solar-profiles`.
4. **Limit Device** — Tambahkan pengecekan jumlah device saat `POST /devices`.
5. **Limit History Days** — Filter query history berdasarkan tier (7/90/unlimited hari).
6. **Frontend Gate** — Tampilkan badge tier di sidebar + lock icon pada fitur yang tidak tersedia.

### Sprint 2: Billing & Upgrade Flow (Minggu 3-4)

**Goal:** User bisa upgrade dan bayar.

1. **Pricing Page** — Halaman perbandingan tier (bisa diakses tanpa login).
2. **Payment Gateway** — Integrasi Midtrans/Xendit untuk pembayaran.
3. **Subscription Management** — Tabel `subscriptions` (user_id, tier, start_date, end_date, status).
4. **Webhook Handler** — Terima callback dari payment gateway untuk aktivasi/deaktivasi tier.
5. **Grace Period** — Jika subscription expired, downgrade ke free setelah 7 hari grace.

### Sprint 3: Fitur Pro Differentiator (Minggu 5-6)

**Goal:** Memberikan nilai tambah signifikan bagi user Pro.

1. **CSV Export** — Tombol download di halaman History.
2. **Telegram & WhatsApp Gate** — Pastikan channel ini hanya aktif untuk tier ≥ Pro.
3. **Trend Analysis** — Grafik perbandingan mingguan forecast vs actual.
4. **Rekomendasi Aksi** — Tambahkan saran operasional berdasarkan weather risk status.

---

## Strategi Konversi Free → Pro

### Trigger Points (Kapan user terdorong upgrade)

| Momen                                | Nudge                                                                                        |
| ------------------------------------ | -------------------------------------------------------------------------------------------- |
| User mencoba tambah profile ke-2     | _"Upgrade ke Pro untuk mengelola hingga 5 site."_                                            |
| User buka history > 7 hari           | _"Lihat data historis hingga 90 hari dengan Pro."_                                           |
| User enable Telegram/WA notification | _"Notifikasi multi-channel tersedia di paket Pro."_                                          |
| 30 hari setelah registrasi           | Email: _"Anda sudah menghemat Rp X dari forecast kami. Tingkatkan akurasi monitoring Anda."_ |
| User klik tombol Export              | _"Export CSV tersedia di paket Pro."_                                                        |

### Pricing Psychology

- **Free** → Rp 0 → Akuisisi user, data crowdsource
- **Pro** → Rp 99.000/bulan (atau Rp 899.000/tahun, hemat 25%) → Target: Pemilik PLTS rumahan/UKM
- **Enterprise** → Rp 499.000/bulan (atau custom quote) → Target: EPC, Pabrik, Korporat

> [!TIP]
> Pricing Rp 99rb/bulan dipilih karena berada di bawah _psychological barrier_ Rp 100rb, yang umum menjadi batas psikologis konsumen Indonesia untuk subscription digital.

---

## Perbandingan dengan Kompetitor

| Platform             | Model           | Harga              | Keunggulan Kita                         |
| -------------------- | --------------- | ------------------ | --------------------------------------- |
| SolarEdge Monitoring | Hardware-locked | $$$$ (bundled)     | Kita hardware-agnostic, ESP32 cukup     |
| Huawei FusionSolar   | Proprietary     | Gratis tapi locked | Kita open API, multi-brand              |
| PVOutput.org         | Community       | Gratis             | Kita punya AI forecast + notification   |
| Enphase Enlighten    | SaaS            | $$/bulan           | Kita fokus pasar Indonesia, harga lokal |

---

## Quick Wins (Bisa Dikerjakan Minggu Ini)

1. **Extend `plan_tier` validation** — Tambahkan [pro](file:///Users/akbarsenawijaya/Akbar-WIT/forcast-solar-panel/internal/scheduler/scheduler.go#164-232) dan `enterprise` sebagai opsi valid
2. **Profile count check** — Tolak pembuatan profile jika melebihi kuota tier
3. **Tier badge di frontend** — Tampilkan label "Free" / "Pro" di sidebar user
4. **Upgrade CTA button** — Tombol "Upgrade ke Pro" di sidebar untuk user Free

> [!IMPORTANT]
> Urutan prioritas: **Tier enforcement dulu, baru billing.** Ini memungkinkan Anda melakukan "manual upgrade" untuk early adopters (via database) sambil payment gateway sedang dibangun. Strategi ini umum dipakai startup tahap awal.
