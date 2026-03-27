# TASK_QUEUE.md — Solar Forecast Agent Task Queue
# Dikelola bersama antara developer dan Claude Code agent.
# Agent membaca file ini setiap session untuk menentukan task berikutnya.
#
# STATUS:
#   [ ] = Pending (belum dikerjakan)
#   [~] = In Progress (agent sedang mengerjakan)
#   [x] = Done (selesai, test pass, sudah commit)
#   [!] = Failed (gagal setelah 3 iterasi, butuh review manusia)
#   [-] = Blocked (menunggu dependency atau jawaban)
#   [s] = Skipped (diputuskan tidak dikerjakan)
#
# ATURAN AGENT:
#   - Ambil task [ ] paling atas yang dependencynya sudah [x]
#   - Kerjakan satu task per session
#   - Update status di file ini sebelum mulai dan setelah selesai
#   - Prompt detail ada di docs/prompts/[TASK-ID].md

---

## 🔴 SPRINT 1 — Tier System Enforcement
### Target: Minggu 1-2 April 2026
### Goal: Platform bisa membedakan Free / Pro / Enterprise secara teknis

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [x] | E0-T1 | Extend `plan_tier` dari `free\|paid` ke `free\|pro\|enterprise` + constants | 0.5 hari | - | docs/prompts/E0-T1.md |
| [x] | E0-T2 | Tier Enforcement Middleware (TierMiddleware + RequireTier) | 2 hari | E0-T1 | docs/prompts/E0-T2.md |
| [x] | E0-T3 | Limit Solar Profile saat POST /solar-profiles | 0.5 hari | E0-T2 | docs/prompts/E0-T3T4.md |
| [x] | E0-T4 | Limit Device saat POST /devices | 0.5 hari | E0-T2 | docs/prompts/E0-T3T4.md |
| [x] | E0-T5 | History Day-Limit per tier di query history | 1 hari | E0-T2 | docs/prompts/E0-T5.md |
| [ ] | E0-T6 | Frontend: tier badge di sidebar (Free/Pro/Enterprise) | 0.5 hari | E0-T1 | docs/prompts/E0-T6T7.md |
| [ ] | E0-T7 | Frontend: lock icon + upgrade CTA pada fitur terkunci | 1.5 hari | E0-T6 | docs/prompts/E0-T6T7.md |
| [ ] | E0-T8 | Frontend: upgrade nudge messages di trigger points | 1 hari | E0-T7 | docs/prompts/E0-T8.md |

## Migrations (M)

| Status | ID | Migration | Est. | Dependency |
|--------|----|-----------|------|------------|
| [x] | E0-M1 | Migration `paid` to `pro` data in DB | 0.2 hari | - |

---

## 🔴 SPRINT 2 — Billing & Subscription
### Target: Minggu 3-4 April 2026
### Goal: User bisa upgrade dan bayar. Manual upgrade via DB tetap bisa selama ini berlangsung.

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [ ] | E0-T9 | Tabel DB `subscriptions` + migration | 1 hari | E0-T1 | docs/prompts/E0-T9.md |
| [ ] | E0-T10 | Pricing page frontend (perbandingan tier, tanpa login) | 1 hari | E0-T1 | docs/prompts/E0-T10.md |
| [ ] | E0-T11 | Payment Gateway Midtrans — checkout flow | 3 hari | E0-T9 | docs/prompts/E0-T11.md |
| [ ] | E0-T12 | Webhook handler payment callback + aktivasi tier | 2 hari | E0-T11 | docs/prompts/E0-T12.md |
| [ ] | E0-T13 | Grace period logic — downgrade ke free setelah 7 hari expired | 0.5 hari | E0-T12 | docs/prompts/E0-T13.md |
| [ ] | E0-T14 | Subscription management page frontend | 1.5 hari | E0-T12 | docs/prompts/E0-T14.md |
| [ ] | E0-T15 | Email notifikasi: konfirmasi upgrade + reminder expired | 1 hari | E0-T12 | docs/prompts/E0-T15.md |

---

## 🟡 EPIC 1 — Monthly Energy Report PDF (Pro)
### Target: Mei 2026
### Gate: Epic 0 Sprint 1 selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [-] | E1-T1 | Setup gofpdf + shared helpers (internal/pdfgen/base.go) | 2 hari | E0-T2 | docs/prompts/E1-T1.md |
| [-] | E1-T2 | API GET /reports/monthly — agregasi kWh, hemat, CO2 + tier gate | 2 hari | E1-T1 | docs/prompts/E1-T2T3.md |
| [-] | E1-T3 | Tier check endpoint: Pro/Enterprise only, 403 + upgrade CTA jika Free | 0.5 hari | E1-T2 | docs/prompts/E1-T2T3.md |
| [-] | E1-T4 | PDF generation laporan bulanan (layout, tabel, KPI summary) | 3 hari | E1-T1 | docs/prompts/E1-T4.md |
| [-] | E1-T5 | CSV Export history harian — Pro/Enterprise only | 1 hari | E0-T2 | docs/prompts/E1-T5.md |
| [-] | E1-T6 | Frontend halaman Reports — filter, download PDF, export CSV | 2 hari | E1-T4 | docs/prompts/E1-T6.md |
| [-] | E1-T7 | Frontend lock state untuk Free + upgrade nudge | 0.5 hari | E1-T6 | docs/prompts/E1-T7.md |
| [-] | E1-T8 | Testing & QA dengan data real | 1 hari | E1-T6 | - |

---

## 🟡 EPIC 2 — Annual Summary & PBB Letter (Pro)
### Target: Juni 2026
### Gate: Epic 1 selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [-] | E2-T1 | API GET /reports/annual — agregasi 12 bulan | 1 hari | E1-T2 | docs/prompts/E2-T1.md |
| [-] | E2-T2 | Template surat keterangan produksi energi (format resmi) | 2 hari | - | docs/prompts/E2-T2.md |
| [-] | E2-T3 | Form input data surat: nomor surat, pejabat, instansi | 1 hari | E2-T2 | docs/prompts/E2-T3.md |
| [-] | E2-T4 | PDF generation: laporan tahunan + surat keterangan | 2 hari | E2-T1 | docs/prompts/E2-T4T5.md |
| [-] | E2-T5 | White-label logo upload — Enterprise only | 1 hari | E0-T2 | docs/prompts/E2-T4T5.md |
| [-] | E2-T6 | Frontend halaman Annual Report + preview surat | 2 hari | E2-T4 | docs/prompts/E2-T6.md |
| [-] | E2-T7 | Testing dengan berbagai skenario format instansi | 1 hari | E2-T6 | - |

---

## 🟢 EPIC 3 — MWh Accumulator & REC Readiness
### Target: Agustus 2026
### Gate: Epic 0 Sprint 1 selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [-] | E3-T1 | Tabel DB `energy_accumulator` + migration | 1 hari | E0-T1 | docs/prompts/E3-T1T2.md |
| [-] | E3-T2 | Background job: update akumulator harian | 2 hari | E3-T1 | docs/prompts/E3-T1T2.md |
| [-] | E3-T3 | API GET /accumulator/rec-readiness | 1 hari | E3-T2 | docs/prompts/E3-T3T4.md |
| [-] | E3-T4 | REC notifikasi — Pro/Enterprise only | 1 hari | E3-T3 | docs/prompts/E3-T3T4.md |
| [-] | E3-T5 | Frontend widget MWh progress bar di dashboard | 2 hari | E3-T3 | docs/prompts/E3-T5T6.md |
| [-] | E3-T6 | Frontend lock detail untuk Free + nudge saat REC pertama | 0.5 hari | E3-T5 | docs/prompts/E3-T5T6.md |
| [-] | E3-T7 | PDF laporan REC readiness — Pro/Enterprise only | 2 hari | E3-T3 | docs/prompts/E3-T7.md |
| [-] | E3-T8 | Testing & validasi akurasi akumulator | 1 hari | E3-T7 | - |

---

## 🟢 EPIC 4 — CO2 Avoided Tracker & MRV Report
### Target: September 2026
### Gate: Epic 3 selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [-] | E4-T1 | Finalisasi faktor emisi grid Indonesia per region | 1 hari | - | docs/prompts/E4-T1.md |
| [-] | E4-T2 | Tabel DB `co2_records` + migration | 1 hari | E4-T1 | docs/prompts/E4-T2T3.md |
| [-] | E4-T3 | Rekalkukasi historis CO2 dengan faktor emisi final | 1 hari | E4-T2 | docs/prompts/E4-T2T3.md |
| [-] | E4-T4 | API GET /co2/summary — CO2 per periode + metodologi | 1 hari | E4-T3 | docs/prompts/E4-T4.md |
| [-] | E4-T5 | PDF MRV Report — Pro/Enterprise only | 3 hari | E4-T4 | docs/prompts/E4-T5T6.md |
| [-] | E4-T6 | White-label branding di PDF — Enterprise only | 1 hari | E4-T5 | docs/prompts/E4-T5T6.md |
| [-] | E4-T7 | Frontend halaman CO2 Tracker + estimasi carbon credit | 2 hari | E4-T4 | docs/prompts/E4-T7T8.md |
| [-] | E4-T8 | Frontend lock PDF download untuk Free + nudge | 0.5 hari | E4-T7 | docs/prompts/E4-T7T8.md |
| [-] | E4-T9 | Testing metodologi vs standar IDX Carbon | 1 hari | E4-T7 | - |

---

## 🟢 EPIC 5 — ESG Dashboard & Report (Enterprise)
### Target: Oktober 2026
### Gate: Epic 3 + Epic 4 selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [-] | E5-T1 | API GET /esg/summary — agregasi semua site user | 2 hari | E3-T3, E4-T4 | docs/prompts/E5-T1T2.md |
| [-] | E5-T2 | Kalkulasi ESG: % energi hijau, total MWh, CO2, REC | 1 hari | E5-T1 | docs/prompts/E5-T1T2.md |
| [-] | E5-T3 | Frontend ESG Dashboard — KPI, grafik, tabel per site | 3 hari | E5-T2 | docs/prompts/E5-T3.md |
| [-] | E5-T4 | PDF ESG Report: cover, executive summary, metodologi | 3 hari | E5-T3 | docs/prompts/E5-T4T5.md |
| [-] | E5-T5 | White-label: upload logo + custom kop surat ESG | 1 hari | E5-T4 | docs/prompts/E5-T4T5.md |
| [-] | E5-T6 | Public share link laporan ESG (toggle on/off) | 1.5 hari | E5-T5 | docs/prompts/E5-T6.md |
| [-] | E5-T7 | Frontend lock halaman ESG untuk Free/Pro + Enterprise CTA | 0.5 hari | E5-T3 | docs/prompts/E5-T7.md |
| [-] | E5-T8 | Testing multi-site (5+ profile) | 1 hari | E5-T6 | - |

---

## 📋 Log Aktivitas Agent

<!-- Agent mengisi bagian ini setiap selesai task -->

| Tanggal | Task ID | Status | Durasi | Catatan |
|---------|---------|--------|--------|---------|
| - | - | - | - | Belum ada aktivitas |

---

## ❓ QUESTIONS.md Pending

<!-- Pertanyaan dari agent yang belum dijawab -->
<!-- Format: [TASK-ID] [tanggal] pertanyaan -->

Belum ada pertanyaan.

---

## 🔴 EPIC 6A — Admin Dashboard: Operasional Harian
### Target: Paralel dengan Epic 0 Sprint 1 (Minggu 1-2 April 2026)
### Goal: Admin bisa manage user dan monitor scheduler tanpa masuk ke DB

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [x] | E6-T1 | Admin auth — role `is_admin`, middleware `RequireAdmin`, route `/admin` | 1 hari | E0-T1 | docs/prompts/E6-T1.md |
| [x] | E6-T2 | API GET /admin/users — list + filter + subscription status | 1 hari | E6-T1 | docs/prompts/E6-T2T3.md |
| [x] | E6-T3 | API PATCH /admin/users/:id/tier — manual upgrade/downgrade | 0.5 hari | E6-T2 | docs/prompts/E6-T2T3.md |
| [x] | E6-T4 | API POST /admin/users/:id/impersonate — login as user | 1 hari | E6-T1 | docs/prompts/E6-T4.md |
| [x] | E6-T5 | API GET /admin/subscriptions/expiring — expiry 7 hari ke depan | 0.5 hari | E0-T9 | docs/prompts/E6-T5.md |
| [x] | E6-T6 | Scheduler logging (scheduler_runs table) + API GET /admin/scheduler/status | 1 hari | - | docs/prompts/E6-T6.md |
| [x] | E6-T7 | Frontend: halaman User Management | 2 hari | E6-T2, E6-T3 | docs/prompts/E6-T7.md |
| [x] | E6-T8 | Frontend: halaman Scheduler Monitor — status, error log, history | 1 hari | E6-T6 | docs/prompts/E6-T8.md |
| [x] | E6-T9 | Frontend: Subscription Expiry alert panel | 0.5 hari | E6-T5 | docs/prompts/E6-T9.md |

---

## 🟡 EPIC 6B — Admin Dashboard: Kualitas Data
### Target: Mei–Juni 2026
### Gate: Epic 6A selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [x] | E6-T10 | API GET /admin/forecast-quality — MAPE per site + flag | 2 hari | E6-T1 | docs/prompts/E6-T10T11.md |
| [x] | E6-T11 | API GET /admin/cold-start-monitor — site > 30 hari masih synthetic | 1 hari | E6-T1 | docs/prompts/E6-T10T11.md |
| [x] | E6-T12 | API GET /admin/notification-log — delivery status per channel | 1 hari | E6-T1 | docs/prompts/E6-T12.md |
| [x] | E6-T13 | API GET /admin/data-anomalies — actual > 1.5× predicted, zero streak, coverage rendah | 1.5 hari | E6-T1 | docs/prompts/E6-T13.md |
| [x] | E6-T14 | Frontend: Forecast Quality table + Cold Start Monitor | 2 hari | E6-T10, E6-T11 | docs/prompts/E6-T14.md |
| [x] | E6-T15 | Frontend: Notification Log + filter by channel | 1.5 hari | E6-T12 | docs/prompts/E6-T15.md |
| [x] | E6-T16 | Frontend: Data Anomaly panel + detail per site | 1.5 hari | E6-T13 | docs/prompts/E6-T16.md |

---

## 🟢 EPIC 6C — Admin Dashboard: Business Intelligence
### Target: Juli–November 2026
### Gate: Epic 0 Sprint 2 (billing) selesai

| Status | ID | Task | Est. | Dependency | Prompt |
|--------|----|------|------|------------|--------|
| [x] | E6-T18 | API GET /admin/revenue — MRR, ARR, konversi, churn | 2 hari | E0-T12 | docs/prompts/E6-T18.md |
| [x] | E6-T19 | API GET /admin/weather-api-health — response time, cache hit, fallback | 1 hari | - | docs/prompts/E6-T19.md |
| [x] | E6-T20 | API GET /admin/accuracy-leaderboard — ranking MAPE terbaik | 1 hari | E6-T10 | docs/prompts/E6-T20.md |
| [x] | E6-T21 | API GET /admin/audit-log — semua aksi admin + export CSV | 1 hari | E6-T1 | docs/prompts/E6-T21.md |
| [x] | E6-T22 | Frontend: Revenue Dashboard — MRR chart, konversi funnel, churn | 3 hari | E6-T18 | docs/prompts/E6-T22.md |
| [x] | E6-T23 | Frontend: Weather API Health + accuracy leaderboard | 2 hari | E6-T19, E6-T20 | docs/prompts/E6-T23.md |
| [x] | E6-T24 | Frontend: Audit Log table + export CSV | 1 hari | E6-T21 | docs/prompts/E6-T24.md |