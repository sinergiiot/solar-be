"use client";

import React, { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { billingService } from "@/services/billing";
import { userService } from "@/services/user";
import { Card } from "@/components/ui/card";
import { 
  Check, 
  Zap, 
  ShieldCheck, 
  Rocket, 
  Loader2, 
  ArrowRight,
  Sparkles,
  Crown,
  CreditCard,
  Clock,
  Calendar,
  AlertCircle,
  FileText,
  ExternalLink
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useToast } from "@/components/providers/toast-provider";

export default function PricingPage() {
  const { showToast } = useToast();
  const searchParams = useSearchParams();
  const [isYearly, setIsYearly] = useState(false);

  // Handle payment redirects
  useEffect(() => {
    const paymentStatus = searchParams.get("payment");
    if (paymentStatus === "success") {
      showToast("Payment successful! Your plan is being updated.", "success");
    } else if (paymentStatus === "cancel") {
      showToast("Payment was cancelled. No charges were made.", "warning");
    }
  }, [searchParams]);

  const { data: user } = useQuery({ 
    queryKey: ["me"], 
    queryFn: () => userService.getMe() 
  });

  const { data: subscription, isLoading: isSubLoading } = useQuery({
    queryKey: ["subscription"],
    queryFn: () => billingService.getSubscription()
  });

  const { data: history, isLoading: isHistoryLoading, refetch: refetchHistory } = useQuery({
    queryKey: ["billing-history"],
    queryFn: () => billingService.getHistory()
  });

  const checkoutMutation = useMutation({
    mutationFn: (plan: string) => billingService.createCheckout(plan, isYearly ? "yearly" : "monthly"),
    onSuccess: (data) => {
      showToast("Redirecting to secure payment gateway...", "info");
      window.location.href = data.checkout_url;
    },
    onError: (err: any) => {
      showToast(err.response?.data?.error || "Failed to initiate checkout. Please try again later.", "error");
    }
  });

  const cancelMutation = useMutation({
    mutationFn: () => billingService.cancelSubscription(),
    onSuccess: () => {
      showToast("Subscription cancelled successfully.", "success");
      refetchHistory();
    },
    onError: (err: any) => {
      showToast(err.response?.data?.error || "Failed to cancel subscription.", "error");
    }
  });

  const plans = [
    {
      name: "Free",
      tier: "free",
      monthlyPrice: 0,
      description: "Perfect for individual home monitoring.",
      features: [
        "1 Solar Site Location",
        "1 IoT Device Support",
        "7-Day Historical Data",
        "Daily Forecast Analytics",
        "Email Notifications",
        "Standard Support"
      ],
      icon: Zap,
      color: "text-slate-500",
      bg: "bg-slate-500/10",
      buttonText: "Current Plan",
      popular: false
    },
    {
      name: "Pro",
      tier: "pro",
      monthlyPrice: 99000,
      description: "Advanced analytics for solar enthusiasts.",
      features: [
        "Up to 5 Solar Sites",
        "Up to 10 IoT Devices",
        "90-Day Data Retention",
        "PDF Report & PBB Letter",
        "REC Readiness Report",
        "CSV History Export",
        "Priority Email Notifications"
      ],
      icon: Rocket,
      color: "text-blue-500",
      bg: "bg-blue-500/10",
      buttonText: "Upgrade to Pro",
      popular: true
    },
    {
      name: "Enterprise",
      tier: "enterprise",
      monthlyPrice: 499000,
      description: "Full suite for commercial solar operations.",
      features: [
        "Unlimited Solar Sites",
        "Unlimited IoT Devices",
        "Lifetime Data History",
        "Multi-site ESG Dashboard",
        "White-label (Custom Logo)",
        "Public ESG Share Link",
        "External API Access",
        "Priority Support & SLA"
      ],
      icon: Crown,
      color: "text-purple-500",
      bg: "bg-purple-500/10",
      buttonText: "Get Enterprise",
      popular: false
    }
  ];

  return (
    <div className="max-w-7xl mx-auto space-y-12 pb-20">
      {/* Current Subscription Status */}
      {subscription && subscription.status !== "free" && (
        <Card className="relative overflow-hidden border-primary/20 shadow-premium bg-primary/[0.02] p-8 rounded-[2.5rem]">
          <div className="absolute top-0 right-0 p-8 opacity-5">
            <Sparkles className="w-32 h-32 text-primary" />
          </div>
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-8 relative z-10">
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <span className="flex h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                <span className="text-xs font-black uppercase tracking-widest text-emerald-500">Active Subscription</span>
              </div>
              <div className="flex items-center gap-4">
                <div className="w-16 h-16 rounded-2xl bg-primary flex items-center justify-center text-white shadow-premium">
                  {subscription.plan_tier === "enterprise" ? <Crown className="w-8 h-8" /> : <Rocket className="w-8 h-8" />}
                </div>
                <div>
                  <h2 className="text-3xl font-black capitalize">{subscription.plan_tier} Plan</h2>
                  <p className="text-muted-foreground font-medium">Billed {subscription.billing_cycle}</p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
              <div className="space-y-1">
                <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Renewal Date</p>
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-primary" />
                  <p className="font-bold">{new Date(subscription.expires_at).toLocaleDateString()}</p>
                </div>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Amount</p>
                <div className="flex items-center gap-2">
                  <CreditCard className="w-4 h-4 text-primary" />
                  <p className="font-bold">IDR {(subscription.amount / 1000).toLocaleString()}k</p>
                </div>
              </div>
              <div className="space-y-1 col-span-2 md:col-span-1">
                <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Status</p>
                <div className="flex items-center gap-2">
                  <div className="px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-500 text-[10px] font-black uppercase">
                    {subscription.status}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Card>
      )}

      <div className="text-center space-y-4 pt-8">
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/10 text-primary text-sm font-bold">
          <Sparkles className="w-4 h-4" />
          Flexible Pricing for Every Scale
        </div>
        <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight">Choose the Perfect Plan</h1>
        <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
          Unlock advanced forecasting, ESG reporting, and enterprise-grade monitoring tools.
        </p>
      </div>

      {/* Billing Toggle */}
      <div className="flex items-center justify-center gap-4">
        <span className={cn("text-sm font-bold", !isYearly ? "text-foreground" : "text-muted-foreground")}>Monthly</span>
        <button 
          onClick={() => setIsYearly(!isYearly)}
          className="w-14 h-8 rounded-full bg-accent border border-border p-1 relative transition-all"
        >
          <div className={cn(
            "w-6 h-6 rounded-full bg-primary shadow-lg transition-all transform",
            isYearly ? "translate-x-6" : "translate-x-0"
          )} />
        </button>
        <span className={cn("text-sm font-bold", isYearly ? "text-foreground" : "text-muted-foreground")}>
          Yearly <span className="ml-1 text-[10px] bg-emerald-500/10 text-emerald-500 px-2 py-0.5 rounded-full uppercase">Save 20%</span>
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {plans.map((plan) => {
          const isCurrentPlan = user?.tier === plan.tier;
          const isHigher = (plan.tier === "pro" && user?.tier === "free") || 
                           (plan.tier === "enterprise" && (user?.tier === "free" || user?.tier === "pro"));

          // Calculate price
          let displayPrice = plan.monthlyPrice;
          if (isYearly && plan.monthlyPrice > 0) {
            displayPrice = Math.floor(plan.monthlyPrice * 0.8);
          }
          
          const formattedPrice = displayPrice === 0 ? "0" : 
                                 displayPrice >= 1000 ? `${(displayPrice / 1000).toFixed(displayPrice % 1000 === 0 ? 0 : 1)}k` : 
                                 displayPrice.toString();

          return (
            <Card 
              key={plan.name}
              className={cn(
                "relative flex flex-col p-8 rounded-[2.5rem] transition-all duration-300 hover:translate-y-[-8px]",
                plan.popular ? "border-primary shadow-premium ring-4 ring-primary/5" : "border-border shadow-sm",
                isCurrentPlan && "bg-accent/30"
              )}
            >
              {plan.popular && (
                <div className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-primary text-white text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-full shadow-lg">
                  Most Popular
                </div>
              )}

              <div className="mb-8">
                <div className={cn("w-14 h-14 rounded-2xl flex items-center justify-center mb-6", plan.bg, plan.color)}>
                  <plan.icon className="w-8 h-8" />
                </div>
                <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                <p className="text-sm text-muted-foreground">{plan.description}</p>
              </div>

              <div className="mb-8">
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-black">IDR {formattedPrice}</span>
                  <span className="text-muted-foreground font-bold">/mo</span>
                </div>
                {isYearly && plan.monthlyPrice > 0 && (
                  <p className="text-[10px] text-emerald-500 font-bold mt-1">
                    Billed annually (IDR {(displayPrice * 12 / 1000).toFixed(0)}k/year)
                  </p>
                )}
              </div>

              <div className="space-y-4 mb-10 flex-1">
                {plan.features.map((feature) => (
                  <div key={feature} className="flex items-start gap-3">
                    <div className="mt-1 w-5 h-5 rounded-full bg-emerald-500/10 flex items-center justify-center shrink-0">
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                    <span className="text-sm font-medium leading-tight">{feature}</span>
                  </div>
                ))}
              </div>

              <button
                disabled={isCurrentPlan || checkoutMutation.isPending || !isHigher}
                onClick={() => checkoutMutation.mutate(plan.tier)}
                className={cn(
                  "w-full py-4 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 group",
                  isCurrentPlan 
                    ? "bg-accent text-muted-foreground cursor-default" 
                    : isHigher
                      ? "bg-primary text-white shadow-premium hover:opacity-90"
                      : "bg-slate-200 text-slate-400 cursor-not-allowed"
                )}
              >
                {checkoutMutation.isPending && checkoutMutation.variables === plan.tier ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : isCurrentPlan ? (
                  "Current Plan"
                ) : (
                  <>
                    {plan.buttonText}
                    <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                  </>
                )}
              </button>
            </Card>
          );
        })}
      </div>

      {/* Billing History Section */}
      <div className="pt-12 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-black tracking-tight">Transaction History</h2>
            <p className="text-muted-foreground font-medium">Detailed log of your billing activities and invoices.</p>
          </div>
          <div className="hidden md:block">
            <div className="flex items-center gap-2 px-4 py-2 rounded-xl bg-accent/50 border border-border text-xs font-bold">
              <ShieldCheck className="w-4 h-4 text-primary" />
              Secure Payment Records
            </div>
          </div>
        </div>

        <Card className="overflow-hidden border-border shadow-premium rounded-[2rem]">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-accent/30 border-b border-border">
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Reference</th>
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Plan / Cycle</th>
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Amount</th>
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Status</th>
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Date</th>
                  <th className="px-8 py-5 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground text-right">Invoice</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {isHistoryLoading ? (
                  <tr>
                    <td colSpan={6} className="px-8 py-20 text-center text-muted-foreground">
                      <div className="flex flex-col items-center gap-4">
                        <Loader2 className="w-10 h-10 animate-spin text-primary" />
                        <p className="font-bold tracking-tight">Syncing transaction data...</p>
                      </div>
                    </td>
                  </tr>
                ) : history?.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-8 py-20 text-center text-muted-foreground">
                      <div className="flex flex-col items-center gap-4 opacity-40">
                        <FileText className="w-16 h-16" />
                        <div>
                          <p className="text-lg font-bold text-foreground">No history yet</p>
                          <p className="text-sm font-medium">Your future transactions will appear here.</p>
                        </div>
                      </div>
                    </td>
                  </tr>
                ) : (
                  history?.map((item) => (
                    <tr key={item.id} className="group hover:bg-primary/[0.01] transition-colors">
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-xl bg-accent flex items-center justify-center border border-border group-hover:border-primary/20 transition-colors">
                            <Clock className="w-5 h-5 text-muted-foreground group-hover:text-primary transition-colors" />
                          </div>
                          <div>
                            <p className="text-sm font-bold tracking-tight">#{item.external_checkout_id?.substring(0, 12)}</p>
                            <p className="text-[10px] font-bold text-muted-foreground uppercase">{item.id.substring(0, 8)}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-3">
                          <div className={cn(
                            "w-8 h-8 rounded-lg flex items-center justify-center",
                            item.plan_tier === "enterprise" ? "bg-purple-500/10 text-purple-500" :
                            item.plan_tier === "pro" ? "bg-blue-500/10 text-blue-500" :
                            "bg-slate-500/10 text-slate-500"
                          )}>
                            {item.plan_tier === "enterprise" ? <Crown className="w-4 h-4" /> : <Rocket className="w-4 h-4" />}
                          </div>
                          <div>
                            <p className="text-sm font-bold capitalize">{item.plan_tier}</p>
                            <p className="text-[10px] text-muted-foreground uppercase font-black tracking-widest">{item.billing_cycle}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <p className="text-sm font-black tracking-tight">IDR {(item.amount / 1000).toLocaleString()}k</p>
                      </td>
                      <td className="px-8 py-5">
                        <span className={cn(
                          "text-[10px] font-black uppercase tracking-[0.1em] px-3 py-1.5 rounded-full border flex items-center gap-2 w-fit",
                          item.status === "active" ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" :
                          item.status === "pending" ? "bg-amber-500/10 text-amber-500 border-amber-500/20" :
                          "bg-red-500/10 text-red-500 border-red-500/20"
                        )}>
                          <span className={cn("w-1.5 h-1.5 rounded-full", 
                            item.status === "active" ? "bg-emerald-500" :
                            item.status === "pending" ? "bg-amber-500" : "bg-red-500"
                          )} />
                          {item.status}
                        </span>
                      </td>
                      <td className="px-8 py-5">
                        <p className="text-sm font-bold text-muted-foreground">
                          {/* @ts-ignore */}
                          {new Date(item.created_at).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}
                        </p>
                      </td>
                      <td className="px-8 py-5 text-right flex items-center justify-end gap-2">
                        {item.status === "pending" && item.payment_url && (
                          <a 
                            href={item.payment_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="p-2.5 rounded-xl border border-amber-500/20 bg-amber-500/10 text-amber-500 hover:bg-amber-500 hover:text-white transition-all group/pay flex items-center gap-2"
                            title="Complete Payment"
                          >
                            <span className="text-[10px] font-black uppercase hidden lg:block">Pay Now</span>
                            <ExternalLink className="w-5 h-5 group-hover/pay:scale-110 transition-transform" />
                          </a>
                        )}
                        <button 
                          onClick={async () => {
                            try {
                              showToast("Generating your invoice...", "info");
                              const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8085"}/billing/invoice/${item.id}`, {
                                headers: {
                                  'Authorization': `Bearer ${localStorage.getItem("access_token")}`
                                }
                              });
                              if (!response.ok) throw new Error("Failed to download invoice");
                              const blob = await response.blob();
                              const url = window.URL.createObjectURL(blob);
                              const a = document.createElement('a');
                              a.href = url;
                              a.download = `Invoice_${item.id.substring(0, 8)}.pdf`;
                              document.body.appendChild(a);
                              a.click();
                              a.remove();
                              showToast("Invoice downloaded successfully", "success");
                            } catch (error) {
                              showToast("Failed to download invoice. Please try again.", "error");
                            }
                          }}
                          className="p-2.5 rounded-xl border border-border text-muted-foreground hover:bg-primary/10 hover:text-primary hover:border-primary/20 transition-all group/btn"
                          title="Download Invoice"
                        >
                          <FileText className="w-5 h-5 group-hover/btn:scale-110 transition-transform" />
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>

        {subscription && subscription.status === "active" && (
          <div className="flex justify-center">
            <button 
              onClick={() => {
                if (confirm("Are you sure you want to cancel your subscription? Your access will revert to Free tier immediately.")) {
                  cancelMutation.mutate();
                }
              }}
              disabled={cancelMutation.isPending}
              className="flex items-center gap-2 text-xs font-bold text-red-500 hover:text-red-600 transition-colors py-2 px-4 rounded-xl hover:bg-red-50"
            >
              {cancelMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <AlertCircle className="w-4 h-4" />}
              Cancel Current Subscription
            </button>
          </div>
        )}
      </div>

      {/* Trust Badges / Footer */}
      <div className="pt-12 border-t border-border flex flex-col md:flex-row items-center justify-center gap-12 grayscale opacity-50">
        <div className="flex items-center gap-3">
          <ShieldCheck className="w-8 h-8" />
          <span className="font-bold">Secure DOKU Payment</span>
        </div>
        <div className="flex items-center gap-3">
          <Sparkles className="w-8 h-8" />
          <span className="font-bold">Trusted by 500+ Solar Pros</span>
        </div>
      </div>
    </div>
  );
}
