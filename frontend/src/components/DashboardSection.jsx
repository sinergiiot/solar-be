import { useEffect, useMemo, useState } from "react";
import { FiCloud, FiCloudRain, FiCpu, FiEdit3, FiSun, FiTrendingDown, FiTrendingUp } from "react-icons/fi";
import { formatDateID, formatIDR } from "../utils";

// DashboardSection renders heartbeat, summary, and latest 7-day snapshot tables.
export default function DashboardSection({ heartbeatSummary, summary, dashboardForecastHistory, dashboardActualHistory, todayForecast, activeProfile }) {
  const forecastByDate = new Map();
  dashboardForecastHistory.slice(0, 7).forEach((item) => {
    forecastByDate.set(item.date, Number(item.predicted_kwh || 0));
  });

  const actualByDate = new Map();
  dashboardActualHistory.slice(0, 7).forEach((item) => {
    actualByDate.set(item.date, Number(item.actual_kwh || 0));
  });

  const chartDates = Array.from(new Set([...forecastByDate.keys(), ...actualByDate.keys()])).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());

  const trendRows = chartDates.map((date) => ({
    date,
    forecast: forecastByDate.get(date) ?? null,
    actual: actualByDate.get(date) ?? null,
  }));

  const trendMax = Math.max(1, ...trendRows.map((row) => row.forecast || 0), ...trendRows.map((row) => row.actual || 0));

  const sourceStats = dashboardActualHistory.slice(0, 7).reduce(
    (acc, item) => {
      if (String(item.source).toLowerCase() === "iot") {
        acc.iot += 1;
      } else {
        acc.manual += 1;
      }
      return acc;
    },
    { manual: 0, iot: 0 },
  );
  const sourceTotal = sourceStats.manual + sourceStats.iot;
  const iotPercent = sourceTotal > 0 ? (sourceStats.iot / sourceTotal) * 100 : 0;
  const accuracyValue = summary ? Number(summary.accuracy_percent) : null;
  const accuracyClass = accuracyValue === null ? "" : accuracyValue >= 90 ? "metric-value-good" : accuracyValue >= 75 ? "metric-value-warn" : "metric-value-bad";
  const latestTrend = [...trendRows].reverse().find((row) => row.forecast !== null || row.actual !== null) || null;
  const latestGapKwh = latestTrend && latestTrend.forecast !== null && latestTrend.actual !== null ? latestTrend.forecast - latestTrend.actual : null;
  const latestGapPct = latestTrend && latestTrend.actual && latestGapKwh !== null ? (latestGapKwh / latestTrend.actual) * 100 : null;
  const latestGapTone = latestGapPct === null ? "neutral" : Math.abs(latestGapPct) <= 10 ? "good" : Math.abs(latestGapPct) <= 20 ? "warn" : "bad";

  const [todayHourlyWeather, setTodayHourlyWeather] = useState([]);
  const [isLoadingHourlyWeather, setIsLoadingHourlyWeather] = useState(false);

  const latitude = Number(activeProfile?.lat);
  const longitude = Number(activeProfile?.lng);
  const hasLocation = Number.isFinite(latitude) && Number.isFinite(longitude);

  const energyEstimate = Number(todayForecast?.predicted_kwh || 0);
  const weatherFactor = Number(todayForecast?.weather_factor || 0);
  const costEstimate = energyEstimate * 1444;

  const conditionLabel = useMemo(() => {
    if (weatherFactor >= 0.9) return "cerah";
    if (weatherFactor >= 0.75) return "berawan";
    if (weatherFactor >= 0.6) return "mendung";
    return "mendung tebal";
  }, [weatherFactor]);

  const conditionImpact = useMemo(() => {
    if (weatherFactor >= 0.9) return "dampak ke produksi rendah, panel berpotensi menghasilkan energi optimal";
    if (weatherFactor >= 0.75) return "dampak ke produksi ringan, output masih cukup baik";
    if (weatherFactor >= 0.6) return "dampak ke produksi sedang, output cenderung turun dibanding hari cerah";
    return "dampak ke produksi tinggi, output berpotensi turun signifikan";
  }, [weatherFactor]);

  useEffect(() => {
    if (!hasLocation || !todayForecast?.date) {
      setTodayHourlyWeather([]);
      return;
    }

    let ignore = false;

    async function loadHourlyWeather() {
      setIsLoadingHourlyWeather(true);
      try {
        const query = new URLSearchParams({
          latitude: String(latitude),
          longitude: String(longitude),
          hourly: "weather_code,cloud_cover,temperature_2m,shortwave_radiation",
          start_date: todayForecast.date,
          end_date: todayForecast.date,
          timezone: "auto",
        });

        const response = await fetch(`https://api.open-meteo.com/v1/forecast?${query.toString()}`);
        if (!response.ok) {
          throw new Error("failed fetch hourly weather");
        }

        const payload = await response.json();
        const times = payload?.hourly?.time || [];
        const weatherCodes = payload?.hourly?.weather_code || [];
        const clouds = payload?.hourly?.cloud_cover || [];
        const temperatures = payload?.hourly?.temperature_2m || [];
        const radiation = payload?.hourly?.shortwave_radiation || [];

        const rows = times
          .map((timestamp, index) => {
            const hour = new Date(timestamp).getHours();
            return {
              hour,
              label: `${String(hour).padStart(2, "0")}:00`,
              weatherCode: Number(weatherCodes[index] ?? 0),
              cloudCover: Number(clouds[index] ?? 0),
              temperature: Number(temperatures[index] ?? 0),
              radiation: Number(radiation[index] ?? 0),
            };
          })
          .filter((row) => row.hour >= 6 && row.hour <= 15)
          .sort((a, b) => a.hour - b.hour);

        if (!ignore) {
          setTodayHourlyWeather(rows);
        }
      } catch {
        if (!ignore) {
          setTodayHourlyWeather([]);
        }
      } finally {
        if (!ignore) {
          setIsLoadingHourlyWeather(false);
        }
      }
    }

    loadHourlyWeather();
    return () => {
      ignore = true;
    };
  }, [hasLocation, latitude, longitude, todayForecast?.date]);

  const renderWeatherCode = (weatherCode) => {
    if ([0].includes(weatherCode)) return { label: "Cerah", icon: <FiSun /> };
    if ([1, 2, 3].includes(weatherCode)) return { label: "Berawan", icon: <FiCloud /> };
    if ([45, 48].includes(weatherCode)) return { label: "Berkabut", icon: <FiCloud /> };
    if ([51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82].includes(weatherCode)) return { label: "Hujan", icon: <FiCloudRain /> };
    return { label: "Variatif", icon: <FiCloud /> };
  };

  const renderMetricHelp = (id, text) => (
    <span className='metric-help-wrap'>
      <button type='button' className='metric-help-button' aria-label={text} aria-describedby={id}>
        i
      </button>
      <span id={id} role='tooltip' className='metric-help-tooltip'>
        {text}
      </span>
    </span>
  );

  return (
    <>
      <section className='panel panel-wide device-status-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Heartbeat</span>
          <h2>Status Device Lapangan</h2>
        </div>
        <div className='device-status-grid'>
          <div className='summary-card'>
            <span className='metric-label'>Field Device</span>
            <strong className='metric-value'>{heartbeatSummary?.has_devices ? "Ada" : "Belum Ada"}</strong>
            <span className='metric-sublabel'>{heartbeatSummary?.total_devices || 0} device terdaftar</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label'>Connected</span>
            <strong className='metric-value'>
              {heartbeatSummary?.connected_devices || 0}/{heartbeatSummary?.active_devices || 0}
            </strong>
            <span className='metric-sublabel'>aktif dalam 24 jam</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label'>Last Heartbeat</span>
            <strong className='metric-value device-heartbeat'>{heartbeatSummary?.latest_seen_at ? new Date(heartbeatSummary.latest_seen_at).toLocaleString() : "Belum ada data"}</strong>
            <span className='metric-sublabel'>update terakhir dari device</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label'>Mode Ingest</span>
            <strong className='metric-value'>12 Jam / Bucket</strong>
            <span className='metric-sublabel'>hemat storage dan ringan agregasi</span>
          </div>
        </div>
      </section>

      <section className='panel panel-wide summary-panel today-condition-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Today's Condition</span>
          <h2>Kondisi Hari Ini (update setelah forecast 06:00)</h2>
        </div>
        {todayForecast && hasLocation ? (
          <>
            <div className='today-condition-layout'>
              <article className='today-condition-story'>
                <span className='today-condition-badge'>Forecast update 06:00</span>
                <p className='dashboard-explainer'>
                  Hari ini berdasarkan ramalan cuaca koordinat{" "}
                  <strong>
                    {latitude.toFixed(4)}, {longitude.toFixed(4)}
                  </strong>
                  , diprediksi <strong>{conditionLabel}</strong>, <strong>{conditionImpact}</strong>, dan estimasi produksi energi harian Anda sekitar <strong>{energyEstimate.toFixed(2)} kWh</strong> dengan potensi penghematan{" "}
                  <strong>{formatIDR(costEstimate)}</strong>.
                </p>
              </article>

              <div className='today-condition-meta'>
                <div className='summary-card'>
                  <span className='metric-label'>Tanggal Forecast</span>
                  <strong className='metric-value'>{todayForecast.date}</strong>
                  <span className='metric-sublabel'>sinkron dari hasil forecast terbaru</span>
                </div>
                <div className='summary-card'>
                  <span className='metric-label'>Estimasi Produksi</span>
                  <strong className='metric-value'>{energyEstimate.toFixed(2)} kWh</strong>
                  <span className='metric-sublabel'>estimasi energi hari ini</span>
                </div>
                <div className='summary-card'>
                  <span className='metric-label'>Weather Factor</span>
                  <strong className='metric-value'>{weatherFactor.toFixed(2)}</strong>
                  <span className='metric-sublabel'>indikasi dampak cuaca ke produksi</span>
                </div>
                <div className='summary-card'>
                  <span className='metric-label'>Estimasi Penghematan</span>
                  <strong className='metric-value'>{formatIDR(costEstimate)}</strong>
                  <span className='metric-sublabel'>asumsi tarif Rp 1.444 / kWh</span>
                </div>
              </div>
            </div>

            <h3 className='today-hourly-title'>Ramalan Cuaca Per Jam (06:00-15:00)</h3>
            {isLoadingHourlyWeather ? (
              <div className='empty-state'>Memuat ramalan cuaca per jam...</div>
            ) : todayHourlyWeather.length > 0 ? (
              <div className='today-hourly-grid'>
                {todayHourlyWeather.map((item) => {
                  const weatherVisual = renderWeatherCode(item.weatherCode);
                  return (
                    <article key={item.label} className='today-hourly-card'>
                      <div className='today-hourly-head'>
                        <span>{item.label}</span>
                        <span className='today-hourly-icon'>{weatherVisual.icon}</span>
                      </div>
                      <strong>{weatherVisual.label}</strong>
                      <p>Awan: {item.cloudCover.toFixed(0)}%</p>
                      <p>Suhu: {item.temperature.toFixed(1)}°C</p>
                      <p>Radiasi: {item.radiation.toFixed(0)} W/m²</p>
                    </article>
                  );
                })}
              </div>
            ) : (
              <div className='empty-state'>Data hourly cuaca belum tersedia untuk hari ini.</div>
            )}
          </>
        ) : (
          <div className='empty-state'>Tambahkan solar profile dengan koordinat valid lalu muat forecast hari ini untuk melihat Today's Condition.</div>
        )}
      </section>

      <section className='panel panel-wide summary-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Overview</span>
          <h2>Ringkasan Performa (90 hari)</h2>
        </div>
        <p className='dashboard-explainer'>Angka di bawah adalah ringkasan 90 hari terakhir. Fokus utama: akurasi mendekati 100% dan gap antara forecast vs actual makin kecil.</p>
        <div className='summary-grid'>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Total Forecast
              {renderMetricHelp("help-total-forecast", "Akumulasi total energi yang diprediksi sistem dalam 90 hari terakhir.")}
            </span>
            <strong className='metric-value'>{summary ? summary.total_forecasted_kwh.toFixed(1) : "--"} kWh</strong>
            <span className='metric-sublabel'>{summary ? summary.forecast_count : 0} hari</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Total Actual
              {renderMetricHelp("help-total-actual", "Akumulasi total energi actual yang tercatat (manual/IoT) dalam 90 hari terakhir.")}
            </span>
            <strong className='metric-value'>{summary ? summary.total_actual_kwh.toFixed(1) : "--"} kWh</strong>
            <span className='metric-sublabel'>{summary ? summary.actual_count : 0} hari</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Avg Forecast
              {renderMetricHelp("help-avg-forecast", "Rata-rata energi forecast per hari. Dipakai untuk melihat baseline prediksi harian.")}
            </span>
            <strong className='metric-value'>{summary ? summary.average_forecast_kwh.toFixed(2) : "--"} kWh</strong>
            <span className='metric-sublabel'>per hari</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Avg Actual
              {renderMetricHelp("help-avg-actual", "Rata-rata energi actual per hari dari data lapangan.")}
            </span>
            <strong className='metric-value'>{summary ? summary.average_actual_kwh.toFixed(2) : "--"} kWh</strong>
            <span className='metric-sublabel'>per hari</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Efficiency
              {renderMetricHelp("help-efficiency", "Faktor performa model/sistem saat menghitung forecast. Semakin stabil, prediksi biasanya makin konsisten.")}
            </span>
            <strong className='metric-value'>{summary ? (summary.current_efficiency * 100).toFixed(1) : "--"}%</strong>
            <span className='metric-sublabel'>adaptif</span>
          </div>
          <div className='summary-card'>
            <span className='metric-label metric-label-row'>
              Akurasi
              {renderMetricHelp("help-accuracy", "Persentase kedekatan forecast terhadap actual. Semakin mendekati 100%, semakin baik kualitas prediksi.")}
            </span>
            <strong className={`metric-value ${accuracyClass}`}>{summary ? summary.accuracy_percent.toFixed(1) : "--"}%</strong>
            <span className='metric-sublabel'>prediksi vs real</span>
          </div>
        </div>
        <p className='metric-tap-hint'>Tip mobile: tap ikon i pada setiap metrik untuk melihat penjelasan singkat.</p>
      </section>

      <section className='panel panel-wide summary-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Highlights</span>
          <h2>Highlight Cepat</h2>
        </div>
        <div className='dashboard-highlights-grid'>
          <article className='dashboard-highlight-card'>
            <div className='dashboard-highlight-top'>
              <span>Akurasi Model</span>
              {accuracyValue !== null && (accuracyValue >= 85 ? <FiTrendingUp /> : <FiTrendingDown />)}
            </div>
            <strong className={`dashboard-highlight-value ${accuracyClass}`}>{accuracyValue !== null ? `${accuracyValue.toFixed(1)}%` : "--"}</strong>
            <p>Target ideal di atas 85% agar forecast mendekati data lapangan.</p>
          </article>

          <article className={`dashboard-highlight-card dashboard-highlight-${latestGapTone}`}>
            <div className='dashboard-highlight-top'>
              <span>Gap Forecast Terakhir</span>
              {latestGapKwh !== null && (latestGapKwh >= 0 ? <FiTrendingUp /> : <FiTrendingDown />)}
            </div>
            <strong className='dashboard-highlight-value'>{latestGapKwh !== null ? `${latestGapKwh >= 0 ? "+" : ""}${latestGapKwh.toFixed(2)} kWh` : "--"}</strong>
            <p>{latestGapPct !== null ? `${latestGapPct >= 0 ? "+" : ""}${latestGapPct.toFixed(1)}% vs actual terbaru` : "Belum ada pasangan data forecast/actual terbaru."}</p>
          </article>

          <article className='dashboard-highlight-card'>
            <div className='dashboard-highlight-top'>
              <span>Kualitas Sumber Data</span>
              <FiCpu />
            </div>
            <strong className='dashboard-highlight-value'>{sourceTotal > 0 ? `${iotPercent.toFixed(0)}% IoT` : "--"}</strong>
            <p>{sourceTotal > 0 ? `${sourceStats.iot} dari ${sourceTotal} hari memakai input IoT.` : "Belum ada data actual untuk menilai kualitas sumber."}</p>
          </article>
        </div>
      </section>

      <section className='panel panel-wide summary-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Insight</span>
          <h2>Chart Ringkas Dashboard</h2>
        </div>
        <p className='dashboard-explainer'>Grafik membantu membaca tren: jika batang Forecast sering jauh di atas/bawah Actual, model perlu kalibrasi atau data actual tambahan.</p>
        <div className='dashboard-chart-grid'>
          <article className='dashboard-chart-card'>
            <h3>Tren Forecast vs Actual (7 hari)</h3>
            {trendRows.length > 0 ? (
              <div className='dashboard-line-chart'>
                {trendRows.map((row) => (
                  <div key={row.date} className='dashboard-line-row'>
                    <span className='dashboard-line-label'>{formatDateID(row.date)}</span>
                    <div className='dashboard-line-track'>
                      {row.forecast !== null && <div className='dashboard-line-bar dashboard-line-forecast' style={{ width: `${(row.forecast / trendMax) * 100}%` }} />}
                      {row.actual !== null && <div className='dashboard-line-bar dashboard-line-actual' style={{ width: `${(row.actual / trendMax) * 100}%` }} />}
                    </div>
                    <span className='dashboard-line-value'>{`${row.forecast !== null ? row.forecast.toFixed(1) : "--"} / ${row.actual !== null ? row.actual.toFixed(1) : "--"}`}</span>
                  </div>
                ))}
                <div className='dashboard-line-legend'>
                  <span>
                    <i className='dashboard-dot dashboard-dot-forecast' /> Forecast
                  </span>
                  <span>
                    <i className='dashboard-dot dashboard-dot-actual' /> Actual
                  </span>
                </div>
              </div>
            ) : (
              <div className='empty-state'>Belum ada data untuk chart tren.</div>
            )}
          </article>

          <article className='dashboard-chart-card'>
            <h3>Komposisi Sumber Data Actual</h3>
            {sourceTotal > 0 ? (
              <div className='dashboard-source-wrap'>
                <div className='dashboard-source-ring' style={{ background: `conic-gradient(#176f46 0% ${iotPercent}%, #d6830f ${iotPercent}% 100%)` }}>
                  <div className='dashboard-source-center'>{sourceTotal}</div>
                </div>
                <div className='dashboard-source-list'>
                  <p>
                    <i className='dashboard-dot dashboard-dot-actual' /> IoT: {sourceStats.iot} hari
                  </p>
                  <p>
                    <i className='dashboard-dot dashboard-dot-forecast' /> Manual: {sourceStats.manual} hari
                  </p>
                </div>
              </div>
            ) : (
              <div className='empty-state'>Belum ada data actual untuk chart komposisi.</div>
            )}
          </article>
        </div>
      </section>

      <section className='panel panel-wide summary-panel history-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Data</span>
          <h2>Forecast & Actual Terbaru (7 hari)</h2>
        </div>
        <p className='dashboard-explainer'>Tabel ini adalah data mentah 7 hari terakhir. Gunakan untuk cek harian apakah prediksi sudah mendekati angka actual dari lapangan.</p>
        <div className='history-grid'>
          <div className='history-column'>
            <h3 className='history-title'>Forecast</h3>
            <div className='history-table'>
              {dashboardForecastHistory.length > 0 ? (
                <table>
                  <thead>
                    <tr>
                      <th>Tanggal</th>
                      <th>Prediksi (kWh)</th>
                      <th>Weather</th>
                    </tr>
                  </thead>
                  <tbody>
                    {dashboardForecastHistory.slice(0, 7).map((f, idx) => (
                      <tr key={idx}>
                        <td>{f.date}</td>
                        <td>{Number(f.predicted_kwh).toFixed(2)}</td>
                        <td>{Number(f.weather_factor).toFixed(2)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className='empty-state'>Belum ada data forecast</div>
              )}
            </div>
          </div>
          <div className='history-column'>
            <h3 className='history-title'>Actual</h3>
            <div className='history-table'>
              {dashboardActualHistory.length > 0 ? (
                <table>
                  <thead>
                    <tr>
                      <th>Tanggal</th>
                      <th>Actual (kWh)</th>
                      <th>Sumber</th>
                    </tr>
                  </thead>
                  <tbody>
                    {dashboardActualHistory.slice(0, 7).map((a, idx) => (
                      <tr key={idx}>
                        <td>{a.date}</td>
                        <td>{Number(a.actual_kwh).toFixed(2)}</td>
                        <td className='source-badge'>
                          <span className='source-icon'>{a.source === "manual" ? <FiEdit3 /> : <FiCpu />}</span>
                          {a.source}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className='empty-state'>Belum ada data actual</div>
              )}
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
