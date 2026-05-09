import api from "@/lib/api";

export interface Device {
  id: string;
  user_id: string;
  solar_profile_id: string;
  name: string;
  device_key: string;
  status: "online" | "offline";
  last_heartbeat: string;
  created_at: string;
}

export interface DeviceHeartbeatSummary {
  total: number;
  online: number;
  offline: number;
}

export const deviceService = {
  listDevices: async (): Promise<Device[]> => {
    const { data } = await api.get("/devices");
    return data.devices;
  },

  createDevice: async (device: Partial<Device>): Promise<Device> => {
    const { data } = await api.post("/devices", device);
    return data;
  },

  updateDevice: async (id: string, device: Partial<Device>): Promise<Device> => {
    const { data } = await api.put(`/devices/${id}`, device);
    return data;
  },

  deleteDevice: async (id: string): Promise<void> => {
    await api.delete(`/devices/${id}`);
  },

  rotateKey: async (id: string): Promise<{ device_key: string }> => {
    const { data } = await api.post(`/devices/${id}/rotate-key`);
    return data;
  },

  getHeartbeatSummary: async (): Promise<DeviceHeartbeatSummary> => {
    const { data } = await api.get("/devices/heartbeat-summary");
    return data;
  }
};
