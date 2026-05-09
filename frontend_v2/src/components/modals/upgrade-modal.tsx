"use client";

import React from "react";
import { 
  X, 
  Sparkles, 
  Rocket, 
  Crown, 
  Check, 
  ArrowRight,
  ShieldCheck,
  Zap
} from "lucide-react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";

interface UpgradeModalProps {
  isOpen: boolean;
  onClose: () => void;
  featureName: string;
  featureDescription: string;
  requiredPlan: "pro" | "enterprise";
}

export function UpgradeModal({ 
  isOpen, 
  onClose, 
  featureName, 
  featureDescription, 
  requiredPlan 
}: UpgradeModalProps) {
  const router = useRouter();

  if (!isOpen) return null;

  const isEnterprise = requiredPlan === "enterprise";

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div 
        className="absolute inset-0 bg-background/80 backdrop-blur-sm transition-opacity" 
        onClick={onClose} 
      />
      
      <div className={cn(
        "relative w-full max-w-lg overflow-hidden rounded-[2.5rem] border shadow-2xl transition-all animate-in fade-in zoom-in duration-300",
        isEnterprise ? "border-purple-500/20 bg-purple-500/[0.02]" : "border-primary/20 bg-primary/[0.02]",
        "bg-card"
      )}>
        {/* Header Banner */}
        <div className={cn(
          "p-8 text-center relative overflow-hidden",
          isEnterprise ? "bg-purple-500" : "bg-primary"
        )}>
          <div className="absolute top-0 right-0 p-8 opacity-10">
            {isEnterprise ? <Crown className="w-24 h-24 text-white" /> : <Rocket className="w-24 h-24 text-white" />}
          </div>
          
          <button 
            onClick={onClose}
            className="absolute top-4 right-4 p-2 rounded-full bg-white/20 text-white hover:bg-white/30 transition-all"
          >
            <X className="w-4 h-4" />
          </button>

          <div className="relative z-10 space-y-2">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/20 text-white text-[10px] font-black uppercase tracking-widest">
              <Sparkles className="w-3 h-3" />
              Premium Feature
            </div>
            <h2 className="text-3xl font-black text-white tracking-tight">{featureName}</h2>
          </div>
        </div>

        {/* Content */}
        <div className="p-8 space-y-8">
          <div className="space-y-4">
            <p className="text-muted-foreground font-medium text-center">
              {featureDescription}
            </p>
            
            <div className="p-6 rounded-2xl bg-accent/50 border border-border/50 space-y-4">
              <p className="text-xs font-black uppercase tracking-widest text-primary">Why Upgrade to {requiredPlan.toUpperCase()}?</p>
              <div className="space-y-3">
                {[
                  "Advanced solar forecasting algorithms",
                  "Comprehensive ESG & carbon tracking",
                  "Premium 24/7 technical support",
                  "Priority data processing"
                ].map((text) => (
                  <div key={text} className="flex items-center gap-3">
                    <div className="w-5 h-5 rounded-full bg-emerald-500/10 flex items-center justify-center shrink-0">
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                    <span className="text-sm font-semibold">{text}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <button 
              onClick={() => {
                onClose();
                router.push("/dashboard/billing");
              }}
              className={cn(
                "w-full py-4 rounded-2xl font-black text-white shadow-premium hover:opacity-90 transition-all flex items-center justify-center gap-2 group",
                isEnterprise ? "bg-purple-500" : "bg-primary"
              )}
            >
              Upgrade to {requiredPlan.charAt(0).toUpperCase() + requiredPlan.slice(1)} Now
              <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </button>
            
            <div className="flex items-center justify-center gap-6 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="w-3.5 h-3.5" />
                Secure Payment
              </div>
              <div className="flex items-center gap-1.5">
                <Zap className="w-3.5 h-3.5" />
                Instant Access
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
