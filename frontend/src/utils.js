// formatIDR renders number to compact Indonesian Rupiah currency string.
export function formatIDR(value) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(value);
}

// formatDateID renders date to Indonesian locale format used in UI.
export function formatDateID(dateValue) {
  return new Date(dateValue).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}

// getWeatherRisk maps weather factor to risk badge metadata.
export function getWeatherRisk(weatherFactor) {
  if (weatherFactor === null || weatherFactor === undefined) {
    return { label: "--", tone: "neutral" };
  }
  const factor = Number(weatherFactor);
  if (factor >= 0.9) return { label: "Risiko Cuaca Rendah", tone: "low" };
  if (factor >= 0.7) return { label: "Risiko Cuaca Sedang", tone: "medium" };
  return { label: "Risiko Cuaca Tinggi", tone: "high" };
}

// getHistoryRowKey creates a stable key for merged history rows.
export function getHistoryRowKey(row) {
  return `${row.date}__${row.solar_profile_id || "none"}`;
}

// getTodayLocalDate returns YYYY-MM-DD based on local calendar date.
export function getTodayLocalDate() {
  const date = new Date();
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

// getDateDaysAgo returns YYYY-MM-DD for N days before today in local timezone.
export function getDateDaysAgo(days) {
  const date = new Date();
  date.setDate(date.getDate() - days);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

// getHourlyDistribution estimates production share per time period.
export function getHourlyDistribution(dateInput, latitude, weatherFactor) {
  const date = dateInput ? new Date(dateInput) : new Date();
  const month = date.getUTCMonth() + 1;
  const latAbs = Math.abs(Number(latitude || 0));

  let morning = 0.26;
  let noon = 0.48;
  let afternoon = 0.26;

  // Seasonal bias for tropical regions: dry months shift a bit to noon.
  if (month >= 4 && month <= 9) {
    morning -= 0.02;
    noon += 0.04;
    afternoon -= 0.02;
  } else {
    morning += 0.02;
    noon -= 0.04;
    afternoon += 0.02;
  }

  // Higher latitude tends to concentrate production around noon.
  if (latAbs >= 10) {
    morning -= 0.01;
    noon += 0.02;
    afternoon -= 0.01;
  }

  // Cloudier conditions push useful generation toward midday windows.
  if (Number(weatherFactor) <= 0.6) {
    morning -= 0.02;
    noon += 0.04;
    afternoon -= 0.02;
  } else if (Number(weatherFactor) <= 0.8) {
    morning -= 0.01;
    noon += 0.02;
    afternoon -= 0.01;
  }

  const sum = morning + noon + afternoon;
  const safeSum = sum > 0 ? sum : 1;

  return [
    { label: "Pagi", share: morning / safeSum },
    { label: "Siang", share: noon / safeSum },
    { label: "Sore", share: afternoon / safeSum },
  ];
}
