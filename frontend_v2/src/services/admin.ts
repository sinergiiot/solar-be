import api from "@/lib/api";

export const adminService = {
  getAllUsers: async (): Promise<any[]> => {
    const { data } = await api.get("/admin/users");
    return data;
  },

  updateUserTier: async (userId: string, tier: string): Promise<void> => {
    await api.put(`/admin/users/${userId}/tier`, { plan_tier: tier });
  },

  getSystemStats: async (): Promise<any> => {
    const { data } = await api.get("/admin/stats");
    return data;
  },

  getSchedulerStatus: async (): Promise<any[]> => {
    const { data } = await api.get("/admin/scheduler/status");
    return data;
  },

  getWeatherHealth: async (): Promise<any> => {
    const { data } = await api.get("/admin/weather/health");
    return data;
  },

  updateUser: async (userId: string, data: { name: string, email: string }): Promise<void> => {
    await api.put(`/admin/users/${userId}`, data);
  },

  deleteUser: async (userId: string): Promise<void> => {
    await api.delete(`/admin/users/${userId}`);
  }
};
