"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { userService, User } from "@/services/user";
import { billingService } from "@/services/billing";
import { notificationService, NotificationPreferences } from "@/services/notification";
import { Card } from "@/components/ui/card";
import { 
  User as UserIcon, 
  Building2, 
  Share2, 
  Bell, 
  Upload, 
  Check, 
  Copy, 
  Loader2,
  ExternalLink,
  ShieldCheck,
  Calendar,
  CreditCard,
  Clock,
  Globe
} from "lucide-react";
import { cn } from "@/lib/utils";
import { usePlan } from "@/components/providers/plan-provider";

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const { checkAccess } = usePlan();
  const [activeTab, setActiveTab] = useState("profile");
  const [copied, setCopied] = useState(false);
  
  const handleTabChange = (tabId: string) => {
    if (tabId === "branding") {
      const hasAccess = checkAccess(
        "enterprise", 
        "Company Branding", 
        "White-label branding and custom logos are Enterprise features. Personalize the platform to match your corporate identity."
      );
      if (!hasAccess) return;
    }
    
    if (tabId === "sharing") {
      const hasAccess = checkAccess(
        "enterprise", 
        "Public ESG Sharing", 
        "Publicly shareable ESG dashboards and external analytics pages are exclusive to Enterprise users. Showcase your sustainability achievements to the world."
      );
      if (!hasAccess) return;
    }

    setActiveTab(tabId);
  };
  
  // Branding State
  const [companyName, setCompanyName] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  // Notification State
  const [prefForm, setPrefForm] = useState<Partial<NotificationPreferences>>({});

  // Queries
  const { data: user, isLoading: isUserLoading } = useQuery({
    queryKey: ["me"],
    queryFn: () => userService.getMe(),
  });

  const { data: sub, isLoading: isSubLoading } = useQuery({
    queryKey: ["subscription"],
    queryFn: () => billingService.getSubscription(),
  });

  const { data: prefs, isLoading: isPrefsLoading } = useQuery({
    queryKey: ["notification-preferences"],
    queryFn: () => notificationService.getPreferences(),
  });

  const isLoading = isUserLoading || isSubLoading || isPrefsLoading;

  useEffect(() => {
    if (user) {
      setCompanyName(user.company_name || "");
      setPreviewUrl(user.company_logo_url || null);
    }
    if (prefs) {
      setPrefForm(prefs);
    }
  }, [user, prefs]);

  // Mutations
  const updateBrandingMutation = useMutation({
    mutationFn: () => userService.updateBranding(companyName, selectedFile || undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
      alert("Branding updated successfully!");
    }
  });

  const updatePrefsMutation = useMutation({
    mutationFn: (payload: Partial<NotificationPreferences>) => notificationService.updatePreferences(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-preferences"] });
    }
  });

  const toggleESGMutation = useMutation({
    mutationFn: (enable: boolean) => enable ? userService.enableESGShare() : userService.disableESGShare(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
    }
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const copyShareLink = () => {
    if (!user?.esg_share_token) return;
    const url = `${window.location.origin}/public/esg/${user.esg_share_token}`;
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center p-20">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
      </div>
    );
  }

  const tabs = [
    { id: "profile", label: "General", icon: UserIcon },
    { id: "branding", label: "Branding", icon: Building2 },
    { id: "sharing", label: "Public Sharing", icon: Share2 },
    { id: "notifications", label: "Notifications", icon: Bell },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Settings</h1>
        <p className="text-muted-foreground">Manage your account, company branding, and preferences.</p>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Sidebar Tabs */}
        <div className="lg:w-64 space-y-2">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id)}
              className={cn(
                "w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-sm font-bold transition-all",
                activeTab === tab.id 
                  ? "bg-primary text-white shadow-premium" 
                  : "text-muted-foreground hover:bg-accent hover:text-foreground"
              )}
            >
              <tab.icon className="w-5 h-5" />
              {tab.label}
            </button>
          ))}
        </div>

        {/* Content Area */}
        <div className="flex-1 max-w-2xl">
          {activeTab === "profile" && (
            <Card className="space-y-6">
              <h2 className="text-xl font-bold mb-6">Personal Profile</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Full Name</label>
                  <div className="p-4 bg-accent/30 border border-border rounded-xl font-bold text-foreground">
                    {user?.name}
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Email Address</label>
                  <div className="p-4 bg-accent/30 border border-border rounded-xl font-bold text-foreground flex items-center gap-2 group">
                    <span className="truncate">{user?.email}</span>
                    {user?.email_verified && (
                      <div className="flex items-center justify-center w-5 h-5 rounded-full bg-emerald-500 text-white shadow-sm shadow-emerald-200 shrink-0" title="Verified Account">
                        <Check className="w-3 h-3 stroke-[4]" />
                      </div>
                    )}
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-bold text-muted-foreground uppercase ml-1">AI Accuracy (Efficiency)</label>
                  <div className="p-4 bg-amber-500/5 border border-amber-500/10 rounded-xl">
                    <p className="text-xl font-black text-amber-600">{(user?.forecast_efficiency || 0) * 100}%</p>
                    <p className="text-[10px] text-amber-600/70 font-bold uppercase mt-1 tracking-tight">Active Forecasting Accuracy</p>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Member Since</label>
                  <div className="p-4 bg-accent/30 border border-border rounded-xl font-bold text-foreground">
                    {user?.created_at ? new Date(user.created_at).toLocaleDateString("id-ID", { 
                      year: "numeric", 
                      month: "long", 
                      day: "numeric" 
                    }) : "-"}
                  </div>
                </div>
              </div>

              <div className="space-y-2 pt-4">
                <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Account Subscription</label>
                <div className={cn(
                  "p-6 rounded-3xl border transition-all duration-500 overflow-hidden relative",
                  sub?.plan_tier === "enterprise" ? "bg-gradient-to-br from-purple-500/10 via-indigo-500/5 to-transparent border-purple-500/20 shadow-xl shadow-purple-500/5" :
                  sub?.plan_tier === "pro" ? "bg-gradient-to-br from-amber-500/10 via-orange-500/5 to-transparent border-amber-500/20 shadow-xl shadow-amber-500/5" :
                  "bg-accent/30 border-border"
                )}>
                  {/* Decorative background icon */}
                  <ShieldCheck className={cn(
                    "absolute -right-8 -bottom-8 w-48 h-48 opacity-[0.03] rotate-12",
                    sub?.plan_tier === "enterprise" ? "text-purple-600" : "text-amber-600"
                  )} />

                  <div className="relative z-10">
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                      <div className="space-y-4">
                        <div className="flex items-center gap-3">
                          <div className={cn(
                            "px-4 py-1.5 rounded-full text-xs font-black tracking-widest uppercase border",
                            sub?.plan_tier === "enterprise" ? "bg-purple-500 text-white border-purple-400" :
                            sub?.plan_tier === "pro" ? "bg-amber-500 text-white border-amber-400" :
                            "bg-accent text-muted-foreground border-border"
                          )}>
                            {sub?.plan_tier || "Free"} Plan
                          </div>
                          {sub?.status === "active" && (
                            <span className="flex items-center gap-1.5 text-[10px] font-bold text-emerald-600 bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20">
                              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                              ACTIVE
                            </span>
                          )}
                        </div>

                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                          <div className="flex items-center gap-3 text-muted-foreground">
                            <div className="w-10 h-10 rounded-xl bg-card border border-border flex items-center justify-center">
                              <CreditCard className="w-5 h-5" />
                            </div>
                            <div>
                              <p className="text-[10px] font-bold uppercase tracking-tight">Billing Amount</p>
                              <p className="text-sm font-black text-foreground">
                                {sub?.currency} {sub?.amount?.toLocaleString()}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-3 text-muted-foreground">
                            <div className="w-10 h-10 rounded-xl bg-card border border-border flex items-center justify-center">
                              <Calendar className="w-5 h-5" />
                            </div>
                            <div>
                              <p className="text-[10px] font-bold uppercase tracking-tight">Next Renewal</p>
                              <p className="text-sm font-black text-foreground">
                                {sub?.expires_at ? new Date(sub.expires_at).toLocaleDateString("id-ID", { 
                                  year: "numeric", 
                                  month: "long", 
                                  day: "numeric" 
                                }) : "Never"}
                              </p>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div className="flex flex-col gap-2">
                        <button className="px-6 py-3 bg-foreground text-background font-bold rounded-2xl hover:opacity-90 transition-all text-sm shadow-xl shadow-black/10">
                          Manage Billing
                        </button>
                        <p className="text-[10px] text-center text-muted-foreground font-medium">Auto-renews on {sub?.expires_at ? new Date(sub.expires_at).toLocaleDateString() : "-"}</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          )}

          {activeTab === "branding" && (
            <Card className="space-y-8">
              <h2 className="text-xl font-bold mb-6">Company Branding</h2>
              
              <div className="flex items-center gap-8">
                <div className="relative group">
                  <div className="w-24 h-24 rounded-3xl bg-accent border border-border overflow-hidden flex items-center justify-center">
                    {previewUrl ? (
                      <img src={previewUrl.startsWith("/") ? `${process.env.NEXT_PUBLIC_API_URL}${previewUrl}` : previewUrl} alt="Logo" className="w-full h-full object-cover" />
                    ) : (
                      <Building2 className="w-10 h-10 text-muted-foreground/30" />
                    )}
                  </div>
                  <label className="absolute -bottom-2 -right-2 p-2 bg-primary text-white rounded-xl shadow-premium cursor-pointer hover:scale-110 transition-all">
                    <Upload className="w-4 h-4" />
                    <input type="file" className="hidden" onChange={handleFileChange} accept="image/*" />
                  </label>
                </div>
                <div className="flex-1 space-y-4">
                  <div className="space-y-2">
                    <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Company Name</label>
                    <input 
                      type="text" 
                      value={companyName}
                      onChange={(e) => setCompanyName(e.target.value)}
                      placeholder="Enter company name"
                      className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold"
                    />
                    <p className="text-xs text-muted-foreground">This name will appear on all generated reports.</p>
                  </div>
                  
                  <div className="p-4 bg-primary/5 rounded-2xl border border-primary/10">
                    <p className="text-xs font-bold text-primary uppercase mb-2">Logo Guide</p>
                    <ul className="text-[10px] space-y-1 text-muted-foreground font-medium">
                      <li>• Format: PNG, JPG, or SVG</li>
                      <li>• Recommended: Square (1:1) or Horizontal</li>
                      <li>• Min Size: 512x512px for best quality</li>
                      <li>• Max File Size: 2MB</li>
                    </ul>
                  </div>
                </div>
              </div>

              <div className="pt-6 border-t border-border flex justify-end">
                <button 
                  onClick={() => updateBrandingMutation.mutate()}
                  disabled={updateBrandingMutation.isPending}
                  className="px-8 py-3 bg-primary text-white font-bold rounded-2xl shadow-premium hover:opacity-90 transition-all flex items-center gap-2"
                >
                  {updateBrandingMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                  Save Branding Settings
                </button>
              </div>
            </Card>
          )}

          {activeTab === "sharing" && (
            <Card className="space-y-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-xl font-bold">Public ESG Page</h2>
                  <p className="text-sm text-muted-foreground">Allow anyone with the link to view your live ESG impact.</p>
                </div>
                <button 
                  onClick={() => toggleESGMutation.mutate(!user?.esg_share_enabled)}
                  className={cn(
                    "relative inline-flex h-7 w-14 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2  focus-visible:ring-white  focus-visible:ring-opacity-75",
                    user?.esg_share_enabled ? "bg-primary" : "bg-accent"
                  )}
                >
                  <span className={cn(
                    "pointer-events-none inline-block h-6 w-6 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out",
                    user?.esg_share_enabled ? "translate-x-7" : "translate-x-0"
                  )} />
                </button>
              </div>

              {user?.esg_share_enabled && (
                <div className="space-y-4 pt-4 border-t border-border animate-in fade-in slide-in-from-top-2">
                  <div className="space-y-2">
                    <label className="text-xs font-bold text-muted-foreground uppercase ml-1">Your Public Sharing Link</label>
                    <div className="flex gap-2">
                      <div className="flex-1 p-3 bg-accent/30 border border-border rounded-xl text-xs font-mono truncate">
                        {`${window.location.origin}/public/esg/${user.esg_share_token}`}
                      </div>
                      <button 
                        onClick={copyShareLink}
                        className="px-4 py-2 bg-card border border-border rounded-xl hover:bg-accent transition-all flex items-center gap-2 text-xs font-bold"
                      >
                        {copied ? <Check className="w-4 h-4 text-emerald-500" /> : <Copy className="w-4 h-4" />}
                        Copy
                      </button>
                    </div>
                  </div>
                  <Link 
                    href={`/public/esg/${user.esg_share_token}`}
                    target="_blank"
                    className="inline-flex items-center gap-2 text-sm font-bold text-primary hover:underline"
                  >
                    <ExternalLink className="w-4 h-4" />
                    Preview Public Page
                  </Link>
                </div>
              )}

              {!user?.esg_share_enabled && (
                <div className="p-6 rounded-2xl bg-accent/30 border border-dashed border-border flex flex-col items-center justify-center text-center">
                  <Share2 className="w-10 h-10 text-muted-foreground/30 mb-3" />
                  <p className="text-sm text-muted-foreground font-medium">Public sharing is currently disabled. Enable it to share your green achievements.</p>
                </div>
              )}
            </Card>
          )}

          {activeTab === "notifications" && (
            <Card className="space-y-8">
              <div>
                <h2 className="text-xl font-bold mb-2">Notification Channels</h2>
                <p className="text-sm text-muted-foreground">Choose how you want to receive your daily reports and alerts.</p>
              </div>

              <div className="space-y-6">
                {/* Email Channel */}
                <div className="p-5 rounded-[2rem] border border-border bg-accent/10 flex items-center justify-between group hover:border-primary/30 transition-all duration-300">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform duration-300">
                      <Bell className="w-6 h-6" />
                    </div>
                    <div>
                      <p className="font-bold text-base">Email Notifications</p>
                      <p className="text-xs text-muted-foreground font-medium">Daily forecast reports & critical system alerts.</p>
                    </div>
                  </div>
                  <button 
                    onClick={() => {
                      const newValue = !prefForm.email_enabled;
                      setPrefForm(prev => ({ ...prev, email_enabled: newValue }));
                      updatePrefsMutation.mutate({ email_enabled: newValue });
                    }}
                    className={cn(
                      "relative inline-flex h-7 w-14 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none",
                      prefForm.email_enabled ? "bg-primary" : "bg-slate-200 dark:bg-slate-700"
                    )}
                  >
                    <span className={cn(
                      "pointer-events-none inline-block h-6 w-6 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out",
                      prefForm.email_enabled ? "translate-x-7" : "translate-x-0"
                    )} />
                  </button>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-6 border-t border-border">
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 mb-1">
                      <Globe className="w-4 h-4 text-muted-foreground" />
                      <label className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Timezone</label>
                    </div>
                    <select 
                      value={prefForm.timezone || "Asia/Jakarta"}
                      onChange={(e) => {
                        const val = e.target.value;
                        setPrefForm(prev => ({ ...prev, timezone: val }));
                        updatePrefsMutation.mutate({ timezone: val });
                      }}
                      className="w-full px-4 py-3 bg-background border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold text-sm appearance-none cursor-pointer"
                    >
                      <option value="Asia/Jakarta">Asia/Jakarta (GMT+7)</option>
                      <option value="Asia/Singapore">Asia/Singapore (GMT+8)</option>
                      <option value="UTC">UTC (Universal)</option>
                    </select>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center gap-2 mb-1">
                      <Clock className="w-4 h-4 text-muted-foreground" />
                      <label className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Preferred Send Time</label>
                    </div>
                    <input 
                      type="time" 
                      value={prefForm.preferred_send_time || "06:00"}
                      step="3600"
                      onChange={(e) => {
                        const val = e.target.value + (e.target.value.length === 5 ? ":00" : "");
                        setPrefForm(prev => ({ ...prev, preferred_send_time: val }));
                        updatePrefsMutation.mutate({ preferred_send_time: val });
                      }}
                      className="w-full px-4 py-3 bg-background border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold text-sm"
                    />
                  </div>
                </div>

                <div className="p-4 rounded-2xl bg-primary/5 border border-primary/10 flex items-start gap-3">
                  <div className="p-2 bg-primary/10 rounded-lg text-primary">
                    <Clock className="w-4 h-4" />
                  </div>
                  <div>
                    <p className="text-xs font-bold text-primary mb-1">Next Delivery Scheduled</p>
                    <p className="text-[10px] text-primary/70 font-medium">
                      Your next daily report will be sent at {prefForm.preferred_send_time?.substring(0, 5)} {prefForm.timezone}.
                    </p>
                  </div>
                </div>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
