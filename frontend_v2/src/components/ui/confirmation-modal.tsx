"use client";

import React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { AlertTriangle, X, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface ConfirmationModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "danger" | "warning" | "primary";
  isLoading?: boolean;
}

export default function ConfirmationModal({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = "Confirm",
  cancelText = "Cancel",
  variant = "primary",
  isLoading = false,
}: ConfirmationModalProps) {
  if (!isOpen) return null;

  const variants = {
    danger: "bg-red-500 hover:bg-red-600 shadow-red-200",
    warning: "bg-amber-500 hover:bg-amber-600 shadow-amber-200",
    primary: "bg-primary hover:bg-primary/90 shadow-primary/20",
  };

  const iconColors = {
    danger: "text-red-500 bg-red-50",
    warning: "text-amber-500 bg-amber-50",
    primary: "text-primary bg-primary/10",
  };

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-[110] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 10 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 10 }}
          className="w-full max-w-md bg-card border border-border rounded-[2rem] shadow-2xl overflow-hidden"
        >
          <div className="p-8">
            <div className="flex items-center justify-between mb-6">
              <div className={cn("w-12 h-12 rounded-2xl flex items-center justify-center", iconColors[variant])}>
                <AlertTriangle className="w-6 h-6" />
              </div>
              <button onClick={onClose} className="p-2 rounded-xl hover:bg-accent transition-all text-muted-foreground">
                <X className="w-5 h-5" />
              </button>
            </div>

            <h3 className="text-xl font-bold mb-2">{title}</h3>
            <p className="text-muted-foreground leading-relaxed">{description}</p>

            <div className="flex gap-3 mt-8">
              <button
                onClick={onClose}
                disabled={isLoading}
                className="flex-1 py-3 font-bold text-muted-foreground bg-accent hover:bg-border transition-all rounded-xl"
              >
                {cancelText}
              </button>
              <button
                onClick={onConfirm}
                disabled={isLoading}
                className={cn(
                  "flex-1 py-3 font-bold text-white transition-all rounded-xl shadow-lg flex items-center justify-center gap-2",
                  variants[variant]
                )}
              >
                {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : confirmText}
              </button>
            </div>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}
