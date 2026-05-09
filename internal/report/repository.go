package report

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	CreateReportHistory(h *ReportHistory) error
	GetReportHistoryByUserID(userID uuid.UUID) ([]*ReportHistory, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateReportHistory(h *ReportHistory) error {
	metadataJSON, _ := json.Marshal(h.Metadata)
	query := `
		INSERT INTO report_history (id, user_id, report_name, report_type, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, h.ID, h.UserID, h.ReportName, h.ReportType, metadataJSON, h.CreatedAt)
	if err != nil {
		return fmt.Errorf("create report history: %w", err)
	}
	return nil
}

func (r *repository) GetReportHistoryByUserID(userID uuid.UUID) ([]*ReportHistory, error) {
	query := `
		SELECT id, user_id, report_name, report_type, metadata, created_at
		FROM report_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("get report history: %w", err)
	}
	defer rows.Close()

	var history []*ReportHistory
	for rows.Next() {
		h := &ReportHistory{}
		var metadataRaw []byte
		if err := rows.Scan(&h.ID, &h.UserID, &h.ReportName, &h.ReportType, &metadataRaw, &h.CreatedAt); err != nil {
			return nil, err
		}
		if len(metadataRaw) > 0 {
			json.Unmarshal(metadataRaw, &h.Metadata)
		}
		history = append(history, h)
	}
	return history, nil
}
