"use client";

import Sidebar from "@/components/layout/sidebar";
import Navbar from "@/components/layout/navbar";
import OnboardingGuide from "@/components/layout/onboarding-guide";
import { useSidebar } from "@/components/providers/sidebar-provider";
import { cn } from "@/lib/utils";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isCollapsed } = useSidebar();

  return (
    <div className="min-h-screen bg-background transition-all duration-300">
      <OnboardingGuide />
      <Sidebar />
      <div className={cn(
        "flex flex-col min-h-screen transition-all duration-300",
        isCollapsed ? "lg:ml-[8rem]" : "lg:ml-[19rem]"
      )}>
        <Navbar />
        <main className="p-8 flex-1">
          {children}
        </main>
      </div>
    </div>
  );
}
