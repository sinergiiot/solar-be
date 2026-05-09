import api from "@/lib/api";

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: {
    id: string;
    name: string;
    email: string;
    role: string;
  };
}

export const authService = {
  login: async (email: string, password: string): Promise<LoginResponse> => {
    const { data } = await api.post("/auth/login", { email, password });
    if (data.access_token) {
      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      localStorage.setItem("user", JSON.stringify(data.user));
    }
    return data;
  },

  register: async (name: string, email: string, password: string): Promise<any> => {
    const { data } = await api.post("/auth/register", { name, email, password });
    return data;
  },

  verifyEmail: async (email: string, code: string): Promise<LoginResponse> => {
    const { data } = await api.post("/auth/verify-email", { email, code });
    if (data.access_token) {
      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      localStorage.setItem("user", JSON.stringify(data.user));
    }
    return data;
  },

  forgotPassword: async (email: string): Promise<any> => {
    const { data } = await api.post("/auth/forgot-password", { email });
    return data;
  },

  resetPassword: async (email: string, code: string, newPassword: string): Promise<any> => {
    const { data } = await api.post("/auth/reset-password", { email, code, new_password: newPassword });
    return data;
  },
  
  logout: () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
  },

  getUser: () => {
    if (typeof window !== "undefined") {
      const user = localStorage.getItem("user");
      return user ? JSON.parse(user) : null;
    }
    return null;
  }
};
