"use client";

import { useQuery } from "@tanstack/react-query";
import { Bell, Search, User } from "lucide-react";
import { ThemeToggle } from "./theme-toggle";
import { userService } from "@/services/user";
import { cn } from "@/lib/utils";

export default function Navbar() {
  const { data: user } = useQuery({
    queryKey: ["me"],
    queryFn: () => userService.getMe(),
  });

  return (
    <nav className="h-20 border border-border bg-card/80 backdrop-blur-md sticky top-6 z-40 mx-8 mt-6 rounded-[2.5rem] px-8 flex items-center justify-between shadow-premium">
      <div className="relative w-96 group">
        <Search className="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
        <input 
          type="text" 
          placeholder="Search for forecasts, devices..." 
          className="w-full pl-12 pr-4 py-2.5 bg-accent/30 border border-border rounded-full text-sm focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
        />
      </div>

      <div className="flex items-center gap-4">
        <ThemeToggle />
        
        <button className="relative p-2.5 text-muted-foreground hover:bg-accent hover:text-foreground rounded-xl transition-all">
          <Bell className="w-5 h-5" />
          <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-red-500 rounded-full border-2 border-card" />
        </button>

        <div className="h-8 w-[1px] bg-border" />

        <div className="flex items-center gap-3 pl-2">
          <div className="text-right hidden sm:block">
            <p className="text-sm font-bold leading-none mb-1">{user?.name || "User Account"}</p>
            <p className={cn(
              "text-[10px] font-black uppercase tracking-widest px-2.5 py-0.5 rounded-full inline-block",
              user?.tier === "enterprise" ? "bg-purple-500 text-white" : 
              user?.tier === "pro" ? "bg-amber-500 text-white" : 
              "bg-primary/10 text-primary"
            )}>
              {user?.tier || "Free"} Plan
            </p>
          </div>
          <button className="w-10 h-10 rounded-xl bg-accent border border-border flex items-center justify-center hover:bg-card transition-all">
            <User className="w-5 h-5 text-muted-foreground" />
          </button>
        </div>
      </div>
    </nav>
  );
}
