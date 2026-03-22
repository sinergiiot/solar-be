# Delta Weather Factor (ΔWF) — Synthetic Baseline dari Open-Meteo Historical

> Dokumen ini dirancang untuk dibaca oleh GitHub Copilot sebagai konteks implementasi.
> Setiap bagian berisi: **WHAT** (apa yang dihitung), **WHY** (alasan teknis), **HOW** (langkah implementasi).

---

## 1. Konteks dan Tujuan

Sistem Solar Forecast memprediksi produksi energi PLTS harian per site.
Faktor cuaca (cloud cover) mempengaruhi hasil prediksi, namun cara menerapkannya
berbeda tergantung apakah site sudah memiliki data actual atau belum.

**Masalah dengan Weather Factor absolut:**
- `WF = 1 - cloud_cover/100` memotong prediksi secara linear
- Cloud cover 80% → prediksi dipotong 80% — terlalu agresif
- Hubungan cloud cover vs produksi aktual tidak linear
- Tidak konsisten antara mode cold start dan mode terkalibrasi

**Solusi: Delta Weather Factor (ΔWF)**
- ΔWF mengukur *deviasi cuaca hari ini* relatif terhadap *baseline historis site*
- Jika cuaca hari ini sama dengan rata-rata historis → ΔWF = 1.0 (tidak ada koreksi)
- Koreksi hanya aktif saat cuaca hari ini signifikan berbeda dari baseline

**ΔWF dipakai di SEMUA mode — termasuk mode terkalibrasi:**
- Mode cold start: ΔWF berbasis synthetic baseline (Open-Meteo Historical 30 hari)
- Mode terkalibrasi: ΔWF berbasis site baseline (hari-hari actual valid milik site)
- Perbedaannya hanya pada *sumber baseline*, bukan pada ada/tidaknya ΔWF

**Mengapa mode terkalibrasi tetap butuh ΔWF:**
- `η_calibrated` menyerap efek cuaca *rata-rata* dari hari-hari yang pernah diobservasi
- Tapi η tidak bisa memprediksi hari ini yang cuacanya jauh di atas/bawah rata-rata tersebut
- ΔWF mengisi gap ini: mengoreksi sisa deviasi cuaca yang tidak tertangkap oleh η
- Tanpa ΔWF, hari sangat cerah akan under-estimate, hari sangat mendung akan over-estimate

---

## 2. Rumus Lengkap

### 2a. Mode Cold Start (belum ada data actual)

Baseline diambil dari rata-rata cloud cover 30 hari terakhir via Open-Meteo Historical API.

```
baseline_cloud_cover = avg(cloud_cover_30_days_historical[lat, lng])

ΔWF_cold = (1 - cloud_cover_today / 100) / (1 - baseline_cloud_cover / 100)

E_pred = P_rated × PSH × η_default × clamp(ΔWF_cold, 0.5, 1.5)
```

**Catatan:**
- `η_default` = nilai awal efficiency, misal 0.80
- `clamp(x, 0.5, 1.5)` = batasi ΔWF agar tidak terlalu ekstrem
- Jika `baseline_cloud_cover >= 95`, gunakan fallback: `ΔWF = 1 - cloud_cover_today / 100`
  (hindari division by zero / nilai baseline yang terlalu kecil)

### 2b. Mode Terkalibrasi (sudah ada data actual, n_actual >= N_THRESHOLD)

Baseline diambil dari rata-rata cloud cover pada hari-hari yang punya data actual valid.
ΔWF **tetap dipakai** di mode ini — bukan dihilangkan.

```
// Site baseline: rata-rata cloud cover pada hari-hari actual valid
baseline_cloud_cover = avg(cloud_cover_mean WHERE actual_kwh > 0
                           AND profile_id = X AND user_id = Y)

// ΔWF sebagai residual correction
ΔWF_cal = (1 - cloud_cover_today / 100) / (1 - baseline_cloud_cover / 100)

E_pred = P_rated × PSH × η_calibrated × clamp(ΔWF_cal, 0.5, 1.5)
```

**Catatan penting — bedanya dengan cold start:**

| Aspek | Cold start | Terkalibrasi |
|-------|-----------|--------------|
| Sumber baseline | Open-Meteo Historical 30 hari | Data actual valid site |
| η yang dipakai | η_default (misal 0.80) | η_calibrated (hasil learning harian) |
| Peran ΔWF | Koreksi utama cuaca | Residual correction (deviasi dari rata-rata) |
| ΔWF saat cuaca normal | ≈ 1.0 | ≈ 1.0 (tidak ada koreksi) |
| ΔWF saat cuaca ekstrem | Koreksi signifikan | Koreksi signifikan |

**Mengapa ΔWF tidak dihilangkan meski η sudah terkalibrasi:**
- `η_calibrated` adalah scalar — satu angka yang merata-ratakan performa historis
- η tidak tahu apakah hari ini lebih cerah atau lebih mendung dari rata-rata historis
- ΔWF = 1.0 saat cuaca normal → tidak mengganggu kalibrasi
- ΔWF ≠ 1.0 hanya saat cuaca menyimpang → koreksi tepat sasaran
- Ini bukan over-correction karena baseline berasal dari data site sendiri, bukan asumsi generik

### 2c. Mode Transisi (data actual ada tapi masih sedikit, 0 < n < N_THRESHOLD)

Blend antara cold start baseline dan site baseline secara gradual.

```
N_threshold = 14  // hari minimum sebelum dianggap "terkalibrasi penuh"
w = n_actual / N_threshold  // weight untuk site baseline (0.0 → 1.0)

baseline_blended = (1 - w) × baseline_cold + w × baseline_site

ΔWF_transition = (1 - cloud_cover_today / 100) / (1 - baseline_blended / 100)

E_pred = P_rated × PSH × η_calibrated × clamp(ΔWF_transition, 0.5, 1.5)
```

---

### Ringkasan tiga mode — referensi cepat untuk Copilot

```
// KONDISI 1: Cold start (n_actual == 0)
//   - baseline  : synthetic (Open-Meteo Historical 30 hari)
//   - efficiency: η_default
//   - ΔWF       : DIPAKAI, baseline = rata-rata cuaca regional
//
// KONDISI 2: Terkalibrasi (n_actual >= 14)
//   - baseline  : site (rata-rata cloud cover saat actual valid)
//   - efficiency: η_calibrated
//   - ΔWF       : DIPAKAI, baseline = rata-rata cuaca site sendiri
//
// KONDISI 3: Transisi (0 < n_actual < 14)
//   - baseline  : blend(synthetic, site) dengan w = n_actual / 14
//   - efficiency: η_calibrated
//   - ΔWF       : DIPAKAI, baseline = interpolasi keduanya
//
// SEMUA KONDISI menggunakan rumus yang sama:
//   E_pred = P_rated × PSH × efficiency × clamp(ΔWF, 0.5, 1.5)
//
// Yang berbeda antar kondisi: HANYA sumber baseline dan nilai efficiency.
// ΔWF = 1.0 jika cuaca hari ini == rata-rata baseline (tidak ada koreksi).
// ΔWF ≠ 1.0 hanya jika cuaca menyimpang signifikan dari baseline.
```

---

## 3. Konversi Radiasi ke PSH

```
PSH = H_daily_MJ / 3.6

// H_daily_MJ = shortwave_radiation_sum dari Open-Meteo (satuan MJ/m²/day)
// PSH = Peak Sun Hours
```

---

## 4. Struktur Data yang Dibutuhkan

### Tabel baru: `weather_baselines`

```sql
CREATE TABLE weather_baselines (
    id              BIGSERIAL PRIMARY KEY,
    profile_id      BIGINT NOT NULL REFERENCES solar_profiles(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    baseline_type   VARCHAR(20) NOT NULL, -- 'synthetic' | 'site'
    cloud_cover_avg FLOAT NOT NULL,       -- rata-rata cloud cover (0-100)
    sample_count    INT NOT NULL,         -- jumlah hari yang dipakai
    date_from       DATE NOT NULL,        -- tanggal awal sampel
    date_to         DATE NOT NULL,        -- tanggal akhir sampel
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(profile_id, user_id, baseline_type)
);
```

### Field tambahan di tabel `weather_daily`

```sql
-- Tambahkan kolom ini ke tabel weather_daily yang sudah ada:
cloud_cover_mean FLOAT  -- rata-rata cloud cover harian (0-100), dari Open-Meteo
```

### Field tambahan di tabel `forecasts`

```sql
-- Tambahkan kolom ini ke tabel forecasts yang sudah ada:
delta_wf        FLOAT,  -- nilai ΔWF yang dipakai saat kalkulasi
baseline_type   VARCHAR(20)  -- 'synthetic' | 'site' | 'blended'
```

---

## 5. Algoritma Lengkap: Komputasi Synthetic Baseline

### Fungsi: `ComputeSyntheticBaseline(profileID, lat, lng)`

```
INPUT:
  profileID   : ID solar profile
  lat, lng    : koordinat lokasi site

OUTPUT:
  baseline_cloud_cover (float, 0-100)
  sample_count (int)

LANGKAH:
  1. Tentukan date range: date_to = today - 1, date_from = today - 30
  2. Panggil Open-Meteo Historical API:
       URL: https://archive-api.open-meteo.com/v1/archive
       params:
         latitude     = lat
         longitude    = lng
         start_date   = date_from (format: YYYY-MM-DD)
         end_date     = date_to   (format: YYYY-MM-DD)
         daily        = ["cloudcover_mean"]  // atau "cloud_cover_mean" tergantung versi API
         timezone     = "Asia/Jakarta"       // sesuaikan timezone site
  3. Dari response, ambil array daily.cloudcover_mean
  4. Filter nilai null / NaN
  5. Hitung rata-rata: baseline = sum(values) / len(values)
  6. Simpan ke tabel weather_baselines dengan baseline_type = 'synthetic'
  7. Return baseline, sample_count
```

**Endpoint Open-Meteo Historical:**
```
GET https://archive-api.open-meteo.com/v1/archive
  ?latitude={lat}
  &longitude={lng}
  &start_date={YYYY-MM-DD}
  &end_date={YYYY-MM-DD}
  &daily=cloud_cover_mean
  &timezone=Asia%2FJakarta
```

**Contoh Response:**
```json
{
  "daily": {
    "time": ["2024-12-01", "2024-12-02", ...],
    "cloud_cover_mean": [72.0, 45.0, 88.0, ...]
  }
}
```

---

## 6. Algoritma Lengkap: Komputasi Site Baseline

### Fungsi: `ComputeSiteBaseline(profileID, userID)`

```
INPUT:
  profileID : ID solar profile
  userID    : ID user

OUTPUT:
  baseline_cloud_cover (float, 0-100)
  sample_count (int)

LANGKAH:
  1. Query join antara:
       - tabel forecasts      (filter: user_id, profile_id, actual_kwh > 0)
       - tabel weather_daily  (join by lat/lng/date)
     untuk mendapatkan cloud_cover_mean pada hari-hari yang ada actual valid
  2. Filter nilai null
  3. Hitung rata-rata
  4. Simpan ke tabel weather_baselines dengan baseline_type = 'site'
  5. Return baseline, sample_count

CATATAN:
  - Jalankan fungsi ini setiap kali ada actual baru yang masuk (event-driven)
    atau via scheduler harian (misalnya jam 01:00 WIB)
  - Minimal 3 data actual valid sebelum site baseline dianggap representatif
```

---

## 7. Algoritma Lengkap: Kalkulasi ΔWF dan Prediksi Energi

### Fungsi: `ComputeDeltaWF(profileID, userID, cloudCoverToday) → deltaWF, baselineType`

```
INPUT:
  profileID       : ID solar profile
  userID          : ID user
  cloudCoverToday : cloud cover hari ini (0-100), dari weather_daily

OUTPUT:
  deltaWF      (float)
  baselineType (string: 'synthetic' | 'site' | 'blended')

LANGKAH:
  1. Hitung n_actual = COUNT(forecasts WHERE user_id = userID
                                       AND profile_id = profileID
                                       AND actual_kwh > 0)

  2. JIKA n_actual == 0:
       // Cold start: pakai synthetic baseline
       baseline = GetOrComputeSyntheticBaseline(profileID, lat, lng)
       baselineType = 'synthetic'

  3. JIKA n_actual >= N_THRESHOLD (default: 14):
       // Terkalibrasi penuh: pakai site baseline
       baseline = GetOrComputeSiteBaseline(profileID, userID)
       baselineType = 'site'

  4. JIKA 0 < n_actual < N_THRESHOLD:
       // Transisi: blend keduanya
       w = n_actual / N_THRESHOLD
       baseline_cold = GetOrComputeSyntheticBaseline(profileID, lat, lng)
       baseline_site = GetOrComputeSiteBaseline(profileID, userID)
       baseline = (1 - w) × baseline_cold + w × baseline_site
       baselineType = 'blended'

  5. Hitung denominator = 1 - baseline / 100
     JIKA denominator < 0.05:
       // Fallback: baseline terlalu tinggi, gunakan WF absolut
       deltaWF = 1 - cloudCoverToday / 100
       RETURN deltaWF, baselineType + '_fallback'

  6. deltaWF = (1 - cloudCoverToday / 100) / denominator

  7. deltaWF = clamp(deltaWF, 0.5, 1.5)

  8. RETURN deltaWF, baselineType
```

### Fungsi: `ComputeForecast(profile, weather, efficiency) → predictedKwh`

```
INPUT:
  profile    : { capacity_kwp, ... }
  weather    : { shortwave_radiation_sum_mj, cloud_cover_mean }
  efficiency : float (η_default atau η_calibrated)

OUTPUT:
  predictedKwh : float

LANGKAH:
  1. psh = weather.shortwave_radiation_sum_mj / 3.6
  2. deltaWF, baselineType = ComputeDeltaWF(profileID, userID, weather.cloud_cover_mean)
  3. predictedKwh = profile.capacity_kwp × psh × efficiency × deltaWF
  4. predictedKwh = max(predictedKwh, 0)  // tidak boleh negatif
  5. RETURN predictedKwh
```

---

## 8. Caching dan Refresh Strategy

```
Synthetic baseline:
  - Cache di tabel weather_baselines (baseline_type = 'synthetic')
  - TTL: 7 hari (refresh mingguan via scheduler)
  - Alasan: data historis 30 hari tidak berubah cepat

Site baseline:
  - Cache di tabel weather_baselines (baseline_type = 'site')
  - Refresh: setiap kali ada actual baru yang masuk
  - Atau: scheduler harian jam 01:00 WIB

Lookup order:
  1. Cek weather_baselines di DB (by profile_id, user_id, baseline_type)
  2. Jika ada dan tidak expired → pakai cached value
  3. Jika tidak ada atau expired → recompute → simpan ke DB → return
```

---

## 9. Open-Meteo Historical API — Detail Endpoint

```
Base URL : https://archive-api.open-meteo.com/v1/archive

Parameter wajib:
  latitude    float   Koordinat lintang
  longitude   float   Koordinat bujur
  start_date  string  Format: YYYY-MM-DD
  end_date    string  Format: YYYY-MM-DD
  daily       string  Variabel yang diminta

Variabel daily yang dipakai:
  cloud_cover_mean       Rata-rata cloud cover harian (%)
  shortwave_radiation_sum Total radiasi harian (MJ/m²)  [opsional, jika ingin PSH historis]

Timezone:
  Asia/Jakarta   Untuk site di WIB
  Asia/Makassar  Untuk site di WITA
  Asia/Jayapura  Untuk site di WIT

Contoh URL lengkap:
  https://archive-api.open-meteo.com/v1/archive
    ?latitude=-7.2575
    &longitude=112.7521
    &start_date=2024-12-01
    &end_date=2024-12-31
    &daily=cloud_cover_mean
    &timezone=Asia%2FJakarta

Response field yang dipakai:
  daily.time[]              Array tanggal (string YYYY-MM-DD)
  daily.cloud_cover_mean[]  Array nilai cloud cover (float, 0-100, nullable)
```

---

## 10. Pseudocode Go — Struktur Fungsi

```go
// weatherbaseline/service.go

// SyntheticBaseline mengambil atau menghitung baseline cloud cover dari Open-Meteo Historical
func (s *Service) GetSyntheticBaseline(ctx context.Context, profileID int64, lat, lng float64) (float64, error) {
    // 1. Cek cache di DB
    cached, err := s.repo.GetBaseline(ctx, profileID, 0, "synthetic")
    if err == nil && !isExpired(cached.ComputedAt, 7*24*time.Hour) {
        return cached.CloudCoverAvg, nil
    }

    // 2. Fetch dari Open-Meteo Historical
    dateTo := time.Now().AddDate(0, 0, -1)
    dateFrom := dateTo.AddDate(0, 0, -30)
    values, err := s.openMeteo.FetchHistoricalCloudCover(ctx, lat, lng, dateFrom, dateTo)
    if err != nil {
        return 0, err
    }

    // 3. Hitung rata-rata, filter null
    avg, count := computeAverage(values)

    // 4. Simpan ke DB
    s.repo.UpsertBaseline(ctx, profileID, 0, "synthetic", avg, count, dateFrom, dateTo)

    return avg, nil
}

// ComputeDeltaWF menghitung delta weather factor berdasarkan mode data actual
func (s *Service) ComputeDeltaWF(ctx context.Context, profileID, userID int64, cloudCoverToday float64, lat, lng float64) (float64, string, error) {
    nActual, _ := s.repo.CountActualDays(ctx, profileID, userID)
    const nThreshold = 14

    var baseline float64
    var baselineType string

    switch {
    case nActual == 0:
        // KONDISI 1: Cold start — pakai synthetic baseline dari Open-Meteo Historical
        // ΔWF mengoreksi prediksi berdasarkan deviasi cuaca hari ini
        // terhadap rata-rata cuaca regional 30 hari terakhir.
        b, err := s.GetSyntheticBaseline(ctx, profileID, lat, lng)
        if err != nil { return 0, "", err }
        baseline, baselineType = b, "synthetic"

    case nActual >= nThreshold:
        // KONDISI 2: Terkalibrasi penuh — pakai site baseline
        // ΔWF tetap dihitung sebagai residual correction.
        // η_calibrated sudah menyerap performa rata-rata, tapi tidak tahu
        // apakah hari ini lebih cerah/mendung dari rata-rata tersebut.
        // ΔWF mengisi gap ini tanpa mengganggu kalibrasi (ΔWF=1.0 saat cuaca normal).
        b, err := s.GetSiteBaseline(ctx, profileID, userID)
        if err != nil { return 0, "", err }
        baseline, baselineType = b, "site"

    default:
        // Transisi: blend
        w := float64(nActual) / float64(nThreshold)
        bCold, _ := s.GetSyntheticBaseline(ctx, profileID, lat, lng)
        bSite, _ := s.GetSiteBaseline(ctx, profileID, userID)
        baseline = (1-w)*bCold + w*bSite
        baselineType = "blended"
    }

    // Hitung ΔWF
    denominator := 1.0 - baseline/100.0
    if denominator < 0.05 {
        // Fallback ke WF absolut
        deltaWF := 1.0 - cloudCoverToday/100.0
        return deltaWF, baselineType + "_fallback", nil
    }

    deltaWF := (1.0 - cloudCoverToday/100.0) / denominator
    deltaWF = clamp(deltaWF, 0.5, 1.5)

    return deltaWF, baselineType, nil
}

// ComputeForecast menghitung prediksi energi harian dengan ΔWF
func (s *Service) ComputeForecast(ctx context.Context, profile SolarProfile, weather WeatherDaily, efficiency float64) (ForecastResult, error) {
    psh := weather.ShortwaveRadiationSum / 3.6

    deltaWF, baselineType, err := s.ComputeDeltaWF(ctx, profile.ID, profile.UserID,
        weather.CloudCoverMean, profile.Lat, profile.Lng)
    if err != nil {
        return ForecastResult{}, err
    }

    predictedKwh := profile.CapacityKwp * psh * efficiency * deltaWF
    if predictedKwh < 0 {
        predictedKwh = 0
    }

    return ForecastResult{
        PredictedKwh: predictedKwh,
        DeltaWF:      deltaWF,
        BaselineType: baselineType,
        PSH:          psh,
    }, nil
}

// Helper
func clamp(val, min, max float64) float64 {
    if val < min { return min }
    if val > max { return max }
    return val
}

func computeAverage(values []float64) (avg float64, count int) {
    for _, v := range values {
        avg += v
        count++
    }
    if count > 0 {
        avg /= float64(count)
    }
    return
}
```

---

## 11. Ringkasan Keputusan Desain

| Keputusan | Pilihan | Alasan |
|-----------|---------|--------|
| ΔWF di cold start | Dipakai, baseline synthetic | Koreksi cuaca sejak hari pertama tanpa data user |
| ΔWF di mode terkalibrasi | Dipakai, baseline site | Residual correction — η tidak cukup untuk hari ekstrem |
| ΔWF di mode transisi | Dipakai, baseline blended | Konsistensi rumus, hanya baseline yang berubah |
| Baseline cold start | Open-Meteo Historical 30 hari | Tersedia tanpa data user, konsisten dengan mode terkalibrasi |
| Transisi cold → calibrated | Blend linear (weighted avg) | Mencegah lompatan prediksi saat switch mode |
| N threshold transisi | 14 hari | Cukup data untuk baseline site yang representatif |
| Clamp ΔWF | 0.5 – 1.5 | Mencegah koreksi ekstrem di edge case cuaca sangat tidak biasa |
| Fallback jika baseline ≥ 95% | WF absolut | Hindari division by zero / nilai tidak stabil |
| Cache TTL synthetic | 7 hari | Data historis stabil, tidak perlu refresh harian |
| Cache TTL site | Setiap ada actual baru | Site baseline harus selalu up-to-date |
