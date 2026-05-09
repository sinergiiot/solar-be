"use client";

import React from "react";
import { useQuery } from "@tanstack/react-query";
import { reportService } from "@/services/report";
import { 
  TreeDeciduous, 
  Cloud, 
  Zap, 
  Car, 
  Loader2,
  Globe,
  Award,
  Leaf
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useParams } from "next/navigation";

export default function PublicESGPage() {
  const { token } = useParams();
  const year = new Date().getFullYear();

  const { data, isLoading, error } = useQuery({
    queryKey: ["public-esg", token, year],
    queryFn: () => reportService.getPublicESGSummary(token as string, year),
    enabled: !!token,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center p-6 text-center">
        <Loader2 className="w-12 h-12 text-primary animate-spin mb-4" />
        <p className="text-muted-foreground font-bold animate-pulse">Loading green impact data...</p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center p-6 text-center">
        <div className="w-20 h-20 rounded-full bg-red-50 flex items-center justify-center text-red-500 mb-6">
          <Globe className="w-10 h-10" />
        </div>
        <h1 className="text-3xl font-bold mb-2">Report Not Found</h1>
        <p className="text-muted-foreground max-w-md">The ESG summary you are looking for does not exist or has been disabled by the owner.</p>
      </div>
    );
  }

  const { summary, company_name, company_logo } = data;

  const co2Kg = (summary?.total_co2_saved_ton || 0) * 1000;
  const treeEq = summary?.total_trees_eq || 0;
  const coalSavedKg = (summary?.total_actual_mwh || 0) * 400; // 0.4kg coal per kWh approx
  const carMilesEq = co2Kg / 0.404; // 0.404kg CO2 per mile approx

  const impactStats = [
    { label: "CO2 Saved", value: `${co2Kg.toLocaleString()} kg`, icon: Cloud, color: "text-blue-500", bg: "bg-blue-500/10" },
    { label: "Coal Saved", value: `${coalSavedKg.toLocaleString()} kg`, icon: Zap, color: "text-amber-500", bg: "bg-amber-500/10" },
    { label: "Trees Equivalent", value: `${treeEq.toLocaleString()} Trees`, icon: TreeDeciduous, color: "text-emerald-500", bg: "bg-emerald-500/10" },
    { label: "Car Miles Saved", value: `${carMilesEq.toLocaleString(undefined, { maximumFractionDigits: 0 })} Miles`, icon: Car, color: "text-purple-500", bg: "bg-purple-500/10" },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground selection:bg-emerald-500 selection:text-white">
      {/* Public Header */}
      <header className="border-b border-border bg-card/50 backdrop-blur-xl sticky top-0 z-50">
        <div className="max-w-5xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            {company_logo ? (
              <img 
                src={company_logo.startsWith("/") ? `${process.env.NEXT_PUBLIC_API_URL}${company_logo}` : company_logo} 
                alt="Logo" 
                className="w-10 h-10 object-contain"
              />
            ) : (
              <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white font-bold">
                <Leaf className="w-6 h-6" />
              </div>
            )}
            <span className="text-lg font-bold tracking-tight">{company_name || "Solar Partner"}</span>
          </div>
          <div className="flex items-center gap-2 text-xs font-bold bg-emerald-500/10 text-emerald-600 px-4 py-2 rounded-full border border-emerald-500/20">
            <Award className="w-4 h-4" />
            Verified Green Impact
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-12 md:py-20">
        <div className="text-center mb-16 space-y-4">
          <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight">
            Our Renewable Energy <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-500 to-teal-600 underline decoration-emerald-200 decoration-8 underline-offset-8">Impact Summary</span>
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto pt-4 leading-relaxed">
            Proudly monitoring our transition to sustainable solar power. 
            Here is our collective contribution to the planet for the year {year}.
          </p>
        </div>

        {summary?.total_actual_mwh === 0 && (
          <div className="p-12 rounded-[2rem] bg-accent/30 border border-dashed border-border text-center mb-16">
            <Globe className="w-12 h-12 text-muted-foreground/30 mx-auto mb-4" />
            <h3 className="text-xl font-bold mb-2">No Data for {year}</h3>
            <p className="text-muted-foreground">We haven't recorded any solar production data for this year yet.</p>
          </div>
        )}

        {/* Impact Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
          {impactStats.map((stat, idx) => (
            <div 
              key={idx} 
              className="p-8 rounded-[2.5rem] bg-card border border-border hover:border-emerald-500/50 hover:shadow-2xl hover:translate-y-[-4px] transition-all duration-300 text-center group"
            >
              <div className={cn("w-16 h-16 rounded-[1.5rem] flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform", stat.bg, stat.color)}>
                <stat.icon className="w-8 h-8" />
              </div>
              <p className="text-sm font-bold text-muted-foreground uppercase tracking-widest mb-2">{stat.label}</p>
              <p className="text-3xl font-black">{stat.value}</p>
            </div>
          ))}
        </div>

        {/* Statement Card */}
        <div className="relative rounded-[3rem] bg-gradient-to-br from-emerald-500 to-teal-700 p-12 md:p-20 text-white overflow-hidden shadow-2xl">
          <div className="absolute top-0 right-0 w-96 h-96 bg-white/10 rounded-full blur-3xl -mr-32 -mt-32" />
          <div className="absolute bottom-0 left-0 w-64 h-64 bg-black/10 rounded-full blur-2xl -ml-20 -mb-20" />
          
          <div className="relative z-10 grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold mb-6">A Shared Vision for a Sustainable Future</h2>
              <p className="text-emerald-50/80 text-lg leading-relaxed mb-8">
                By harnessing the power of the sun, we are reducing our dependence on fossil fuels and actively participating in the global fight against climate change. Every kilowatt produced brings us closer to a net-zero future.
              </p>
              <div className="flex flex-wrap gap-4">
                <div className="px-4 py-2 bg-white/10 backdrop-blur-md rounded-xl border border-white/20 text-sm font-bold">
                  ⚡ {((summary?.total_actual_mwh || 0) * 1000).toLocaleString()} kWh Generated
                </div>
                <div className="px-4 py-2 bg-white/10 backdrop-blur-md rounded-xl border border-white/20 text-sm font-bold">
                  🌱 Sustainable Operations
                </div>
              </div>
            </div>
            <div className="hidden md:flex justify-center">
              <div className="w-64 h-64 rounded-full border-8 border-white/20 flex items-center justify-center">
                <div className="w-48 h-48 rounded-full border-8 border-white/40 flex items-center justify-center animate-pulse">
                  <Leaf className="w-24 h-24 text-white" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      <footer className="border-t border-border py-12 bg-card/30">
        <div className="max-w-5xl mx-auto px-6 text-center">
          <p className="text-sm text-muted-foreground font-medium mb-4">
            Powered by <span className="text-primary font-bold">SolarForecast Platform</span>
          </p>
          <div className="flex justify-center gap-6 text-xs text-muted-foreground font-bold uppercase tracking-widest">
            <a href="#" className="hover:text-primary">Terms</a>
            <a href="#" className="hover:text-primary">Privacy</a>
            <a href="#" className="hover:text-primary">Audit Log</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
