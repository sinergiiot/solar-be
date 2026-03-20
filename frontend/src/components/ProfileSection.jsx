import { FiCrosshair } from "react-icons/fi";

// ProfileSection renders solar profile form and list of existing profiles.
export default function ProfileSection({
  profile,
  profiles,
  profileForm,
  setProfileForm,
  handleCreateProfile,
  isSavingProfile,
  editingProfileID,
  handleEditProfile,
  handleDeleteProfile,
  deletingProfileID,
  handleCancelProfileEdit,
  handleUseDeviceLocation,
  isLocating,
  setProfile,
  setSelectedForecastProfileID,
}) {
  return (
    <>
      <section className='panel panel-form panel-wide'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Setup</span>
          <h2>{profile ? "Tambah / update solar profile" : "Isi solar profile"}</h2>
        </div>
        {profile && <p className='form-note'>Sekarang Anda bisa punya lebih dari satu profile per plant/titik pembangkit.</p>}
        <form className='stack stack-grid' onSubmit={handleCreateProfile}>
          <label>
            <span>Nama Site</span>
            <input type='text' value={profileForm.site_name} onChange={(event) => setProfileForm({ ...profileForm, site_name: event.target.value })} placeholder='Atap Timur / Inverter A' required />
          </label>
          <label>
            <span>Capacity (kWp)</span>
            <input type='number' min='0' step='0.1' value={profileForm.capacity_kwp} onChange={(event) => setProfileForm({ ...profileForm, capacity_kwp: event.target.value })} placeholder='3.5' required />
          </label>
          <label>
            <span>Latitude</span>
            <div className='input-with-action'>
              <input type='number' step='0.0001' value={profileForm.lat} onChange={(event) => setProfileForm({ ...profileForm, lat: event.target.value })} placeholder='-6.2000' required />
              <button className='icon-button' type='button' onClick={handleUseDeviceLocation} disabled={isLocating} title='Isi otomatis dari lokasi device'>
                {isLocating ? "..." : <FiCrosshair />}
              </button>
            </div>
          </label>
          <label>
            <span>Longitude</span>
            <div className='input-with-action'>
              <input type='number' step='0.0001' value={profileForm.lng} onChange={(event) => setProfileForm({ ...profileForm, lng: event.target.value })} placeholder='106.8166' required />
              <button className='icon-button' type='button' onClick={handleUseDeviceLocation} disabled={isLocating} title='Isi otomatis dari lokasi device'>
                {isLocating ? "..." : <FiCrosshair />}
              </button>
            </div>
          </label>
          <label>
            <span>Tilt</span>
            <input type='number' step='0.1' value={profileForm.tilt} onChange={(event) => setProfileForm({ ...profileForm, tilt: event.target.value })} placeholder='15' />
          </label>
          <label>
            <span>Azimuth</span>
            <input type='number' step='0.1' value={profileForm.azimuth} onChange={(event) => setProfileForm({ ...profileForm, azimuth: event.target.value })} placeholder='180' />
          </label>
          <button className='primary-button' disabled={isSavingProfile} type='submit'>
            {isSavingProfile ? "Menyimpan..." : editingProfileID ? "Update Profile" : "Simpan Profile"}
          </button>
          {editingProfileID && (
            <button className='secondary-button' type='button' onClick={handleCancelProfileEdit}>
              Batal Edit
            </button>
          )}
        </form>
      </section>

      <section className='panel panel-data panel-wide'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Data</span>
          <h2>Daftar Solar Profile Akun</h2>
        </div>
        {profiles.length > 0 ? (
          <div className='device-list'>
            {profiles.map((p) => (
              <article className='device-row' key={p.id}>
                <div>
                  <strong>{p.site_name}</strong>
                  <p>ID: {p.id}</p>
                  <p>Capacity: {Number(p.capacity_kwp).toFixed(2)} kWp</p>
                  <p>
                    Koordinat: {p.lat}, {p.lng}
                  </p>
                </div>
                <div className='device-row-actions'>
                  <button className='secondary-button' type='button' onClick={() => handleEditProfile(p)}>
                    Edit
                  </button>
                  <button
                    className='secondary-button'
                    type='button'
                    onClick={() => {
                      setProfile(p);
                      setSelectedForecastProfileID(p.id);
                    }}>
                    Pilih
                  </button>
                  <button className='secondary-button danger-button' type='button' onClick={() => handleDeleteProfile(p.id)} disabled={deletingProfileID === p.id}>
                    {deletingProfileID === p.id ? "Menghapus..." : "Delete"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className='empty-state'>Akun ini belum memiliki solar profile.</div>
        )}
      </section>
    </>
  );
}
