import { FiActivity, FiArrowLeft, FiCpu, FiMapPin, FiShield, FiSun, FiTrendingUp, FiZap } from "react-icons/fi";

const TIME_PRESETS = [
  { id: "morning", label: "Pagi", psh: 3.2 },
  { id: "noon", label: "Siang", psh: 4.5 },
  { id: "afternoon", label: "Sore", psh: 3.8 },
];

// getPshByPreset maps simulation time preset to PSH assumption.
function getPshByPreset(timePreset) {
  const matched = TIME_PRESETS.find((preset) => preset.id === timePreset);
  return matched ? matched.psh : 4.5;
}

// getSimWeatherFactor converts cloud cover percentage to a weather factor.
function getSimWeatherFactor(cloudCover) {
  if (cloudCover <= 20) return 1.0;
  if (cloudCover <= 50) return 0.8;
  if (cloudCover <= 80) return 0.6;
  return 0.4;
}

// getCloudLabel returns a human-readable sky condition label.
function getCloudLabel(cloudCover) {
  if (cloudCover <= 20) return "Cerah ☀️";
  if (cloudCover <= 50) return "Berawan Sebagian ⛅";
  if (cloudCover <= 80) return "Mendung 🌥️";
  return "Sangat Mendung / Hujan 🌧️";
}

// AuthFlow renders landing page, register, and login screens for unauthenticated users.
export default function AuthFlow({
  authPage,
  setAuthPage,
  error,
  setError,
  feedback,
  setFeedback,
  feedbackFading,
  isAuthLoading,
  registerForm,
  setRegisterForm,
  handleRegister,
  loginForm,
  setLoginForm,
  handleLogin,
  simCapacity,
  setSimCapacity,
  simCloudCover,
  setSimCloudCover,
  simTimePreset,
  setSimTimePreset,
}) {
  // ── Register page ──────────────────────────────────────────────────────
  if (authPage === "register") {
    return (
      <div className='shell landing-shell auth-subpage'>
        <div className='backdrop backdrop-left' />
        <div className='backdrop backdrop-right' />
        <div className='auth-subpage-wrap'>
          <button
            className='auth-back-btn'
            type='button'
            onClick={() => {
              setError("");
              setFeedback("");
              setAuthPage("landing");
            }}>
            <FiArrowLeft /> Kembali
          </button>
          {(error || feedback) && (
            <div className={`banner ${error ? "banner-error" : "banner-success"}${feedbackFading && !error ? " banner-fading" : ""}`}>
              <span>{error || feedback}</span>
              <button className='banner-close' type='button' onClick={() => (error ? setError("") : setFeedback(""))} aria-label='Tutup'>
                ×
              </button>
            </div>
          )}
          <section className='panel auth-subpage-card'>
            <div className='panel-heading'>
              <p className='eyebrow'>Green Energy Intelligence</p>
              <h2>Buat Akun Gratis</h2>
              <p className='hero-copy' style={{ marginTop: 6 }}>
                Mulai kelola forecast PLTS Anda dalam hitungan menit.
              </p>
            </div>
            <form className='stack' onSubmit={handleRegister}>
              <label>
                <span>Nama</span>
                <input value={registerForm.name} onChange={(e) => setRegisterForm({ ...registerForm, name: e.target.value })} placeholder='Nama Anda / Nama Tim' required />
              </label>
              <label>
                <span>Email</span>
                <input type='email' value={registerForm.email} onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })} placeholder='tim-energy@company.com' required />
              </label>
              <label>
                <span>Password</span>
                <input type='password' value={registerForm.password} onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })} placeholder='Minimal 8 karakter' minLength={8} required />
              </label>
              <button className='primary-button' disabled={isAuthLoading} type='submit'>
                {isAuthLoading ? "Membuat akun..." : "Aktifkan Workspace"}
              </button>
            </form>
            <p className='auth-subpage-switch'>
              Sudah punya akun?{" "}
              <button
                type='button'
                className='link-btn'
                onClick={() => {
                  setError("");
                  setFeedback("");
                  setAuthPage("login");
                }}>
                Masuk di sini
              </button>
            </p>
          </section>
        </div>
      </div>
    );
  }

  // ── Login page ─────────────────────────────────────────────────────────
  if (authPage === "login") {
    return (
      <div className='shell landing-shell auth-subpage'>
        <div className='backdrop backdrop-left' />
        <div className='backdrop backdrop-right' />
        <div className='auth-subpage-wrap'>
          <button
            className='auth-back-btn'
            type='button'
            onClick={() => {
              setError("");
              setFeedback("");
              setAuthPage("landing");
            }}>
            <FiArrowLeft /> Kembali
          </button>
          {(error || feedback) && (
            <div className={`banner ${error ? "banner-error" : "banner-success"}${feedbackFading && !error ? " banner-fading" : ""}`}>
              <span>{error || feedback}</span>
              <button className='banner-close' type='button' onClick={() => (error ? setError("") : setFeedback(""))} aria-label='Tutup'>
                ×
              </button>
            </div>
          )}
          <section className='panel auth-subpage-card'>
            <div className='panel-heading'>
              <p className='eyebrow'>Welcome Back</p>
              <h2>Masuk ke Dashboard</h2>
              <p className='hero-copy' style={{ marginTop: 6 }}>
                Pantau performa PLTS Anda hari ini.
              </p>
            </div>
            <form className='stack' onSubmit={handleLogin}>
              <label>
                <span>Email</span>
                <input type='email' value={loginForm.email} onChange={(e) => setLoginForm({ ...loginForm, email: e.target.value })} placeholder='tim-energy@company.com' required />
              </label>
              <label>
                <span>Password</span>
                <input type='password' value={loginForm.password} onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })} placeholder='Password akun' required />
              </label>
              <button className='primary-button' disabled={isAuthLoading} type='submit'>
                {isAuthLoading ? "Memproses login..." : "Masuk Sekarang"}
              </button>
            </form>
            <p className='auth-subpage-switch'>
              Belum punya akun?{" "}
              <button
                type='button'
                className='link-btn'
                onClick={() => {
                  setError("");
                  setFeedback("");
                  setAuthPage("register");
                }}>
                Daftar gratis
              </button>
            </p>
          </section>
        </div>
      </div>
    );
  }

  // ── Landing page ───────────────────────────────────────────────────────
  const simPsh = getPshByPreset(simTimePreset);
  const simWeatherFactor = getSimWeatherFactor(simCloudCover);
  const simPredictedKwh = (simCapacity * simPsh * 0.8 * simWeatherFactor).toFixed(2);
  const simChartScenarios = [20, 50, 80, 100].map((cloudValue) => {
    const factor = getSimWeatherFactor(cloudValue);
    const kwh = simCapacity * simPsh * 0.8 * factor;
    return { label: `${cloudValue}%`, kwh };
  });
  const simChartMaxKwh = Math.max(...simChartScenarios.map((item) => item.kwh), 1);
  const simActiveChartLabel = simCloudCover <= 20 ? "20%" : simCloudCover <= 50 ? "50%" : simCloudCover <= 80 ? "80%" : "100%";

  return (
    <div className='shell landing-shell'>
      <div className='backdrop backdrop-left' />
      <div className='backdrop backdrop-right' />

      <header className='hero landing-hero'>
        <div>
          <p className='eyebrow'>Green Energy Intelligence</p>
          <h1>Naikkan performa PLTS Anda dengan forecast yang actionable.</h1>
          <p className='hero-copy'>Solar Forecast membantu tim operasional memprediksi produksi energi harian, memantau akurasi, dan menghubungkan data device lapangan dalam satu platform ringan.</p>
          <div className='landing-cta-row'>
            <button
              className='primary-button'
              type='button'
              onClick={() => {
                setError("");
                setFeedback("");
                setAuthPage("register");
              }}>
              Mulai Gratis
            </button>
            <button
              className='secondary-button'
              type='button'
              onClick={() => {
                setError("");
                setFeedback("");
                setAuthPage("login");
              }}>
              Saya Sudah Punya Akun
            </button>
          </div>
        </div>

        <div className='hero-card landing-proof'>
          <span className='hero-card-label'>Kenapa tim energy pilih ini?</span>
          <div className='landing-proof-item'>
            <FiTrendingUp />
            <div>
              <strong>Forecast Lebih Presisi</strong>
              <span>Adaptive efficiency belajar dari actual data Anda.</span>
            </div>
          </div>
          <div className='landing-proof-item'>
            <FiCpu />
            <div>
              <strong>Siap Integrasi Device</strong>
              <span>Ingestion telemetry dengan API key per device.</span>
            </div>
          </div>
          <div className='landing-proof-item'>
            <FiShield />
            <div>
              <strong>Data Aman per User</strong>
              <span>JWT auth, data terisolasi, dan endpoint terproteksi.</span>
            </div>
          </div>
        </div>
      </header>

      {(error || feedback) && (
        <div className={`banner ${error ? "banner-error" : "banner-success"}${feedbackFading && !error ? " banner-fading" : ""}`}>
          <span>{error || feedback}</span>
          <button className='banner-close' type='button' onClick={() => (error ? setError("") : setFeedback(""))} aria-label='Tutup'>
            ×
          </button>
        </div>
      )}

      <section className='landing-highlights'>
        <article className='landing-highlight-card'>
          <span className='landing-highlight-icon'>
            <FiMapPin />
          </span>
          <h3>Efisiensi Green Energy</h3>
          <p>Optimalkan output panel surya per site dan kurangi loss operasional harian.</p>
        </article>
        <article className='landing-highlight-card'>
          <span className='landing-highlight-icon'>
            <FiSun />
          </span>
          <h3>Forecast Harian Cepat</h3>
          <p>Dapatkan prediksi energi berdasarkan weather factor dan profil panel.</p>
        </article>
        <article className='landing-highlight-card'>
          <span className='landing-highlight-icon'>
            <FiActivity />
          </span>
          <h3>Monitoring Berkelanjutan</h3>
          <p>Pantau trend performa forecast vs actual langsung dari dashboard.</p>
        </article>
        <article className='landing-highlight-card'>
          <span className='landing-highlight-icon'>
            <FiZap />
          </span>
          <h3>Skalabel Multi-Site</h3>
          <p>Kelola banyak solar profile dan device lapangan dalam satu akun.</p>
        </article>
      </section>

      {/* Simulation card */}
      <section className='landing-sim-wrap'>
        <div className='landing-sim-card'>
          <div className='landing-sim-left'>
            <p className='eyebrow'>Coba Sekarang — Tanpa Daftar</p>
            <h2 className='landing-sim-title'>Simulasi Forecast PLTS</h2>
            <p className='landing-sim-desc'>Masukkan kapasitas panel dan kondisi cuaca untuk melihat estimasi energi harian yang akan dihasilkan sistem Anda.</p>

            <div className='sim-time-presets'>
              <span className='sim-time-label'>Preset Waktu (PSH)</span>
              <div className='sim-time-actions'>
                {TIME_PRESETS.map((preset) => (
                  <button key={preset.id} className={`sim-time-btn ${simTimePreset === preset.id ? "active" : ""}`} type='button' onClick={() => setSimTimePreset(preset.id)}>
                    {preset.label} · {preset.psh}h
                  </button>
                ))}
              </div>
            </div>

            <div className='landing-sim-controls'>
              <label className='landing-sim-label'>
                <span>
                  Kapasitas Panel <strong>{simCapacity} kWp</strong>
                </span>
                <input className='sim-range' type='range' min={1} max={100} value={simCapacity} onChange={(e) => setSimCapacity(Number(e.target.value))} />
                <div className='sim-range-ticks'>
                  <span>1 kWp</span>
                  <span>50 kWp</span>
                  <span>100 kWp</span>
                </div>
              </label>

              <label className='landing-sim-label'>
                <span>
                  Tutupan Awan <strong>{simCloudCover}%</strong> — {getCloudLabel(simCloudCover)}
                </span>
                <input className='sim-range' type='range' min={0} max={100} value={simCloudCover} onChange={(e) => setSimCloudCover(Number(e.target.value))} />
                <div className='sim-range-ticks'>
                  <span>0%</span>
                  <span>50%</span>
                  <span>100%</span>
                </div>
              </label>
            </div>
          </div>

          <div className='landing-sim-right'>
            <div className='sim-result-box'>
              <p className='sim-result-label'>Estimasi Energi Harian</p>
              <p className='sim-result-kwh'>
                {simPredictedKwh} <span>kWh</span>
              </p>
              <div className='sim-result-meta'>
                <div>
                  <span>Kapasitas</span>
                  <strong>{simCapacity} kWp</strong>
                </div>
                <div>
                  <span>PSH</span>
                  <strong>{simPsh.toFixed(1)} jam</strong>
                </div>
                <div>
                  <span>Efisiensi</span>
                  <strong>80%</strong>
                </div>
                <div>
                  <span>Weather Factor</span>
                  <strong>{simWeatherFactor.toFixed(1)}×</strong>
                </div>
              </div>
              <p className='sim-result-formula'>
                {simCapacity} × {simPsh.toFixed(1)} × 0.8 × {simWeatherFactor.toFixed(1)} = <strong>{simPredictedKwh} kWh</strong>
              </p>

              <div className='sim-mini-chart'>
                <p className='sim-mini-chart-title'>Perbandingan Output vs Cloud Cover</p>
                <div className='sim-mini-chart-grid'>
                  {simChartScenarios.map((item) => (
                    <div key={item.label} className={`sim-mini-bar-wrap ${item.label === simActiveChartLabel ? "active" : ""}`}>
                      <div className='sim-mini-bar-track'>
                        <div className='sim-mini-bar-fill' style={{ height: `${Math.max((item.kwh / simChartMaxKwh) * 100, 6)}%` }} />
                      </div>
                      <span className='sim-mini-bar-value'>{item.kwh.toFixed(1)} kWh</span>
                      <span className='sim-mini-bar-label'>{item.label}</span>
                    </div>
                  ))}
                </div>
                <p className='sim-mini-chart-note'>Semakin tinggi cloud cover, semakin rendah estimasi output energi.</p>
              </div>
            </div>
            <button
              className='primary-button sim-cta-btn'
              type='button'
              onClick={() => {
                setError("");
                setFeedback("");
                setAuthPage("register");
              }}>
              Buat Akun &amp; Simpan Profil Ini
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
