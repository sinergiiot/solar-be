import { FiCpu, FiEdit3 } from "react-icons/fi";
import { formatDateID, formatIDR, getHistoryRowKey } from "../utils";

// HistorySection renders filterable history tables and detail insight panel.
export default function HistorySection({
  profiles,
  historyProfileID,
  setHistoryProfileID,
  historyStartDate,
  setHistoryStartDate,
  historyEndDate,
  setHistoryEndDate,
  handleApplyHistoryFilter,
  isLoadingHistory,
  historyKpi,
  historyComparisonRows,
  selectedHistoryRow,
  setSelectedHistoryRowKey,
  forecastHistory,
  actualHistory,
  selectedHistoryProfile,
  selectedHistoryRisk,
  selectedHistoryHourly,
  electricityTariff,
}) {
  return (
    <>
      <section className='panel panel-form panel-wide'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Data</span>
          <h2>Forecast & Actual History</h2>
        </div>
        <form className='history-filter-form' onSubmit={handleApplyHistoryFilter}>
          <label>
            <span>Profile</span>
            <select value={historyProfileID} onChange={(event) => setHistoryProfileID(event.target.value)}>
              <option value=''>Semua Profile</option>
              {profiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.site_name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Start Date</span>
            <input type='date' value={historyStartDate} onChange={(event) => setHistoryStartDate(event.target.value)} />
          </label>
          <label>
            <span>End Date</span>
            <input type='date' value={historyEndDate} onChange={(event) => setHistoryEndDate(event.target.value)} />
          </label>
          <button className='secondary-button history-filter-button' type='submit' disabled={isLoadingHistory}>
            {isLoadingHistory ? "Memuat..." : "Terapkan Filter"}
          </button>
        </form>

        <div className='history-summary-grid'>
          <article className='history-summary-card'>
            <span>Hari dengan data lengkap</span>
            <strong>{historyKpi.pairedDays}</strong>
          </article>
          <article className='history-summary-card'>
            <span>Rata-rata akurasi</span>
            <strong>{historyKpi.avgAccuracy !== null ? `${historyKpi.avgAccuracy.toFixed(1)}%` : "--"}</strong>
          </article>
          <article className='history-summary-card'>
            <span>Rata-rata selisih absolut</span>
            <strong>{historyKpi.avgAbsDelta !== null ? `${historyKpi.avgAbsDelta.toFixed(2)} kWh` : "--"}</strong>
          </article>
          <article className='history-summary-card'>
            <span>Hari terbaik</span>
            <strong>{historyKpi.bestDay ? `${new Date(historyKpi.bestDay.date).toLocaleDateString("id-ID", { day: "2-digit", month: "short" })} (${(historyKpi.bestDay.accuracy || 0).toFixed(1)}%)` : "--"}</strong>
          </article>
        </div>

        <div className='history-comparison-wrap'>
          <h3 className='history-title'>Perbandingan Forecast vs Actual</h3>
          <div className='history-table history-comparison-table'>
            {historyComparisonRows.length > 0 ? (
              <table>
                <thead>
                  <tr>
                    <th>Tanggal</th>
                    <th>Profile</th>
                    <th>Forecast</th>
                    <th>Actual</th>
                    <th>Selisih</th>
                    <th>Akurasi</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {historyComparisonRows.map((row, idx) => {
                    const deltaTone = row.delta === null ? "neutral" : row.delta >= 0 ? "positive" : "negative";
                    const rowKey = getHistoryRowKey(row);
                    return (
                      <tr key={`${row.date}-${row.solar_profile_id || "none"}-${idx}`} className={selectedHistoryRow && getHistoryRowKey(selectedHistoryRow) === rowKey ? "history-row-active" : ""} onClick={() => setSelectedHistoryRowKey(rowKey)}>
                        <td>{formatDateID(row.date)}</td>
                        <td>{profiles.find((p) => p.id === row.solar_profile_id)?.site_name || "-"}</td>
                        <td>{row.predicted_kwh !== null ? row.predicted_kwh.toFixed(2) : "-"}</td>
                        <td>{row.actual_kwh !== null ? row.actual_kwh.toFixed(2) : "-"}</td>
                        <td>
                          {row.delta !== null ? (
                            <span className={`delta-chip delta-chip-${deltaTone}`}>
                              {`${row.delta >= 0 ? "+" : ""}${row.delta.toFixed(2)} kWh`}
                              {row.deltaPct !== null ? ` (${row.deltaPct >= 0 ? "+" : ""}${row.deltaPct.toFixed(1)}%)` : ""}
                            </span>
                          ) : (
                            "-"
                          )}
                        </td>
                        <td>{row.accuracy !== null ? `${row.accuracy.toFixed(1)}%` : "-"}</td>
                        <td>{row.predicted_kwh !== null && row.actual_kwh !== null ? <span className='match-status match-status-ready'>Lengkap</span> : <span className='match-status match-status-pending'>Belum lengkap</span>}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            ) : (
              <div className='empty-state'>Belum ada data perbandingan</div>
            )}
          </div>
        </div>

        <div className='history-grid'>
          <div className='history-column'>
            <h3 className='history-title'>Forecast</h3>
            <div className='history-table'>
              {forecastHistory.length > 0 ? (
                <table>
                  <thead>
                    <tr>
                      <th>Tanggal</th>
                      <th>Profile</th>
                      <th>Prediksi (kWh)</th>
                      <th>Weather</th>
                    </tr>
                  </thead>
                  <tbody>
                    {forecastHistory.map((f, idx) => (
                      <tr key={idx}>
                        <td>{f.date}</td>
                        <td>{profiles.find((p) => p.id === f.solar_profile_id)?.site_name || "-"}</td>
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
              {actualHistory.length > 0 ? (
                <table>
                  <thead>
                    <tr>
                      <th>Tanggal</th>
                      <th>Profile</th>
                      <th>Actual (kWh)</th>
                      <th>Sumber</th>
                    </tr>
                  </thead>
                  <tbody>
                    {actualHistory.map((a, idx) => (
                      <tr key={idx}>
                        <td>{a.date}</td>
                        <td>{profiles.find((p) => p.id === a.solar_profile_id)?.site_name || "-"}</td>
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

      <section className='panel panel-data panel-wide'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Details</span>
          <h2>Detail Insight History</h2>
        </div>
        {selectedHistoryRow && selectedHistoryRow.predicted_kwh !== null ? (
          <div className='history-detail-grid'>
            <div className='forecast-card history-detail-card'>
              <div>
                <span className='metric-label'>Prediksi energi</span>
                <strong className='metric-highlight'>{`${Number(selectedHistoryRow.predicted_kwh).toFixed(2)} kWh`}</strong>
              </div>
              <div>
                <span className='metric-label'>
                  Weather factor
                  <span className='metric-help' title='Faktor pengali berdasarkan kondisi cuaca (utama: cloud cover).'>
                    ?
                  </span>
                </span>
                <strong>{selectedHistoryRow.weather_factor !== null ? Number(selectedHistoryRow.weather_factor).toFixed(2) : "--"}</strong>
              </div>
              <div>
                <span className='metric-label'>
                  Efficiency
                  <span className='metric-help' title='Efisiensi model sistem PLTS saat forecast tanggal ini dihitung.'>
                    ?
                  </span>
                </span>
                <strong>{selectedHistoryRow.efficiency !== null ? `${(Number(selectedHistoryRow.efficiency) * 100).toFixed(1)}%` : "--"}</strong>
              </div>
              <div>
                <span className='metric-label'>Tanggal forecast</span>
                <strong>{formatDateID(selectedHistoryRow.date)}</strong>
              </div>
              <div>
                <span className='metric-label'>Solar profile aktif</span>
                <strong>{selectedHistoryProfile?.site_name || "-"}</strong>
              </div>
              <div>
                <span className='metric-label'>Estimasi hemat biaya</span>
                <strong>{formatIDR(Number(selectedHistoryRow.predicted_kwh) * Number(electricityTariff || 0))}</strong>
              </div>
              <div>
                <span className='metric-label'>
                  Estimasi CO2 dihindari
                  <span className='metric-help' title='Estimasi sederhana: prediksi kWh × 0.85 kgCO2 per kWh.'>
                    ?
                  </span>
                </span>
                <strong>{`${(Number(selectedHistoryRow.predicted_kwh) * 0.85).toFixed(2)} kgCO2`}</strong>
              </div>
              <div>
                <span className='metric-label'>
                  Deviasi vs actual referensi
                  <span className='metric-help' title='Menggunakan actual pada baris yang sama jika tersedia.'>
                    ?
                  </span>
                </span>
                <strong>{selectedHistoryRow.deltaPct !== null ? `${selectedHistoryRow.deltaPct >= 0 ? "+" : ""}${selectedHistoryRow.deltaPct.toFixed(1)}%` : "--"}</strong>
                <span className='metric-note'>Referensi: {selectedHistoryRow.actual_kwh !== null ? `actual ${formatDateID(selectedHistoryRow.date)}` : "actual referensi"}</span>
              </div>
              <div>
                <span className='metric-label'>Status risiko cuaca</span>
                <strong className={`forecast-risk-badge forecast-risk-${selectedHistoryRisk.tone}`}>{selectedHistoryRisk.label}</strong>
              </div>
            </div>

            <div className='forecast-hourly-card history-hourly-detail'>
              <p className='forecast-hourly-title'>Estimasi Produksi per Periode Waktu (Dinamis)</p>
              <div className='forecast-hourly-grid'>
                {selectedHistoryHourly.map((slot) => (
                  <div key={slot.label} className='forecast-hourly-item'>
                    <span>{slot.label}</span>
                    <strong>{slot.value.toFixed(2)} kWh</strong>
                    <small>{(slot.share * 100).toFixed(0)}% porsi</small>
                    <div className='forecast-hourly-bar-track'>
                      <div className='forecast-hourly-bar-fill' style={{ width: `${slot.share * 100}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : (
          <div className='empty-state'>Klik baris yang memiliki data forecast untuk melihat detail metrik.</div>
        )}
      </section>
    </>
  );
}
