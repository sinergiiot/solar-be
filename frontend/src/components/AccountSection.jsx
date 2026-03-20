// AccountSection renders account info and notification preference form.
export default function AccountSection({ currentUser, notificationPreference, setNotificationPreference, handleSaveNotificationPreference, isSavingNotificationPreference, isLoadingNotificationPreference }) {
  return (
    <section className='panel panel-wide summary-panel'>
      <div className='panel-heading'>
        <span className='panel-kicker'>Session</span>
        <h2>Account Information</h2>
      </div>
      <div className='account-info'>
        <div className='info-group'>
          <span className='info-label'>Nama</span>
          <strong className='info-value'>{currentUser?.name || "-"}</strong>
        </div>
        <div className='info-group'>
          <span className='info-label'>Email</span>
          <strong className='info-value'>{currentUser?.email || "-"}</strong>
        </div>
        <div className='info-group'>
          <span className='info-label'>Status</span>
          <strong className='info-value' style={{ color: "var(--success)" }}>
            Authenticated
          </strong>
        </div>
      </div>

      <div className='account-notification-block'>
        <div className='panel-heading'>
          <span className='panel-kicker'>Notification</span>
          <h2>Notifikasi Otomatis 06:00</h2>
        </div>

        {isLoadingNotificationPreference ? (
          <div className='empty-state'>Memuat pengaturan notifikasi...</div>
        ) : (
          <form className='stack account-notification-form' onSubmit={handleSaveNotificationPreference}>
            <label>
              <span>Plan tier</span>
              <select
                value={notificationPreference.plan_tier}
                onChange={(event) => {
                  const nextPlan = event.target.value;
                  setNotificationPreference((current) => ({
                    ...current,
                    plan_tier: nextPlan,
                    ...(nextPlan === "free"
                      ? {
                          whatsapp_enabled: false,
                          whatsapp_opted_in: false,
                          ...(current.primary_channel === "whatsapp" ? { primary_channel: "email" } : {}),
                        }
                      : {}),
                  }));
                }}>
                <option value='free'>Free (Email/Telegram)</option>
                <option value='paid'>Paid (WhatsApp + fallback)</option>
              </select>
            </label>

            <label>
              <span>Primary channel</span>
              <select value={notificationPreference.primary_channel} onChange={(event) => setNotificationPreference((current) => ({ ...current, primary_channel: event.target.value }))}>
                <option value='email'>Email</option>
                <option value='telegram'>Telegram</option>
                {notificationPreference.plan_tier === "paid" && <option value='whatsapp'>WhatsApp</option>}
              </select>
            </label>

            <label>
              <span>Timezone</span>
              <input value={notificationPreference.timezone} onChange={(event) => setNotificationPreference((current) => ({ ...current, timezone: event.target.value }))} placeholder='Asia/Jakarta' />
            </label>

            <label>
              <span>Jam kirim (HH:MM:SS)</span>
              <input value={notificationPreference.preferred_send_time} onChange={(event) => setNotificationPreference((current) => ({ ...current, preferred_send_time: event.target.value }))} placeholder='06:00:00' />
            </label>

            <div className='notification-toggle-grid'>
              <label className='notification-toggle'>
                <input type='checkbox' checked={notificationPreference.email_enabled} onChange={(event) => setNotificationPreference((current) => ({ ...current, email_enabled: event.target.checked }))} />
                <span>Email enabled</span>
              </label>

              <label className='notification-toggle'>
                <input type='checkbox' checked={notificationPreference.telegram_enabled} onChange={(event) => setNotificationPreference((current) => ({ ...current, telegram_enabled: event.target.checked }))} />
                <span>Telegram enabled</span>
              </label>

              <label className='notification-toggle'>
                <input
                  type='checkbox'
                  checked={notificationPreference.whatsapp_enabled}
                  disabled={notificationPreference.plan_tier !== "paid"}
                  onChange={(event) => setNotificationPreference((current) => ({ ...current, whatsapp_enabled: event.target.checked }))}
                />
                <span>WhatsApp enabled (paid)</span>
              </label>
            </div>

            <label>
              <span>Telegram chat ID (opsional)</span>
              <input value={notificationPreference.telegram_chat_id} onChange={(event) => setNotificationPreference((current) => ({ ...current, telegram_chat_id: event.target.value }))} placeholder='123456789' />
            </label>

            <label>
              <span>WhatsApp phone E.164 (opsional)</span>
              <input value={notificationPreference.whatsapp_phone_e164} onChange={(event) => setNotificationPreference((current) => ({ ...current, whatsapp_phone_e164: event.target.value }))} placeholder='62812xxxxxxx' />
            </label>

            <label className='notification-toggle'>
              <input
                type='checkbox'
                checked={notificationPreference.whatsapp_opted_in}
                disabled={notificationPreference.plan_tier !== "paid"}
                onChange={(event) => setNotificationPreference((current) => ({ ...current, whatsapp_opted_in: event.target.checked }))}
              />
              <span>WhatsApp opt-in confirmed</span>
            </label>

            <button className='primary-button' type='submit' disabled={isSavingNotificationPreference}>
              {isSavingNotificationPreference ? "Menyimpan..." : "Simpan Pengaturan Notifikasi"}
            </button>
          </form>
        )}
      </div>
    </section>
  );
}
