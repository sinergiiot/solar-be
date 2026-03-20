# VPS Deployment Guide

Panduan ini memakai pendekatan native deploy: binary Go + PostgreSQL + systemd + nginx untuk backend, dan static build React untuk frontend.

## 1. Siapkan server

Contoh target: Ubuntu 24.04.

Install dependency:

```bash
sudo apt update
sudo apt install -y golang-go postgresql postgresql-contrib nginx git
```

## 2. Buat user aplikasi

```bash
sudo useradd -r -s /bin/false -m -d /opt/solar-forecast solar
sudo mkdir -p /opt/solar-forecast
sudo chown -R solar:solar /opt/solar-forecast
```

## 3. Clone source dan build binary

```bash
cd /opt/solar-forecast
git clone <repo-url> app
cd app
go build -o solar-forecast ./cmd/api
```

## 4. Setup PostgreSQL

```bash
sudo -u postgres psql
```

Lalu buat database dan user:

```sql
CREATE DATABASE solar_forecast;
CREATE USER solar_user WITH ENCRYPTED PASSWORD 'replace-this-password';
GRANT ALL PRIVILEGES ON DATABASE solar_forecast TO solar_user;
```

## 5. Setup environment file

```bash
cd /opt/solar-forecast/app
cp .env.example .env
```

Isi minimal:

```env
DATABASE_URL=postgres://solar_user:replace-this-password@127.0.0.1:5432/solar_forecast?sslmode=disable
PORT=8080
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your_email@gmail.com
WEATHER_BASE_URL=https://api.open-meteo.com/v1
```

## 6. Pasang systemd service

Copy file contoh:

```bash
sudo cp deploy/systemd/solar-forecast.service /etc/systemd/system/solar-forecast.service
sudo systemctl daemon-reload
sudo systemctl enable solar-forecast
sudo systemctl start solar-forecast
```

Verifikasi:

```bash
sudo systemctl status solar-forecast
journalctl -u solar-forecast -f
```

## 7. Build frontend React

```bash
cd /opt/solar-forecast/app/frontend
cp .env.example .env
npm install
npm run build
```

Jika domain API dan domain frontend sama, isi `.env` frontend dengan:

```env
VITE_API_BASE_URL=https://your-domain.com
```

## 8. Pasang nginx reverse proxy dan static frontend

Copy config contoh:

```bash
sudo cp deploy/nginx/solar-forecast.conf /etc/nginx/sites-available/solar-forecast
sudo ln -s /etc/nginx/sites-available/solar-forecast /etc/nginx/sites-enabled/solar-forecast
sudo nginx -t
sudo systemctl reload nginx
```

## 9. Uji endpoint

```bash
curl http://127.0.0.1:8080/health
curl https://your-domain.com/health
```

Frontend akan disajikan dari `frontend/dist` oleh nginx.

## 10. SSL dengan Let's Encrypt

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## 11. Update aplikasi

```bash
cd /opt/solar-forecast/app
git pull
go build -o solar-forecast ./cmd/api
sudo systemctl restart solar-forecast

cd frontend
npm install
npm run build
```

## Catatan operasional

- Scheduler berjalan setiap hari jam `06:00 UTC`.
- Migrasi dijalankan otomatis saat service start.
- Pastikan kredensial SMTP valid, jika tidak email notification akan gagal walau API tetap hidup.
- Service mencari folder `migrations` secara aman saat dijalankan dari binary di VPS.
- Samakan `FRONTEND_ORIGIN` backend dengan origin browser yang diizinkan.
