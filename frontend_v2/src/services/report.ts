import api from "@/lib/api";

export interface MonthlySummary {
  month: string;
  actual_kwh: number;
  savings_idr: number;
}

export interface OfficialDetails {
  letter_number: string;
  signatory: string;
  title: string;
  organization: string;
  official_date: string;
}

export interface EnergyReport {
  user_id: string;
  solar_profile_id?: string;
  period_start: string;
  period_end: string;
  total_forecasted_kwh: number;
  total_actual_kwh: number;
  total_savings_idr: number;
  total_co2_avoided_kg: number;
  data_coverage_pct: number;
  plan_tier: string;
  created_at: string;
  is_annual: boolean;
  monthly_breakdown?: MonthlySummary[];
  official_details?: OfficialDetails;
  total_rec: number;
}

export interface ESGMonth {
  month: string;
  actual_mwh: number;
  co2_saved_ton: number;
}

export interface SiteESG {
  profile_id: string;
  profile_name: string;
  location: string;
  actual_mwh: number;
  co2_saved_ton: number;
  rec_reached: number;
}

export interface ESGSummary {
  user_id: string;
  total_actual_mwh: number;
  total_co2_saved_ton: number;
  total_trees_eq: number;
  total_rec_count: number;
  total_savings_idr: number;
  clean_energy_pct: number;
  site_breakdown: SiteESG[];
  yearly_trend: ESGMonth[];
}

export interface CO2Day {
  date: string;
  actual_kwh: number;
  co2_avoided_kg: number;
}

export interface CO2Summary {
  user_id: string;
  period_start: string;
  period_end: string;
  total_actual_kwh: number;
  total_co2_avoided_kg: number;
  total_co2_avoided_ton: number;
  carbon_credit_idr: number;
  carbon_credit_usd: number;
  emission_factor_kg_per_kwh: number;
  grid_region: string;
  methodology_standard: string;
  daily_breakdown?: CO2Day[];
  plan_tier: string;
}

export interface SiteREC {
  profile_name: string;
  capacity_kwp: number;
  location: string;
  total_actual_mwh: number;
  rec_contribution: number;
}

export interface RECMonth {
  month: string;
  actual_mwh: number;
}

export interface RECReadinessReport {
  user_id: string;
  total_actual_mwh: number;
  total_rec: number;
  site_breakdown: SiteREC[];
  monthly_history: RECMonth[];
  generated_at: string;
}

export interface ReportHistory {
  id: string;
  user_id: string;
  report_name: string;
  report_type: string;
  metadata?: any;
  created_at: string;
}

export const reportService = {
  getEnergyReport: async (params: any): Promise<EnergyReport> => {
    const { data } = await api.get("/report/energy", { params });
    return data;
  },

  getESGSummary: async (year: number): Promise<ESGSummary> => {
    const { data } = await api.get("/report/esg", { params: { year } });
    return data;
  },

  getCO2Summary: async (params: any): Promise<CO2Summary> => {
    const { data } = await api.get("/report/co2", { params });
    return data;
  },

  getRECReadinessReport: async (): Promise<RECReadinessReport> => {
    const { data } = await api.get("/report/rec");
    return data;
  },

  getReportHistory: async (): Promise<ReportHistory[]> => {
    const { data } = await api.get("/report/history");
    return data;
  },

  downloadEnergyReportPDF: async (params: any) => {
    const response = await api.get("/report/energy/pdf", { 
      params,
      responseType: "blob" 
    });
    return response.data;
  },

  downloadESGReportPDF: async (year: number) => {
    const response = await api.get("/report/esg/pdf", { 
      params: { year },
      responseType: "blob" 
    });
    return response.data;
  },

  downloadMRVPDF: async (params: any) => {
    const response = await api.get("/report/co2/pdf", { 
      params,
      responseType: "blob" 
    });
    return response.data;
  },

  downloadHistoryCSV: async (params: any) => {
    const response = await api.get("/report/history/csv", { 
      params,
      responseType: "blob" 
    });
    return response.data;
  },

  downloadRECPDF: async (type: "report" | "certificate" = "report") => {
    const response = await api.get("/report/rec/pdf", { 
      params: { type },
      responseType: "blob" 
    });
    return response.data;
  },

  downloadMonthlySummaryPDF: async (month: string, year: number) => {
    const response = await api.get("/report/monthly/pdf", {
      params: { month, year },
      responseType: "blob"
    });
    return response.data;
  },

  downloadSiteAuditPDF: async (solarProfileId: string) => {
    const response = await api.get("/report/site-audit/pdf", {
      params: { solar_profile_id: solarProfileId },
      responseType: "blob"
    });
    return response.data;
  },

  getPublicESGSummary: async (token: string, year: number): Promise<any> => {
    const { data } = await api.get(`/public/esg/${token}`, { params: { year } });
    return data;
  }
};
