import api from "@/lib/api";

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  tier: string;
  email_verified: boolean;
  forecast_efficiency: number;
  company_name: string;
  company_logo_url: string;
  esg_share_enabled: boolean;
  esg_share_token: string;
  created_at: string;
}

export const userService = {
  getMe: async (): Promise<User> => {
    const { data } = await api.get("/auth/me"); // Assuming /auth/me exists based on auth patterns
    return data;
  },

  updateBranding: async (companyName: string, logo?: File): Promise<any> => {
    const formData = new FormData();
    formData.append("company_name", companyName);
    if (logo) {
      formData.append("logo", logo);
    }
    const { data } = await api.post("/users/me/branding", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    return data;
  },

  getESGShareStatus: async (): Promise<{ enabled: boolean; token: string }> => {
    const { data } = await api.get("/users/me/esg-share");
    return data;
  },

  enableESGShare: async (): Promise<{ enabled: boolean; token: string }> => {
    const { data } = await api.post("/users/me/esg-share/enable");
    return data;
  },

  disableESGShare: async (): Promise<{ enabled: boolean }> => {
    const { data } = await api.post("/users/me/esg-share/disable");
    return data;
  }
};
