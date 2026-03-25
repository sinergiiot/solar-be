package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"gopkg.in/gomail.v2"
)

// Service defines notification sending operations
type Service interface {
	SendForecastEmail(payload EmailPayload) error
	GetPreference(userID uuid.UUID) (*NotificationPreference, error)
	GetAllPreferences() ([]*NotificationPreference, error)
	UpsertPreference(userID uuid.UUID, req UpsertPreferenceRequest) (*NotificationPreference, error)
	DispatchDailyForecast(payload DispatchPayload) error
	MarkDailyForecastSent(userID uuid.UUID, forecastDate time.Time, sentAt time.Time) error
	SendRECMilestoneEmail(toEmail string, userName string, mwh float64) error
	SetPlanTier(userID uuid.UUID, tier string, expiresAt *time.Time) error
	SendUpgradeConfirmationEmail(toEmail, userName, tier string, expiresAt time.Time) error
	SendSubscriptionExpiringEmail(toEmail, userName, tier string, expiresAt time.Time) error
}

type service struct {
	repo             Repository
	httpClient       *http.Client
	host             string
	port             int
	username         string
	password         string
	from             string
	telegramBotToken string
	whatsAppToken    string
	whatsAppPhoneID  string
	whatsAppTemplate string
	whatsAppLanguage string
}

// NewService creates a new email notification service
func NewService(repo Repository, host, portStr, username, password, from, telegramBotToken, whatsAppToken, whatsAppPhoneID, whatsAppTemplate, whatsAppLanguage string) Service {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587
	}
	return &service{
		repo:             repo,
		httpClient:       &http.Client{Timeout: 12 * time.Second},
		host:             host,
		port:             port,
		username:         username,
		password:         password,
		from:             from,
		telegramBotToken: strings.TrimSpace(telegramBotToken),
		whatsAppToken:    strings.TrimSpace(whatsAppToken),
		whatsAppPhoneID:  strings.TrimSpace(whatsAppPhoneID),
		whatsAppTemplate: strings.TrimSpace(whatsAppTemplate),
		whatsAppLanguage: strings.TrimSpace(whatsAppLanguage),
	}
}

// GetPreference returns one user's notification preference and auto-initializes default when missing.
func (s *service) GetPreference(userID uuid.UUID) (*NotificationPreference, error) {
	pref, err := s.repo.GetPreference(userID)
	if err != nil {
		return nil, err
	}
	if pref != nil {
		return pref, nil
	}

	defaultPref := s.defaultPreference(userID)
	if err := s.repo.UpsertPreference(defaultPref); err != nil {
		return nil, err
	}

	created, err := s.repo.GetPreference(userID)
	if err != nil {
		return nil, err
	}
	if created == nil {
		return defaultPref, nil
	}

	return created, nil
}

// GetAllPreferences returns all persisted notification preferences without creating defaults.
func (s *service) GetAllPreferences() ([]*NotificationPreference, error) {
	return s.repo.GetAllPreferences()
}

// UpsertPreference validates and stores one user's notification preference.
func (s *service) UpsertPreference(userID uuid.UUID, req UpsertPreferenceRequest) (*NotificationPreference, error) {
	// PlanTier should NOT be editable by user here.
	// We first fetch current preference (which also ensures default exists).
	current, err := s.GetPreference(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current preference: %w", err)
	}
	planTier := current.PlanTier

	primaryChannel := strings.TrimSpace(strings.ToLower(req.PrimaryChannel))
	if primaryChannel == "" {
		primaryChannel = ChannelEmail
	}
	if primaryChannel != ChannelEmail && primaryChannel != ChannelTelegram && primaryChannel != ChannelWhatsApp {
		return nil, fmt.Errorf("primary_channel must be email, telegram, or whatsapp")
	}

	timezone := strings.TrimSpace(req.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}

	preferredSendTime := strings.TrimSpace(req.PreferredSendTime)
	if preferredSendTime == "" {
		preferredSendTime = "06:00:00"
	}
	if _, err := time.Parse("15:04:05", preferredSendTime); err != nil {
		if _, shortErr := time.Parse("15:04", preferredSendTime); shortErr == nil {
			preferredSendTime = preferredSendTime + ":00"
		} else {
			return nil, fmt.Errorf("preferred_send_time must use HH:MM or HH:MM:SS")
		}
	}

	pref := &NotificationPreference{
		UserID:            userID,
		PlanTier:          planTier,
		PrimaryChannel:    primaryChannel,
		EmailEnabled:      req.EmailEnabled,
		TelegramEnabled:   req.TelegramEnabled,
		WhatsAppEnabled:   req.WhatsAppEnabled,
		TelegramChatID:    strings.TrimSpace(req.TelegramChatID),
		WhatsAppPhoneE164: strings.TrimSpace(req.WhatsAppPhoneE164),
		WhatsAppOptedIn:   req.WhatsAppOptedIn,
		Timezone:          timezone,
		PreferredSendTime: preferredSendTime,
	}

	if !pref.EmailEnabled && !pref.TelegramEnabled && !pref.WhatsAppEnabled {
		pref.EmailEnabled = true
	}

	if pref.PlanTier == PlanFree {
		pref.WhatsAppEnabled = false
		if pref.PrimaryChannel == ChannelWhatsApp {
			pref.PrimaryChannel = ChannelEmail
		}
	}

	if err := s.repo.UpsertPreference(pref); err != nil {
		return nil, err
	}

	return s.GetPreference(userID)
}

// SetPlanTier updates the plan tier for a user. This is intended to be called by the billing service.
func (s *service) SetPlanTier(userID uuid.UUID, tier string, expiresAt *time.Time) error {
	tier = strings.TrimSpace(strings.ToLower(tier))
	if tier == "" {
		tier = PlanFree
	}
	if tier != PlanFree && tier != PlanPro && tier != PlanEnterprise && tier != PlanPaid {
		return fmt.Errorf("invalid plan tier: %s", tier)
	}
	if tier == PlanPaid {
		tier = PlanPro
	}

	pref, err := s.GetPreference(userID)
	if err != nil {
		return err
	}
	pref.PlanTier = tier
	pref.PlanExpiresAt = expiresAt
	pref.UpdatedAt = time.Now()
	return s.repo.UpsertPreference(pref)
}

func (s *service) SendUpgradeConfirmationEmail(toEmail, userName, tier string, expiresAt time.Time) error {
	subject := fmt.Sprintf("Pembayaran Berhasil: Upgrade Paket %s Aktif!", strings.Title(tier))

	localExpiry := expiresAt.Format("02 January 2006")
	body := buildBaseEmailTemplate(fmt.Sprintf(`
		<div style="text-align: center; padding: 20px 0;">
			<div style="background: #ecfdf5; color: #059669; width: 64px; height: 64px; line-height: 64px; border-radius: 50%%; font-size: 32px; margin: 0 auto 20px;">✓</div>
			<h2 style="color: #111827; margin: 0;">Pembayaran Berhasil!</h2>
			<p style="color: #6b7280; font-size: 16px;">Akun Anda kini telah resmi ditingkatkan.</p>
		</div>
		<div style="background: #f9fafb; border-radius: 12px; padding: 20px; margin-bottom: 24px;">
			<table style="width: 100%%; border-collapse: collapse;">
				<tr>
					<td style="color: #6b7280; padding: 8px 0;">Paket Baru</td>
					<td style="text-align: right; font-weight: 600; color: #111827;">%s</td>
				</tr>
				<tr>
					<td style="color: #6b7280; padding: 8px 0;">Masa Berlaku</td>
					<td style="text-align: right; font-weight: 600; color: #111827;">%s</td>
				</tr>
			</table>
		</div>
		<p style="color: #4b5563; line-height: 1.6;">
			Halo <strong>%s</strong>, terima kasih telah mempercayakan transisi energi Anda kepada kami. 
			Seluruh fitur Green Compliance sekarang telah terbuka untuk Anda.
		</p>
		<div style="text-align: center; margin-top: 32px;">
			<a href="http://localhost:5173/dashboard" style="background: #10b981; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-weight: 600; display: inline-block;">Masuk ke Dashboard</a>
		</div>
	`, strings.Title(tier), localExpiry, userName))

	return s.sendEmail(toEmail, subject, body)
}

// SendSubscriptionExpiringEmail sends a reminder 7 days before subscription expires.
func (s *service) SendSubscriptionExpiringEmail(toEmail, userName, tier string, expiresAt time.Time) error {
	subject := "Peringatan: Langganan Solar Forecast Segera Berakhir"
	localExpiry := expiresAt.Format("02 January 2006")

	body := buildBaseEmailTemplate(fmt.Sprintf(`
		<div style="text-align: center; padding: 20px 0;">
			<div style="background: #fffbeb; color: #d97706; width: 64px; height: 64px; line-height: 64px; border-radius: 50%%; font-size: 32px; margin: 0 auto 20px;">!</div>
			<h2 style="color: #111827; margin: 0;">Langganan Segera Berakhir</h2>
			<p style="color: #6b7280; font-size: 16px;">Pastikan akses Anda tetap aktif tanpa gangguan.</p>
		</div>
		<p style="color: #4b5563; line-height: 1.6;">
			Halo <strong>%s</strong>, masa berlaku paket <strong>%s</strong> Anda akan habis pada tanggal <strong>%s</strong>.
		</p>
		<p style="color: #4b5563; line-height: 1.6;">
			Perpanjang sekarang untuk menjaga akses laporan PDF bulanan, tracker CO2, dan data historis tanpa batas Anda.
		</p>
		<div style="text-align: center; margin-top: 32px;">
			<a href="http://localhost:5173/dashboard" style="background: #10b981; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-weight: 600; display: inline-block;">Perpanjang Sekarang</a>
		</div>
	`, userName, strings.Title(tier), localExpiry))

	return s.sendEmail(toEmail, subject, body)
}

// DispatchDailyForecast routes one forecast payload through user-preferred channels with fallback policy.
func (s *service) DispatchDailyForecast(payload DispatchPayload) error {
	pref, err := s.GetPreference(payload.UserID)
	if err != nil {
		return err
	}

	channels := s.resolveChannelOrder(pref)
	if len(channels) == 0 {
		channels = []string{ChannelEmail}
	}

	var errors []string
	for _, channel := range channels {
		err := s.sendByChannel(channel, pref, payload)
		if err == nil {
			return nil
		}
		errors = append(errors, fmt.Sprintf("%s: %v", channel, err))
	}

	return fmt.Errorf("all notification channels failed: %s", strings.Join(errors, "; "))
}

// MarkDailyForecastSent records a successful scheduled delivery for one local forecast date.
func (s *service) MarkDailyForecastSent(userID uuid.UUID, forecastDate time.Time, sentAt time.Time) error {
	return s.repo.MarkDailyForecastSent(userID, forecastDate, sentAt)
}

// SendForecastEmail composes and sends a daily solar forecast email to the user
func (s *service) SendForecastEmail(payload EmailPayload) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", payload.ToEmail)
	m.SetHeader("Subject", fmt.Sprintf("☀️ Solar Forecast for %s", payload.Date))
	m.SetBody("text/html", buildEmailBody(payload))

	dialer := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("send forecast email to %s: %w", payload.ToEmail, err)
	}
	return nil
}

func (s *service) SendRECMilestoneEmail(toEmail string, userName string, mwh float64) error {
	subject := "🎉 REC Readiness Milestone Tercapai!"

	body := buildBaseEmailTemplate(fmt.Sprintf(`
		<div style="text-align: center; padding: 20px 0;">
			<div style="background: #fef3c7; color: #d97706; width: 64px; height: 64px; line-height: 64px; border-radius: 50%%; font-size: 32px; margin: 0 auto 20px;">🏆</div>
			<h2 style="color: #111827; margin: 0;">Pencapaian Baru!</h2>
			<p style="color: #6b7280; font-size: 16px;">Produksi Energi Bersih Anda Luar Biasa.</p>
		</div>
		<p style="color: #4b5563; line-height: 1.6;">
			Selamat <strong>%s</strong>! Akumulasi produksi energi PLTS Anda telah mencapai <strong>%.2f MWh</strong>.
		</p>
		<p style="color: #4b5563; line-height: 1.6;">
			Dengan angka ini, Anda kini memenuhi syarat untuk mengklaim <strong>Renewable Energy Certificates (REC)</strong> sebagai bukti kontribusi nyata Anda pada transisi energi hijau.
		</p>
		<div style="text-align: center; margin-top: 32px;">
			<a href="http://localhost:5173/dashboard" style="background: #10b981; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-weight: 600; display: inline-block;">Lihat Laporan REC</a>
		</div>
	`, userName, mwh))

	if err := s.sendEmail(toEmail, subject, body); err != nil {
		return fmt.Errorf("send REC email to %s: %w", toEmail, err)
	}
	return nil
}

func (s *service) sendEmail(toEmail, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(s.host, s.port, s.username, s.password)
	return dialer.DialAndSend(m)
}

// resolveChannelOrder returns delivery channel priority based on plan and user preference.
func (s *service) resolveChannelOrder(pref *NotificationPreference) []string {
	enabled := map[string]bool{
		ChannelEmail:    pref.EmailEnabled,
		ChannelTelegram: pref.TelegramEnabled,
		ChannelWhatsApp: pref.WhatsAppEnabled && pref.WhatsAppOptedIn,
	}

	if pref.PlanTier != PlanPaid {
		enabled[ChannelWhatsApp] = false
	}

	ordered := []string{}
	addChannel := func(channel string) {
		if enabled[channel] {
			ordered = append(ordered, channel)
		}
	}

	addChannel(pref.PrimaryChannel)
	addChannel(ChannelEmail)
	addChannel(ChannelTelegram)
	if pref.PlanTier == PlanPaid {
		addChannel(ChannelWhatsApp)
	}

	unique := make([]string, 0, len(ordered))
	seen := map[string]bool{}
	for _, ch := range ordered {
		if !seen[ch] {
			unique = append(unique, ch)
			seen[ch] = true
		}
	}

	return unique
}

// sendByChannel dispatches one payload using the specified channel.
func (s *service) sendByChannel(channel string, pref *NotificationPreference, payload DispatchPayload) error {
	switch channel {
	case ChannelEmail:
		if strings.TrimSpace(payload.ToEmail) == "" {
			return fmt.Errorf("missing email address")
		}
		return s.SendForecastEmail(EmailPayload{
			ToName:           payload.ToName,
			ToEmail:          payload.ToEmail,
			Date:             payload.Date,
			PredictedKwh:     payload.PredictedKwh,
			CloudCover:       payload.CloudCover,
			BaselineType:     payload.BaselineType,
			WeatherFactor:    payload.WeatherFactor,
			Efficiency:       payload.Efficiency,
			SolarProfileName: payload.SolarProfileName,
			EstimatedCost:    payload.EstimatedCost,
			EstimatedCO2Kg:   payload.EstimatedCO2Kg,
			DeviationPct:     payload.DeviationPct,
			ReferenceLabel:   payload.ReferenceLabel,
			WeatherRisk:      payload.WeatherRisk,
			Lat:              payload.Lat,
			Lng:              payload.Lng,
			ConditionLabel:   payload.ConditionLabel,
			ConditionImpact:  payload.ConditionImpact,
		})
	case ChannelTelegram:
		return s.sendForecastTelegram(pref.TelegramChatID, payload)
	case ChannelWhatsApp:
		return s.sendForecastWhatsApp(pref.WhatsAppPhoneE164, payload)
	default:
		return fmt.Errorf("unsupported channel %s", channel)
	}
}

// sendForecastTelegram sends one forecast message through Telegram bot API.
func (s *service) sendForecastTelegram(chatID string, payload DispatchPayload) error {
	if s.telegramBotToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	if strings.TrimSpace(chatID) == "" {
		return fmt.Errorf("telegram_chat_id is required")
	}

	message := fmt.Sprintf(
		"☀️ Forecast %s\n\nPrediksi energi: %.2f kWh\nCloud Cover: %d%%\nWeather Factor (Transmittance): %.2f\nBaseline Type: %s\nEfficiency: %.1f%%\nTanggal forecast: %s\nSolar profile aktif: %s\nEstimasi hemat biaya: %s\nEstimasi CO2 dihindari: %.2f kgCO2\nDeviasi vs actual referensi: %s\nReferensi: %s\nStatus risiko cuaca: %s\n\nHari ini berdasarkan ramalan cuaca koordinat %.4f, %.4f, diprediksi %s, %s, dan estimasi produksi energi harian Anda sekitar %.2f kWh dengan potensi penghematan %s.",
		payload.Date,
		payload.PredictedKwh,
		payload.CloudCover,
		payload.WeatherFactor,
		payload.BaselineType,
		payload.Efficiency*100,
		formatForecastDate(payload.Date),
		emptyFallback(payload.SolarProfileName, "-"),
		formatCurrency(payload.EstimatedCost),
		payload.EstimatedCO2Kg,
		formatDeviation(payload.DeviationPct),
		emptyFallback(payload.ReferenceLabel, "actual referensi"),
		payload.WeatherRisk,
		payload.Lat,
		payload.Lng,
		payload.ConditionLabel,
		payload.ConditionImpact,
		payload.PredictedKwh,
		formatCurrency(payload.EstimatedCost),
	)

	body := map[string]any{
		"chat_id": chatID,
		"text":    message,
	}
	rawBody, _ := json.Marshal(body)

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramBotToken)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram send failed with status %d", resp.StatusCode)
	}

	return nil
}

// sendForecastWhatsApp sends one detailed forecast text message through WhatsApp Cloud API.
func (s *service) sendForecastWhatsApp(phoneE164 string, payload DispatchPayload) error {
	if s.whatsAppToken == "" || s.whatsAppPhoneID == "" {
		return fmt.Errorf("whatsapp provider is not configured")
	}
	if strings.TrimSpace(phoneE164) == "" {
		return fmt.Errorf("whatsapp_phone_e164 is required")
	}

	message := fmt.Sprintf(
		"☀️ Forecast %s\n\nPrediksi energi: %.2f kWh\nCloud Cover: %d%%\nWeather Factor (Transmittance): %.2f\nBaseline Type: %s\nEfficiency: %.1f%%\nTanggal forecast: %s\nSolar profile aktif: %s\nEstimasi hemat biaya: %s\nEstimasi CO2 dihindari: %.2f kgCO2\nDeviasi vs actual referensi: %s\nReferensi: %s\nStatus risiko cuaca: %s\n\nHari ini berdasarkan ramalan cuaca koordinat %.4f, %.4f, diprediksi %s, %s, dan estimasi produksi energi harian Anda sekitar %.2f kWh dengan potensi penghematan %s.",
		payload.Date,
		payload.PredictedKwh,
		payload.CloudCover,
		payload.WeatherFactor,
		payload.BaselineType,
		payload.Efficiency*100,
		formatForecastDate(payload.Date),
		emptyFallback(payload.SolarProfileName, "-"),
		formatCurrency(payload.EstimatedCost),
		payload.EstimatedCO2Kg,
		formatDeviation(payload.DeviationPct),
		emptyFallback(payload.ReferenceLabel, "actual referensi"),
		payload.WeatherRisk,
		payload.Lat,
		payload.Lng,
		payload.ConditionLabel,
		payload.ConditionImpact,
		payload.PredictedKwh,
		formatCurrency(payload.EstimatedCost),
	)

	body := map[string]any{
		"messaging_product": "whatsapp",
		"to":                phoneE164,
		"type":              "text",
		"text": map[string]any{
			"preview_url": false,
			"body":        message,
		},
	}

	rawBody, _ := json.Marshal(body)
	endpoint := fmt.Sprintf("https://graph.facebook.com/v20.0/%s/messages", s.whatsAppPhoneID)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("create whatsapp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.whatsAppToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send whatsapp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp send failed with status %d", resp.StatusCode)
	}

	return nil
}

// defaultPreference returns the default free-tier email configuration.
func (s *service) defaultPreference(userID uuid.UUID) *NotificationPreference {
	return &NotificationPreference{
		UserID:            userID,
		PlanTier:          PlanFree,
		PrimaryChannel:    ChannelEmail,
		EmailEnabled:      true,
		TelegramEnabled:   false,
		WhatsAppEnabled:   false,
		TelegramChatID:    "",
		WhatsAppPhoneE164: "",
		WhatsAppOptedIn:   false,
		Timezone:          "UTC",
		PreferredSendTime: "06:00:00",
	}
}

// formatCurrency produces an IDR-like string for text channels.
func formatCurrency(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 0, 64)
	parts := []byte(formatted)
	if len(parts) <= 3 {
		return "Rp " + formatted
	}

	var out []byte
	count := 0
	for i := len(parts) - 1; i >= 0; i-- {
		out = append([]byte{parts[i]}, out...)
		count++
		if count%3 == 0 && i != 0 {
			out = append([]byte{'.'}, out...)
		}
	}

	return "Rp " + string(out)
}

// formatDeviation renders optional deviation percentage for text channels.
func formatDeviation(value *float64) string {
	if value == nil {
		return "--"
	}
	if *value >= 0 {
		return fmt.Sprintf("+%.1f%%", *value)
	}
	return fmt.Sprintf("%.1f%%", *value)
}

// formatForecastDate converts YYYY-MM-DD to a more human-readable date.
func formatForecastDate(raw string) string {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	return parsed.Format("02 Jan 2006")
}

// emptyFallback returns fallback when value is blank.
func emptyFallback(value string, fallback string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return fallback
	}
	return v
}

// buildEmailBody generates the HTML content for the forecast email
func buildEmailBody(p EmailPayload) string {
	deviation := formatDeviation(p.DeviationPct)
	reference := emptyFallback(p.ReferenceLabel, "actual referensi")
	solarProfile := emptyFallback(p.SolarProfileName, "-")
	forecastDate := formatForecastDate(p.Date)

	content := fmt.Sprintf(`
		<div style="margin-bottom: 24px;">
			<div style="float: right; background: #e5e7eb; padding: 4px 12px; border-radius: 20px; font-size: 12px; color: #4b5563; font-weight: 600;">%s</div>
			<h2 style="color: #111827; margin: 0 0 8px 0;">☀️ Produksi PLTS Hari Ini</h2>
			<p style="color: #6b7280; font-size: 16px; margin: 0;">Laporan harian untuk lokasi: <strong>%s</strong></p>
		</div>

		<div style="background: #f9fafb; border-radius: 12px; padding: 24px; margin-bottom: 24px;">
			<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px;">
				<div style="background: white; padding: 16px; border-radius: 8px; border: 1px solid #e5e7eb; margin-bottom: 12px;">
					<div style="color: #6b7280; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em;">Prediksi Energi</div>
					<div style="font-size: 24px; font-weight: 700; color: #10b981;">%.2f <span style="font-size: 14px; font-weight: 400;">kWh</span></div>
				</div>
				<div style="background: white; padding: 16px; border-radius: 8px; border: 1px solid #e5e7eb; margin-bottom: 12px;">
					<div style="color: #6b7280; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em;">CO2 Avoided</div>
					<div style="font-size: 24px; font-weight: 700; color: #3b82f6;">%.2f <span style="font-size: 14px; font-weight: 400;">kg</span></div>
				</div>
			</div>

			<table style="width: 100%%; border-collapse: collapse; margin-top: 12px;">
				<tr style="border-bottom: 1px solid #e5e7eb;">
					<td style="padding: 12px 0; color: #6b7280;">Kondisi Cuaca</td>
					<td style="padding: 12px 0; text-align: right; font-weight: 600;">%s</td>
				</tr>
				<tr style="border-bottom: 1px solid #e5e7eb;">
					<td style="padding: 12px 0; color: #6b7280;">Cloud Cover</td>
					<td style="padding: 12px 0; text-align: right; font-weight: 600;">%d%%</td>
				</tr>
				<tr style="border-bottom: 1px solid #e5e7eb;">
					<td style="padding: 12px 0; color: #6b7280;">Efisiensi Panel</td>
					<td style="padding: 12px 0; text-align: right; font-weight: 600;">%.1f%%</td>
				</tr>
				<tr>
					<td style="padding: 12px 0; color: #6b7280;">Deviasi vs Referensi</td>
					<td style="padding: 12px 0; text-align: right; font-weight: 600; color: #111827;">%s (%s)</td>
				</tr>
			</table>
		</div>

		<p style="color: #4b5563; line-height: 1.6; font-size: 15px;">
			Berdasarkan koordinat lokasi PLTS Anda, cuaca diprediksi <strong>%s</strong> hari ini. 
			Produksi energi diproyeksikan memberikan penghematan sebesar <strong>%s</strong>.
		</p>

		<div style="text-align: center; margin-top: 32px;">
			<a href="http://localhost:5173/dashboard" style="background: #10b981; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-weight: 600; display: inline-block;">Analisis Lengkap di Dashboard</a>
		</div>
	`, forecastDate, solarProfile, p.PredictedKwh, p.EstimatedCO2Kg, p.ConditionLabel, p.CloudCover, p.Efficiency*100, deviation, reference, p.ConditionLabel, formatCurrency(p.EstimatedCost))

	return buildBaseEmailTemplate(content)
}

func buildBaseEmailTemplate(content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Solar Forecast Notification</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f3f4f6; font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;">
	<table width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #f3f4f6; padding: 40px 20px;">
		<tr>
			<td align="center">
				<table width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">
					<!-- Header / Brand -->
					<tr>
						<td style="background-color: #111827; padding: 32px; text-align: center;">
							<div style="color: #10b981; font-size: 24px; font-weight: 800; letter-spacing: -0.025em;">
								SOLAR<span style="color: #ffffff;">FORECAST</span>
							</div>
							<div style="color: #9ca3af; font-size: 12px; text-transform: uppercase; letter-spacing: 0.1em; margin-top: 4px;">Energy Transition Monitor</div>
						</td>
					</tr>
					<!-- Main Content -->
					<tr>
						<td style="padding: 40px 32px;">
							%s
						</td>
					</tr>
					<!-- Footer -->
					<tr>
						<td style="background-color: #f9fafb; padding: 32px; text-align: center; border-top: 1px solid #e5e7eb;">
							<p style="color: #9ca3af; font-size: 14px; margin: 0 0 16px 0;">Solar Forecast Sinergi IoT Nusantara</p>
							<div style="margin-bottom: 16px;">
								<a href="#" style="color: #6b7280; text-decoration: none; margin: 0 8px; font-size: 13px;">Privacy Policy</a>
								<a href="#" style="color: #6b7280; text-decoration: none; margin: 0 8px; font-size: 13px;">Support</a>
								<a href="#" style="color: #6b7280; text-decoration: none; margin: 0 8px; font-size: 13px;">Portal</a>
							</div>
							<p style="color: #d1d5db; font-size: 12px; margin: 0;">&copy; %d Solar Forecast. All rights reserved.</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
	`, content, time.Now().Year())
}
