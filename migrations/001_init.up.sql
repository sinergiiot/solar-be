-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Solar profiles table
CREATE TABLE IF NOT EXISTS solar_profiles (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capacity_kwp FLOAT NOT NULL,
    lat          FLOAT NOT NULL,
    lng          FLOAT NOT NULL,
    tilt         FLOAT,
    azimuth      FLOAT,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Weather daily cache table
CREATE TABLE IF NOT EXISTS weather_daily (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date        DATE NOT NULL,
    lat         FLOAT NOT NULL,
    lng         FLOAT NOT NULL,
    cloud_cover INT NOT NULL,
    temperature FLOAT,
    raw_json    TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (date, lat, lng)
);

-- Forecasts table
CREATE TABLE IF NOT EXISTS forecasts (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date          DATE NOT NULL,
    predicted_kwh FLOAT NOT NULL,
    weather_factor FLOAT NOT NULL,
    efficiency    FLOAT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);
