# ⚡ Solar Forecast SaaS — Copilot Dev Guide

## 🎯 Project Goal

Build a backend system that:

- Calculates **daily solar energy forecast (kWh)**
- Based on:
  - Solar panel profile
  - Location
  - Weather data

- Sends result via notification (email first)

---

# 🧱 Architecture Principles

- Use **Modular Monolith (Golang)**
- Keep everything **simple & readable**
- Avoid overengineering
- Prefer **small functions**
- Prefer **explicit naming**

---

# 📦 Project Structure

```
solar-forecast/
│
├── cmd/api/
│   └── main.go
│
├── internal/
│   ├── config/
│   ├── user/
│   ├── solar/
│   ├── weather/
│   ├── forecast/
│   ├── notification/
│   └── scheduler/
│
├── pkg/utils/
├── migrations/
├── .env
└── go.mod
```

---

# 🧠 Domain Rules

## Solar Profile

- capacity_kwp (required)
- tilt (optional)
- azimuth (optional)
- lat/lng (required)

## Forecast

- calculated per user per day
- stored in DB
- sent via notification

---

# ⚙️ Forecast Formula

```
Energy (kWh) = Capacity × PSH × Efficiency × WeatherFactor
```

### Default Values (MVP)

- PSH = 4.5
- Efficiency = 0.8

### Weather Factor (based on cloud cover)

| Cloud % | Factor |
| ------- | ------ |
| 0–20    | 1.0    |
| 20–50   | 0.8    |
| 50–80   | 0.6    |
| >80     | 0.4    |

---

# 🗄️ Database Schema

## users

- id (uuid)
- name
- email
- created_at

## solar_profiles

- id
- user_id
- capacity_kwp
- lat
- lng
- tilt
- azimuth

## weather_daily

- id
- date
- lat
- lng
- cloud_cover
- temperature
- raw_json

## forecasts

- id
- user_id
- date
- predicted_kwh
- weather_factor
- efficiency

---

# 🧩 Service Design Pattern

Each domain must contain:

```
<domain>/
├── model.go
├── repository.go
├── service.go
├── handler.go
```

---

# 🧠 Coding Rules (IMPORTANT FOR COPILOT)

## 1. Always write comment before function

Example:

```go
// Calculate solar forecast based on system capacity and weather
func CalculateForecast(...) {}
```

---

## 2. Use clear naming

Good:

- GetUserByID
- CreateForecast
- FetchWeather
- SendEmailNotification

Bad:

- DoStuff
- HandleData

---

## 3. Keep functions small

Bad:

```go
func ProcessEverything() {}
```

Good:

```go
GetWeather()
CalculateForecast()
SaveForecast()
SendNotification()
```

---

## 4. Use explicit structs

```go
type SolarProfile struct {
    CapacityKwp float64
    Lat         float64
    Lng         float64
}
```

---

# 🌦️ Weather Service Rules

- Fetch weather once per location per day
- Cache result in `weather_daily`
- Use cloud_cover as main factor

---

# ⚡ Forecast Service Flow

```text
Get Users
   ↓
Get Solar Profile
   ↓
Fetch Weather
   ↓
Calculate Forecast
   ↓
Save Forecast
   ↓
Send Notification
```

---

# ⏰ Scheduler Rules

- Run daily at 06:00 AM
- Generate forecast for all users

---

# 📡 API Endpoints

```
POST   /users
POST   /solar-profiles
GET    /forecast/today
```

---

# 📬 Notification Rules

MVP:

- Email only

Future:

- WhatsApp
- Push notification

---

# 🚀 MVP Scope (STRICT)

Include:

- User CRUD
- Solar profile CRUD
- Weather fetch
- Forecast calculation
- Daily scheduler
- Email notification

Exclude:

- AI/ML
- IoT integration
- Complex analytics
- Dashboard UI

---

# 🧠 Copilot Prompting Tips

When coding, always start with:

```go
// TODO: Fetch weather data from external API and map to internal struct
```

```go
// TODO: Calculate solar energy forecast using capacity, PSH, efficiency, and weather factor
```

```go
// TODO: Send daily forecast email to user
```

---

# 🧪 Example Core Function

```go
func CalculateForecast(capacity float64, cloudCover int) float64 {
    psh := 4.5
    efficiency := 0.8

    weatherFactor := getWeatherFactor(cloudCover)

    return capacity * psh * efficiency * weatherFactor
}
```

---

# 🔮 Future Extensions (DO NOT BUILD NOW)

- Real IoT data ingestion
- Forecast vs actual comparison
- AI-based efficiency tuning
- Battery optimization
- B2B dashboard

---

# 🧭 Philosophy

- Build fast
- Keep simple
- Validate with users
- Iterate later

---

# ✅ Definition of Done (MVP)

- User can register
- User adds solar profile
- System generates daily forecast
- User receives email with result

---

END OF GUIDE
