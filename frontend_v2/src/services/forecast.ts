import api from "@/lib/api";

export interface ForecastSummary {
  total_forecasted_kwh: number;
  total_actual_kwh: number;
  average_forecast_kwh: number;
  average_actual_kwh: number;
  current_efficiency: number;
  accuracy_percent: number;
  forecast_count: number;
  actual_count: number;
}

export interface ForecastItem {
  date: string;
  solar_profile_id?: string;
  predicted_kwh: number;
  actual_kwh?: number;
  weather_factor: number;
  efficiency: number;
}

export const forecastService = {
  getSummary: async (): Promise<ForecastSummary> => {
    const { data } = await api.get("/forecast/summary");
    return data;
  },
  
  getHistory: async (days: number = 7): Promise<ForecastItem[]> => {
    const { data } = await api.get(`/forecast/history?days=${days}`);
    return data.forecasts;
  },

  getActualHistory: async (days: number = 7): Promise<any[]> => {
    const { data } = await api.get(`/forecast/actuals/history?days=${days}`);
    return data.actuals;
  },

  getHistoryList: async (payload: any): Promise<any> => {
    const { data } = await api.post("/forecast/history/list", payload);
    return data;
  },

  getActualHistoryList: async (payload: any): Promise<any> => {
    const { data } = await api.post("/forecast/actuals/history/list", payload);
    return data;
  },
  
  recordActual: async (payload: { solar_profile_id: string, date: string, actual_kwh: number, source?: string }): Promise<any> => {
    const { data } = await api.post("/forecast/actual", payload);
    return data;
  }
};
