import { formatIDR, getWeatherRisk, getHourlyDistribution } from "../utils";

// ForecastSection renders today's forecast panel and actual daily input form.
export default function ForecastSection({
  profiles,
  profile,
  selectedForecastProfileID,
  setSelectedForecastProfileID,
  isLoadingForecast,
  isLoading,
  handleLoadForecast,
  electricityTariff,
  setElectricityTariff,
  forecast,
  actualHistory,
  handleRecordActual,
  actualForm,
  setActualForm,
  isSavingActual,
}) {
  const forecastDateKey = forecast ? new Date(forecast.date).toISOString().slice(0, 10) : "";
  const activeForecastProfileID = forecast?.solar_profile_id || selectedForecastProfileID || profile?.id || "";

  const filteredActuals = actualHistory
    .filter((item) => {
      const profileMatch = !activeForecastProfileID || item.solar_profile_id === activeForecastProfileID;
      if (!profileMatch) return false;
      if (!forecastDateKey) return true;
      return new Date(item.date).toISOString().slice(0, 10) <= forecastDateKey;
    })
    .slice()
    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());

  const sameDateActual = filteredActuals.find((item) => new Date(item.date).toISOString().slice(0, 10) === forecastDateKey);
  const latestBeforeActual = filteredActuals.find((item) => new Date(item.date).toISOString().slice(0, 10) !== forecastDateKey);
  const comparisonActual = sameDateActual || latestBeforeActual || null;
  const comparisonLabel = sameDateActual ? "actual di tanggal yang sama" : comparisonActual ? `actual ${new Date(comparisonActual.date).toLocaleDateString("id-ID", { day: "2-digit", month: "short" })}` : "actual referensi";

  const predictedToday = forecast ? Number(forecast.predicted_kwh) : 0;
  const deviationVsReference = forecast && comparisonActual && Number(comparisonActual.actual_kwh) > 0 ? ((predictedToday - Number(comparisonActual.actual_kwh)) / Number(comparisonActual.actual_kwh)) * 100 : null;

  const weatherRisk = !forecast ? { label: "--", tone: "neutral" } : getWeatherRisk(Number(forecast.weather_factor));
  const activeProfile = profiles.find((p) => p.id === activeForecastProfileID) || profile || null;
  const hourlyEstimate = forecast
    ? getHourlyDistribution(forecast.date, activeProfile?.lat, Number(forecast.weather_factor)).map((slot) => ({
        label: slot.label,
        share: slot.share,
        value: predictedToday * slot.share,
      }))
    : [];

  return (
    <>
      <section className='panel forecast-primary-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Generate</span>
          <h2>Forecast Hari Ini</h2>
        </div>
        <div className='filter-inline'>
          <label>
            <span>Solar Profile</span>
            <select value={selectedForecastProfileID} onChange={(event) => setSelectedForecastProfileID(event.target.value)}>
              <option value=''>Auto (profile terbaru)</option>
              {profiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.site_name}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className='forecast-actions'>
          <button className='primary-button' onClick={handleLoadForecast} disabled={isLoadingForecast || isLoading} type='button'>
            {isLoadingForecast ? "Menghitung forecast..." : "Ambil Forecast Hari Ini"}
          </button>
        </div>
        <div className='filter-inline'>
          <label>
            <span>Tarif listrik (IDR/kWh) untuk estimasi hemat biaya</span>
            <input type='number' min='0' step='1' value={electricityTariff} onChange={(event) => setElectricityTariff(Number(event.target.value || 0))} />
          </label>
        </div>
        <div className='forecast-card'>
          <div>
            <span className='metric-label'>Prediksi energi</span>
            <strong className='metric-highlight'>{forecast ? `${Number(forecast.predicted_kwh).toFixed(2)} kWh` : "--"}</strong>
          </div>
          <div>
            <span className='metric-label'>
              Weather factor
              <span className='metric-help' title='Faktor pengali berdasarkan kondisi cuaca (utama: cloud cover).'>
                ?
              </span>
            </span>
            <strong>{forecast ? Number(forecast.weather_factor).toFixed(2) : "--"}</strong>
          </div>
          <div>
            <span className='metric-label'>
              Efficiency
              <span className='metric-help' title='Efisiensi model sistem PLTS yang dipakai pada perhitungan prediksi.'>
                ?
              </span>
            </span>
            <strong>{forecast ? (Number(forecast.efficiency) * 100).toFixed(1) : "--"}%</strong>
          </div>
          <div>
            <span className='metric-label'>Tanggal forecast</span>
            <strong>{forecast ? new Date(forecast.date).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" }) : "--"}</strong>
          </div>
          <div>
            <span className='metric-label'>Solar profile aktif</span>
            <strong>{forecast && forecast.solar_profile_id ? profiles.find((p) => p.id === forecast.solar_profile_id)?.site_name || "Profile terpilih" : profile?.site_name || "Auto (profile terbaru)"}</strong>
          </div>
          <div>
            <span className='metric-label'>Estimasi hemat biaya</span>
            <strong>{forecast ? formatIDR(Number(forecast.predicted_kwh) * Number(electricityTariff || 0)) : "--"}</strong>
          </div>
          <div>
            <span className='metric-label'>
              Estimasi CO2 dihindari
              <span className='metric-help' title='Estimasi sederhana: prediksi kWh × 0.85 kgCO2 per kWh.'>
                ?
              </span>
            </span>
            <strong>{forecast ? `${(Number(forecast.predicted_kwh) * 0.85).toFixed(2)} kgCO2` : "--"}</strong>
          </div>
          <div>
            <span className='metric-label'>
              Deviasi vs actual referensi
              <span className='metric-help' title='Diprioritaskan actual pada tanggal yang sama; jika belum ada, pakai actual terbaru sebelum tanggal forecast.'>
                ?
              </span>
            </span>
            <strong>{deviationVsReference !== null ? `${deviationVsReference >= 0 ? "+" : ""}${deviationVsReference.toFixed(1)}%` : "--"}</strong>
            <span className='metric-note'>Referensi: {comparisonLabel}</span>
          </div>
          <div>
            <span className='metric-label'>Status risiko cuaca</span>
            <strong className={`forecast-risk-badge forecast-risk-${weatherRisk.tone}`}>{weatherRisk.label}</strong>
          </div>
        </div>
        <div className='forecast-hourly-card'>
          <p className='forecast-hourly-title'>Estimasi Produksi per Periode Waktu (Dinamis)</p>
          {hourlyEstimate.length > 0 ? (
            <div className='forecast-hourly-grid'>
              {hourlyEstimate.map((slot) => (
                <div key={slot.label} className='forecast-hourly-item'>
                  <span>{slot.label}</span>
                  <strong>{slot.value.toFixed(2)} kWh</strong>
                  <small>{(slot.share * 100).toFixed(0)}% porsi</small>
                  <div className='forecast-hourly-bar-track'>
                    <div className='forecast-hourly-bar-fill' style={{ width: `${(slot.value / Math.max(predictedToday, 1)) * 100}%` }} />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className='forecast-hourly-empty'>Ambil forecast dulu untuk menampilkan estimasi per periode.</p>
          )}
          <p className='forecast-hourly-note'>Model membagi energi berdasarkan musim, lintang profile, dan weather factor.</p>
        </div>
      </section>

      <section className='panel forecast-secondary-panel'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Input</span>
          <h2>Input Actual Harian</h2>
        </div>
        <form className='stack' onSubmit={handleRecordActual}>
          <label>
            <span>Tanggal actual</span>
            <input type='date' value={actualForm.date} onChange={(event) => setActualForm({ ...actualForm, date: event.target.value })} required />
          </label>
          <label>
            <span>Actual energi (kWh)</span>
            <input type='number' min='0' step='0.01' value={actualForm.actual_kwh} onChange={(event) => setActualForm({ ...actualForm, actual_kwh: event.target.value })} placeholder='11.60' required />
          </label>
          <label>
            <span>Sumber data</span>
            <select value={actualForm.source} onChange={(event) => setActualForm({ ...actualForm, source: event.target.value })}>
              <option value='manual'>Manual</option>
              <option value='iot'>IoT</option>
            </select>
          </label>
          <button className='primary-button' disabled={isSavingActual} type='submit'>
            {isSavingActual ? "Menyimpan..." : "Simpan Actual"}
          </button>
        </form>
      </section>
    </>
  );
}
