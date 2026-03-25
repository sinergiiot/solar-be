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

| ID    | Task                                                                                             | Est.     | Dependency |
| ----- | ------------------------------------------------------------------------------------------------ | -------- | ---------- |
| E0-T1 | Extend `plan_tier` dari `free\|paid` menjadi `free\|pro\|enterprise` di DB + validation          | [x] |
| E0-T2 | Tier Enforcement Middleware — baca `plan_tier` dari user context, gate fitur                     | [x] |
| E0-T3 | Limit Solar Profile — cek jumlah profile saat `POST /solar-profiles`                             | [x] |
| E0-T4 | Limit Device — cek jumlah device saat `POST /devices`                                            | [x] |
| E0-T5 | History Day-Limit — filter query history: 7 hari (free) / 90 hari (pro) / unlimited (enterprise) | [x] |
| E0-T6 | Frontend: tier badge di sidebar (Free / Pro / Enterprise label)                                  | [x] |
| E0-T7 | Frontend: lock icon + upgrade CTA pada fitur yang tidak tersedia di tier aktif                   | [x] |
| E0-T8 | Frontend: upgrade nudge messages di trigger points (profile ke-2, history > 7 hari, dll)         | [x] |

### Sprint 2 — Billing & Subscription (Minggu 3-4)

| ID     | Task                                                                                | Est.     | Dependency |
| ------ | ----------------------------------------------------------------------------------- | -------- | ---------- |
| E0-T9  | Tabel DB `subscriptions` (user_id, tier, start_date, end_date, status, payment_ref) | [x] | E0-T1      |
| E0-T10 | Pricing page frontend — perbandingan tier, bisa diakses tanpa login                 | [x] |
| E0-T11 | Payment Gateway integration (Midtrans / Xendit) — checkout flow                     | [x] | E0-T9      |
| E0-T12 | Webhook handler — terima callback payment, aktivasi tier otomatis                   | [x] | E0-T11     |
| E0-T13 | Grace period logic — downgrade ke free setelah 7 hari expired                       | [x] | E0-T12     |
| E0-T14 | Subscription management page — status, tanggal expired, tombol upgrade/cancel       | [x] | E0-T12     |
| E0-T15 | Email notifikasi: konfirmasi upgrade, reminder 7 hari sebelum expired               | [x] | E0-T12     |

**Quick Win (bisa dikerjakan hari ini tanpa menunggu billing):**

- E0-T1 + E0-T3 + E0-T4 → profile/device limit aktif
- E0-T6 + E0-T7 → UI tier badge + lock icon
- Manual upgrade via DB untuk early adopters selama billing belum siap

---

## Epic 1 — Monthly Energy Report PDF (Pro)

**Target selesai:** Mei 2026 (4 minggu)  
**Tier gate:** Pro & Enterprise  
**Nilai:** User Pro bisa download laporan bulanan siap kirim ke instansi — alasan utama upgrade

| ID    | Task                                                                            | Est.     | Dependency   |
| ----- | ------------------------------------------------------------------------------- | -------- | ------------ |
| E1-T1 | Setup PDF generation (gofpdf) + shared helper (internal/pdfgen/base.go)         | [x] |
| E1-T2 | API `GET /reports/monthly` — agregasi kWh, hemat, CO2 per bulan                 | [x] |
| E1-T3 | Tier check di endpoint: Pro/Enterprise only, return 403 + upgrade CTA jika Free | [x] |
| E1-T4 | PDF generation: laporan bulanan per site (layout, tabel harian, KPI summary)    | [x] |
| E1-T5 | CSV Export history harian — Pro/Enterprise only                                 | [x] |
| E1-T6 | Frontend: halaman Reports, filter bulan, tombol Download PDF + Export CSV       | [x] |
| E1-T7 | Frontend: lock state untuk user Free + nudge upgrade                            | [x] |
| E1-T8 | Testing & QA dengan data real                                                   | [x] |

---

## Epic 2 — Annual Summary & PBB Letter (Pro)

**Target selesai:** Juni 2026 (4 minggu)  
**Tier gate:** Pro & Enterprise  
**Nilai:** User bisa klaim insentif PBB ke pemda dengan surat resmi

| ID    | Task                                                                       | Est.   | Dependency     |
| ----- | -------------------------------------------------------------------------- | ------ | -------------- |
| E2-T1 | API `GET /reports/annual` — agregasi 12 bulan                              | [x] |
| E2-T2 | Template surat keterangan produksi energi (format resmi Indonesia)         | [x] |
| E2-T3 | Form input data surat: nomor surat, nama pejabat, instansi                 | [x] |
| E2-T4 | PDF generation: laporan tahunan + surat keterangan (2 dokumen dalam 1 PDF) | [x] |
| E2-T5 | White-label logo upload — Enterprise only (custom kop surat)               | [x] |
| E2-T6 | Frontend: halaman Annual Report, preview surat, tombol download            | [x] |
| E2-T7 | Testing dengan berbagai skenario format instansi                           | [x] |

---

## Epic 3 — MWh Accumulator & REC Readiness (Free angka / Pro laporan)

**Target selesai:** Agustus 2026 (6 minggu)  
**Tier gate:** Free = angka saja di dashboard | Pro = notifikasi + laporan PDF | Enterprise = semua

| ID    | Task                                                                                    | Est.     | Dependency   |
| ----- | --------------------------------------------------------------------------------------- | -------- | ------------ |
| E3-T1 | Table `mwh_accumulators` (Link: user_id, profile_id)                    | [x] |
| E3-T2 | Background job: update cumulative kWh tiap kali data actual masuk          | [x] |
| E3-T3 | API `GET /accumulator/rec-readiness` — MWh kumulatif + REC count            | [x] |
| E3-T4 | REC notifikasi — Pro/Enterprise: kirim alert email+telegram saat REC baru tercapai      | [x] |
| E3-T5 | Frontend: widget MWh progress bar di dashboard (semua tier, angka saja untuk Free)      | [x] |
| E3-T6 | Frontend: lock state widget detail untuk Free + nudge upgrade saat REC pertama tercapai | [x] |
| E3-T7 | PDF laporan REC readiness — Pro/Enterprise only                                         | [x] |

---

## Epic 4 — CO2 Avoided Tracker & MRV Report (Free angka / Pro PDF)

**Target selesai:** September 2026 (6 minggu)  
**Tier gate:** Free = angka di dashboard | Pro = PDF bulanan | Enterprise = PDF + white-label

| ID    | Task                                                                                 | Est.     | Dependency   |
| ----- | ------------------------------------------------------------------------------------ | -------- | ------------ |
| E4-T1 | Finalisasi faktor emisi grid Indonesia per region                                    | [x] |
| E4-T2 | Tabel DB `co2_records` (Calculated on-the-fly via getEmissionFactor)                | [x] |
| E4-T3 | Rekalkukasi historis CO2 dengan faktor emisi final                                   | [x] |
| E4-T4 | API `GET /report/co2` — CO2 per periode + metodologi                                 | [x] |
| E4-T5 | PDF MRV Report — Pro/Enterprise only (3 seksi: Measurement, Reporting, Verification) | [x] |
| E4-T6 | White-label branding di PDF — Enterprise only                                        | [x] |
| E4-T7 | Frontend: halaman CO2 Tracker + estimasi nilai carbon credit (Rp)                    | [x] |
| E4-T8 | Frontend: lock PDF download untuk Free + nudge upgrade                               | [x] |
| E4-T9 | Testing metodologi vs standar IDX Carbon                                             | [x] |

---

## Epic 5 — ESG Dashboard & Report (Enterprise)

**Target selesai:** Oktober 2026 (6 minggu)  
**Tier gate:** Enterprise only  
**Nilai:** Dashboard ESG multi-site + laporan PDF white-label untuk tender dan CSR

| ID    | Task                                                                    | Est.     | Dependency        |
| ----- | ----------------------------------------------------------------------- | -------- | ----------------- |
| E5-T1 | API `GET /report/esg` — agregasi semua site user, KPI ESG               | [x] |
| E5-T2 | Kalkulasi ESG: % energi dari PLTS, total MWh, total CO2, total REC      | [x] |
| E5-T3 | Frontend: ESG Dashboard — KPI cards, grafik, tabel per site, SDG badges | [x] |
| E5-T4 | PDF ESG Report: cover, executive summary, site detail, metodologi       | [x] |
| E5-T5 | White-label: upload logo perusahaan, custom kop surat ESG               | [x] |
| E5-T6 | Public share link laporan ESG (toggle on/off)                           | [x] |
| E5-T7 | Frontend: lock halaman ESG untuk Free/Pro + upgrade CTA ke Enterprise   | [x] |
| E5-T8 | Testing dengan skenario multi-site (5+ profile)                         | [x] |

---

## Milestone Summary

| Milestone | Tanggal      | Deliverable                                               | Gate |
| --------- | ------------ | --------------------------------------------------------- | ---- |
| M0a       | Minggu 2 Apr | Tier enforcement aktif, profile/device limit live         | -    |
| M0b       | Minggu 4 Apr | Billing + payment gateway live, manual upgrade tetap bisa | M0a  |
| M1        | Mei 2026     | Monthly Report PDF + CSV Export live                      | M0a  |
| M2        | Jun 2026     | Annual Report + Surat PBB live                            | M1   |
| M3        | Agt 2026     | MWh Accumulator + REC Readiness live                      | M0a  |
| M4        | Sep 2026     | CO2 MRV Report live                                       | M3   |
| M5        | Okt 2026     | ESG Dashboard + ESG Report live                           | M4   |
