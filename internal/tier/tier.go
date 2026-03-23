package tier

import "fmt"

const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// ProfileLimit defines how many solar profiles a user can create.
var ProfileLimit = map[string]int{
	PlanFree:       1,
	PlanPro:        5,
	PlanEnterprise: -1, // -1 means unlimited
}

// DeviceLimit defines how many field devices a user can register.
var DeviceLimit = map[string]int{
	PlanFree:       1,
	PlanPro:        10,
	PlanEnterprise: -1,
}

// HistoryDaysLimit defines how many days of history a user can see.
var HistoryDaysLimit = map[string]int{
	PlanFree:       7,
	PlanPro:        90,
	PlanEnterprise: -1,
}

// Features that can be gated
const (
	FeatureTelegramNotif = "telegram_notif"
	FeatureWhatsAppNotif = "whatsapp_notif"
	FeatureCSVExport     = "csv_export"
	FeatureMonthlyPDF    = "monthly_pdf"
	FeatureAnnualPDF     = "annual_pdf"
	FeatureRECPDF        = "rec_pdf"
	FeatureMRVPDF        = "mrv_pdf"
	FeatureESGDashboard  = "esg_dashboard"
	FeatureWhiteLabel    = "white_label"
	FeatureAPIAccess     = "api_access"
	FeatureGreenReport    = "green_report"
)

// CanAccess returns true if the given tier can access the given feature.
func CanAccess(userTier, feature string) bool {
	access := map[string][]string{
		FeatureTelegramNotif: {PlanPro, PlanEnterprise},
		FeatureWhatsAppNotif: {PlanPro, PlanEnterprise},
		FeatureCSVExport:     {PlanPro, PlanEnterprise},
		FeatureMonthlyPDF:    {PlanPro, PlanEnterprise},
		FeatureAnnualPDF:     {PlanPro, PlanEnterprise},
		FeatureRECPDF:        {PlanPro, PlanEnterprise},
		FeatureMRVPDF:        {PlanPro, PlanEnterprise},
		FeatureESGDashboard:  {PlanEnterprise},
		FeatureWhiteLabel:    {PlanEnterprise},
		FeatureAPIAccess:     {PlanPro, PlanEnterprise},
		FeatureGreenReport:    {PlanPro, PlanEnterprise},
	}

	allowed, ok := access[feature]
	if !ok {
		return true // public feature
	}

	for _, t := range allowed {
		if userTier == t {
			return true
		}
	}
	return false
}

// LimitError is returned when a tier limit is reached.
type LimitError struct {
	Feature string `json:"feature"`
	Current int    `json:"current"`
	Limit   int    `json:"limit"`
	Tier    string `json:"tier"`
	Message string `json:"message"`
}

func (e *LimitError) Error() string {
	return e.Message
}

func NewLimitError(feature string, current, limit int, tier string) *LimitError {
	return &LimitError{
		Feature: feature,
		Current: current,
		Limit:   limit,
		Tier:    tier,
		Message: fmt.Sprintf("Paket %s hanya mendukung %d %s. Upgrade untuk menambah lebih banyak.", tier, limit, feature),
	}
}
