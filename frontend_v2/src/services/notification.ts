import api from "@/lib/api";

export interface NotificationPreferences {
  user_id: string;
  plan_tier: string;
  primary_channel: string;
  email_enabled: boolean;
  telegram_enabled: boolean;
  whatsapp_enabled: boolean;
  whatsapp_opted_in: boolean;
  timezone: string;
  preferred_send_time: string;
  last_daily_forecast_sent_at: string;
  last_daily_forecast_sent_for_date: string;
  created_at: string;
  updated_at: string;
}

export const notificationService = {
  getPreferences: async (): Promise<NotificationPreferences> => {
    const { data } = await api.get("/notifications/preferences");
    return data;
  },

  updatePreferences: async (payload: Partial<NotificationPreferences>): Promise<NotificationPreferences> => {
    const { data } = await api.put("/notifications/preferences", payload);
    return data;
  }
};
