# Solar Forecast SaaS

Platform sederhana untuk menghitung forecast energi panel surya harian. Repository ini sekarang berisi backend Go dan frontend React.

## Fitur MVP

- User CRUD sederhana
- Solar profile CRUD sederhana
- Fetch weather harian dari Open-Meteo
- Hitung forecast harian dengan rumus MVP
- Simpan cache weather dan forecast ke PostgreSQL
- Scheduler harian jam 06:00 UTC
- Email notification via SMTP
- Dashboard React untuk operasional user, profile, dan forecast

## Endpoint

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me` (auth)
- `POST /solar-profiles`
- `PUT /solar-profiles/{profileID}` (auth)
- `DELETE /solar-profiles/{profileID}` (auth)
- `GET /solar-profiles/me` (auth)
- `GET /solar-profiles` (auth)
- `GET /solar-profiles/{profileID}` (auth)
- `GET /forecast/today` (auth)
- `POST /forecast/actual`
- `GET /forecast/history` (auth)
- `GET /forecast/actuals/history` (auth)
- `GET /forecast/summary` (auth)
- `GET /devices` (auth)
- `POST /devices` (auth)
- `PUT /devices/{deviceID}` (auth)
- `DELETE /devices/{deviceID}` (auth)
- `POST /devices/{deviceID}/rotate-key` (auth)
- `POST /ingest/telemetry` (device api key)
- `GET /health`

## Menjalankan Lokal

1. Copy `.env.example` menjadi `.env`.
2. Siapkan PostgreSQL kosong.
3. Update `DATABASE_URL`, `FRONTEND_ORIGIN`, dan konfigurasi SMTP di `.env`.
4. Jalankan backend:

```bash
go run ./cmd/api
```

5. Jalankan frontend React:

```bash
cd frontend
npm install
npm run dev
```

Backend default berjalan di `:8080`, frontend di `:5173`.

## Trial Cepat

Kalau tujuanmu hanya ingin trial cepat:

1. Jalankan PostgreSQL lokal.
2. Isi `.env` backend dari [.env.example](.env.example).
3. Jalankan `make run-api`.
4. Jalankan `make run-frontend`.
5. Buka `http://localhost:5173`.
6. Buat user, isi solar profile, lalu klik ambil forecast.

Contoh request manual juga tersedia di [requests/trial.http](requests/trial.http).
Contoh request device tersedia di [requests/device.http](requests/device.http).

## Integrasi Device Lapangan

Flow singkat:

1. Login user, lalu buat device via `POST /devices`.
   Anda bisa isi `solar_profile_id` (opsional) untuk mengikat device ke titik pembangkit tertentu.
2. Simpan `api_key` yang dikembalikan (hanya ditampilkan sekali).
3. Device mengirim telemetry ke `POST /ingest/telemetry` dengan header `X-Device-Key`.
   Rekomendasi interval: `12 jam` per kirim untuk efisiensi storage.
4. Sistem simpan raw telemetry, hindari duplikasi, lalu agregasi ke `actual_daily` (`source=iot`).

Catatan arsitektur saat ini:

- Device sudah bisa lebih dari satu per user.
- Relasi `device -> solar_profile` sudah didukung.
- Solar profile sekarang bisa lebih dari satu per user (multi-site / multi-titik pembangkit).

Contoh jalankan simulator telemetry:

```bash
DEVICE_KEY=dvc_xxx DEVICE_ID=plant-A-01 make simulate-telemetry
```

Opsi env tambahan:

- `POINTS` (default `6`)
- `INTERVAL_MINUTES` (default `720` = 12 jam)
- `BASE_URL` (default `http://localhost:8080`)

## Frontend

Frontend React ada di [frontend/package.json](frontend/package.json). Konfigurasi API browser memakai variable `VITE_API_BASE_URL`, default ke `http://localhost:8080`.

Contoh file environment frontend:

```bash
cd frontend
cp .env.example .env
```

## Formula Forecast

`Energy (kWh) = Capacity × PSH × Efficiency`

- `PSH` diturunkan dari data cuaca harian Open-Meteo:
  - `PSH = shortwave_radiation_sum (MJ/m2/day) / 3.6`
- `Efficiency` berasal dari `users.forecast_efficiency` (default awal `0.8`, lalu adaptif lewat learning).
- `WeatherFactor` tetap dihitung untuk indikator risiko cuaca:
  - `weather_factor = 1 - cloud_cover/100`

## Endpoint Debug Internal

Endpoint audit perhitungan forecast tersedia untuk tim internal:

- `GET /forecast/debug/calculate` (auth + internal token)

Proteksi endpoint ini:

- Wajib login user (Bearer JWT).
- Wajib kirim header `X-Debug-Token`.
- Nilai header harus sama dengan env backend `FORECAST_DEBUG_TOKEN`.
- Jika `FORECAST_DEBUG_TOKEN` kosong, endpoint otomatis nonaktif.

Contoh query minimum:

- `user_id` (UUID, wajib)
- `date` (format YYYY-MM-DD, wajib)
- `profile_id` (UUID, opsional)

## Deploy ke VPS

Panduan deploy native tanpa Docker ada di [deploy/VPS.md](deploy/VPS.md).
