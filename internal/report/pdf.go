package report

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// GenerateReportPDF writes a professional energy report to the provided writer.
func (s *service) GenerateReportPDF(report *EnergyReport, userName string, writer io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFillColor(21, 150, 90) // Brand Green
	pdf.Rect(0, 0, 210, 40, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 24)
	pdf.Text(15, 20, "Energy Performance Report")
	
	pdf.SetFont("Arial", "", 12)
	pdf.Text(15, 30, fmt.Sprintf("PT Sinergi IoT Nusantara — %s", time.Now().Format("January 2006")))

	// Move cursor down
	pdf.SetY(50)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Ringkasan Eksekutif")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, fmt.Sprintf("Laporan ini merangkum kinerja sistem PLTS untuk user %s dalam periode %s hingga %s. Data diukur berdasarkan integrasi IoT via platform Solar Forecast.", 
		userName, report.PeriodStart.Format(time.DateOnly), report.PeriodEnd.Format(time.DateOnly)), "", "L", false)
	pdf.Ln(10)

	// KPI Cards grid (simulated by positioning)
	// kWh
	drawKPICard(pdf, 15, 80, "Total Produksi", fmt.Sprintf("%.1f kWh", report.TotalActualKwh))
	// CO2
	drawKPICard(pdf, 110, 80, "Emisi Karbon Dihindari", fmt.Sprintf("%.2f kg CO2", report.TotalCO2AvoidedKg))
	// Savings
	drawKPICard(pdf, 15, 115, "Estimasi Penghematan", fmt.Sprintf("Rp %s", formatIDR(report.TotalSavingsIDR)))
	// Accuracy
	drawKPICard(pdf, 110, 115, "Akurasi Prediksi", fmt.Sprintf("%.1f%%", calculateAccuracy(report)))

	// Details Section
	pdf.SetY(160)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Detail Metrik Kinerja")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(40, 8, "Metrik", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, "Nilai Aktual", "1", 0, "R", false, 0, "")
	pdf.CellFormat(50, 8, "Target (Forecast)", "1", 1, "R", false, 0, "")

	pdf.CellFormat(40, 8, "Energi (kWh)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", report.TotalActualKwh), "1", 0, "R", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", report.TotalForecastedKwh), "1", 1, "R", false, 0, "")

	pdf.CellFormat(40, 8, "Coverage Data", "1", 0, "L", false, 0, "")
	pdf.CellFormat(100, 8, fmt.Sprintf("%.1f%% dari total hari periode", report.DataCoveragePct), "1", 1, "L", false, 0, "")

	// Regulatory Context
	pdf.SetY(210)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(0, 5, "Informasi CO2 avoided dihitung berdasarkan faktor emisi regional di Indonesia (ESDM/KLHK). Laporan ini dapat digunakan sebagai lampiran pengajuan Surat Keterangan Produksi Energi Terbarukan bagi kepentingan fiskal (seperti diskon PBB) atau laporan keberlanjutan (ESG).", "", "C", false)

	// Footer
	pdf.SetY(270)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s | User Tier: %s | Report ID: %s", time.Now().Format(time.RFC822), report.PlanTier, report.UserID.String()[:8]), "T", 0, "C", false, 0, "")

	return pdf.Output(writer)
}

func drawKPICard(pdf *gofpdf.Fpdf, x, y float64, label, value string) {
	pdf.Rect(x, y, 90, 30, "D")
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Text(x+5, y+10, label)
	
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(21, 150, 90)
	pdf.Text(x+5, y+22, value)
	pdf.SetTextColor(50, 50, 50)
}

func formatIDR(v float64) string {
	// Simple comma separator
	s := fmt.Sprintf("%.0f", v)
	if len(s) <= 3 {
		return s
	}
	res := ""
	for i, c := range s {
		if (len(s)-i)%3 == 0 && i != 0 {
			res += "."
		}
		res += string(c)
	}
	return res
}

func calculateAccuracy(r *EnergyReport) float64 {
	if r.TotalActualKwh == 0 {
		return 0
	}
	diff := math.Abs(r.TotalActualKwh - r.TotalForecastedKwh)
	acc := 100 - (diff / r.TotalActualKwh * 100)
	if acc < 0 {
		return 0
	}
	return acc
}
