"use client";

import React, { createContext, useContext, useState, ReactNode } from "react";
import { UpgradeModal } from "@/components/modals/upgrade-modal";
import { useQuery } from "@tanstack/react-query";
import { userService } from "@/services/user";

interface PlanContextType {
  checkAccess: (requiredPlan: "pro" | "enterprise", featureName: string, featureDescription: string) => boolean;
  userTier: string;
}

const PlanContext = createContext<PlanContextType | undefined>(undefined);

export function PlanProvider({ children }: { children: ReactNode }) {
  const [modalState, setModalState] = useState<{
    isOpen: boolean;
    featureName: string;
    featureDescription: string;
    requiredPlan: "pro" | "enterprise";
  }>({
    isOpen: false,
    featureName: "",
    featureDescription: "",
    requiredPlan: "pro"
  });

  const { data: user } = useQuery({ 
    queryKey: ["me"], 
    queryFn: () => userService.getMe() 
  });

  const userTier = user?.tier || "free";

  const checkAccess = (requiredPlan: "pro" | "enterprise", featureName: string, featureDescription: string) => {
    const tiers = ["free", "pro", "enterprise"];
    const userIndex = tiers.indexOf(userTier);
    const requiredIndex = tiers.indexOf(requiredPlan);

    if (userIndex < requiredIndex) {
      setModalState({
        isOpen: true,
        featureName,
        featureDescription,
        requiredPlan
      });
      return false;
    }
    return true;
  };

  return (
    <PlanContext.Provider value={{ checkAccess, userTier }}>
      {children}
      <UpgradeModal 
        isOpen={modalState.isOpen}
        onClose={() => setModalState(prev => ({ ...prev, isOpen: false }))}
        featureName={modalState.featureName}
        featureDescription={modalState.featureDescription}
        requiredPlan={modalState.requiredPlan}
      />
    </PlanContext.Provider>
  );
}

export function usePlan() {
  const context = useContext(PlanContext);
  if (context === undefined) {
    throw new Error("usePlan must be used within a PlanProvider");
  }
  return context;
}
