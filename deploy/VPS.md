# VPS Deployment Guide

Setup lengkap untuk deploy Solar Forecast ke 2 domain terpisah:

- **Backend API**: `be-forecast.thingsid.com` → repo `solar-be`
- **Frontend SPA**: `solar-forecast.thingsid.com` → repo `solar-fe`

Native deploy: binary Go + PostgreSQL + systemd + nginx.

## 1. Siapkan server

Target: Ubuntu 24.04 LTS.

```bash
sudo apt update
sudo apt install -y golang-go postgresql postgresql-contrib nginx git certbot python3-certbot-nginx
```

## 2. Buat user aplikasi

```bash
sudo useradd -r -s /bin/false -m -d /opt/solar-forecast solar
sudo mkdir -p /opt/solar-forecast
sudo chown -R solar:solar /opt/solar-forecast
```

## 3. Clone & build backend (solar-be)

```bash
cd /opt/solar-forecast
sudo -u solar git clone https://github.com/<your-org>/solar-be.git be
cd be
go build -o solar-forecast ./cmd/api
```

## 4. Setup PostgreSQL

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE solar_forecast;
CREATE USER solar_user WITH ENCRYPTED PASSWORD 'replace-this-password';
GRANT ALL PRIVILEGES ON DATABASE solar_forecast TO solar_user;
```

## 5. Setup .env backend

```bash
cd /opt/solar-forecast/be
cp .env.example .env
nano .env
```

Isi:

```env
DATABASE_URL=postgres://solar_user:replace-this-password@127.0.0.1:5432/solar_forecast?sslmode=disable
PORT=8080
FRONTEND_ORIGIN=https://solar-forecast.thingsid.com
AUTH_JWT_SECRET=ganti-ini-rahasia
AUTH_TOKEN_EXPIRY_HOURS=24
AUTH_REFRESH_TOKEN_EXPIRY_DAYS=7
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com
WEATHER_BASE_URL=https://api.open-meteo.com/v1
```

## 6. Pasang systemd service (backend)

```bash
sudo cp /opt/solar-forecast/be/deploy/systemd/solar-forecast.service /etc/systemd/system/solar-forecast.service
sudo systemctl daemon-reload
sudo systemctl enable solar-forecast
sudo systemctl start solar-forecast
```

Verifikasi:

```bash
sudo systemctl status solar-forecast
journalctl -u solar-forecast -f
```

## 7. Clone & build frontend (solar-fe)

```bash
cd /opt/solar-forecast
sudo -u solar git clone https://github.com/<your-org>/solar-fe.git fe
cd fe
cp .env.example .env
```

Edit `.env`:

```env
VITE_API_BASE_URL=https://be-forecast.thingsid.com
```

Install dan build:

```bash
npm install
npm run build
```

Hasil build ada di `/opt/solar-forecast/fe/dist`.

## 8. Pasang nginx — backend

```bash
sudo cp /opt/solar-forecast/be/deploy/nginx/be-forecast.thingsid.com.conf \
    /etc/nginx/sites-available/be-forecast.thingsid.com
sudo ln -s /etc/nginx/sites-available/be-forecast.thingsid.com \
    /etc/nginx/sites-enabled/be-forecast.thingsid.com
```

## 9. Pasang nginx — frontend

```bash
sudo cp /opt/solar-forecast/be/deploy/nginx/solar-forecast.thingsid.com.conf \
    /etc/nginx/sites-available/solar-forecast.thingsid.com
sudo ln -s /etc/nginx/sites-available/solar-forecast.thingsid.com \
    /etc/nginx/sites-enabled/solar-forecast.thingsid.com
```

Pastikan path `root` di config sesuai:

```nginx
root /opt/solar-forecast/fe/dist;
```

Test dan reload:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 10. SSL dengan Let's Encrypt

```bash
sudo certbot --nginx -d be-forecast.thingsid.com
sudo certbot --nginx -d solar-forecast.thingsid.com
```

## 11. Uji endpoint

```bash
curl https://be-forecast.thingsid.com/health
curl https://solar-forecast.thingsid.com
```

## 12. Update aplikasi

### Backend

```bash
cd /opt/solar-forecast/be
sudo -u solar git pull
go build -o solar-forecast ./cmd/api
sudo systemctl restart solar-forecast
```

### Frontend

```bash
cd /opt/solar-forecast/fe
sudo -u solar git pull
npm install
npm run build
```

Frontend langsung aktif setelah build — nginx serving static files, tidak perlu restart.

## Catatan operasional

- Scheduler berjalan setiap hari jam `06:00 UTC`.
- Migrasi dijalankan otomatis saat service start.
- Pastikan `FRONTEND_ORIGIN` backend sesuai dengan `https://solar-forecast.thingsid.com` agar CORS tidak block.
- Pastikan kredensial SMTP valid agar email notification berjalan.
