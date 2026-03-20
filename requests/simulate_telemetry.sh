#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
DEVICE_KEY="${DEVICE_KEY:-}"
DEVICE_ID="${DEVICE_ID:-plant-A-01}"
POINTS="${POINTS:-6}"
INTERVAL_MINUTES="${INTERVAL_MINUTES:-720}"
START_ENERGY_KWH="${START_ENERGY_KWH:-0.30}"

if [[ -z "$DEVICE_KEY" ]]; then
  echo "DEVICE_KEY is required"
  echo "Example: DEVICE_KEY=dvc_xxx DEVICE_ID=plant-A-01 ./requests/simulate_telemetry.sh"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for parsing response. Please install jq first."
  exit 1
fi

echo "Sending $POINTS telemetry points to $BASE_URL for device $DEVICE_ID"

for ((i=0; i<POINTS; i++)); do
  minutes_back=$(( (POINTS - 1 - i) * INTERVAL_MINUTES ))
  timestamp=$(date -u -v-"${minutes_back}"M +"%Y-%m-%dT%H:%M:%SZ")

  # Increment sample energy gradually to mimic production curve.
  energy=$(awk -v base="$START_ENERGY_KWH" -v idx="$i" 'BEGIN { printf "%.3f", (base + (idx * 0.02)) }')
  power=$(awk -v idx="$i" 'BEGIN { printf "%.0f", (1300 + (idx * 75)) }')

  payload=$(cat <<JSON
{
  "device_id": "$DEVICE_ID",
  "timestamp": "$timestamp",
  "energy_kwh": $energy,
  "power_w": $power,
  "lat": -6.2000,
  "lng": 106.8166
}
JSON
)

  response=$(curl -sS -X POST "$BASE_URL/ingest/telemetry" \
    -H "X-Device-Key: $DEVICE_KEY" \
    -H "Content-Type: application/json" \
    -d "$payload")

  accepted=$(echo "$response" | jq -r '.accepted // false')
  duplicate=$(echo "$response" | jq -r '.duplicate // false')
  actual_date=$(echo "$response" | jq -r '.actual_date // "-"')
  message=$(echo "$response" | jq -r '.message // "-"')

  echo "[$((i+1))/$POINTS] $timestamp -> accepted=$accepted duplicate=$duplicate date=$actual_date msg=$message"
done

echo "Done. Check dashboard history and summary in frontend."
