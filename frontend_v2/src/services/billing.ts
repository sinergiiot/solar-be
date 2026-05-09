import api from "@/lib/api";

export interface Subscription {
  id: string;
  user_id: string;
  plan_tier: string;
  status: string;
  billing_cycle: string;
  amount: number;
  currency: string;
  expires_at: string;
  created_at: string;
  external_checkout_id: string;
  payment_url?: string;
  next_billing_at?: string;
  last_payment_at?: string;
}

export interface CheckoutResponse {
  checkout_url: string;
  id: string;
}

export const billingService = {
  getSubscription: async (): Promise<Subscription> => {
    const { data } = await api.get("/billing/subscription");
    return data;
  },

  createCheckout: async (planTier: string, billingCycle: string = "monthly"): Promise<CheckoutResponse> => {
    const { data } = await api.post("/billing/checkout", { plan_tier: planTier, billing_cycle: billingCycle });
    return data;
  },

  getHistory: async (): Promise<Subscription[]> => {
    const { data } = await api.get("/billing/history");
    return data;
  },

  cancelSubscription: async (): Promise<void> => {
    await api.post("/billing/subscription/cancel");
  }
};
