# AI Agent Prompt Catalog v2 — Solar Forecast Green Compliance + Freemium
**Untuk:** GitHub Copilot, Claude, Gemini, ChatGPT, Antigravity  
**Stack:** Go + Chi + PostgreSQL + React + Vite + Tailwind  
**Konvensi:** modular monolith, chi router, pgx/v5, JWT auth, UUID sebagai primary key

> Setiap prompt berdiri sendiri dan dapat dijalankan langsung di VSCode (Copilot/Antigravity)
> atau di chat interface (Claude/Gemini/ChatGPT).
> Sertakan file konteks yang relevan sebelum menjalankan prompt.
> **Epic 0 harus selesai sebelum menjalankan prompt Epic 1–5.**

---

# EPIC 0 — Tier System & Billing

---

## Prompt E0-T1 — Extend Plan Tier

```
Context:
Go backend, PostgreSQL with pgx/v5.
Current plan_tier field in table `notification_preferences` only accepts 'free' | 'paid'.
We need to extend this to support three tiers: 'free' | 'pro' | 'enterprise'.

Also, users table may store plan_tier directly. Check both tables.

Task:
1. Create migration: migrations/NNNN_extend_plan_tier.sql

   -- Update CHECK constraint on notification_preferences
   ALTER TABLE notification_preferences
     DROP CONSTRAINT IF EXISTS notification_preferences_plan_tier_check;
   ALTER TABLE notification_preferences
     ADD CONSTRAINT notification_preferences_plan_tier_check
     CHECK (plan_tier IN ('free', 'pro', 'enterprise'));

   -- Update existing 'paid' values to 'pro' (migration of old data)
   UPDATE notification_preferences SET plan_tier = 'pro' WHERE plan_tier = 'paid';

   -- If users table also has plan_tier:
   ALTER TABLE users
     DROP CONSTRAINT IF EXISTS users_plan_tier_check;
   ALTER TABLE users
     ADD CONSTRAINT users_plan_tier_check
     CHECK (plan_tier IN ('free', 'pro', 'enterprise'));
   UPDATE users SET plan_tier = 'pro' WHERE plan_tier = 'paid';

2. Update Go constants in internal/tier/tier.go (create new file):

   package tier

   const (
     Free       = "free"
     Pro        = "pro"
     Enterprise = "enterprise"
   )

   // Limits per tier
   var ProfileLimit = map[string]int{
     Free:       1,
     Pro:        5,
     Enterprise: -1, // unlimited (-1 = no limit)
   }

   var DeviceLimit = map[string]int{
     Free:       1,
     Pro:        10,
     Enterprise: -1,
   }

   var HistoryDaysLimit = map[string]int{
     Free:       7,
     Pro:        90,
     Enterprise: -1,
   }

   // CanAccess returns true if the given tier can access the given feature
   // Features: "telegram_notif", "whatsapp_notif", "csv_export",
   //           "monthly_pdf", "annual_pdf", "rec_pdf", "mrv_pdf",
   //           "esg_dashboard", "white_label", "api_access"
   func CanAccess(userTier, feature string) bool {
     access := map[string][]string{
       "telegram_notif":  {Pro, Enterprise},
       "whatsapp_notif":  {Pro, Enterprise},
       "csv_export":      {Pro, Enterprise},
       "monthly_pdf":     {Pro, Enterprise},
       "annual_pdf":      {Pro, Enterprise},
       "rec_pdf":         {Pro, Enterprise},
       "mrv_pdf":         {Pro, Enterprise},
       "esg_dashboard":   {Enterprise},
       "white_label":     {Enterprise},
       "api_access":      {Pro, Enterprise},
     }
     allowed, ok := access[feature]
     if !ok { return true } // unknown feature = public
     for _, t := range allowed {
       if userTier == t { return true }
     }
     return false
   }

Do not change any other logic.
```

---

## Prompt E0-T2 — Tier Enforcement Middleware

```
Context:
Go backend, Chi router, JWT auth middleware already exists.
JWT payload contains user_id. User's plan_tier is stored in DB.
New file: internal/tier/tier.go with CanAccess() and limit constants (see E0-T1).

Task:
Create internal/middleware/tier.go with two middleware functions:

1. TierMiddleware — injects plan_tier into request context
   - Read user_id from JWT context (existing pattern)
   - Query plan_tier from users table (or notification_preferences)
   - Set into context: ctx = context.WithValue(ctx, tierContextKey, planTier)
   - Call next handler

2. RequireTier(features ...string) func(http.Handler) http.Handler
   - Reads plan_tier from context (set by TierMiddleware)
   - Checks tier.CanAccess(userTier, feature) for ALL given features
   - If any check fails: return JSON 403
     {
       "error": "feature_not_available",
       "message": "Fitur ini tidak tersedia di paket Anda.",
       "required_tier": "pro",  // minimum tier that has access
       "upgrade_url": "/pricing"
     }
   - If all checks pass: call next handler

Usage example in router:
  r.With(middleware.TierMiddleware, middleware.RequireTier("monthly_pdf")).
    Get("/reports/monthly/pdf", reportsHandler.MonthlyPDF)

Also create helper function GetTierFromContext(ctx context.Context) string
to be used in service layer for limit checks (profile count, history days).

Do not change existing auth middleware.
```

---

## Prompt E0-T3&T4 — Profile & Device Count Limit

```
Context:
Go backend, Chi router, pgx/v5.
Tier limits defined in internal/tier/tier.go:
  tier.ProfileLimit = {"free": 1, "pro": 5, "enterprise": -1}
  tier.DeviceLimit  = {"free": 1, "pro": 10, "enterprise": -1}
TierMiddleware injects plan_tier into request context.

Task:
In the existing solar_profiles handler/service, add limit check to POST /solar-profiles:

  func (s *service) CreateSolarProfile(ctx context.Context, userID uuid.UUID, input CreateProfileInput) (*SolarProfile, error) {
    // Get user tier from context
    userTier := middleware.GetTierFromContext(ctx)

    // Check limit
    limit := tier.ProfileLimit[userTier]
    if limit != -1 {
      count, err := s.repo.CountProfilesByUser(ctx, userID)
      if err != nil { return nil, err }
      if count >= limit {
        return nil, &tier.LimitError{
          Feature: "solar_profile",
          Current: count,
          Limit:   limit,
          Tier:    userTier,
          Message: fmt.Sprintf("Paket %s hanya mendukung %d site. Upgrade untuk menambah lebih banyak.", userTier, limit),
        }
      }
    }
    // ... existing create logic
  }

Do the same for POST /devices using tier.DeviceLimit.

Create internal/tier/errors.go:
  type LimitError struct {
    Feature string
    Current int
    Limit   int
    Tier    string
    Message string
  }
  func (e *LimitError) Error() string { return e.Message }

In HTTP handler, check for *tier.LimitError and return:
  HTTP 403 JSON:
  {
    "error": "tier_limit_reached",
    "message": "[LimitError.Message]",
    "current": [current],
    "limit": [limit],
    "upgrade_url": "/pricing"
  }
```

---

## Prompt E0-T5 — History Day-Limit per Tier

```
Context:
Go backend, pgx/v5.
Tier history limits in internal/tier/tier.go:
  tier.HistoryDaysLimit = {"free": 7, "pro": 90, "enterprise": -1}
TierMiddleware injects plan_tier into request context.

Task:
In the existing forecast/history service, apply day-limit filter:

  func (s *service) GetForecastHistory(ctx context.Context, userID uuid.UUID, profileID uuid.UUID, from, to time.Time) ([]Forecast, error) {
    userTier := middleware.GetTierFromContext(ctx)
    dayLimit := tier.HistoryDaysLimit[userTier]

    if dayLimit != -1 {
      minAllowedDate := time.Now().AddDate(0, 0, -dayLimit)
      if from.Before(minAllowedDate) {
        from = minAllowedDate
      }
    }
    // ... existing query with from/to filter
  }

Apply the same pattern to GetActualHistory.

Also add a response field `history_limit_days` in the history API response:
  {
    "data": [...],
    "history_limit_days": 7,   // or 90 or -1
    "tier": "free"
  }

This allows frontend to show the correct "upgrade to see more" message.
```

---

## Prompt E0-T6&T7 — Frontend Tier Badge & Lock UI

```
Context:
React + Tailwind + Zustand.
Auth store (src/stores/authStore.ts) contains user object with plan_tier field.
Sidebar component: src/components/Sidebar.tsx

Task:
1. Add TierBadge component (src/components/TierBadge.tsx):
   - Free: gray badge "Free"
   - Pro: yellow/amber badge "Pro ⭐"
   - Enterprise: blue badge "Enterprise 🏢"
   Show in sidebar below user name.

2. Add FeatureLock component (src/components/FeatureLock.tsx):
   Props: { feature: string, requiredTier: 'pro' | 'enterprise', children: ReactNode }
   - If user has access: render children normally
   - If user does NOT have access:
     - Render children with opacity-40 + pointer-events-none overlay
     - Show lock icon (🔒) in top-right corner
     - On click overlay: show upgrade modal

3. UpgradeModal component (src/components/UpgradeModal.tsx):
   - Title: "Fitur ini tersedia di paket [requiredTier]"
   - Brief feature description
   - CTA button: "Lihat Paket" → navigate to /pricing
   - Secondary: "Nanti saja" → close modal

4. Usage example in Reports page:
   <FeatureLock feature="monthly_pdf" requiredTier="pro">
     <button onClick={downloadPDF}>Download PDF</button>
   </FeatureLock>

Helper hook: src/hooks/useTierAccess.ts
  const { canAccess, userTier } = useTierAccess()
  canAccess('monthly_pdf') → boolean

Feature access map (mirrors backend tier.CanAccess):
  const FEATURE_TIERS = {
    telegram_notif: ['pro', 'enterprise'],
    whatsapp_notif: ['pro', 'enterprise'],
    csv_export:     ['pro', 'enterprise'],
    monthly_pdf:    ['pro', 'enterprise'],
    annual_pdf:     ['pro', 'enterprise'],
    rec_pdf:        ['pro', 'enterprise'],
    mrv_pdf:        ['pro', 'enterprise'],
    esg_dashboard:  ['enterprise'],
    white_label:    ['enterprise'],
  }
```

---

## Prompt E0-T9&T11 — Subscriptions Table & Payment Integration

```
Context:
Go backend, PostgreSQL, Chi router.
Payment gateway: Midtrans (preferred) or Xendit.
Tier values: 'free' | 'pro' | 'enterprise'.

Task 1 — DB Migration:
CREATE TABLE subscriptions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES users(id),
  tier            VARCHAR(20) NOT NULL CHECK (tier IN ('pro', 'enterprise')),
  status          VARCHAR(20) NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'grace')),
  start_date      DATE NOT NULL,
  end_date        DATE NOT NULL,
  payment_ref     VARCHAR(255),   -- payment gateway order ID
  payment_method  VARCHAR(50),
  amount_idr      INT NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

Task 2 — Checkout API:
POST /billing/checkout
Request: { "tier": "pro" | "enterprise", "period": "monthly" | "annual" }
Response: { "payment_url": "https://...", "order_id": "..." }

Logic:
  1. Calculate amount:
     pro+monthly = 99000, pro+annual = 899000
     enterprise+monthly = 499000, enterprise+annual = custom (return error for now)
  2. Create pending record in subscriptions
  3. Create Midtrans transaction, return payment_url

Task 3 — Webhook Handler:
POST /billing/webhook (no auth, verify Midtrans signature)
  - On payment success: UPDATE subscriptions SET status='active'
  - UPDATE users SET plan_tier = subscription.tier
  - Send confirmation email via existing notification service

Task 4 — Grace Period Scheduler:
Add to existing scheduler (daily job at 00:00 WIB):
  - Find subscriptions WHERE end_date < today AND status = 'active'
  - Set status = 'grace'
  - Find subscriptions WHERE end_date < today - 7 AND status = 'grace'
  - Set status = 'expired'
  - UPDATE users SET plan_tier = 'free' for expired subscriptions
  - Send expiry warning email 7 days before end_date
```

---
---

# EPIC 1 — Monthly Energy Report PDF (Pro)

---

## Prompt E1-T2&T3 — Monthly Report API with Tier Gate

```
Context:
Go backend, Chi router, pgx/v5, JWT auth.
Tier middleware already exists (see E0-T2).
Existing tables: forecasts (id, user_id, solar_profile_id, date, predicted_kwh,
actual_kwh, efficiency, delta_wf, weather_factor, baseline_type, weather_risk_status,
created_at), solar_profiles (id, user_id, name, capacity_kwp, lat, lng).

Task:
Create internal/reports package with:

handler.go — register routes:
  GET /reports/monthly        → JSON data (all tiers, but Free limited to 7 days)
  GET /reports/monthly/pdf    → PDF download (Pro/Enterprise only, tier gate)
  GET /reports/monthly/csv    → CSV download (Pro/Enterprise only, tier gate)

service.go — MonthlyReport(ctx, userID, profileID, year, month):
  1. Validate profile ownership
  2. Apply history day limit from tier (Free: only current month if within 7 days)
  3. Query forecasts for the period
  4. Calculate:
     total_predicted_kwh  = SUM(predicted_kwh)
     total_actual_kwh     = SUM(actual_kwh) WHERE actual_kwh > 0
     total_saving_idr     = total_actual_kwh * 1444
     total_co2_avoided_kg = total_actual_kwh * 0.87
     data_coverage_pct    = COUNT(actual_kwh > 0) / total_days * 100
  5. Return MonthlyReportResponse

Constants (internal/reports/constants.go):
  PLNTariffRpPerKwh    = 1444.0
  CO2FactorKgPerKwh    = 0.87
  CarbonCreditRpPerTon = 96000.0

Response JSON: same as previous E1-T2 spec + add:
  "tier":              "free" | "pro" | "enterprise"
  "pdf_available":     bool  (true if tier >= pro)
  "csv_available":     bool  (true if tier >= pro)

For PDF/CSV endpoints: use RequireTier("monthly_pdf") middleware.
Return 403 with upgrade_url if Free tier tries to access.
```

---

## Prompt E1-T4 — PDF Generation (gofpdf)

```
Context:
Go backend. Use github.com/jung-kurt/gofpdf.
MonthlyReportResponse struct defined in internal/reports/service.go.
Tier info available: Enterprise gets white-label logo option.

Task:
Create internal/pdfgen/base.go (shared helpers) and internal/reports/pdf.go.

internal/pdfgen/base.go:
  package pdfgen

  const (
    PageWidth    = 210.0  // A4 mm
    MarginLeft   = 15.0
    MarginRight  = 15.0
    MarginTop    = 20.0
    MarginBottom = 20.0
    ContentWidth = PageWidth - MarginLeft - MarginRight
  )

  func NewA4() *gofpdf.Fpdf  // A4 portrait, Arial font, standard margins
  func FormatRupiah(amount float64) string   // "Rp 1.234.567"
  func FormatFloat2(val float64) string      // "40.82"
  func IndonesianMonthName(month int) string // "Januari", "Februari", dst
  func FormatDateID(t time.Time) string      // "23 Maret 2026"

internal/reports/pdf.go:
  func GenerateMonthlyReportPDF(data MonthlyReportResponse, logoPath string) ([]byte, error)
  // logoPath: path to white-label logo (empty string = use default platform header)

PDF Layout (A4):
  HEADER
    - If logoPath != "": show custom logo (Enterprise white-label)
    - Else: "SOLAR FORECAST PLATFORM" text header
    - Title: "LAPORAN PRODUKSI ENERGI SURYA"
    - Subtitle: "[IndonesianMonthName] [Year]"

  SITE INFO BOX (bordered)
    Nama Site: | Kapasitas: | Periode: | Tanggal Cetak:

  KPI SUMMARY (4 boxes horizontal)
    Total Produksi Aktual | Estimasi Hemat | CO2 Dihindari | Coverage Data

  DAILY TABLE
    Tanggal | Prediksi (kWh) | Aktual (kWh) | Cloud Cover | Risiko Cuaca
    Alternating gray/white rows. "-" for missing actual.
    Page break if needed.

  FOOTER
    Metodologi note: tarif PLN Rp 1.444/kWh, faktor emisi 0.87 kgCO2/kWh
    "Laporan dihasilkan oleh Solar Forecast Platform — [generated_at]"

Return []byte of the PDF.
```

---

## Prompt E1-T5 — CSV Export

```
Context:
Go backend, Chi router.
MonthlyReportResponse already defined in internal/reports/service.go.
Tier gate: Pro/Enterprise only.

Task:
Add to internal/reports/handler.go:

GET /reports/monthly/csv?profile_id=&year=&month=
  - Tier gate: RequireTier("csv_export")
  - Generate CSV with headers:
    Tanggal,Prediksi (kWh),Aktual (kWh),Cloud Cover (%),Delta WF,Risiko Cuaca
  - Content-Type: text/csv
  - Content-Disposition: attachment; filename="data-[profile_name]-[year]-[month].csv"
  - Stream response directly (no temp file)

Also add annual CSV:
GET /reports/annual/csv?profile_id=&year=
  Headers: Bulan,Total Prediksi (kWh),Total Aktual (kWh),Hemat (Rp),CO2 (kg)
```

---
---

# EPIC 2 — Annual Summary & PBB Letter (Pro)
# [Prompts sama seperti v1, tambahkan tier gate dan white-label]

## Prompt E2-T4&T5 — Annual PDF with Tier Gate & White-label

```
Context:
Go backend, gofpdf, internal/pdfgen/base.go already exists.
AnnualReportResponse and SuratKeteranganInput structs defined in internal/reports/service.go.
Enterprise tier gets white-label: custom logo on PDF header.

Task:
Create in internal/reports/pdf.go:

  func GenerateAnnualReportPDF(
    data AnnualReportResponse,
    surat SuratKeteranganInput,
    logoPath string,  // empty = default, path = Enterprise white-label
  ) ([]byte, error)

PDF contains TWO documents concatenated (same as v1 spec).
Add logoPath support: if not empty, replace "SOLAR FORECAST PLATFORM" header
with the provided logo image (use gofpdf ImageOptions).

Also add endpoint:
  GET /reports/annual/pdf?profile_id=&year=&nomor_surat=&nama_pejabat=&jabatan=
  Tier gate: RequireTier("annual_pdf")
  Enterprise: also accept logo from user's stored white-label asset
```

---
---

# EPIC 3 — MWh Accumulator & REC Readiness
# [Sama seperti v1, tambahkan tier gate pada PDF]

## Prompt E3-T4 — REC Notification (Pro/Enterprise only)

```
Context:
Go backend. Existing notification service supports email + telegram + whatsapp.
Tier gate: Pro/Enterprise only for REC notifications.
When rec_claimable increases, send notification.

Task:
In internal/accumulator/service.go, after UpdateAccumulator:

  func (s *Service) CheckAndNotifyREC(ctx context.Context, profileID, userID uuid.UUID) error {
    // 1. Get current rec_claimable
    // 2. Get previous rec_claimable (from last snapshot or compare with previous run)
    // 3. If rec_claimable increased:
    //    a. Check user tier — only notify if Pro or Enterprise
    //    b. Send notification via existing notification service:
    //       Subject: "🌞 Site [name] berhasil mencapai REC baru!"
    //       Body: "Site [name] telah menghasilkan [n] MWh energi terbarukan.
    //              Total REC yang dapat diklaim: [rec_claimable] unit.
    //              Upgrade ke Pro untuk mendapatkan laporan PDF REC Anda."
    //              (remove upgrade message if already Pro/Enterprise)
    // 4. For Free tier: skip notification, but still update accumulator

  }
```

---
---

# EPIC 4 — CO2 Avoided Tracker & MRV Report
# [Sama seperti v1, tambahkan tier gate]

## Prompt E4-T5&T6 — MRV PDF with Tier Gate & White-label

```
Context:
Go backend, gofpdf.
Tier gate: Pro/Enterprise for MRV PDF.
Enterprise: white-label logo.
CO2SummaryResponse defined in internal/co2/service.go.

Task:
Create internal/co2/pdf.go:
  func GenerateMRVReportPDF(data CO2SummaryResponse, logoPath string) ([]byte, error)

Same 3-section layout as v1 spec (Measurement / Reporting / Verification).
Add logoPath support for Enterprise white-label.

Endpoint:
  GET /co2/report/pdf?profile_id=&period=
  Tier gate: RequireTier("mrv_pdf")
  Enterprise: fetch user's stored logo path for white-label
```

---
---

# EPIC 5 — ESG Dashboard (Enterprise)

---

## Prompt E5-T1&T2 — ESG Summary API (Enterprise only)

```
Context:
Go backend, Chi router, pgx/v5.
Tier gate: Enterprise only.
Data aggregated from: energy_accumulator, forecasts (actual_kwh), co2 calculations.

Task:
Create internal/esg package.

Endpoint:
  GET /esg/summary
  Tier gate: RequireTier("esg_dashboard")
  Auth: JWT. Aggregates ALL profiles owned by the authenticated user.

Response: same as v1 ESGSummaryResponse spec.

Also add:
  GET /esg/report/pdf?year=
  Tier gate: RequireTier("esg_dashboard")
  Enterprise: white-label logo from user's stored asset
  Returns ESG Report PDF (same layout as v1 spec).
```

---

## Prompt E5-T3 — ESG Dashboard Frontend (Enterprise gate)

```
Context:
React + Tailwind + TanStack Query + Recharts.
useTierAccess hook exists (see E0-T6&T7).
FeatureLock component exists.

Task:
Create src/pages/ESGDashboard.tsx

Wrap entire page content with Enterprise tier check:
  const { canAccess, userTier } = useTierAccess()

  if (!canAccess('esg_dashboard')) {
    return <EnterpriseLockScreen />  // see below
  }

EnterpriseLockScreen component (src/components/EnterpriseLockScreen.tsx):
  - Full page overlay (not just a modal)
  - Title: "ESG Dashboard — Paket Enterprise"
  - Description: "Kelola dan laporkan dampak energi terbarukan dari semua site Anda
    dalam satu dashboard. Siap untuk tender, CSR reporting, dan klaim carbon credit."
  - Feature list (checklist):
    ✅ Dashboard multi-site agregasi
    ✅ ESG Report PDF white-label
    ✅ CO2 MRV Report
    ✅ Public share link laporan
  - CTA: "Upgrade ke Enterprise" → /pricing#enterprise
  - Secondary: "Hubungi Kami" → WA link

If canAccess: render full ESG Dashboard (same layout as v1 spec).
Add route /esg → <ESGDashboard /> with "ESG" in sidebar (always visible, but locked for non-Enterprise).
```

---
---

# SHARED — Pricing Page

---

## Prompt SHARED-PRICING — Pricing Page Frontend

```
Context:
React + Tailwind. No auth required to view this page.
Route: /pricing
Tiers: Free (Rp 0) | Pro (Rp 99rb/bln) | Enterprise (Rp 499rb/bln)

Task:
Create src/pages/Pricing.tsx

Layout:
  1. Header: "Pilih Paket yang Tepat untuk Anda"
     Subtitle: "Mulai gratis, upgrade kapan saja."

  2. Billing toggle: "Bulanan" | "Tahunan (hemat 25%)"
     When annual selected, show annual prices.

  3. Three pricing cards (horizontal, Pro card slightly elevated/highlighted):

     FREE — Rp 0/bulan
     [List features from tier matrix]
     CTA: "Mulai Gratis" → /register

     PRO ⭐ — Rp 99.000/bulan (atau Rp 899.000/tahun)
     [List features, highlighted ones that Free doesn't have]
     CTA: "Upgrade ke Pro" → POST /billing/checkout {tier: "pro"}
     Badge: "Paling Populer"

     ENTERPRISE 🏢 — Rp 499.000/bulan
     [List features, highlighted Enterprise-exclusive]
     CTA: "Hubungi Kami" → WA link or contact form
     Note: "Tersedia annual quote untuk fleet besar"

  4. FAQ section (accordion):
     - "Apakah data saya aman jika downgrade?"
     - "Bagaimana cara upgrade?"
     - "Apa itu REC dan carbon credit?"
     - "Apakah bisa trial Pro?"

  5. Compare table (full feature matrix, collapsible):
     All features from tier matrix with ✅ / ❌ / text per tier.

Add "Lihat Paket" link in Sidebar for Free users.
Add "Upgrade" badge/button in Sidebar next to tier badge for Free/Pro users.
```

---
---

# EPIC 6 — Admin Dashboard

---

## Prompt E6-T1 — Admin Auth Middleware

```
Context:
Go backend, Chi router, JWT auth already exists.
Current JWT payload contains: user_id, email.
We need to add admin role support.

Task:
1. Add `is_admin` boolean field to users table:
   ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

2. Add `is_admin` claim to JWT payload when generating tokens.
   In internal/auth/service.go, include is_admin in claims:
     claims["is_admin"] = user.IsAdmin

3. Create internal/middleware/admin.go:

   // RequireAdmin middleware — blocks non-admin users with 403
   func RequireAdmin(next http.Handler) http.Handler {
     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       claims := GetClaimsFromContext(r.Context())
       isAdmin, _ := claims["is_admin"].(bool)
       if !isAdmin {
         http.Error(w, `{"error":"forbidden","message":"Admin access required"}`,
           http.StatusForbidden)
         return
       }
       next.ServeHTTP(w, r)
     })
   }

4. Register admin routes in main router under /admin prefix:
   r.Route("/admin", func(r chi.Router) {
     r.Use(middleware.AuthMiddleware)   // JWT required
     r.Use(middleware.RequireAdmin)     // admin role required
     // admin handlers registered here
   })

5. Set first admin via SQL (manual, one-time):
   UPDATE users SET is_admin = TRUE WHERE email = 'your@email.com';

Do not change existing auth middleware or user-facing routes.
```

---

## Prompt E6-T2&T3 — User Management API

```
Context:
Go backend, Chi router, pgx/v5.
Admin auth middleware exists (RequireAdmin).
Tables: users, subscriptions, solar_profiles, devices.
Tier values: 'free' | 'pro' | 'enterprise'.

Task:
Create internal/admin/users_handler.go and internal/admin/users_service.go

Endpoint 1 — List users:
  GET /admin/users
  Query params:
    tier     string  (optional: free|pro|enterprise)
    status   string  (optional: active|expired|grace|free)
    search   string  (optional: email or name)
    page     int     (default: 1)
    per_page int     (default: 50)

  Response JSON:
  {
    "users": [
      {
        "id":              "uuid",
        "email":           "string",
        "name":            "string",
        "plan_tier":       "free|pro|enterprise",
        "is_admin":        bool,
        "profile_count":   int,
        "device_count":    int,
        "subscription": {
          "status":        "active|expired|grace|none",
          "tier":          "string",
          "end_date":      "YYYY-MM-DD",
          "days_left":     int
        },
        "last_forecast_at": "ISO8601 or null",
        "created_at":     "ISO8601"
      }
    ],
    "total": int,
    "page":  int,
    "per_page": int
  }

Endpoint 2 — Manual tier upgrade/downgrade:
  PATCH /admin/users/:id/tier
  Request: { "tier": "free|pro|enterprise", "reason": "string" }
  Logic:
    1. Update users SET plan_tier = $tier WHERE id = $id
    2. If upgrading to pro/enterprise: create or extend subscription
       end_date = today + 30 days (monthly) — admin manual override
    3. If downgrading to free: set subscription status = 'cancelled'
    4. Insert into audit_log: admin_id, action='tier_change', target_user_id,
       old_tier, new_tier, reason, timestamp
  Response: updated user object

Endpoint 3 — Extend subscription:
  PATCH /admin/users/:id/subscription/extend
  Request: { "days": int }
  Logic: UPDATE subscriptions SET end_date = end_date + $days WHERE user_id = $id
```

---

## Prompt E6-T4 — Impersonate User

```
Context:
Go backend, Chi router, JWT auth.
Admin can generate a temporary token to login as any user for debugging.

Task:
Create endpoint:
  POST /admin/users/:id/impersonate

Logic:
  1. Verify requester is admin (RequireAdmin middleware)
  2. Fetch target user by id
  3. Generate a short-lived JWT (15 minutes) with target user's claims
     + extra claim: "impersonated_by": admin_user_id
  4. Insert into audit_log: action='impersonate', admin_id, target_user_id, timestamp
  5. Return:
     { "token": "jwt_string", "expires_in": 900, "user": { id, email, plan_tier } }

Security notes:
  - Impersonate token expires in 15 minutes max (hardcode, not configurable)
  - Impersonate token cannot impersonate another admin (check target is_admin=false)
  - All actions done with impersonate token are logged with impersonated_by field
  - Add warning header to all responses when impersonate token is used:
    X-Impersonated-By: [admin_email]
```

---

## Prompt E6-T6 — Scheduler Status API

```
Context:
Go backend. Existing scheduler runs daily at 06:00 WIB.
We need to expose scheduler run history to admin.

Task:
1. Create table: scheduler_runs
   CREATE TABLE scheduler_runs (
     id           BIGSERIAL PRIMARY KEY,
     run_date     DATE NOT NULL,
     started_at   TIMESTAMPTZ NOT NULL,
     finished_at  TIMESTAMPTZ,
     total_sites  INT NOT NULL DEFAULT 0,
     success      INT NOT NULL DEFAULT 0,
     failed       INT NOT NULL DEFAULT 0,
     status       VARCHAR(20) NOT NULL DEFAULT 'running',
     -- status: 'running' | 'completed' | 'partial' | 'failed'
     UNIQUE(run_date)
   );

   CREATE TABLE scheduler_run_errors (
     id              BIGSERIAL PRIMARY KEY,
     run_id          BIGINT REFERENCES scheduler_runs(id),
     solar_profile_id UUID,
     user_id         UUID,
     error_message   TEXT,
     created_at      TIMESTAMPTZ DEFAULT NOW()
   );

2. Instrument existing scheduler to log into these tables:
   - INSERT scheduler_runs at start of daily run
   - UPDATE after each site: increment success or failed
   - INSERT scheduler_run_errors for each failed site
   - UPDATE status and finished_at when done

3. Create API:
   GET /admin/scheduler/status
   Response:
   {
     "today": {
       "run_date":    "YYYY-MM-DD",
       "status":      "completed|running|partial|failed|not_run",
       "started_at":  "ISO8601",
       "finished_at": "ISO8601",
       "total_sites": int,
       "success":     int,
       "failed":      int,
       "errors": [
         { "profile_id": "uuid", "user_id": "uuid", "error": "string" }
       ]
     },
     "history": [  // last 7 days
       { "run_date", "status", "success", "failed", "total_sites" }
     ]
   }
```

---

## Prompt E6-T10&T11 — Forecast Quality & Cold Start Monitor API

```
Context:
Go backend, pgx/v5.
Tables: forecasts (predicted_kwh, actual_kwh, date, solar_profile_id, baseline_type),
        solar_profiles (id, name, user_id, capacity_kwp).

Task:
Create internal/admin/quality_handler.go

Endpoint 1 — Forecast Quality per site:
  GET /admin/forecast-quality?days=7

  For each active site (has forecast in last 30 days):
    MAPE = avg(|actual - predicted| / actual * 100) for last N days
    Only include days where actual_kwh > 0

  Response:
  {
    "sites": [
      {
        "profile_id":    "uuid",
        "profile_name":  "string",
        "user_email":    "string",
        "capacity_kwp":  float64,
        "baseline_type": "synthetic|blended|site",
        "efficiency":    float64,
        "mape_7d":       float64,
        "mape_flag":     "good|warning|critical",
        // good: mape < 15%, warning: 15-30%, critical: > 30%
        "actual_days_count":  int,   // total days with actual
        "last_actual_date":   "YYYY-MM-DD or null",
        "days_since_actual":  int
      }
    ],
    "summary": {
      "total_sites":    int,
      "avg_mape":       float64,
      "critical_count": int,
      "warning_count":  int,
      "good_count":     int
    }
  }

Endpoint 2 — Cold Start Monitor:
  GET /admin/cold-start-monitor?min_days=30

  Return sites where:
    baseline_type = 'synthetic' AND first forecast > min_days ago
    (these users have never submitted actual data)

  Response:
  {
    "sites": [
      {
        "profile_id":       "uuid",
        "profile_name":     "string",
        "user_email":       "string",
        "plan_tier":        "string",
        "first_forecast":   "YYYY-MM-DD",
        "days_since_start": int,
        "forecast_count":   int,
        "actual_count":     0
      }
    ],
    "total": int
  }
```

---

## Prompt E6-T13 — Data Anomaly Detector API

```
Context:
Go backend, pgx/v5.
Tables: forecasts (predicted_kwh, actual_kwh, date, solar_profile_id).

Task:
Create endpoint: GET /admin/data-anomalies

Detect three types of anomalies:

TYPE 1 — actual_kwh suspiciously high:
  actual_kwh > predicted_kwh * 1.5
  Possible cause: wrong manual input (e.g. user entered Wh instead of kWh)

TYPE 2 — actual_kwh = 0 or near-zero for many consecutive days:
  actual_kwh < 0.01 for 7+ consecutive days
  Possible cause: sensor offline, user stopped inputting

TYPE 3 — Coverage too low:
  Days with actual_kwh > 0 / total forecast days < 50% in last 30 days
  Possible cause: user not engaged, IoT device disconnected

Response:
{
  "anomalies": [
    {
      "type":         "high_actual|zero_streak|low_coverage",
      "profile_id":   "uuid",
      "profile_name": "string",
      "user_email":   "string",
      "plan_tier":    "string",
      "detail": {
        // for high_actual:
        "date":          "YYYY-MM-DD",
        "predicted_kwh": float64,
        "actual_kwh":    float64,
        "ratio":         float64,
        // for zero_streak:
        "streak_days":   int,
        "streak_from":   "YYYY-MM-DD",
        // for low_coverage:
        "coverage_pct":  float64,
        "actual_days":   int,
        "total_days":    int
      }
    }
  ],
  "total": int,
  "generated_at": "ISO8601"
}
```

---

## Prompt E6-T7 — Frontend Admin User Management Page

```
Context:
React + Tailwind + TanStack Query.
Admin routes under /admin — only accessible with is_admin=true JWT claim.
Add admin check in router: if (!user.is_admin) redirect to /dashboard.

Task:
Create src/pages/admin/UserManagement.tsx

Layout:
  1. Page header: "User Management" + badge "Admin"
  2. Filter bar:
     - Search input (email or name)
     - Dropdown filter: Tier (All / Free / Pro / Enterprise)
     - Dropdown filter: Status (All / Active / Expired / Grace)
  3. Stats row (4 cards):
     - Total Users | Free | Pro | Enterprise
  4. User table:
     Columns:
       Email | Name | Tier badge | Subscription status | Sites | Devices |
       Last forecast | Actions
     Tier badge colors:
       Free = gray | Pro = amber | Enterprise = blue
     Subscription status badge:
       active = green | expired = red | grace = orange | none = gray
     Actions column (dropdown per row):
       - "Upgrade ke Pro"
       - "Upgrade ke Enterprise"
       - "Downgrade ke Free"
       - "Extend 30 hari"
       - "Login sebagai user" (impersonate)
  5. Upgrade/Downgrade modal:
     - Confirm dialog: "Ubah tier [email] dari [old] ke [new]?"
     - Input: Reason (text, optional)
     - Button: Konfirmasi
  6. Pagination (50 per page)

API calls:
  GET /admin/users?tier=&status=&search=&page=
  PATCH /admin/users/:id/tier
  POST /admin/users/:id/impersonate → store token in sessionStorage, redirect to /dashboard

Add /admin route group in src/router.tsx with AdminGuard wrapper.
Add "Admin" link in sidebar — only visible if user.is_admin === true.
```

---

## Prompt E6-T8 — Frontend Scheduler Monitor

```
Context:
React + Tailwind + TanStack Query.
API: GET /admin/scheduler/status

Task:
Create src/pages/admin/SchedulerMonitor.tsx

Layout:
  1. Header: "Scheduler Monitor" + last updated timestamp + auto-refresh button
  2. Today's run card (large, full width):
     Status badge (large):
       running    = blue pulsing
       completed  = green
       partial    = orange (some failed)
       failed     = red
       not_run    = gray
     Stats: [success] / [total_sites] sites berhasil
     Time: started_at → finished_at (or "sedang berjalan...")
  3. Error list (if any failed sites):
     Table: Profile Name | User Email | Error Message
     Collapsible if > 5 errors
  4. History table (last 7 days):
     Date | Status badge | Success | Failed | Total | Duration
  5. Auto-refresh every 30 seconds when status = 'running'

Note: if status = 'not_run' and time > 07:00 WIB, show warning banner:
  "⚠️ Scheduler belum berjalan hari ini. Periksa server."
```

---

## Prompt E6-T14 — Frontend Forecast Quality Dashboard

```
Context:
React + Tailwind + TanStack Query + Recharts.
APIs:
  GET /admin/forecast-quality?days=7
  GET /admin/cold-start-monitor?min_days=30

Task:
Create src/pages/admin/ForecastQuality.tsx

Layout:
  SECTION 1 — Forecast Quality
  1. Header: "Forecast Quality Monitor"
  2. Summary cards (3):
     ✅ Good (MAPE < 15%) | ⚠️ Warning (15–30%) | ❌ Critical (> 30%)
  3. Filter: Days selector (7 / 14 / 30)
  4. Quality table:
     Profile | User | Capacity | Tier | Baseline Type | MAPE | Flag | Efficiency | Last Actual | Days Since
     Flag column: color-coded badge (good=green / warning=yellow / critical=red)
     Sort by MAPE descending (worst first)
     Click row → expand: show last 7 days predicted vs actual inline chart

  SECTION 2 — Cold Start Monitor
  5. Header: "Cold Start Monitor"
     Subtitle: "Site yang sudah > 30 hari tapi belum pernah input actual"
  6. Table:
     Profile | User | Tier | First Forecast | Days Active | Forecasts Count | Actual Count
     Action: "Kirim Reminder" (opens email draft modal)
  7. Export button: download CSV list untuk outreach manual

Both sections on same page, separated by divider.
```