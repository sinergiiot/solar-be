import { useState } from "react";
import { FiActivity, FiBarChart2, FiClock, FiCopy, FiCpu, FiMapPin, FiRefreshCw, FiSettings, FiSun, FiZap } from "react-icons/fi";

// IntegrationSection renders device registration form, API key display, and device list.
export default function IntegrationSection({
  profiles,
  deviceForm,
  setDeviceForm,
  handleCreateDevice,
  isSavingDevice,
  latestDeviceKey,
  copyLatestDeviceKey,
  isLoadingDevices,
  devices,
  editingDeviceID,
  editingDeviceIsActive,
  setEditingDeviceIsActive,
  handleEditDevice,
  handleDeleteDevice,
  handleCancelDeviceEdit,
  deletingDeviceID,
  rotatingDeviceID,
  handleRotateDeviceKey,
}) {
  const [activeInsightTab, setActiveInsightTab] = useState("overview");

  return (
    <>
      <section className='panel panel-form panel-wide'>
        <div className='panel-heading'>
          <span className='panel-kicker'>IoT</span>
          <h2>Integrasi Device Lapangan</h2>
        </div>

        <div className='integration-manage-grid'>
          <form className='integration-manage-card stack' onSubmit={handleCreateDevice}>
            <h3>{editingDeviceID ? "Edit Device" : "Daftarkan Device"}</h3>
            <label>
              <span>Nama device</span>
              <input value={deviceForm.name} onChange={(event) => setDeviceForm({ ...deviceForm, name: event.target.value })} placeholder='Inverter Atap Timur' required />
            </label>
            <label>
              <span>Device ID lapangan</span>
              <input value={deviceForm.external_id} onChange={(event) => setDeviceForm({ ...deviceForm, external_id: event.target.value })} placeholder='plant-A-01' required />
            </label>
            <label>
              <span>Solar profile ID (opsional)</span>
              <select value={deviceForm.solar_profile_id} onChange={(event) => setDeviceForm({ ...deviceForm, solar_profile_id: event.target.value })}>
                <option value=''>Auto dari profile terbaru</option>
                {profiles.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.site_name} ({p.id.slice(0, 8)})
                  </option>
                ))}
              </select>
            </label>
            {editingDeviceID && (
              <label>
                <span>Status device</span>
                <select value={editingDeviceIsActive ? "active" : "inactive"} onChange={(event) => setEditingDeviceIsActive(event.target.value === "active")}>
                  <option value='active'>Active</option>
                  <option value='inactive'>Inactive</option>
                </select>
              </label>
            )}
            <button className='primary-button' disabled={isSavingDevice} type='submit'>
              {isSavingDevice ? "Menyimpan..." : editingDeviceID ? "Update Device" : "Buat Device"}
            </button>
            {editingDeviceID && (
              <button className='secondary-button' type='button' onClick={handleCancelDeviceEdit}>
                Batal Edit
              </button>
            )}
          </form>

          <div className='integration-manage-card'>
            <h3>API Key Terbaru</h3>
            <p className='integration-note'>Key hanya tampil sekali saat buat/rotate. Simpan ke device Anda.</p>
            <div className='key-box'>{latestDeviceKey || "Belum ada key terbaru"}</div>
            <button className='secondary-button' type='button' onClick={copyLatestDeviceKey} disabled={!latestDeviceKey}>
              <FiCopy /> Copy API Key
            </button>
          </div>
        </div>

        <div className='integration-tabs'>
          <div className='integration-tab-list'>
            <button className={`integration-tab-button ${activeInsightTab === "overview" ? "active" : ""}`} type='button' onClick={() => setActiveInsightTab("overview")}>
              Overview
            </button>
            <button className={`integration-tab-button ${activeInsightTab === "architecture" ? "active" : ""}`} type='button' onClick={() => setActiveInsightTab("architecture")}>
              Arsitektur
            </button>
            <button className={`integration-tab-button ${activeInsightTab === "payload" ? "active" : ""}`} type='button' onClick={() => setActiveInsightTab("payload")}>
              Payload
            </button>
            <button className={`integration-tab-button ${activeInsightTab === "checklist" ? "active" : ""}`} type='button' onClick={() => setActiveInsightTab("checklist")}>
              Checklist
            </button>
          </div>

          {activeInsightTab === "overview" && (
            <div className='integration-grid'>
              <article className='integration-card'>
                <div className='integration-icon'>
                  <FiCpu />
                </div>
                <h3>Mode Pengiriman</h3>
                <p>Direkomendasikan device kirim data per 12 jam. Data disimpan per bucket 12 jam agar hemat storage dan ringan agregasi.</p>
                <span className='status-pill status-pill-active'>Recommended</span>
              </article>

              <article className='integration-card'>
                <div className='integration-icon'>
                  <FiClock />
                </div>
                <h3>Agregasi Harian</h3>
                <p>Data ingestion akan diproses ke metrik harian untuk evaluasi akurasi forecast dan adaptive efficiency.</p>
                <span className='status-pill'>Daily Pipeline</span>
              </article>

              <article className='integration-card'>
                <div className='integration-icon'>
                  <FiActivity />
                </div>
                <h3>Reliability</h3>
                <p>Gunakan buffering lokal di device saat offline lalu kirim ulang batch data ketika jaringan pulih.</p>
                <span className='status-pill'>Edge Safe</span>
              </article>
            </div>
          )}

          {activeInsightTab === "architecture" && (
            <div className='integration-flow'>
              <h3>Arsitektur MVP</h3>
              <div className='flow-row'>
                <div className='flow-item'>
                  <FiCpu />
                  Device
                </div>
                <div className='flow-arrow'>→</div>
                <div className='flow-item'>
                  <FiZap />
                  API Ingestion
                </div>
                <div className='flow-arrow'>→</div>
                <div className='flow-item'>
                  <FiBarChart2 />
                  Aggregation
                </div>
                <div className='flow-arrow'>→</div>
                <div className='flow-item'>
                  <FiSun />
                  Forecast Accuracy
                </div>
              </div>
            </div>
          )}

          {activeInsightTab === "payload" && (
            <div className='integration-spec'>
              <h3>Payload Contoh</h3>
              <pre>{`{
  "device_id": "plant-A-01",
  "timestamp": "2026-03-19T08:15:00Z",
  "energy_kwh": 4.32,
  "power_w": 1850,
  "lat": -6.2,
  "lng": 106.8
}`}</pre>
              <p className='integration-note'>Endpoint yang disarankan: POST /ingest/telemetry dengan API key per-device.</p>
              <p className='integration-note'>Header wajib: X-Device-Key: &lt;api_key_device&gt;</p>
            </div>
          )}

          {activeInsightTab === "checklist" && (
            <div className='integration-checklist'>
              <h3>Checklist Integrasi</h3>
              <ul>
                <li>
                  <FiMapPin /> Device terdaftar ke user
                </li>
                <li>
                  <FiSettings /> API key unik per device
                </li>
                <li>
                  <FiClock /> Idempotency device_id + timestamp
                </li>
                <li>
                  <FiActivity /> Retry saat jaringan kembali normal
                </li>
              </ul>
            </div>
          )}
        </div>
      </section>

      <section className='panel panel-data panel-wide'>
        <div className='panel-heading'>
          <div className='integration-devices-heading'>
            <div>
              <span className='panel-kicker'>Data</span>
              <h2>Daftar Device Terdaftar</h2>
            </div>
            <span className='status-pill status-pill-active'>{devices.length} device</span>
          </div>
        </div>
        {isLoadingDevices ? (
          <div className='empty-state'>Memuat device...</div>
        ) : devices.length === 0 ? (
          <div className='empty-state'>Belum ada device terdaftar.</div>
        ) : (
          <div className='device-list integration-device-list'>
            {devices.map((device) => (
              <article key={device.id} className='device-row integration-device-row'>
                <span className={`status-pill integration-device-status ${device.is_active ? "status-pill-active" : ""}`}>{device.is_active ? "Active" : "Inactive"}</span>
                <div className='integration-device-content'>
                  <strong>{device.name}</strong>
                  <p>ID: {device.external_id}</p>
                  <p>Solar profile: {device.solar_profile_id ? profiles.find((p) => p.id === device.solar_profile_id)?.site_name || device.solar_profile_id : "(belum di-link)"}</p>
                  <p>Key prefix: {device.api_key_prefix}</p>
                </div>
                <div className='device-row-actions'>
                  <button className='secondary-button' type='button' onClick={() => handleEditDevice(device)}>
                    Edit
                  </button>
                  <button className='secondary-button' type='button' disabled={rotatingDeviceID === device.id} onClick={() => handleRotateDeviceKey(device.id)}>
                    <FiRefreshCw /> {rotatingDeviceID === device.id ? "Rotating..." : "Rotate Key"}
                  </button>
                  <button className='secondary-button danger-button' type='button' disabled={deletingDeviceID === device.id} onClick={() => handleDeleteDevice(device.id)}>
                    {deletingDeviceID === device.id ? "Menghapus..." : "Delete"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </>
  );
}
