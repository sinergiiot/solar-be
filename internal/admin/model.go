package admin

import (
	"context"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/billing"
	"github.com/google/uuid"
)

type AdminUserRow struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	PlanTier      string     `json:"plan_tier"`
	PlanExpiresAt *time.Time `json:"plan_expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ForecastQuality struct {
	ProfileID   uuid.UUID `json:"profile_id"`
	SiteName    string    `json:"site_name"`
	UserEmail   string    `json:"user_email"`
	MAPE        float64   `json:"mape"`          // Mean Absolute Percentage Error
	SampleCount int       `json:"sample_count"`  // Number of days compared
	Status      string    `json:"status"`        // 'excellent', 'good', 'poor' (>30%)
}

type ColdStartSite struct {
	ProfileID       uuid.UUID `json:"profile_id"`
	SiteName        string    `json:"site_name"`
	UserEmail       string    `json:"user_email"`
	ActualDays      int       `json:"actual_days"`
	CurrentBaseline string    `json:"current_baseline"`
	CreatedAt       time.Time `json:"created_at"`
}

type NotificationLog struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	Channel      string    `json:"channel"`
	Status       string    `json:"status"`
	SentAt       time.Time `json:"sent_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type DataAnomaly struct {
	ProfileID   uuid.UUID `json:"profile_id"`
	SiteName    string    `json:"site_name"`
	Date        time.Time `json:"date"`
	AnomalyType string    `json:"anomaly_type"` // 'high_actual', 'zero_streak', 'low_coverage'
	Predicted   float64   `json:"predicted"`
	Actual      float64   `json:"actual"`
	Ratio       float64   `json:"ratio"`
}

type SystemStats struct {
	TotalUsers      int     `json:"total_users"`
	ProUsers        int     `json:"pro_users"`
	EnterpriseUsers int     `json:"enterprise_users"`
	TotalKwh        float64 `json:"total_kwh"`
	TotalProfiles   int     `json:"total_profiles"`
}

type SchedulerRun struct {
	ID           int       `json:"id"`
	JobName      string    `json:"job_name"`
	Status       string    `json:"status"`
	DurationMs   int64     `json:"duration_ms"`
	ErrorMessage string    `json:"error_message,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
}

type AggregateStats struct {
	Date          time.Time `json:"date"`
	TotalActual   float64   `json:"total_actual"`
	TotalPredicted float64   `json:"total_predicted"`
}

type SiteRanking struct {
	ProfileID uuid.UUID `json:"profile_id"`
	SiteName  string    `json:"site_name"`
	UserEmail string    `json:"user_email"`
	AvgActual float64   `json:"avg_actual"`
	AvgMAPE   float64   `json:"avg_mape"`
}

type TierDistribution struct {
	Tier  string  `json:"tier"`
	Count int     `json:"count"`
	Pct   float64 `json:"pct"`
}

type WeatherHealth struct {
	Provider       string  `json:"provider"`
	AvgResponseTime float64 `json:"avg_response_time"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	TotalRequests  int     `json:"total_requests"`
	SuccessRate    float64 `json:"success_rate"`
}

type AuditLog struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"admin_id"`
	AdminEmail string    `json:"admin_email"`
	Action    string    `json:"action"`
	TargetID  uuid.UUID `json:"target_id"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	GetAllUsersWithTiers(ctx context.Context) ([]AdminUserRow, error)
	UpdateUserTier(ctx context.Context, userID uuid.UUID, newTier string, adminID uuid.UUID, ip string) error
	GetSystemStats(ctx context.Context) (*SystemStats, error)
	GenerateImpersonationToken(ctx context.Context, userID uuid.UUID) (string, error)
	GetExpiringSubscriptions(ctx context.Context, days int) ([]billing.Subscription, error)
	GetSchedulerRuns(ctx context.Context, limit int) ([]SchedulerRun, error)
	GetForecastQuality(ctx context.Context) ([]ForecastQuality, error)
	GetColdStartSites(ctx context.Context) ([]ColdStartSite, error)
	GetNotificationLogs(ctx context.Context, limit int) ([]NotificationLog, error)
	GetDataAnomalies(ctx context.Context) ([]DataAnomaly, error)

	// BI Analytics
	GetAggregateAnalytics(ctx context.Context, days int) ([]AggregateStats, error)
	GetSiteRankings(ctx context.Context, limit int) ([]SiteRanking, error)
	GetTierDistribution(ctx context.Context) ([]TierDistribution, error)

	// Sprint C Extension
	GetWeatherAPIHealth(ctx context.Context) ([]WeatherHealth, error)
	GetAuditLogs(ctx context.Context, limit int) ([]AuditLog, error)
	LogAdminAction(ctx context.Context, adminID uuid.UUID, action string, targetID uuid.UUID, details string, ip string) error
}
