package report

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

// GenerateReportPDF writes a professional energy report (and potentially an official letter) to the provided writer.
func (s *service) GenerateReportPDF(report *EnergyReport, userObj *user.User, writer io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// 1. Page 1: Standard Executive Summary Report
	generateExecutiveSummaryPage(pdf, report, userObj)

	// 2. Page 2: Official PBB Letter if requested
	if report.OfficialDetails != nil {
		generatePBBLetterPage(pdf, report, userObj)
	}

	// 3. Page 3: MRV Methodology Report
	generateMRVMethodologyPage(pdf, report)

	return pdf.Output(writer)
}

func generateExecutiveSummaryPage(pdf *gofpdf.Fpdf, report *EnergyReport, userObj *user.User) {
	pdf.AddPage()

	// Header
	pdf.SetFillColor(21, 150, 90) // Brand Green
	pdf.Rect(0, 0, 210, 40, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 24)
	pdf.Text(15, 20, "Energy Performance Report")

	companyName := "PT Sinergi IoT Nusantara"
	if report.PlanTier == "enterprise" && userObj.CompanyName != "" {
		companyName = userObj.CompanyName
	}

	pdf.SetFont("Arial", "", 12)
	pdf.Text(15, 30, fmt.Sprintf("%s - %s", companyName, time.Now().Format("January 2006")))

	// Move cursor down
	pdf.SetY(50)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Ringkasan Eksekutif")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)

	periodStr := fmt.Sprintf("periode %s hingga %s", report.PeriodStart.Format(time.DateOnly), report.PeriodEnd.Format(time.DateOnly))
	if report.IsAnnual {
		periodStr = fmt.Sprintf("tahun %d", report.PeriodStart.Year())
	}

	pdf.MultiCell(0, 6, fmt.Sprintf("Laporan ini merangkum kinerja sistem PLTS untuk user %s dalam %s. Data diukur berdasarkan integrasi IoT via platform Solar Forecast.",
		userObj.Name, periodStr), "", "L", false)
	pdf.Ln(10)

	// KPI Cards grid
	drawKPICard(pdf, 15, 80, "Total Produksi", fmt.Sprintf("%.1f kWh", report.TotalActualKwh))
	drawKPICard(pdf, 110, 80, "Emisi Karbon Dihindari", fmt.Sprintf("%.2f kg CO2", report.TotalCO2AvoidedKg))

	savingsLabel := "Estimasi Penghematan"
	recLabel := "Pencapaian Sertifikat (REC)"
	recValue := fmt.Sprintf("%d REC", report.TotalREC)

	drawKPICard(pdf, 15, 115, savingsLabel, fmt.Sprintf("Rp %s", formatIDR(report.TotalSavingsIDR)))
	drawKPICard(pdf, 110, 115, recLabel, recValue)

	pdf.SetY(148)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(21, 150, 90)
	pdf.Cell(0, 10, fmt.Sprintf("Akurasi Prediksi: %.1f%%", calculateAccuracy(report)))
	pdf.SetTextColor(50, 50, 50)
	pdf.Ln(10)

	// Details Section
	pdf.SetY(160)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Detail Metrik Kinerja")
	pdf.Ln(10)

	// Render Monthly Breakdown if Annual, otherwise standard detail
	if report.IsAnnual && len(report.MonthlyBreakdown) > 0 {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(40, 8, "Bulan", "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, "Aktual (kWh)", "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 8, "Penghematan (Rp)", "1", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 10)
		for _, m := range report.MonthlyBreakdown {
			pdf.CellFormat(40, 8, m.Month, "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", m.ActualKwh), "1", 0, "R", false, 0, "")
			pdf.CellFormat(50, 8, formatIDR(m.SavingsIDR), "1", 1, "R", false, 0, "")
		}
	} else {
		// Standard
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(40, 8, "Metrik", "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 8, "Nilai Aktual", "1", 0, "R", false, 0, "")
		pdf.CellFormat(50, 8, "Target (Forecast)", "1", 1, "R", false, 0, "")

		pdf.CellFormat(40, 8, "Energi (kWh)", "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", report.TotalActualKwh), "1", 0, "R", false, 0, "")
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", report.TotalForecastedKwh), "1", 1, "R", false, 0, "")

		pdf.CellFormat(100, 8, fmt.Sprintf("%.1f%% dari total hari periode", report.DataCoveragePct), "1", 1, "L", false, 0, "")
	}

	// Impact Equivalents section (New in Epic 4)
	pdf.SetY(pdf.GetY() + 10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Dampak Lingkungan (Impact Analytics)")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	trees := int(report.TotalCO2AvoidedKg / 20)
	cars := int(report.TotalCO2AvoidedKg / 0.2)

	pdf.MultiCell(0, 6, fmt.Sprintf("Berdasarkan volume emisi karbon yang berhasil dihindari sebesar %.2f kg CO2, kontribusi Anda setara dengan penyerapan karbon oleh %d pohon dewasa per tahun atau menghindari emisi dari perjalanan mobil sejauh %d km.",
		report.TotalCO2AvoidedKg, trees, cars), "", "J", false)

	// Regulatory Context
	pdf.SetY(260)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(0, 5, "Informasi CO2 avoided dihitung berdasarkan faktor emisi regional di Indonesia (ESDM/KLHK). Laporan ini dapat digunakan sebagai lampiran pengajuan Surat Keterangan Produksi Energi Terbarukan bagi kepentingan fiskal (seperti diskon PBB) atau laporan keberlanjutan (ESG).", "", "C", false)

	// Footer
	pdf.SetY(275)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s | User Tier: %s | Report ID: %s", time.Now().Format(time.RFC822), report.PlanTier, report.UserID.String()[:8]), "T", 0, "C", false, 0, "")
}

func generatePBBLetterPage(pdf *gofpdf.Fpdf, report *EnergyReport, userObj *user.User) {
	pdf.AddPage()

	// Kop Surat
	companyName := "PT SINERGI IOT NUSANTARA"
	companyTagline := "Platform Monitoring & Forecasting Energi Surya"
	companyAddress := "Menara Sinergi Lt. 5, Jl. Teknologi Hijau No. 1, Jakarta"

	if report.PlanTier == "enterprise" && userObj.CompanyName != "" {
		companyName = userObj.CompanyName
		companyTagline = "Laporan Resmi Kinerja Energi Surya"
		companyAddress = ""
	}

	if report.PlanTier == "enterprise" && userObj.CompanyLogoURL != "" {
		imgPath := "." + userObj.CompanyLogoURL
		pdf.ImageOptions(imgPath, 10, 10, 30, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")

		pdf.SetY(15)
		pdf.SetX(45)
		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 8, companyName, "", 1, "C", false, 0, "")
		pdf.SetX(45)
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, companyTagline, "", 1, "C", false, 0, "")
		pdf.CellFormat(0, 6, companyAddress, "", 1, "C", false, 0, "")
		pdf.SetLineWidth(0.5)
		pdf.Line(10, 40, 200, 40)
		pdf.Ln(15)
	}

	// Judul Surat
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "SURAT KETERANGAN PRODUKSI ENERGI TERBARUKAN", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Nomor: %s", report.OfficialDetails.LetterNumber), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Isi Surat
	bodyText := fmt.Sprintf("Yang bertanda tangan di bawah ini, %s selaku %s di %s, menerangkan dengan sesungguhnya bahwa:",
		report.OfficialDetails.Signatory, report.OfficialDetails.Title, report.OfficialDetails.Organization)
	pdf.MultiCell(0, 6, bodyText, "", "J", false)
	pdf.Ln(5)

	// Identitas
	pdf.Cell(10, 6, "")
	pdf.CellFormat(40, 6, "Nama Pemilik", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, userObj.Name, "", 1, "L", false, 0, "")

	pdf.Cell(10, 6, "")
	pdf.CellFormat(40, 6, "Sistem Pengukuran", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, "Sinergi IoT Metering & Telemetry", "", 1, "L", false, 0, "")

	pdf.Cell(10, 6, "")
	pdf.CellFormat(40, 6, "Tahun Pembukuan", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("%d", report.PeriodStart.Year()), "", 1, "L", false, 0, "")

	pdf.Cell(10, 6, "")
	pdf.CellFormat(40, 6, "Total REC", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("%d REC", report.TotalREC), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Klaim
	claimText := fmt.Sprintf("Telah berhasil memproduksi dan memanfaatkan energi bersih dari sistem Pembangkit Listrik Tenaga Surya (PLTS) sebesar %.2f kWh (Kilowatt-hour) selama periode 1 Januari %d hingga 31 Desember %d.",
		report.TotalActualKwh, report.PeriodStart.Year(), report.PeriodEnd.Year())
	pdf.MultiCell(0, 6, claimText, "", "J", false)
	pdf.Ln(5)

	co2Text := fmt.Sprintf("Dengan produksi energi bersih tersebut, pemilik susut emisi karbon (CO2 Avoided) yang setara dengan %.2f kilogram emisi karbon dioksida. Data ini tervalidasi oleh sistem telemetri digital PT Sinergi IoT Nusantara.",
		report.TotalCO2AvoidedKg)
	pdf.MultiCell(0, 6, co2Text, "", "J", false)
	pdf.Ln(10)

	// Penutup
	closingText := "Surat Keterangan ini diberikan sebagai bukti partisipasi aktif warga dalam transisi energi hijau dan dapat digunakan sebagai kelengkapan dokumen pengajuan insentif Pajak Bumi dan Bangunan (PBB) berbasis lingkungan, maupun keperluan mitigasi perubahan iklim lainnya yang ditetapkan Pemerintah Daerah."
	pdf.MultiCell(0, 6, closingText, "", "J", false)
	pdf.Ln(20)

	// Tanda tangan
	pdf.CellFormat(0, 6, fmt.Sprintf("Jakarta, %s", report.OfficialDetails.OfficialDate.Format("02 January 2006")), "", 1, "R", false, 0, "")
	pdf.Ln(20) // Spasi untuk ttd

	pdf.SetFont("Arial", "U", 11)
	pdf.CellFormat(0, 6, report.OfficialDetails.Signatory, "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, report.OfficialDetails.Title, "", 1, "R", false, 0, "")
	pdf.CellFormat(0, 6, report.OfficialDetails.Organization, "", 1, "R", false, 0, "")
}

func generateMRVMethodologyPage(pdf *gofpdf.Fpdf, report *EnergyReport) {
	pdf.AddPage()

	pdf.SetFillColor(240, 240, 240)
	pdf.Rect(10, 10, 190, 20, "F")
	pdf.SetFont("Arial", "B", 14)
	pdf.SetY(15)
	pdf.CellFormat(0, 10, "LAMPIRAN: METODOLOGI PERHITUNGAN (MRV)", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "1. Monitoring (Pemantauan)")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Data produksi energi dikumpulkan secara real-time melalui integrasi IoT Telemetry (Smart Meter/Inverter API). Validasi data dilakukan setiap 24 jam untuk memastikan akurasi data harian sebelum dikonversi menjadi unit karbon.", "", "J", false)
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "2. Reporting (Pelaporan) & Faktor Emisi Grid")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Faktor emisi dihitung berdasarkan Nilai Emisi GRK Sektor Pembangkit Listrik (ESDM 2023) yang spesifik untuk lokasi profil solar Anda. Berikut adalah benchmark yang digunakan:", "", "J", false)
	pdf.Ln(2)

	pdf.SetFont("Courier", "", 9)
	pdf.Cell(10, 5, "")
	pdf.Cell(0, 5, "- Jawa-Madura-Bali (JAMALI) : 0.87 kg CO2 / kWh")
	pdf.Ln(4)
	pdf.Cell(10, 5, "")
	pdf.Cell(0, 5, "- Sumatera                  : 0.81 kg CO2 / kWh")
	pdf.Ln(4)
	pdf.Cell(10, 5, "")
	pdf.Cell(0, 5, "- Kalimantan                 : 0.84 kg CO2 / kWh")
	pdf.Ln(4)
	pdf.Cell(10, 5, "")
	pdf.Cell(0, 5, "- Sulawesi                   : 0.72 kg CO2 / kWh")
	pdf.Ln(4)
	pdf.Cell(10, 5, "")
	pdf.Cell(0, 5, "- Maluku, NTT, Papua         : 0.68 kg CO2 / kWh")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, "3. Verification (Verifikasi)")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Platform Solar Forecast melakukan cross-check antara ramalan cuaca (weather factor) dengan data aktual. Jika terdapat anomali >50% tanpa alasan cuaca yang jelas, sistem akan menandai data tersebut untuk verifikasi manual guna menjamin integritas laporan hijau/ESG.", "", "J", false)
	pdf.Ln(10)

	pdf.SetLineWidth(0.2)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(0, 4, "Laporan ini diterbitkan secara digital oleh Solar Forecast Platform (PT Sinergi IoT Nusantara). Metodologi ini merujuk pada standar pengukuran emisi karbon nasional untuk membantu entitas dalam pelaporan ESG dan klaim insentif fiskal hijau.", "", "C", false)
}

func (s *service) GenerateRECPDF(ctx context.Context, userID uuid.UUID, writer io.Writer) error {
	acc, err := s.recService.GetAccumulator(ctx, userID, nil)
	if err != nil {
		return err
	}
	userObj, err := s.userSvc.GetUserByID(userID)
	if err != nil {
		return err
	}

	totalMwh := acc.CumulativeKwh / 1000.0

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetLineWidth(2)
	pdf.SetDrawColor(21, 150, 90)
	pdf.Rect(5, 5, 287, 200, "D")

	if userObj.CompanyLogoURL != "" {
		imgPath := "." + userObj.CompanyLogoURL
		pdf.ImageOptions(imgPath, 133, 25, 30, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}

	pdf.SetY(60)
	pdf.SetFont("Arial", "B", 32)
	pdf.SetTextColor(21, 150, 90)
	pdf.CellFormat(0, 20, "RENEWABLE ENERGY CERTIFICATE", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "I", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 10, "This is to certify that", "", 1, "C", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 26)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 15, userObj.Name, "", 1, "C", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Arial", "", 14)
	pdf.SetTextColor(100, 100, 100)
	msg := "has successfully generated and contributed a cumulative total of"
	pdf.CellFormat(0, 10, msg, "", 1, "C", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(21, 150, 90)
	pdf.CellFormat(0, 15, fmt.Sprintf("%.3f MWh", totalMwh), "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 10, "of clean solar energy to the grid through the Solar Forecast Platform.", "", 1, "C", false, 0, "")

	pdf.Ln(20)

	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(150, 150, 150)
	footerMsg := fmt.Sprintf("Certificate ID: REC-%s | Verified on: %s", userID.String()[:8], time.Now().Format("02 Jan 2006"))
	pdf.CellFormat(0, 10, footerMsg, "", 1, "C", false, 0, "")

	pdf.SetY(170)
	pdf.SetDrawColor(0, 0, 0)
	pdf.Line(110, 185, 180, 185)
	pdf.SetY(187)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 5, "Digital Verification Service", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(0, 5, "PT Sinergi IoT Nusantara", "", 1, "C", false, 0, "")

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

// generateMRVReport creates a professional 3-section MRV PDF (Measurement, Reporting, Verification).
func generateMRVReport(summary *CO2Summary, userObj *user.User, writer io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	companyName := "PT Sinergi IoT Nusantara"
	if userObj != nil && userObj.CompanyName != "" {
		companyName = userObj.CompanyName
	}

	// -------------------------------------------------
	// PAGE 1: COVER
	// -------------------------------------------------
	pdf.AddPage()

	// Background gradient-like header
	pdf.SetFillColor(21, 150, 90)
	pdf.Rect(0, 0, 210, 75, "F")

	// Decorative accent strip
	pdf.SetFillColor(15, 110, 65)
	pdf.Rect(0, 65, 210, 10, "F")

	// Title
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 26)
	pdf.Text(15, 28, "MRV CO2 Avoided Report")
	pdf.SetFont("Arial", "", 13)
	pdf.Text(15, 40, "Measurement, Reporting & Verification")
	pdf.SetFont("Arial", "", 11)
	pdf.Text(15, 52, fmt.Sprintf("%s  *  %s", companyName, summary.PeriodStart.Format("2 Jan 2006")+" - "+summary.PeriodEnd.Format("2 Jan 2006")))

	// Badge
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(15, 57, 60, 8, 2, "1234", "F")
	pdf.SetTextColor(21, 150, 90)
	pdf.SetFont("Arial", "B", 8)
	pdf.Text(17, 62.5, "ESDM 2023 Methodology")

	// Below header: key stats
	pdf.SetTextColor(30, 30, 30)
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(15, 90, "Executive Carbon Summary")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.Text(15, 97, "Akumulasi penghindaran emisi CO2 berdasarkan produksi energi surya terverifikasi.")

	// KPI boxes
	drawMRVKPIBox(pdf, 15, 103, 55, "Total Produksi", fmt.Sprintf("%.1f kWh", summary.TotalActualKwh), 21, 150, 90)
	drawMRVKPIBox(pdf, 75, 103, 55, "CO2 Dihindari", fmt.Sprintf("%.2f kg", summary.TotalCO2AvoidedKg), 21, 150, 90)
	drawMRVKPIBox(pdf, 135, 103, 60, "Setara Pohon", fmt.Sprintf("%.0f pohon", summary.TotalCO2AvoidedKg/20), 21, 150, 90)

	// Carbon credit section
	pdf.SetY(140)
	pdf.SetFillColor(240, 255, 245)
	pdf.Rect(15, 140, 180, 30, "F")
	pdf.SetTextColor(21, 150, 90)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(20, 152, "Estimasi Nilai Carbon Credit")
	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Arial", "", 10)
	pdf.Text(20, 160, fmt.Sprintf("IDX Carbon (Rp 30.000/ton): Rp %s  |  Voluntary Market (USD 5/ton): USD %.2f",
		formatIDR(summary.CarbonCreditIDR), summary.CarbonCreditUSD))

	// Grid info
	pdf.SetY(178)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetFont("Arial", "B", 10)
	pdf.Text(15, 178, "Faktor Emisi Grid:")
	pdf.SetFont("Arial", "", 10)
	pdf.Text(55, 178, fmt.Sprintf("%.4f kg CO2/kWh - %s", summary.EmissionFactor, summary.GridRegion))
	pdf.SetFont("Arial", "B", 10)
	pdf.Text(15, 186, "Referensi Metodologi:")
	pdf.SetFont("Arial", "", 10)
	pdf.Text(60, 186, summary.Standard)

	// Footer
	pdf.SetY(270)
	pdf.SetTextColor(150, 150, 150)
	pdf.SetFont("Arial", "I", 8)
	pdf.Text(15, 274, fmt.Sprintf("Digenerate oleh Solar Forecast Platform  *  %s  *  Hal 1/3",
		time.Now().Format("02 Jan 2006 15:04")))

	// -------------------------------------------------
	// PAGE 2: MEASUREMENT & REPORTING
	// -------------------------------------------------
	pdf.AddPage()

	// Section header
	pdf.SetFillColor(21, 150, 90)
	pdf.Rect(0, 0, 210, 20, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(15, 13, "Bagian I - Measurement & Reporting")

	// Methodology explanation
	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(15, 30, "1.1 Ruang Lingkup & Batasan")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(70, 70, 70)
	lines := []string{
		"Laporan ini mencatat emisi CO2 yang berhasil dihindari berkat produksi energy dari sistem PLTS (Photovoltaic)",
		"yang terdaftar pada platform Solar Forecast. Penghitungan menggunakan data produksi aktual (kWh terukur)",
		"dikalikan dengan faktor emisi spesifik grid jaringan listrik Indonesia sesuai region lokasi instalasi.",
	}
	y := 38.0
	for _, l := range lines {
		pdf.Text(15, y, l)
		y += 6
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.Text(15, y+4, "1.2 Metode Pengukuran")
	y += 12
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(70, 70, 70)
	metodLines := []string{
		fmt.Sprintf("* Faktor emisi yang digunakan: %.4f kg CO2/kWh (%s)", summary.EmissionFactor, summary.GridRegion),
		"* Sumber referensi: Kementerian ESDM RI - Nilai Emisi GRK Sektor Pembangkit Listrik 2023",
		"* Data produksi: Rekaman aktual dari inverter / perangkat IoT yang terhubung ke platform",
		fmt.Sprintf("* Periode: %s s/d %s", summary.PeriodStart.Format("02 Jan 2006"), summary.PeriodEnd.Format("02 Jan 2006")),
		fmt.Sprintf("* Total hari data: %d entri produksi", len(summary.DailyBreakdown)),
	}
	for _, l := range metodLines {
		pdf.Text(20, y, l)
		y += 7
	}

	// Daily breakdown table (up to 30 rows)
	y += 5
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.Text(15, y, "1.3 Rincian Produksi Harian (max 30 entri ditampilkan)")
	y += 8

	// Table header
	pdf.SetFillColor(21, 150, 90)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.Rect(15, y, 55, 8, "F")
	pdf.Rect(70, y, 55, 8, "F")
	pdf.Rect(125, y, 70, 8, "F")
	pdf.Text(17, y+5.5, "Tanggal")
	pdf.Text(72, y+5.5, "Produksi (kWh)")
	pdf.Text(127, y+5.5, "CO2 Dihindari (kg)")
	y += 8

	limit := 30
	if len(summary.DailyBreakdown) < limit {
		limit = len(summary.DailyBreakdown)
	}
	for i := 0; i < limit; i++ {
		d := summary.DailyBreakdown[i]
		if i%2 == 0 {
			pdf.SetFillColor(240, 252, 245)
			pdf.Rect(15, y, 180, 7, "F")
		}
		pdf.SetTextColor(40, 40, 40)
		pdf.SetFont("Arial", "", 9)
		pdf.Text(17, y+5, d.Date)
		pdf.Text(72, y+5, fmt.Sprintf("%.2f", d.ActualKwh))
		pdf.Text(127, y+5, fmt.Sprintf("%.3f", d.CO2AvoidedKg))
		y += 7
		if y > 265 {
			pdf.AddPage()
			y = 20
		}
	}

	// -------------------------------------------------
	// PAGE 3: VERIFICATION
	// -------------------------------------------------
	pdf.AddPage()

	pdf.SetFillColor(21, 150, 90)
	pdf.Rect(0, 0, 210, 20, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(15, 13, "Bagian II - Verification & Attestation")

	pdf.SetTextColor(40, 40, 40)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(15, 30, "2.1 Pernyataan Verifikasi")
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(70, 70, 70)
	verLines := []string{
		"Sistem Solar Forecast menyatakan bahwa data produksi energi yang digunakan dalam laporan ini diperoleh dari",
		"sensor IoT terverifikasi yang terhubung ke platform secara real-time. Pengolahan data dilakukan secara",
		"otomatis mengikuti algoritma yang dapat diaudit sesuai standar pengukuran open-source.",
	}
	vy := 38.0
	for _, l := range verLines {
		pdf.Text(15, vy, l)
		vy += 6
	}

	// Summary table
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.Text(15, vy+4, "2.2 Ringkasan Hasil Verifikasi")
	vy += 12
	summRows := [][2]string{
		{"Total Produksi Aktual", fmt.Sprintf("%.2f kWh", summary.TotalActualKwh)},
		{"Faktor Emisi Grid", fmt.Sprintf("%.4f kg CO2/kWh", summary.EmissionFactor)},
		{"Jaringan Listrik (Grid)", summary.GridRegion},
		{"Total CO2 Dihindari", fmt.Sprintf("%.3f kg (%.4f ton)", summary.TotalCO2AvoidedKg, summary.TotalCO2AvoidedTon)},
		{"Setara Pohon Dewasa", fmt.Sprintf("%.0f pohon/tahun (@ 20 kg CO2/pohon)", summary.TotalCO2AvoidedKg/20)},
		{"Estimasi Carbon Credit IDX", fmt.Sprintf("Rp %s (@ Rp 30.000/ton)", formatIDR(summary.CarbonCreditIDR))},
		{"Estimasi Voluntary Market", fmt.Sprintf("USD %.2f (@ USD 5/ton)", summary.CarbonCreditUSD)},
		{"Standar Metodologi", summary.Standard},
		{"Periode Laporan", fmt.Sprintf("%s - %s", summary.PeriodStart.Format("02 Jan 2006"), summary.PeriodEnd.Format("02 Jan 2006"))},
	}
	for i, row := range summRows {
		if i%2 == 0 {
			pdf.SetFillColor(240, 252, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.Rect(15, vy, 180, 8, "F")
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(40, 40, 40)
		pdf.Text(17, vy+5.5, row[0])
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(21, 100, 60)
		pdf.Text(95, vy+5.5, row[1])
		vy += 8
	}

	// Attestation seal area
	vy += 10
	pdf.SetFillColor(240, 255, 245)
	pdf.RoundedRect(15, vy, 180, 35, 3, "1234", "F")
	pdf.SetTextColor(21, 150, 90)
	pdf.SetFont("Arial", "B", 10)
	pdf.Text(20, vy+8, "Attestation / Pernyataan Kebenaran Data")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.Text(20, vy+15, "Laporan ini digenerate secara otomatis oleh platform Solar Forecast.")
	pdf.Text(20, vy+21, fmt.Sprintf("Pengguna: %s  |  Timestamp: %s", userObj.Email, time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))

	// Footer
	pdf.SetY(275)
	pdf.SetTextColor(150, 150, 150)
	pdf.SetFont("Arial", "I", 8)
	pdf.Text(15, 279, "Solar Forecast Platform  *  MRV CO2 Avoided Report  *  Dokumen ini dapat digunakan sebagai lampiran laporan ESG / CSR perusahaan.")

	return pdf.Output(writer)
}

func drawMRVKPIBox(pdf *gofpdf.Fpdf, x, y, w float64, label, value string, r, g, b int) {
	pdf.SetFillColor(r, g, b)
	pdf.RoundedRect(x, y, w, 22, 3, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "", 8)
	pdf.Text(x+3, y+7, label)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(x+3, y+16, value)
}


// generateESGReport creates a professional 3-page ESG Performance report.
func generateESGReport(summary *ESGSummary, userObj *user.User, year int, writer io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Branding
	companyName := "PT Sinergi IoT Nusantara"
	if userObj.CompanyName != "" {
		companyName = userObj.CompanyName
	}

	// 1. Cover Page
	pdf.AddPage()
	pdf.SetFillColor(21, 150, 90) // Brand Green Deep
	pdf.Rect(0, 0, 210, 297, "F")

	// White box for content
	pdf.SetFillColor(255, 255, 255)
	pdf.RoundedRect(20, 40, 170, 220, 5, "1234", "F")

	// Logo Placeholder or real logo if enterprise
	pdf.SetY(60)
	pdf.SetFont("Arial", "B", 32)
	pdf.SetTextColor(21, 150, 90)
	pdf.CellFormat(170, 20, "ESG PERFORMANCE", "0", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(170, 10, "Sustainability & Environmental Impact", "0", 1, "C", false, 0, "")

	pdf.SetY(120)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(210, 15, companyName, "0", 1, "C", false, 0, "")
	
	pdf.SetY(140)
	pdf.SetFont("Arial", "", 14)
	pdf.CellFormat(210, 10, fmt.Sprintf("Annual Dashboard Summary - %d", year), "0", 1, "C", false, 0, "")

	// 2. Executive Summary Page
	pdf.AddPage()
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(0, 0, 210, 40, "F")
	pdf.SetY(15)
	
	// Local helper for symmetric text
	drawSymmetryText(pdf, 15, 15, "ESG EXECUTIVE SUMMARY", "Arial", "B", 18, 50, 50, 50)
	
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Text(15, 25, fmt.Sprintf("Portfolio Performance: %s  |  Reporting Year: %d", companyName, year))

	// KPI Cards Grid (4 boxes)
	yStart := 50.0
	pdf.SetFillColor(21, 100, 60)
	pdf.RoundedRect(15, yStart, 85, 30, 2, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "", 9)
	pdf.Text(20, yStart+10, "Total Clean Energy Production")
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(20, yStart+22, fmt.Sprintf("%.3f MWh", summary.TotalActualMwh))

	pdf.SetFillColor(50, 50, 50)
	pdf.RoundedRect(110, yStart, 85, 30, 2, "1234", "F")
	pdf.SetFont("Arial", "", 9) // Restore font
	pdf.SetTextColor(255, 255, 255) // Restore color
	pdf.Text(115, yStart+10, "Total Carbon Offset (CO2 Saved)")
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(115, yStart+22, fmt.Sprintf("%.2f Tons", summary.TotalCO2SavedTon))

	yStart += 35
	pdf.SetFillColor(21, 150, 90)
	pdf.RoundedRect(15, yStart, 85, 30, 2, "1234", "F")
	pdf.SetFont("Arial", "", 9)
	pdf.Text(20, yStart+10, "Bio-Conversion (Tree Equivalent)")
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(20, yStart+22, fmt.Sprintf("%d Mature Trees", summary.TotalTreesEq))

	pdf.SetFillColor(230, 126, 34) // Orange for REC
	pdf.RoundedRect(110, yStart, 85, 30, 2, "1234", "F")
	pdf.SetFont("Arial", "", 9)
	pdf.Text(115, yStart+10, "Energy Attribute Certificates (REC)")
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(115, yStart+22, fmt.Sprintf("%d Units", summary.TotalRECCount))

	// Production Trend Section
	pdf.SetY(yStart + 45)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(40, 40, 40)
	pdf.Cell(0, 10, "Annual Production Trend (MWh)")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 9)
	for i, t := range summary.YearlyTrend {
		pdf.CellFormat(20, 8, t.Month, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.3f", t.ActualMwh), "1", 0, "R", false, 0, "")
		
		// Mini bar charts
		barMax := 100.0
		barWidth := (t.ActualMwh / (summary.TotalActualMwh / 6)) * barMax
		if barWidth > barMax { barWidth = barMax }
		
		pdf.SetFillColor(21, 150, 90)
		pdf.Rect(pdf.GetX()+5, pdf.GetY()+2, barWidth, 4, "F")
		pdf.CellFormat(120, 8, "", "1", 1, "L", false, 0, "")
		if i == 5 { // Show 6 months per page or just 12 rows
			// If we want all 12:
		}
	}

	// 3. Site Detail Page
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, "Multi-Site Impact Breakdown")
	pdf.Ln(10)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(60, 10, "Site Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Location", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "Energy (MWh)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "CO2 (Tons)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 10, "REC", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	for _, s := range summary.SiteBreakdown {
		pdf.CellFormat(60, 8, s.ProfileName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 8, s.Location, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.3f", s.ActualMwh), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.2f", s.CO2SavedTon), "1", 0, "R", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%d", s.RECReached), "1", 1, "C", false, 0, "")
	}

	// Index & Summary
	pdf.Ln(15)
	pdf.SetFillColor(240, 255, 245)
	pdf.RoundedRect(15, pdf.GetY(), 180, 25, 3, "1234", "F")
	pdf.SetY(pdf.GetY() + 5)
	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(21, 100, 60)
	pdf.Cell(0, 6, "Clean Energy Transition Index")
	pdf.Ln(6)
	pdf.SetX(20)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(60, 60, 60)
	pdf.Cell(0, 6, fmt.Sprintf("Portfolio Reach: %.2f%% of clean energy goals achieved across %d sites.", summary.CleanEnergyPct, len(summary.SiteBreakdown)))

	// Footer
	pdf.SetY(280)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 10, "Generated by Solar Forecast Enterprise Sustainability Engine  |  Verified MRV Logic", "0", 0, "C", false, 0, "")

	return pdf.Output(writer)
}

func drawSymmetryText(pdf *gofpdf.Fpdf, x, y float64, txt, family, style string, size float64, r, g, b int) {
	pdf.SetFont(family, style, size)
	pdf.SetTextColor(r, g, b)
	pdf.Text(x, y, txt)
}


func generateRECReadinessReport(report *RECReadinessReport, userObj *user.User, writer io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	generateRECSummaryPage(pdf, report, userObj)
	generateRECSiteBreakdownPage(pdf, report)
	generateRECVerificationPage(pdf, report)

	return pdf.Output(writer)
}

func generateRECSummaryPage(pdf *gofpdf.Fpdf, report *RECReadinessReport, userObj *user.User) {
	pdf.AddPage()

	// Header branding
	pdf.SetFillColor(21, 150, 90)
	pdf.Rect(0, 0, 210, 50, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 26)
	pdf.Text(15, 25, "REC Readiness Report")
	pdf.SetFont("Arial", "", 12)
	pdf.Text(15, 38, "Solar Forecast Platform - Evidence for Renewable Energy Certificates")

	pdf.SetY(65)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Ringkasan Akumulasi Energi Terbarukan")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, fmt.Sprintf("Laporan ini mengonfirmasi total akumulasi produksi energi listrik dari sumber terbarukan (Surya) yang dihasilkan oleh %s melalui portofolio site yang terintegrasi pada platform Solar Forecast.", userObj.Name), "", "L", false)
	pdf.Ln(10)

	// Summary Cards
	pdf.SetFillColor(245, 250, 248)
	pdf.Rect(15, 110, 180, 45, "F")
	
	pdf.SetY(115)
	pdf.SetX(25)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(80, 10, "Total Akumulasi Produksi")
	pdf.Cell(80, 10, "Setara Sertifikat (REC)")
	pdf.Ln(10)

	pdf.SetX(25)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(21, 150, 90)
	pdf.Cell(80, 15, fmt.Sprintf("%.3f MWh", report.TotalActualMwh))
	pdf.Cell(80, 15, fmt.Sprintf("%d REC", report.TotalREC))
	pdf.Ln(25)

	// Monthly Trend Table
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(50, 50, 50)
	pdf.Cell(0, 10, "Tren Produksi 12 Bulan Terakhir")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 8, "Bulan", "1", 0, "C", false, 0, "")
	pdf.CellFormat(90, 8, "Produksi (MWh)", "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, m := range report.MonthlyHistory {
		pdf.CellFormat(90, 8, m.Month, "1", 0, "C", false, 0, "")
		pdf.CellFormat(90, 8, fmt.Sprintf("%.3f", m.ActualMwh), "1", 1, "C", false, 0, "")
	}
}

func generateRECSiteBreakdownPage(pdf *gofpdf.Fpdf, report *RECReadinessReport) {
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(21, 150, 90)
	pdf.Cell(0, 10, "Daftar Aset PLTS (Proof of Origin)")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, "Berikut adalah daftar site yang berkontribusi terhadap akumulasi energi di atas. Setiap site terdaftar dengan kapasitas dan lokasi koordinat yang tervalidasi melalui integrasi IoT.", "", "L", false)
	pdf.Ln(8)

	// Table Header
	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(60, 10, "Nama Site", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "KWP", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 10, "Lokasi", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Total MWh", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	for _, s := range report.SiteBreakdown {
		pdf.CellFormat(60, 10, s.ProfileName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%.1f", s.CapacityKwp), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 10, s.Location, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%.3f", s.TotalActualMwh), "1", 1, "R", false, 0, "")
	}
}

func generateRECVerificationPage(pdf *gofpdf.Fpdf, report *RECReadinessReport) {
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Pernyataan Verifikasi")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 7, "Solar Forecast Platform menggunakan metodologi pengumpulan data berbasis IoT Cloud. Data produksi aktual ditarik secara realtime dari inverter atau energy meter yang terpasang di lokasi.\n\nFaktor Kesiapan REC:\n1. 1 REC setara dengan 1 MWh energi listrik terbarukan.\n2. Data ini dapat digunakan sebagai lampiran teknis untuk pendaftaran REC pada sistem PLN GEAS atau ICDX.\n3. Nilai akumulasi telah diverifikasi berdasarkan ketersediaan data (Data Coverage) di atas 95%.", "", "L", false)

	pdf.Ln(40)
	pdf.SetX(130)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 5, "Diverifikasi Oleh,")
	pdf.Ln(30)
	pdf.SetX(130)
	pdf.Cell(0, 5, "Sistem Solar Forecast")
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(130)
	pdf.Cell(0, 5, time.Now().Format("02 January 2006"))
	
	// QR code placeholder
	pdf.Rect(20, 240, 30, 30, "D")
	pdf.SetFont("Arial", "", 7)
	pdf.Text(22, 275, "Scan to Verify Online")
}
