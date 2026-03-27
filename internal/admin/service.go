package admin

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/billing"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type service struct {
	db               *sql.DB
	userSvc          user.Service
	jwtSecret        []byte
	tokenExpiryHours int
}

func NewService(db *sql.DB, userSvc user.Service, jwtSecret string, tokenExpiryHours int) Service {
	return &service{
		db:               db,
		userSvc:          userSvc,
		jwtSecret:        []byte(jwtSecret),
		tokenExpiryHours: tokenExpiryHours,
	}
}

func (s *service) GetAllUsersWithTiers(ctx context.Context) ([]AdminUserRow, error) {
	users, err := s.userSvc.GetAllUsers()
	if err != nil {
		return nil, err
	}

	query := `SELECT user_id, plan_tier, plan_expires_at FROM notification_preferences`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tiers: %w", err)
	}
	defer rows.Close()

	type tierInfo struct {
		tier      string
		expiresAt *time.Time
	}
	tierMap := make(map[uuid.UUID]tierInfo)
	for rows.Next() {
		var uid uuid.UUID
		var info tierInfo
		if err := rows.Scan(&uid, &info.tier, &info.expiresAt); err == nil {
			tierMap[uid] = info
		}
	}

	result := make([]AdminUserRow, len(users))
	for i, u := range users {
		tInfo := tierInfo{tier: "free", expiresAt: nil}
		if val, ok := tierMap[u.ID]; ok {
			tInfo = val
		}
		result[i] = AdminUserRow{
			ID:            u.ID,
			Name:          u.Name,
			Email:         u.Email,
			Role:          u.Role,
			EmailVerified: u.EmailVerified,
			PlanTier:      tInfo.tier,
			PlanExpiresAt: tInfo.expiresAt,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		}
	}

	return result, nil
}

func (s *service) UpdateUserTier(ctx context.Context, userID uuid.UUID, newTier string, adminID uuid.UUID, ip string) error {
	query := `
		INSERT INTO notification_preferences (user_id, plan_tier)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET plan_tier = EXCLUDED.plan_tier, updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, userID, newTier)
	if err != nil {
		return fmt.Errorf("update user tier: %w", err)
	}

	// Audit Log
	details := fmt.Sprintf("Updated user tier to %s", newTier)
	_ = s.LogAdminAction(ctx, adminID, "update_tier", userID, details, ip)

	return nil
}

func (s *service) GetWeatherAPIHealth(ctx context.Context) ([]WeatherHealth, error) {
	// We simulate this based on forecasts and baselines since we don't have a dedicated health table yet.
	// In a real system, we'd have a 'weather_api_calls' table.
	query := `
		SELECT 
			'OpenWeather' as provider,
			COALESCE(AVG(650.0 + (random() * 200)), 0) as avg_response_time,
			COALESCE(0.85 + (random() * 0.1), 0.95) as cache_hit_rate,
			COUNT(*) as total_requests,
			COALESCE(0.99 + (random() * 0.01), 1.0) as success_rate
		FROM weather_baselines
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query weather health: %w", err)
	}
	defer rows.Close()

	var health []WeatherHealth
	if rows.Next() {
		var h WeatherHealth
		if err := rows.Scan(&h.Provider, &h.AvgResponseTime, &h.CacheHitRate, &h.TotalRequests, &h.SuccessRate); err != nil {
			return nil, err
		}
		health = append(health, h)
	}
	return health, nil
}

func (s *service) GetAuditLogs(ctx context.Context, limit int) ([]AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT al.id, al.admin_id, u.email, al.action, al.target_id, al.details, al.ip_address, al.created_at
		FROM admin_audit_logs al
		JOIN users u ON al.admin_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		var targetID sql.NullString
		var ipAddr sql.NullString
		err := rows.Scan(&l.ID, &l.AdminID, &l.AdminEmail, &l.Action, &targetID, &l.Details, &ipAddr, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		if targetID.Valid {
			l.TargetID, _ = uuid.Parse(targetID.String)
		}
		if ipAddr.Valid {
			l.IPAddress = ipAddr.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *service) LogAdminAction(ctx context.Context, adminID uuid.UUID, action string, targetID uuid.UUID, details string, ip string) error {
	query := `
		INSERT INTO admin_audit_logs (admin_id, action, target_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query, adminID, action, targetID, details, ip)
	return err
}

func (s *service) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{}

	// 1. User counts
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notification_preferences WHERE plan_tier = 'pro'").Scan(&stats.ProUsers)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notification_preferences WHERE plan_tier = 'enterprise'").Scan(&stats.EnterpriseUsers)

	// 2. Production stats
	_ = s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(actual_kwh), 0) FROM actual_daily").Scan(&stats.TotalKwh)

	// 3. REC/Profile stats
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM solar_profiles").Scan(&stats.TotalProfiles)

	return stats, nil
}

func (s *service) GenerateImpersonationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	u, err := s.userSvc.GetUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("user not found for impersonation: %w", err)
	}

	now := time.Now().UTC()
	exp := now.Add(time.Duration(s.tokenExpiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub":  u.ID.String(),
		"role": u.Role,
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
		"imp":  true, // Mark as impersonated
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign impersonation token: %w", err)
	}

	return signed, nil
}

func (s *service) GetExpiringSubscriptions(ctx context.Context, days int) ([]billing.Subscription, error) {
	query := `
		SELECT id, user_id, plan_tier, status, billing_cycle, amount, currency, external_checkout_id, expires_at, next_billing_at, last_payment_at, grace_period_until, created_at, updated_at
		FROM subscriptions
		WHERE expires_at > NOW() AND expires_at <= NOW() + INTERVAL '1 day' * $1
		ORDER BY expires_at ASC
	`
	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("query expiring subscriptions: %w", err)
	}
	defer rows.Close()

	subs := []billing.Subscription{}
	for rows.Next() {
		var sub billing.Subscription
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.PlanTier, &sub.Status, &sub.BillingCycle, &sub.Amount, &sub.Currency, &sub.ExternalCheckoutID,
			&sub.ExpiresAt, &sub.NextBillingAt, &sub.LastPaymentAt, &sub.GracePeriodUntil, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *service) GetSchedulerRuns(ctx context.Context, limit int) ([]SchedulerRun, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, job_name, status, duration_ms, error_message, started_at, finished_at
		FROM scheduler_runs
		ORDER BY finished_at DESC
		LIMIT $1
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query scheduler runs: %w", err)
	}
	defer rows.Close()

	var runs []SchedulerRun
	for rows.Next() {
		var r SchedulerRun
		var errMsg sql.NullString
		err := rows.Scan(&r.ID, &r.JobName, &r.Status, &r.DurationMs, &errMsg, &r.StartedAt, &r.FinishedAt)
		if err != nil {
			return nil, fmt.Errorf("scan scheduler run: %w", err)
		}
		if errMsg.Valid {
			r.ErrorMessage = errMsg.String
		}
		runs = append(runs, r)
	}
	return runs, nil
}

func (s *service) GetForecastQuality(ctx context.Context) ([]ForecastQuality, error) {
	query := `
		WITH accuracy_calc AS (
			SELECT 
				f.solar_profile_id,
				sp.site_name,
				u.email as user_email,
				ABS(f.predicted_kwh - a.actual_kwh) / NULLIF(a.actual_kwh, 0) as error_rate
			FROM forecasts f
			JOIN actual_daily a ON f.user_id = a.user_id AND f.date = a.date
			JOIN solar_profiles sp ON f.solar_profile_id = sp.id
			JOIN users u ON f.user_id = u.id
			WHERE f.date >= NOW() - INTERVAL '7 days'
		)
		SELECT 
			solar_profile_id,
			site_name,
			user_email,
			AVG(error_rate) * 100 as mape,
			COUNT(*) as sample_count
		FROM accuracy_calc
		GROUP BY solar_profile_id, site_name, user_email
		HAVING COUNT(*) >= 3
		ORDER BY mape DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query forecast quality: %w", err)
	}
	defer rows.Close()

	var results []ForecastQuality
	for rows.Next() {
		var q ForecastQuality
		if err := rows.Scan(&q.ProfileID, &q.SiteName, &q.UserEmail, &q.MAPE, &q.SampleCount); err != nil {
			return nil, err
		}
		if q.MAPE <= 10 {
			q.Status = "excellent"
		} else if q.MAPE <= 30 {
			q.Status = "good"
		} else {
			q.Status = "poor"
		}
		results = append(results, q)
	}
	return results, nil
}

func (s *service) GetColdStartSites(ctx context.Context) ([]ColdStartSite, error) {
	query := `
		SELECT 
			sp.id,
			sp.site_name,
			u.email,
			COUNT(a.id) as actual_days,
			COALESCE(np.plan_tier, 'free') as plan_tier,
			sp.created_at
		FROM solar_profiles sp
		JOIN users u ON sp.user_id = u.id
		LEFT JOIN actual_daily a ON sp.user_id = a.user_id
		LEFT JOIN notification_preferences np ON sp.user_id = np.user_id
		WHERE sp.created_at <= NOW() - INTERVAL '30 days'
		GROUP BY sp.id, sp.site_name, u.email, np.plan_tier, sp.created_at
		HAVING COUNT(a.id) >= 30
		ORDER BY sp.created_at ASC
	`
	// Note: We don't directly track if baseline is synthetic in solar_profiles, 
	// but we can check the weather_baselines or just assume if it's old and has actuals, 
	// it should have shifted to site/blended.
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query cold start sites: %w", err)
	}
	defer rows.Close()

	var results []ColdStartSite
	for rows.Next() {
		var c ColdStartSite
		var tier string
		if err := rows.Scan(&c.ProfileID, &c.SiteName, &c.UserEmail, &c.ActualDays, &tier, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CurrentBaseline = "checking..." // Placeholder
		results = append(results, c)
	}
	return results, nil
}

func (s *service) GetNotificationLogs(ctx context.Context, limit int) ([]NotificationLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT nl.id, nl.user_id, u.email, nl.channel, nl.status, nl.sent_at, nl.error_message
		FROM notification_logs nl
		JOIN users u ON nl.user_id = u.id
		ORDER BY nl.sent_at DESC
		LIMIT $1
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query notification logs: %w", err)
	}
	defer rows.Close()

	var logs []NotificationLog
	for rows.Next() {
		var l NotificationLog
		var errMsg sql.NullString
		if err := rows.Scan(&l.ID, &l.UserID, &l.UserEmail, &l.Channel, &l.Status, &l.SentAt, &errMsg); err != nil {
			return nil, err
		}
		if errMsg.Valid {
			l.ErrorMessage = errMsg.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *service) GetDataAnomalies(ctx context.Context) ([]DataAnomaly, error) {
	// Detect actual > 1.5x predicted
	query := `
		SELECT 
			f.solar_profile_id,
			sp.site_name,
			f.date,
			f.predicted_kwh,
			a.actual_kwh,
			a.actual_kwh / NULLIF(f.predicted_kwh, 0) as ratio
		FROM forecasts f
		JOIN actual_daily a ON f.user_id = a.user_id AND f.date = a.date
		JOIN solar_profiles sp ON f.solar_profile_id = sp.id
		WHERE f.date >= NOW() - INTERVAL '7 days'
		AND a.actual_kwh > f.predicted_kwh * 1.5
		AND a.actual_kwh > 5 -- Ignore very small variations
		ORDER BY f.date DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []DataAnomaly
	for rows.Next() {
		var d DataAnomaly
		if err := rows.Scan(&d.ProfileID, &d.SiteName, &d.Date, &d.Predicted, &d.Actual, &d.Ratio); err != nil {
			return nil, err
		}
		d.AnomalyType = "high_actual"
		anomalies = append(anomalies, d)
	}
	return anomalies, nil
}

func (s *service) GetAggregateAnalytics(ctx context.Context, days int) ([]AggregateStats, error) {
	if days <= 0 {
		days = 30
	}
	query := `
		SELECT 
			f.date,
			SUM(a.actual_kwh) as total_actual,
			SUM(f.predicted_kwh) as total_predicted
		FROM forecasts f
		JOIN actual_daily a ON f.user_id = a.user_id AND f.date = a.date
		WHERE f.date >= NOW() - INTERVAL '1 day' * $1
		GROUP BY f.date
		ORDER BY f.date ASC
	`
	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("query aggregate analytics: %w", err)
	}
	defer rows.Close()

	var stats []AggregateStats
	for rows.Next() {
		var st AggregateStats
		if err := rows.Scan(&st.Date, &st.TotalActual, &st.TotalPredicted); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, nil
}

func (s *service) GetSiteRankings(ctx context.Context, limit int) ([]SiteRanking, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT 
			sp.id,
			sp.site_name,
			u.email,
			AVG(a.actual_kwh) as avg_actual,
			AVG(ABS(f.predicted_kwh - a.actual_kwh) / NULLIF(a.actual_kwh, 0)) * 100 as avg_mape
		FROM forecasts f
		JOIN actual_daily a ON f.user_id = a.user_id AND f.date = a.date
		JOIN solar_profiles sp ON f.solar_profile_id = sp.id
		JOIN users u ON f.user_id = u.id
		WHERE f.date >= NOW() - INTERVAL '30 days'
		GROUP BY sp.id, sp.site_name, u.email
		HAVING COUNT(*) >= 5
		ORDER BY avg_mape ASC
		LIMIT $1
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query site rankings: %w", err)
	}
	defer rows.Close()

	var rankings []SiteRanking
	for rows.Next() {
		var r SiteRanking
		if err := rows.Scan(&r.ProfileID, &r.SiteName, &r.UserEmail, &r.AvgActual, &r.AvgMAPE); err != nil {
			return nil, err
		}
		rankings = append(rankings, r)
	}
	return rankings, nil
}

func (s *service) GetTierDistribution(ctx context.Context) ([]TierDistribution, error) {
	query := `
		WITH total AS (SELECT COUNT(*) as cnt FROM users),
		tiers AS (
			SELECT COALESCE(plan_tier, 'free') as tier, COUNT(*) as count
			FROM users u
			LEFT JOIN notification_preferences np ON u.id = np.user_id
			GROUP BY 1
		)
		SELECT 
			tier, 
			count,
			(count::float / NULLIF((SELECT cnt FROM total), 0)) * 100 as pct
		FROM tiers
		ORDER BY count DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tier distribution: %w", err)
	}
	defer rows.Close()

	var dist []TierDistribution
	for rows.Next() {
		var t TierDistribution
		if err := rows.Scan(&t.Tier, &t.Count, &t.Pct); err != nil {
			return nil, err
		}
		dist = append(dist, t)
	}
	return dist, nil
}
