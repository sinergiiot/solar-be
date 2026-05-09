"use client";

import React, { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  CheckCircle2, 
  Circle, 
  Sun, 
  CloudSun, 
  PenLine, 
  BellRing,
  ChevronDown,
  ChevronUp,
  X
} from "lucide-react";
import { cn } from "@/lib/utils";
import Link from "next/link";

const tasks = [
  {
    id: "profile",
    title: "Lengkapi Solar Profile",
    description: "Isi kapasitas panel (kWp) dan lokasi koordinat agar sistem bisa menghitung radiasi matahari.",
    icon: Sun,
    href: "/dashboard/profiles",
    completed: false,
  },
  {
    id: "forecast",
    title: "Generate Forecast Pertama",
    description: "Klik ambil forecast untuk pertama kalinya agar model mulai memproses data cuaca hari ini.",
    icon: CloudSun,
    href: "/dashboard/forecasts",
    completed: false,
  },
  {
    id: "actual",
    title: "Input Data Actual",
    description: "Masukkan angka kWh yang dihasilkan panel Anda untuk melatih akurasi AI.",
    icon: PenLine,
    href: "/dashboard/actuals",
    completed: false,
  },
  {
    id: "notifications",
    title: "Aktifkan Notifikasi",
    description: "Terima laporan forecast harian setiap pagi melalui Email atau Telegram.",
    icon: BellRing,
    href: "/dashboard/settings",
    completed: false,
  }
];

export default function GettingStartedChecklist() {
  const [isMinimized, setIsMinimized] = useState(false);
  const [isVisible, setIsVisible] = useState(true);

  const completedCount = tasks.filter(t => t.completed).length;
  const progress = (completedCount / tasks.length) * 100;

  if (!isVisible) return null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn(
        "bg-card border border-border rounded-[2rem] shadow-premium overflow-hidden transition-all duration-500 mb-8",
        isMinimized ? "max-h-[80px]" : "max-h-[600px]"
      )}
    >
      {/* Header */}
      <div className="p-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <div>
            <h3 className="font-bold text-lg">Getting Started</h3>
            <p className="text-xs text-muted-foreground font-medium">
              {completedCount} of {tasks.length} tasks completed ({progress}%)
            </p>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <button 
            onClick={() => setIsMinimized(!isMinimized)}
            className="p-2 rounded-xl hover:bg-accent text-muted-foreground transition-all"
          >
            {isMinimized ? <ChevronDown className="w-5 h-5" /> : <ChevronUp className="w-5 h-5" />}
          </button>
          <button 
            onClick={() => setIsVisible(false)}
            className="p-2 rounded-xl hover:bg-red-50 text-muted-foreground hover:text-red-500 transition-all"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="h-1 w-full bg-accent">
        <motion.div 
          initial={{ width: 0 }}
          animate={{ width: `${progress}%` }}
          className="h-full bg-primary transition-all duration-500"
        />
      </div>

      <AnimatePresence>
        {!isMinimized && (
          <motion.div 
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"
          >
            {tasks.map((task) => (
              <Link 
                key={task.id}
                href={task.href}
                className={cn(
                  "p-5 rounded-2xl border transition-all group text-left",
                  task.completed 
                    ? "bg-emerald-50/30 border-emerald-100 dark:bg-emerald-500/5 dark:border-emerald-500/20" 
                    : "bg-background/50 border-border hover:border-primary hover:bg-card hover:shadow-md"
                )}
              >
                <div className="flex items-center justify-between mb-4">
                  <div className={cn(
                    "w-10 h-10 rounded-xl flex items-center justify-center transition-colors",
                    task.completed ? "bg-emerald-500 text-white" : "bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white"
                  )}>
                    <task.icon className="w-5 h-5" />
                  </div>
                  {task.completed ? (
                    <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                  ) : (
                    <Circle className="w-5 h-5 text-muted-foreground" />
                  )}
                </div>
                <h4 className="font-bold mb-2">{task.title}</h4>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {task.description}
                </p>
              </Link>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}
