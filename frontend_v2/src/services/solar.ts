import api from "@/lib/api";

export interface SolarProfile {
  id: string;
  user_id: string;
  site_name: string;
  capacity_kwp: number;
  lat: number;
  lng: number;
  tilt: number;
  azimuth: number;
}

export const solarService = {
  getProfiles: async (): Promise<SolarProfile[]> => {
    const { data } = await api.get("/solar-profiles");
    return data.profiles;
  },

  createProfile: async (profile: Omit<SolarProfile, "id" | "user_id">): Promise<SolarProfile> => {
    const { data } = await api.post("/solar-profiles", profile);
    return data;
  },

  updateProfile: async (id: string, profile: Partial<SolarProfile>): Promise<SolarProfile> => {
    const { data } = await api.put(`/solar-profiles/${id}`, profile);
    return data;
  },

  deleteProfile: async (id: string): Promise<void> => {
    await api.delete(`/solar-profiles/${id}`);
  }
};
