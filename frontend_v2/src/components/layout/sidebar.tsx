"use client";
 
import React, { useState, useEffect } from "react";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { 
  LayoutDashboard, 
  Cpu, 
  Sun, 
  FileBarChart, 
  Settings, 
  ShieldCheck, 
  Database,
  Users,
  LogOut,
  CreditCard
} from "lucide-react";
import { cn } from "@/lib/utils";
import { authService } from "@/services/auth";
import { useSidebar } from "@/components/providers/sidebar-provider";
import { ChevronLeft, ChevronRight } from "lucide-react";

const menuItems = [
  { icon: LayoutDashboard, label: "Overview", href: "/dashboard" },
  { icon: Sun, label: "Solar Profiles", href: "/dashboard/profiles" },
  { icon: Cpu, label: "Devices", href: "/dashboard/devices" },
  { icon: FileBarChart, label: "Reports", href: "/dashboard/reports" },
  { icon: Database, label: "Documentation", href: "/docs" },
  { icon: CreditCard, label: "Upgrade Plan", href: "/dashboard/billing" },
  { icon: ShieldCheck, label: "Admin Panel", href: "/dashboard/admin", admin: true },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { isCollapsed, toggleSidebar, isHovered, setIsHovered } = useSidebar();
  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    setUser(authService.getUser());
  }, []);

  const handleLogout = () => {
    authService.logout();
    router.push("/login");
  };

  const showFull = !isCollapsed || isHovered;

  const filteredMenuItems = menuItems.filter(item => {
    if (item.admin) {
      return user?.role === "admin";
    }
    return true;
  });

  return (
    <aside 
      onMouseEnter={() => isCollapsed && setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className={cn(
        "fixed left-6 top-6 bottom-6 bg-card/80 backdrop-blur-xl border border-border rounded-[2.5rem] hidden lg:flex flex-col z-50 shadow-premium transition-all duration-500 ease-in-out",
        showFull ? "w-64" : "w-20"
      )}
    >
      <div className="p-6 flex items-center justify-between relative min-h-[80px]">
        <Link href="/dashboard" className={cn("flex items-center gap-3 transition-all duration-300", !showFull && "opacity-0 invisible w-0 overflow-hidden")}>
          <div className="flex items-center justify-center min-w-[40px] h-10 rounded-xl bg-primary shadow-premium">
            <Sun className="text-white w-6 h-6" />
          </div>
          <span className="text-xl font-bold tracking-tight whitespace-nowrap">SolarForecast</span>
        </Link>
        
        {!showFull && (
          <div className="flex items-center justify-center w-full">
            <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-primary shadow-premium">
              <Sun className="text-white w-6 h-6" />
            </div>
          </div>
        )}

        <button 
          onClick={(e) => {
            e.stopPropagation();
            toggleSidebar();
          }}
          className={cn(
            "absolute -right-3 top-1/2 -translate-y-1/2 p-1.5 rounded-full bg-primary text-white shadow-premium border-2 border-background hover:scale-110 transition-all z-50",
            !showFull && !isHovered && "opacity-0" // Hide when collapsed and not hovered
          )}
        >
          {isCollapsed ? <ChevronRight className="w-3 h-3" /> : <ChevronLeft className="w-3 h-3" />}
        </button>
      </div>

      <nav className="flex-1 px-3 space-y-2 mt-4">
        {filteredMenuItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center rounded-2xl text-sm font-semibold transition-all duration-200 group relative h-12",
                isActive 
                  ? "bg-primary text-white shadow-premium" 
                  : "text-muted-foreground hover:bg-accent hover:text-foreground",
                showFull ? "px-4 gap-3" : "px-0 justify-center"
              )}
            >
              <item.icon className={cn(
                "w-5 h-5 min-w-[20px] transition-transform duration-300",
                isActive ? "text-white" : "text-muted-foreground group-hover:text-primary transition-colors",
                !showFull && "scale-110"
              )} />
              
              <span className={cn(
                "transition-all duration-300 whitespace-nowrap overflow-hidden",
                showFull ? "w-auto opacity-100" : "w-0 opacity-0 invisible"
              )}>
                {item.label}
              </span>

              {item.admin && showFull && (
                <span className="ml-auto text-[10px] bg-secondary/20 text-secondary px-2 py-0.5 rounded-full uppercase font-bold">
                  Admin
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      <div className={cn("p-4 mt-auto transition-all duration-300", !showFull && "opacity-0 invisible h-0 overflow-hidden")}>
        <div className="p-4 rounded-2xl bg-accent/50 border border-border/50">
          <p className="text-xs font-bold text-foreground uppercase tracking-wider mb-3">Settings</p>
          <Link 
            href="/dashboard/settings"
            className="flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold text-foreground hover:bg-card hover:text-foreground transition-all"
          >
            <Settings className="w-5 h-5" />
            Preferences
          </Link>
          <button 
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-4 py-3 mt-1 rounded-xl text-sm font-semibold text-red-500 hover:bg-red-50 transition-all"
          >
            <LogOut className="w-5 h-5" />
            Sign Out
          </button>
        </div>
      </div>
    </aside>
  );
}
