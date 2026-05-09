"use client";

import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Sun, 
  ChevronRight, 
  X, 
  Zap, 
  LineChart, 
  ShieldCheck 
} from "lucide-react";
import { cn } from "@/lib/utils";

const steps = [
  {
    title: "Welcome to SolarForecast",
    description: "Your next-generation energy intelligence platform. Let's take a quick tour of your new dashboard.",
    icon: Sun,
    color: "bg-amber-500",
  },
  {
    title: "Monitor Real-time Production",
    description: "Track actual energy output vs. AI predictions. Our models update every 15 minutes for maximum precision.",
    icon: Zap,
    color: "bg-primary",
  },
  {
    title: "Deep Analytics & ESG",
    description: "Generate detailed energy reports and track your carbon offset impact with automated ESG summaries.",
    icon: LineChart,
    color: "bg-emerald-500",
  },
  {
    title: "Secure Device Ingestion",
    description: "Connect your IoT devices securely with military-grade encryption and automated key rotation.",
    icon: ShieldCheck,
    color: "bg-blue-500",
  }
];

export default function OnboardingGuide() {
  const [isVisible, setIsVisible] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);

  useEffect(() => {
    const hasSeenOnboarding = localStorage.getItem("hasSeenOnboarding");
    if (!hasSeenOnboarding) {
      setIsVisible(true);
    }
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    localStorage.setItem("hasSeenOnboarding", "true");
  };

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      handleClose();
    }
  };

  if (!isVisible) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
        <motion.div
          initial={{ opacity: 0, scale: 0.9, y: 20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.9, y: 20 }}
          className="relative w-full max-w-2xl bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden"
        >
          {/* Close Button */}
          <button 
            onClick={handleClose}
            className="absolute top-6 right-6 p-2 rounded-xl bg-accent/50 text-muted-foreground hover:text-foreground transition-all z-10"
          >
            <X className="w-5 h-5" />
          </button>

          <div className="flex flex-col md:flex-row h-full">
            {/* Visual Side */}
            <div className={cn(
              "md:w-5/12 p-12 flex items-center justify-center text-white transition-colors duration-500 min-h-[200px] md:min-h-full",
              steps[currentStep].color
            )}>
              <motion.div
                key={currentStep}
                initial={{ opacity: 0, rotate: -10, scale: 0.8 }}
                animate={{ opacity: 1, rotate: 0, scale: 1 }}
                transition={{ type: "spring", damping: 12 }}
              >
                {/* Dynamic Icon/Visual */}
                <div className="w-24 h-24 rounded-[2rem] bg-white/20 backdrop-blur-md flex items-center justify-center shadow-xl">
                  {React.createElement(steps[currentStep].icon, { className: "w-12 h-12" })}
                </div>
              </motion.div>
            </div>

            {/* Content Side */}
            <div className="md:w-7/12 p-10 md:p-12 flex flex-col justify-between">
              <div>
                <div className="flex gap-1 mb-6">
                  {steps.map((_, idx) => (
                    <div 
                      key={idx}
                      className={cn(
                        "h-1.5 rounded-full transition-all duration-300",
                        idx === currentStep ? "w-8 bg-primary" : "w-2 bg-accent"
                      )}
                    />
                  ))}
                </div>

                <AnimatePresence mode="wait">
                  <motion.div
                    key={currentStep}
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: -20 }}
                    transition={{ duration: 0.3 }}
                  >
                    <h2 className="text-3xl font-bold tracking-tight mb-4">
                      {steps[currentStep].title}
                    </h2>
                    <p className="text-lg text-muted-foreground leading-relaxed">
                      {steps[currentStep].description}
                    </p>
                  </motion.div>
                </AnimatePresence>
              </div>

              <div className="flex items-center justify-between mt-12">
                <button 
                  onClick={handleClose}
                  className="text-sm font-bold text-muted-foreground hover:text-foreground transition-all"
                >
                  Skip Guide
                </button>
                <button 
                  onClick={handleNext}
                  className="flex items-center gap-2 px-8 py-4 bg-primary text-white font-bold rounded-2xl shadow-premium hover:translate-y-[-2px] transition-all group"
                >
                  {currentStep === steps.length - 1 ? "Get Started" : "Next Step"}
                  <ChevronRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                </button>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}
