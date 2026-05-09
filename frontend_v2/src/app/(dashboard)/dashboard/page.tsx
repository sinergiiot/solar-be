"use client";

import { useState, useMemo } from "react";
import { Card } from "@/components/ui/card";
import { 
  Zap, 
  TrendingUp, 
  CloudSun, 
  Award,
  Loader2,
  Calendar,
  ChevronRight,
  ArrowRight,
  Sparkles,
  Rocket
} from "lucide-react";
import dynamic from "next/dynamic";
import { useQuery } from "@tanstack/react-query";
import { forecastService } from "@/services/forecast";
import { solarService } from "@/services/solar";
import { userService } from "@/services/user";
import { cn } from "@/lib/utils";
import GettingStartedChecklist from "@/components/layout/getting-started-checklist";

// Helper for formatting dates to YYYY-MM-DD
const formatDate = (date: Date) => {
  return date.toISOString().split('T')[0];
};

// Dynamic import for ApexCharts
const Chart = dynamic(() => import("react-apexcharts"), { ssr: false });

import { useTheme } from "next-themes";
import { usePlan } from "@/components/providers/plan-provider";

export default function DashboardPage() {
  const { theme } = useTheme();
  const { checkAccess } = usePlan();
  const [days, setDays] = useState(7);
  const [dateRange, setDateRange] = useState({
    start: formatDate(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)),
    end: formatDate(new Date())
  });
  
  // Fetch Summary Stats
  const { data: summary, isLoading: isSummaryLoading } = useQuery({
    queryKey: ["forecast-summary"],
    queryFn: () => forecastService.getSummary(),
  });

  const { data: user } = useQuery({ 
    queryKey: ["me"], 
    queryFn: () => userService.getMe() 
  });

  // Fetch History for Chart
  const { data: historyData, isLoading: isHistoryLoading, isFetching: isHistoryFetching } = useQuery({
    queryKey: ["forecast-history", days, dateRange],
    queryFn: () => forecastService.getHistoryList({
      page: 1,
      limit: days === 0 ? 100 : days,
      sort: "date",
      order: "asc",
      start_date: days === 0 ? dateRange.start : "",
      end_date: days === 0 ? dateRange.end : ""
    }),
    staleTime: 0,
  });

  // Fetch Actuals for Chart
  const { data: actualsData, isLoading: isActualsLoading, isFetching: isActualsFetching } = useQuery({
    queryKey: ["actual-history", days, dateRange],
    queryFn: () => forecastService.getActualHistoryList({
      page: 1,
      limit: days === 0 ? 100 : days,
      sort: "date",
      order: "asc",
      start_date: days === 0 ? dateRange.start : "",
      end_date: days === 0 ? dateRange.end : ""
    }),
    staleTime: 0,
  });

  const history = historyData?.items || [];
  const actuals = actualsData?.items || [];

  // Fetch Profiles
  const { data: profiles } = useQuery({
    queryKey: ["solar-profiles"],
    queryFn: () => solarService.getProfiles(),
  });

  const isFetchingData = isHistoryFetching || isActualsFetching;
  const isLoading = isSummaryLoading || isHistoryLoading || isActualsLoading;

  const stats = [
    { label: "Total Forecasted", value: `${summary?.total_forecasted_kwh?.toFixed(2) || "0.00"} kWh`, icon: CloudSun, color: "text-blue-500", bg: "bg-blue-500/10" },
    { label: "Total Actual", value: `${summary?.total_actual_kwh?.toFixed(2) || "0.00"} kWh`, icon: Zap, color: "text-amber-500", bg: "bg-amber-500/10" },
    { label: "Accuracy Rate", value: `${summary?.accuracy_percent?.toFixed(2) || "0.00"}%`, icon: TrendingUp, color: "text-emerald-500", bg: "bg-emerald-500/10" },
    { label: "CO2 Avoided", value: `${((summary?.total_actual_kwh || 0) * 0.45).toFixed(2)} kg`, icon: Award, color: "text-purple-500", bg: "bg-purple-500/10" },
  ];

  // Prepare Chart Data
  const chartOptions: any = useMemo(() => ({
    chart: {
      id: "production-forecast",
      toolbar: { show: false },
      fontFamily: "var(--font-outfit)",
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: 800,
      }
    },
    theme: {
      mode: theme === "dark" ? "dark" : "light"
    },
    colors: ["#f09235", "#fee75c"],
    stroke: { curve: "smooth", width: 3 },
    fill: {
      type: "gradient",
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.45,
        opacityTo: 0.05,
        stops: [20, 100, 100, 100]
      }
    },
    dataLabels: {
      enabled: false
    },
    xaxis: {
      categories: (history as any[])?.map(h => new Date(h.date).toLocaleDateString("en-US", { month: "short", day: "numeric" })) || [],
      axisBorder: { show: false },
      axisTicks: { show: false },
      labels: {
        style: {
          colors: theme === "dark" ? "#94a3b8" : "#64748b",
          fontWeight: 600
        }
      }
    },
    yaxis: {
      labels: {
        formatter: (val: number) => val?.toFixed(1) + " kWh",
        style: {
          colors: theme === "dark" ? "#94a3b8" : "#64748b",
          fontWeight: 600
        }
      }
    },
    grid: {
      borderColor: theme === "dark" ? "#1e293b" : "#f1f5f9",
      strokeDashArray: 4,
    },
    legend: {
      position: "top",
      horizontalAlign: "right",
    },
    tooltip: {
      y: {
        formatter: (val: number) => val?.toFixed(2) + " kWh"
      }
    }
  }), [theme, history, days]);

  const chartSeries = useMemo(() => [
    {
      name: "Forecast",
      data: (history as any[])?.map(h => h.predicted_kwh) || []
    },
    {
      name: "Actual",
      data: (history as any[])?.map(h => {
        const actual = (actuals as any[])?.find(a => a.date === h.date);
        return actual ? actual.actual_kwh : 0;
      }) || []
    }
  ], [history, actuals, days]);

  if (isLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-3xl font-bold tracking-tight mb-2">Welcome Back, Akbar</h1>
          <p className="text-muted-foreground">Here is what&apos;s happening with your energy production today.</p>
        </div>
        <div className="text-sm font-bold bg-secondary/10 text-secondary px-4 py-2 rounded-xl">
          Live Data Active
        </div>
      </div>

      {/* Upgrade Nudge for Free Users */}
      {user?.tier === "free" && (
        <Card className="relative overflow-hidden border-none shadow-premium bg-gradient-to-r from-primary via-primary/90 to-primary/80 text-white p-8 rounded-[2rem]">
          <div className="absolute top-0 right-0 p-6 opacity-10">
            <Sparkles className="w-40 h-40" />
          </div>
          <div className="flex flex-col md:flex-row items-center justify-between gap-6 relative z-10">
            <div className="flex items-center gap-6">
              <div className="w-16 h-16 rounded-2xl bg-white/20 backdrop-blur-md flex items-center justify-center">
                <Rocket className="w-8 h-8 text-white" />
              </div>
              <div className="space-y-1">
                <h2 className="text-2xl font-black tracking-tight">Unlock AI Power</h2>
                <p className="text-white/80 font-medium">Get 99% accuracy with AI Forecasting and detailed ESG Reports.</p>
              </div>
            </div>
            <a 
              href="/dashboard/billing" 
              className="px-8 py-4 bg-white text-primary rounded-2xl font-black text-sm flex items-center gap-2 hover:scale-105 transition-all shadow-xl"
            >
              Explore Pro Plans
              <ArrowRight className="w-4 h-4" />
            </a>
          </div>
        </Card>
      )}

      <GettingStartedChecklist />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat, idx) => (
          <Card key={idx} className="relative overflow-hidden group hover:translate-y-[-4px] transition-all duration-300">
            <div className="flex items-center gap-4">
              <div className={cn("w-12 h-12 rounded-2xl flex items-center justify-center", stat.bg)}>
                <stat.icon className={cn("w-6 h-6", stat.color)} />
              </div>
              <div>
                <p className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">{stat.label}</p>
                <p className="text-2xl font-bold">{stat.value}</p>
              </div>
            </div>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <Card className="lg:col-span-2">
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-3">
              <h2 className="text-xl font-bold">Production vs Forecast</h2>
              {isFetchingData && (
                <Loader2 className="w-4 h-4 text-primary animate-spin" />
              )}
            </div>
            <div className="flex flex-wrap items-center gap-4">
              <div className="flex items-center gap-2 bg-card border border-border p-1.5 rounded-2xl">
                {[7, 90, 0].map((d) => (
                  <button
                    key={d}
                    onClick={() => {
                      if (d === 90 || d === 0) {
                        const hasAccess = checkAccess(
                          "pro", 
                          "Historical Data", 
                          "Access to 90-day history and custom date ranges is reserved for Pro users. Gain a comprehensive long-term view of your energy performance."
                        );
                        if (!hasAccess) return;
                      }
                      setDays(d);
                    }}
                    className={cn(
                      "px-5 py-1.5 rounded-full text-xs font-bold transition-all duration-200",
                      days === d 
                        ? "bg-primary text-white shadow-sm" 
                        : "text-slate-500 hover:text-slate-700 dark:text-muted-foreground dark:hover:text-foreground"
                    )}
                  >
                    {d === 0 ? "Custom Range" : `${d} Days`}
                  </button>
                ))}
              </div>

              {days === 0 && (
                <div className="flex items-center gap-2 animate-in fade-in zoom-in duration-300">
                  <div className="flex items-center gap-2 bg-card border border-border p-1.5 rounded-2xl">
                    <input 
                      type="date" 
                      value={dateRange.start}
                      onChange={(e) => setDateRange(prev => ({ ...prev, start: e.target.value }))}
                      className="bg-transparent border-none text-xs font-bold focus:ring-0 outline-none w-28 text-slate-700 dark:text-foreground"
                    />
                    <ChevronRight className="w-3 h-3 mx-1 text-slate-400" />
                    <input 
                      type="date" 
                      value={dateRange.end}
                      onChange={(e) => setDateRange(prev => ({ ...prev, end: e.target.value }))}
                      className="bg-transparent border-none text-xs font-bold focus:ring-0 outline-none w-28 text-slate-700 dark:text-foreground"
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
          <div className="h-[350px] w-full">
            <Chart 
              key={`chart-${days}-${theme}`}
              options={chartOptions}
              series={chartSeries}
              type="area"
              height="100%"
            />
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-bold mb-6">Active Profiles</h2>
          <div className="space-y-6">
            {profiles?.map((profile, idx) => {
              // Find latest efficiency for this profile from history
              const profileForecasts = (history as any[])?.filter(h => h.solar_profile_id === profile.id);
              const latestEfficiency = profileForecasts && profileForecasts.length > 0 
                ? profileForecasts[0].efficiency 
                : 0.8; // Fallback to 80%

              const efficiencyPercent = (latestEfficiency * 100).toFixed(1);

              return (
                <div key={idx} className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="font-bold text-foreground">{profile.site_name}</span>
                    <span className="font-bold text-primary">{profile.capacity_kwp} kWp</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground font-medium">
                    <CloudSun className="w-3 h-3" />
                    Real-time Efficiency: {efficiencyPercent}%
                  </div>
                  <div className="h-1.5 w-full bg-accent rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-primary transition-all duration-1000" 
                      style={{ width: `${efficiencyPercent}%` }}
                    />
                  </div>
                </div>
              );
            })}
            {(!profiles || profiles.length === 0) && (
              <p className="text-sm text-muted-foreground italic">No solar profiles found.</p>
            )}
          </div>
          <div className="mt-10 p-4 rounded-2xl bg-primary/5 border border-primary/10">
            <p className="text-xs font-medium text-primary leading-relaxed">
              <strong>ESG Impact:</strong> You have avoided approximately {((summary?.total_actual_kwh || 0) * 0.45).toFixed(2)}kg of CO2 this period.
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}
