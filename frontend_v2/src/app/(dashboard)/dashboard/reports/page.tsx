"use client";

import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { reportService } from "@/services/report";
import { Card } from "@/components/ui/card";
import { 
  FileText, 
  Download, 
  TreeDeciduous, 
  Cloud, 
  Zap, 
  Car, 
  FileBarChart,
  Loader2,
  ChevronRight,
  Calendar
} from "lucide-react";
import { cn } from "@/lib/utils";

import { useToast } from "@/components/providers/toast-provider";
import { usePlan } from "@/components/providers/plan-provider";

export default function ReportsPage() {
  const { showToast } = useToast();
  const { checkAccess } = usePlan();
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear());
  const [selectedMonth, setSelectedMonth] = useState(new Date().toLocaleString('en-US', { month: 'long' }));
  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0],
    end: new Date().toISOString().split('T')[0]
  });
  const [isDownloading, setIsDownloading] = useState<string | null>(null);

  const months = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

  const { data: esgSummary } = useQuery({
    queryKey: ["esg-summary", selectedYear],
    queryFn: () => reportService.getESGSummary(selectedYear),
  });

  const { data: history, refetch: refetchHistory } = useQuery({
    queryKey: ["report-history"],
    queryFn: () => reportService.getReportHistory(),
  });

  const handleDownload = async (type: string, metadata?: any) => {
    const hasAccess = checkAccess(
      "pro", 
      "Report Exports", 
      "Professional PDF and CSV reporting is a Pro feature. Gain deeper insights into your solar production and ESG impact with automated document generation."
    );
    
    if (!hasAccess) return;

    setIsDownloading(type);
    showToast(`Preparing your ${type.replace('-', ' ')}...`, "info");
    try {
      let blob;
      let filename = "";

      if (type === "esg-pdf") {
        blob = await reportService.downloadESGReportPDF(selectedYear);
        filename = `ESG_Report_${selectedYear}.pdf`;
      } else if (type === "energy-pdf") {
        blob = await reportService.downloadEnergyReportPDF({ year: selectedYear, is_annual: true });
        filename = `Energy_Audit_${selectedYear}.pdf`;
      } else if (type === "history-csv") {
        blob = await reportService.downloadHistoryCSV({ start_date: dateRange.start, end_date: dateRange.end });
        filename = `Telemetry_${dateRange.start}_to_${dateRange.end}.csv`;
      } else if (type === "rec-pdf") {
        blob = await reportService.downloadRECPDF();
        filename = `REC_Readiness_${new Date().toISOString().split('T')[0]}.pdf`;
      } else if (type === "monthly-pdf") {
        const month = metadata?.month || selectedMonth;
        blob = await reportService.downloadMonthlySummaryPDF(month, selectedYear);
        filename = `Summary_${month}_${selectedYear}.pdf`;
      } else if (type === "site-audit-pdf") {
        blob = await reportService.downloadSiteAuditPDF(metadata?.solarProfileId);
        filename = `Site_Audit_${metadata?.solarProfileId?.substring(0, 8)}.pdf`;
      }

      if (blob) {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
        
        showToast("Report downloaded successfully!", "success");
        // Refresh history after a successful download
        setTimeout(refetchHistory, 1000);
      }
    } catch (error) {
      console.error("Download failed", error);
      showToast("Download failed. Higher tier subscription may be required.", "error");
    } finally {
      setIsDownloading(null);
    }
  };

  const impactStats = [
    { label: "CO2 Saved", value: `${(esgSummary?.total_co2_saved_ton || 0).toLocaleString()} Tons`, icon: Cloud, color: "text-blue-500", bg: "bg-blue-500/10" },
    { label: "Clean Energy Pct", value: `${(esgSummary?.clean_energy_pct || 0).toFixed(1)}%`, icon: Zap, color: "text-amber-500", bg: "bg-amber-500/10" },
    { label: "Tree Equivalent", value: `${esgSummary?.total_trees_eq?.toLocaleString() || 0} Trees`, icon: TreeDeciduous, color: "text-emerald-500", bg: "bg-emerald-500/10" },
    { label: "Total REC", value: `${esgSummary?.total_rec_count?.toLocaleString() || 0} Units`, icon: FileText, color: "text-purple-500", bg: "bg-purple-500/10" },
  ];

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight mb-2">Reports & Analytics</h1>
          <p className="text-muted-foreground">Detailed energy production and ESG impact summaries.</p>
        </div>
        
        <div className="flex flex-wrap items-center gap-4">
          {/* Year Selector */}
          <div className="flex items-center gap-2 bg-card border border-border p-1.5 rounded-2xl">
            {[2024, 2025, 2026].map(year => (
              <button
                key={year}
                onClick={() => setSelectedYear(year)}
                className={cn(
                  "px-4 py-2 text-sm font-bold rounded-xl transition-all",
                  selectedYear === year ? "bg-primary text-white shadow-premium" : "text-muted-foreground hover:bg-accent"
                )}
              >
                {year}
              </button>
            ))}
          </div>

          {/* Month Selector for Monthly Summary */}
          <select 
            value={selectedMonth}
            onChange={(e) => setSelectedMonth(e.target.value)}
            className="bg-card border border-border px-4 py-3 rounded-2xl text-sm font-bold focus:ring-2 focus:ring-primary outline-none cursor-pointer"
          >
            {months.map(m => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>

          {/* Date Range for CSV */}
          <div className="flex items-center gap-2 bg-card border border-border p-1.5 rounded-2xl">
            <div className="flex items-center px-3 gap-2 border-r border-border">
              <Calendar className="w-4 h-4 text-muted-foreground" />
              <input 
                type="date" 
                value={dateRange.start}
                onChange={(e) => setDateRange(prev => ({ ...prev, start: e.target.value }))}
                className="bg-transparent border-none text-xs font-bold focus:ring-0 outline-none w-28"
              />
            </div>
            <div className="flex items-center px-3 gap-2">
              <input 
                type="date" 
                value={dateRange.end}
                onChange={(e) => setDateRange(prev => ({ ...prev, end: e.target.value }))}
                className="bg-transparent border-none text-xs font-bold focus:ring-0 outline-none w-28"
              />
            </div>
          </div>
        </div>
      </div>

      {/* ESG Impact Visual Section */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <Card className="lg:col-span-4 bg-gradient-to-br from-emerald-500 to-teal-600 text-white border-none overflow-hidden relative">
          <div className="absolute top-0 right-0 w-96 h-96 bg-white/10 rounded-full blur-3xl -mr-20 -mt-20" />
          <div className="relative z-10 flex flex-col md:flex-row items-center gap-12">
            <div className="flex-1">
              <h2 className="text-3xl font-bold mb-4">ESG Impact Summary {selectedYear}</h2>
              <p className="text-emerald-50/80 text-lg max-w-xl mb-8 leading-relaxed">
                Your commitment to renewable energy has significantly reduced carbon emissions. 
                Keep monitoring to track your ongoing contribution to a greener planet.
              </p>
              <div className="flex gap-4">
                <button 
                  onClick={() => handleDownload("esg-pdf")}
                  disabled={!!isDownloading}
                  className="flex items-center gap-2 px-6 py-3 bg-white text-emerald-600 font-bold rounded-2xl shadow-xl hover:scale-105 transition-all"
                >
                  {isDownloading === "esg-pdf" ? <Loader2 className="w-5 h-5 animate-spin" /> : <Download className="w-5 h-5" />}
                  Download ESG PDF
                </button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 w-full md:w-auto">
              {impactStats.map((stat, idx) => (
                <div key={idx} className="p-6 rounded-3xl bg-white/10 backdrop-blur-md border border-white/10 text-center min-w-[160px]">
                  <stat.icon className="w-8 h-8 mx-auto mb-3 text-white/90" />
                  <p className="text-[10px] font-bold uppercase tracking-wider text-white/60 mb-1">{stat.label}</p>
                  <p className="text-xl font-bold">{stat.value}</p>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="md:col-span-2 space-y-6">
          <h2 className="text-xl font-bold">Standard Reports</h2>
          <div className="grid grid-cols-1 gap-4">
            {[
              { id: "energy-pdf", title: "Annual Energy Audit", desc: "Detailed breakdown of forecasted vs actual production across all sites.", icon: FileBarChart },
              { id: "monthly-pdf", title: "Monthly Summary", desc: "A snapshot of your production and savings for the current month.", icon: Calendar },
              { id: "history-csv", title: "Raw Telemetry Export", desc: "CSV export of all 15-minute interval data for custom analysis.", icon: FileText },
              { id: "rec-pdf", title: "REC Readiness Check", desc: "Validate if your production meets Renewable Energy Certificate standards.", icon: TreeDeciduous },
            ].map((report) => (
              <Card key={report.id} className="group hover:border-primary transition-all duration-300">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-5">
                    <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center text-muted-foreground group-hover:bg-primary/10 group-hover:text-primary transition-colors">
                      <report.icon className="w-7 h-7" />
                    </div>
                    <div>
                      <h3 className="font-bold text-lg">{report.title}</h3>
                      <p className="text-sm text-muted-foreground">{report.desc}</p>
                    </div>
                  </div>
                  <button 
                    onClick={() => handleDownload(report.id)}
                    disabled={!!isDownloading}
                    className="p-3 rounded-xl bg-accent/50 text-muted-foreground hover:text-primary hover:bg-primary/10 transition-all"
                  >
                    {isDownloading === report.id ? <Loader2 className="w-5 h-5 animate-spin" /> : <Download className="w-5 h-5" />}
                  </button>
                </div>
              </Card>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          <h2 className="text-xl font-bold">Recent Activity</h2>
          <Card className="p-0 overflow-hidden">
            <div className="divide-y divide-border">
              {history && history.length > 0 ? (
                history.slice(0, 10).map((item, idx) => (
                  <div 
                    key={idx} 
                    onClick={() => handleDownload(item.report_type.replace('_', '-'), item.metadata)}
                    className="p-5 flex items-center justify-between hover:bg-accent/5 transition-colors cursor-pointer group active:scale-[0.98]"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-600">
                        <FileText className="w-4 h-4" />
                      </div>
                      <div>
                        <p className="text-sm font-bold group-hover:text-primary transition-colors">{item.report_name}</p>
                        <p className="text-xs text-muted-foreground">{new Date(item.created_at).toLocaleDateString()}</p>
                      </div>
                    </div>
                    <Download className="w-4 h-4 text-muted-foreground group-hover:text-primary" />
                  </div>
                ))
              ) : (
                <div className="p-8 text-center text-muted-foreground">
                  <p className="text-sm italic">No recent activities found.</p>
                </div>
              )}
            </div>
            {history && history.length > 10 && (
              <div className="p-4 bg-accent/30 text-center">
                <button 
                  className="text-xs font-bold text-primary hover:underline"
                >
                  View All History
                </button>
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
