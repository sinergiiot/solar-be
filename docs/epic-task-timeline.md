# Epic-Task Timeline — Solar Forecast Green Compliance + Freemium
**Horizon:** April 2026 – Oktober 2026  
**Metodologi:** Iteratif, per Epic. Epic 0 adalah fondasi wajib sebelum fitur berbayar dirilis.

---

## Prinsip Prioritasi

1. **Tier enforcement dulu, baru billing** — memungkinkan manual upgrade untuk early adopters
2. **Data sudah ada** — kWh, hemat biaya, CO2 avoided sudah dihitung di platform
3. **Nilai bisnis tercepat** — laporan PDF (Pro) bisa langsung jadi alasan upgrade
4. **Fondasi dulu** — akumulator MWh harus ada sebelum REC readiness ditampilkan
5. **Kompleksitas naik bertahap** — ESG Dashboard (Enterprise) paling terakhir

---

## Timeline Overview

```
Apr 2026      Mei 2026      Jun 2026      Jul 2026      Agt 2026      Sep 2026      Okt 2026
|-------------|-------------|-------------|-------------|-------------|-------------|
[====Epic 0 (Tier + Billing)====]
              [===Epic 1 (Monthly PDF)===]
                            [====Epic 2 (Annual + PBB)====]
                                          [======Epic 3 (MWh + REC)======]
                                                        [======Epic 4 (CO2 + MRV)======]
                                                                      [=======Epic 5 (ESG)=======]
```

---

## Epic 0 — Tier System & Billing (Fondasi)
**Target selesai:** April 2026 — Minggu 1-4  
**Nilai:** Platform bisa membedakan Free / Pro / Enterprise secara teknis dan komersial  
**Catatan:** Semua Epic berikutnya bergantung pada ini. Kerjakan paralel Sprint 1 + Sprint 2.

### Sprint 1 — Tier Enforcement (Minggu 1-2)

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E0-T1 | Extend `plan_tier` dari `free\|paid` menjadi `free\|pro\|enterprise` di DB + validation | 0.5 hari | - |
| E0-T2 | Tier Enforcement Middleware — baca `plan_tier` dari user context, gate fitur | 2 hari | E0-T1 |
| E0-T3 | Limit Solar Profile — cek jumlah profile saat `POST /solar-profiles` | 0.5 hari | E0-T2 |
| E0-T4 | Limit Device — cek jumlah device saat `POST /devices` | 0.5 hari | E0-T2 |
| E0-T5 | History Day-Limit — filter query history: 7 hari (free) / 90 hari (pro) / unlimited (enterprise) | 1 hari | E0-T2 |
| E0-T6 | Frontend: tier badge di sidebar (Free / Pro / Enterprise label) | 0.5 hari | E0-T1 |
| E0-T7 | Frontend: lock icon + upgrade CTA pada fitur yang tidak tersedia di tier aktif | 1.5 hari | E0-T6 |
| E0-T8 | Frontend: upgrade nudge messages di trigger points (profile ke-2, history > 7 hari, dll) | 1 hari | E0-T7 |

### Sprint 2 — Billing & Subscription (Minggu 3-4)

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E0-T9 | Tabel DB `subscriptions` (user_id, tier, start_date, end_date, status, payment_ref) | 1 hari | E0-T1 |
| E0-T10 | Pricing page frontend — perbandingan tier, bisa diakses tanpa login | 1 hari | - |
| E0-T11 | Payment Gateway integration (Midtrans / Xendit) — checkout flow | 3 hari | E0-T9 |
| E0-T12 | Webhook handler — terima callback payment, aktivasi tier otomatis | 2 hari | E0-T11 |
| E0-T13 | Grace period logic — downgrade ke free setelah 7 hari expired | 0.5 hari | E0-T12 |
| E0-T14 | Subscription management page — status, tanggal expired, tombol upgrade/cancel | 1.5 hari | E0-T12 |
| E0-T15 | Email notifikasi: konfirmasi upgrade, reminder 7 hari sebelum expired | 1 hari | E0-T12 |

**Quick Win (bisa dikerjakan hari ini tanpa menunggu billing):**
- E0-T1 + E0-T3 + E0-T4 → profile/device limit aktif
- E0-T6 + E0-T7 → UI tier badge + lock icon
- Manual upgrade via DB untuk early adopters selama billing belum siap

---

## Epic 1 — Monthly Energy Report PDF (Pro)
**Target selesai:** Mei 2026 (4 minggu)  
**Tier gate:** Pro & Enterprise  
**Nilai:** User Pro bisa download laporan bulanan siap kirim ke instansi — alasan utama upgrade

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E1-T1 | Setup PDF generation (gofpdf) + shared helper (internal/pdfgen/base.go) | 2 hari | - |
| E1-T2 | API `GET /reports/monthly` — agregasi kWh, hemat, CO2 per bulan | 2 hari | E0-T2 |
| E1-T3 | Tier check di endpoint: Pro/Enterprise only, return 403 + upgrade CTA jika Free | 0.5 hari | E0-T2 |
| E1-T4 | PDF generation: laporan bulanan per site (layout, tabel harian, KPI summary) | 3 hari | E1-T1, E1-T2 |
| E1-T5 | CSV Export history harian — Pro/Enterprise only | 1 hari | E0-T2 |
| E1-T6 | Frontend: halaman Reports, filter bulan, tombol Download PDF + Export CSV | 2 hari | E1-T4, E1-T5 |
| E1-T7 | Frontend: lock state untuk user Free + nudge upgrade | 0.5 hari | E1-T6 |
| E1-T8 | Testing & QA dengan data real | 1 hari | E1-T6 |

---

## Epic 2 — Annual Summary & PBB Letter (Pro)
**Target selesai:** Juni 2026 (4 minggu)  
**Tier gate:** Pro & Enterprise  
**Nilai:** User bisa klaim insentif PBB ke pemda dengan surat resmi

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E2-T1 | API `GET /reports/annual` — agregasi 12 bulan | 1 hari | Epic 1 selesai |
| E2-T2 | Template surat keterangan produksi energi (format resmi Indonesia) | 2 hari | - |
| E2-T3 | Form input data surat: nomor surat, nama pejabat, instansi | 1 hari | E2-T2 |
| E2-T4 | PDF generation: laporan tahunan + surat keterangan (2 dokumen dalam 1 PDF) | 2 hari | E2-T1, E2-T3 |
| E2-T5 | White-label logo upload — Enterprise only (custom kop surat) | 1 hari | E0-T2 |
| E2-T6 | Frontend: halaman Annual Report, preview surat, tombol download | 2 hari | E2-T4 |
| E2-T7 | Testing dengan berbagai skenario format instansi | 1 hari | E2-T6 |

---

## Epic 3 — MWh Accumulator & REC Readiness (Free angka / Pro laporan)
**Target selesai:** Agustus 2026 (6 minggu)  
**Tier gate:** Free = angka saja di dashboard | Pro = notifikasi + laporan PDF | Enterprise = semua

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E3-T1 | Tabel DB `energy_accumulator` (kumulatif MWh per site per periode) | 1 hari | - |
| E3-T2 | Background job: update akumulator harian + trigger saat actual baru masuk | 2 hari | E3-T1 |
| E3-T3 | API `GET /accumulator/rec-readiness` — MWh kumulatif + REC count | 1 hari | E3-T2 |
| E3-T4 | REC notifikasi — Pro/Enterprise: kirim alert email+telegram saat REC baru tercapai | 1 hari | E0-T2, E3-T3 |
| E3-T5 | Frontend: widget MWh progress bar di dashboard (semua tier, angka saja untuk Free) | 2 hari | E3-T3 |
| E3-T6 | Frontend: lock state widget detail untuk Free + nudge upgrade saat REC pertama tercapai | 0.5 hari | E3-T5 |
| E3-T7 | PDF laporan REC readiness — Pro/Enterprise only | 2 hari | E0-T2, E3-T3 |
| E3-T8 | Testing & validasi akurasi akumulator | 1 hari | E3-T7 |

---

## Epic 4 — CO2 Avoided Tracker & MRV Report (Free angka / Pro PDF)
**Target selesai:** September 2026 (6 minggu)  
**Tier gate:** Free = angka di dashboard | Pro = PDF bulanan | Enterprise = PDF + white-label

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E4-T1 | Finalisasi faktor emisi grid Indonesia per region | 1 hari | - |
| E4-T2 | Tabel DB `co2_records` (CO2 avoided harian, faktor emisi yang dipakai) | 1 hari | E4-T1 |
| E4-T3 | Rekalkukasi historis CO2 dengan faktor emisi final | 1 hari | E4-T2 |
| E4-T4 | API `GET /co2/summary` — CO2 per periode + metodologi | 1 hari | E4-T3 |
| E4-T5 | PDF MRV Report — Pro/Enterprise only (3 seksi: Measurement, Reporting, Verification) | 3 hari | E0-T2, E4-T4 |
| E4-T6 | White-label branding di PDF — Enterprise only | 1 hari | E0-T2, E4-T5 |
| E4-T7 | Frontend: halaman CO2 Tracker + estimasi nilai carbon credit (Rp) | 2 hari | E4-T4 |
| E4-T8 | Frontend: lock PDF download untuk Free + nudge upgrade | 0.5 hari | E4-T7 |
| E4-T9 | Testing metodologi vs standar IDX Carbon | 1 hari | E4-T7 |

---

## Epic 5 — ESG Dashboard & Report (Enterprise)
**Target selesai:** Oktober 2026 (6 minggu)  
**Tier gate:** Enterprise only  
**Nilai:** Dashboard ESG multi-site + laporan PDF white-label untuk tender dan CSR

| ID | Task | Est. | Dependency |
|----|------|------|------------|
| E5-T1 | API `GET /esg/summary` — agregasi semua site user, KPI ESG | 2 hari | Epic 3, 4 selesai |
| E5-T2 | Kalkulasi ESG: % energi dari PLTS, total MWh, total CO2, total REC | 1 hari | E5-T1 |
| E5-T3 | Frontend: ESG Dashboard — KPI cards, grafik, tabel per site, SDG badges | 3 hari | E5-T2 |
| E5-T4 | PDF ESG Report: cover, executive summary, site detail, metodologi | 3 hari | E5-T3 |
| E5-T5 | White-label: upload logo perusahaan, custom kop surat ESG | 1 hari | E5-T4 |
| E5-T6 | Public share link laporan ESG (toggle on/off) | 1.5 hari | E5-T5 |
| E5-T7 | Frontend: lock halaman ESG untuk Free/Pro + upgrade CTA ke Enterprise | 0.5 hari | E5-T3 |
| E5-T8 | Testing dengan skenario multi-site (5+ profile) | 1 hari | E5-T6 |

---

## Milestone Summary

| Milestone | Tanggal | Deliverable | Gate |
|-----------|---------|-------------|------|
| M0a | Minggu 2 Apr | Tier enforcement aktif, profile/device limit live | - |
| M0b | Minggu 4 Apr | Billing + payment gateway live, manual upgrade tetap bisa | M0a |
| M1 | Mei 2026 | Monthly Report PDF + CSV Export live | M0a |
| M2 | Jun 2026 | Annual Report + Surat PBB live | M1 |
| M3 | Agt 2026 | MWh Accumulator + REC Readiness live | M0a |
| M4 | Sep 2026 | CO2 MRV Report live | M3 |
| M5 | Okt 2026 | ESG Dashboard + ESG Report live | M4 |

---

## Epic 6 — Admin Dashboard
**Target selesai:** Paralel dengan Epic 0 Sprint 1 (April 2026) untuk fitur kritikal,
dilanjutkan bertahap hingga November 2026 untuk fitur BI.
**Tier gate:** Internal admin only — route terpisah `/admin`, JWT role `admin`

### Sprint A — Operasional Harian (Minggu 1-2 April, paralel Epic 0)

| ID | Task | Est. | Dependency | Status |
|----|------|------|------------|--------|
| E6-T1 | Admin auth — role `admin` di JWT, middleware `RequireAdmin` | 1 hari | - | Done |
| E6-T2 | API `GET /admin/users` — list semua user + tier + subscription status | 1 hari | E6-T1 | Done |
| E6-T3 | API `PATCH /admin/users/:id/tier` — manual upgrade/downgrade tier | 0.5 hari | E6-T2 | Done |
| E6-T4 | API `POST /admin/users/:id/impersonate` — generate token login as user | 1 hari | E6-T1 | Done |
| E6-T5 | API `GET /admin/subscriptions/expiring` — list expiry 7 hari ke depan | 0.5 hari | E0-T9 | Done |
| E6-T6 | API `GET /admin/scheduler/status` — status run harian per site | 1 hari | - | Done |
| E6-T7 | Frontend admin: halaman User Management (list, filter, upgrade/downgrade) | 2 hari | E6-T2, E6-T3 | Done |
| E6-T8 | Frontend admin: halaman Scheduler Monitor (status, error log) | 1 hari | E6-T6 | Done |
| E6-T9 | Frontend admin: Subscription Expiry alert panel | 0.5 hari | E6-T5 | Done |

### Sprint B — Kualitas Data (Mei–Juni 2026)

| ID | Task | Est. | Dependency | Status |
|----|------|------|------------|--------|
| E6-T10 | API `GET /admin/forecast-quality` — MAPE per site 7 hari terakhir | 2 hari | E6-T1 | Done |
| E6-T11 | API `GET /admin/cold-start-monitor` — site > 30 hari masih synthetic | 1 hari | E6-T1 | Done |
| E6-T12 | API `GET /admin/notification-log` — delivery status email/tg/wa | 1 hari | E6-T1 | Done |
| E6-T13 | API `GET /admin/data-anomalies` — actual > 1.5× predicted, coverage < 50% | 1.5 hari | E6-T1 | Done |
| E6-T14 | Frontend admin: Forecast Quality table + flag merah MAPE > 30% | 2 hari | E6-T10 | Done |
| E6-T15 | Frontend admin: Cold Start Monitor + export list untuk outreach | 1 hari | E6-T11 | Done |
| E6-T16 | Frontend admin: Notification Delivery log + filter by channel | 1.5 hari | E6-T12 | Done |
| E6-T17 | Frontend admin: Data Anomaly panel + detail per site | 1.5 hari | E6-T13 |

### Sprint C — Business Intelligence (Juli–November 2026)

| ID | Task | Est. | Dependency | Status |
|----|------|------|------------|--------|
| E6-T18 | API `GET /admin/analytics/aggregate` — System-wide prod/forecast | 2 hari | E6-T1 | Done |
| E6-T19 | API `GET /admin/analytics/ranking` — Top vs Bottom performers | 1 hari | E6-T1 | Done |
| E6-T20 | API `GET /admin/analytics/tier-distribution` — Free vs Pro vs Ent | 1 hari | E6-T1 | Done |
| E6-T21 | Frontend admin: Intelligence Home — Aggregated Charts | 2 hari | E6-T18 | Done |
| E6-T22 | Frontend admin: Performance Ranking — Leaderboard | 1 hari | E6-T19 | Done |
| E6-T23 | Frontend admin: Weather API Health + accuracy leaderboard | 2 hari | E6-T19, E6-T20 | Done |
| E6-T24 | Frontend admin: Audit Log table + export CSV | 1 hari | E6-T21 | Done |

### Milestone Admin Dashboard

| Milestone | Tanggal | Deliverable |
|-----------|---------|-------------|
| M6a | Minggu 2 Apr | User management + manual upgrade live |
| M6b | Minggu 4 Apr | Scheduler monitor + subscription expiry live |
| M6c | Jun 2026 | Forecast quality + cold start + notification log live |
| M6d | Nov 2026 | Revenue BI + leaderboard + audit log live |