package tier

const (
	Free       = "free"
	Pro        = "pro"
	Enterprise = "enterprise"
)

// ProfileLimit defines how many solar profiles a user can create.
var ProfileLimit = map[string]int{
	Free:       1,
	Pro:        5,
	Enterprise: -1, // -1 means unlimited
}

// DeviceLimit defines how many field devices a user can register.
var DeviceLimit = map[string]int{
	Free:       1,
	Pro:        10,
	Enterprise: -1,
}

// HistoryDaysLimit defines how many days of history a user can see.
var HistoryDaysLimit = map[string]int{
	Free:       7,
	Pro:        90,
	Enterprise: -1,
}

// CanAccess returns true if the given tier can access the given feature.
// Features: "telegram_notif", "whatsapp_notif", "csv_export",
//           "monthly_pdf", "annual_pdf", "rec_pdf", "mrv_pdf",
//           "esg_dashboard", "white_label", "api_access"
func CanAccess(userTier, feature string) bool {
	access := map[string][]string{
		"telegram_notif": {Pro, Enterprise},
		"whatsapp_notif": {Pro, Enterprise},
		"csv_export":     {Pro, Enterprise},
		"monthly_pdf":    {Pro, Enterprise},
		"annual_pdf":     {Pro, Enterprise},
		"rec_pdf":        {Pro, Enterprise},
		"mrv_pdf":        {Pro, Enterprise},
		"esg_dashboard":  {Enterprise},
		"white_label":    {Enterprise},
		"api_access":     {Pro, Enterprise},
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
