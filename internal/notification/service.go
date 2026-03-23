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
	planTier := strings.TrimSpace(strings.ToLower(req.PlanTier))
	if planTier == "" {
		planTier = PlanFree
	}
	if planTier != PlanFree && planTier != PlanPro && planTier != PlanEnterprise && planTier != PlanPaid {
		return nil, fmt.Errorf("plan_tier must be free, pro, or enterprise")
	}

	// Normalize deprecated 'paid' to 'pro'
	if planTier == PlanPaid {
		planTier = PlanPro
	}

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

// SendRECMilestoneEmail sends a congratulatory email to a user who reaches a MWh milestone.
func (s *service) SendRECMilestoneEmail(toEmail string, userName string, mwh float64) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "🎉 REC Readiness Milestone Reached!")
	
	body := fmt.Sprintf(`
	<h2>Congratulations %s!</h2>
	<p>You have successfully accumulated <strong>%.2f MWh</strong> of solar energy production.</p>
	<p>You are now eligible to claim Renewable Energy Certificates (RECs) for your green energy contribution.</p>
	<p>Log in to your dashboard to download your REC-ready production report.</p>
	`, userName, mwh)
	
	m.SetBody("text/html", body)

	dialer := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("send REC email to %s: %w", toEmail, err)
	}
	return nil
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

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; padding: 20px;">
  <h2>☀️ Daily Solar Forecast</h2>
  <p>Hi <strong>%s</strong>,</p>
  <p>Berikut adalah laporan harian dari profil <strong>%s</strong>:</p>
  <table style="border-collapse: collapse; width: 400px; text-align: left;">
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Prediksi energi</td>
      <td style="padding: 8px; border: 1px solid #ddd;"><strong>%.2f kWh</strong></td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Cloud Cover?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%d%%</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Weather Factor (Transmittance)?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%.2f</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Baseline Type?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Efficiency?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%.1f%%</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Tanggal forecast</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Solar profile aktif</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Estimasi hemat biaya</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Estimasi CO2 dihindari?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%.2f kgCO2</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Deviasi vs actual referensi?</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Referensi</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
    <tr>
      <td style="padding: 8px; border: 1px solid #ddd;">Status risiko cuaca</td>
      <td style="padding: 8px; border: 1px solid #ddd;">%s</td>
    </tr>
  </table>
  <br/>
  <p style="color: #333; line-height: 1.5; max-width: 600px;">
    Hari ini berdasarkan ramalan cuaca koordinat %.4f, %.4f, diprediksi %s, %s, dan estimasi produksi energi harian Anda sekitar %.2f kWh dengan potensi penghematan %s.
  </p>
  <br/>
  <p style="color: #888; font-size: 12px;">Solar Forecast System — automated daily report</p>
</body>
</html>
`, p.ToName, solarProfile, p.PredictedKwh, p.CloudCover, p.WeatherFactor, p.BaselineType, p.Efficiency*100, forecastDate, solarProfile, formatCurrency(p.EstimatedCost), p.EstimatedCO2Kg, deviation, reference, p.WeatherRisk, p.Lat, p.Lng, p.ConditionLabel, p.ConditionImpact, p.PredictedKwh, formatCurrency(p.EstimatedCost))
}
