"use client";

import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { deviceService, Device } from "@/services/device";
import { solarService } from "@/services/solar";
import { Card } from "@/components/ui/card";
import { 
  Plus, 
  Cpu, 
  Trash2, 
  Edit2, 
  RefreshCw, 
  Loader2,
  Key,
  Activity,
  Circle,
  Copy,
  Check,
  X
} from "lucide-react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

import ConfirmationModal from "@/components/ui/confirmation-modal";
import { usePlan } from "@/components/providers/plan-provider";

export default function DevicesPage() {
  const queryClient = useQueryClient();
  const { checkAccess, userTier } = usePlan();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  
  // Confirmation Modal State
  const [confirmConfig, setConfirmConfig] = useState<{
    isOpen: boolean;
    type: "create" | "delete" | "rotate" | null;
    id?: string;
  }>({ isOpen: false, type: null });

  // Form State
  const [formData, setFormData] = useState({
    name: "",
    solar_profile_id: ""
  });

  // Queries
  const { data: devices, isLoading: isDevicesLoading } = useQuery({
    queryKey: ["devices"],
    queryFn: () => deviceService.listDevices(),
  });

  const { data: profiles } = useQuery({
    queryKey: ["solar-profiles"],
    queryFn: () => solarService.getProfiles(),
  });

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: any) => deviceService.createDevice(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["devices"] });
      closeModal();
    }
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deviceService.deleteDevice(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["devices"] });
    }
  });

  const rotateKeyMutation = useMutation({
    mutationFn: (id: string) => deviceService.rotateKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["devices"] });
    }
  });

  const openModal = (device?: Device) => {
    if (!device) {
      // Logic for device limits
      const currentCount = devices?.length || 0;
      if (userTier === "free" && currentCount >= 1) {
        checkAccess(
          "pro", 
          "Connect More Devices", 
          "The Free plan is limited to 1 IoT device. Upgrade to Pro to manage up to 10 devices and unlock advanced telemetry features."
        );
        return;
      }
      if (userTier === "pro" && currentCount >= 10) {
        checkAccess(
          "enterprise", 
          "Enterprise Fleet", 
          "You've reached the 10-device limit on Pro. Upgrade to Enterprise for unlimited device registration and fleet management."
        );
        return;
      }

      setEditingDevice(null);
      setFormData({
        name: "",
        solar_profile_id: profiles?.[0]?.id || ""
      });
    } else {
      setEditingDevice(device);
      setFormData({
        name: device.name,
        solar_profile_id: device.solar_profile_id
      });
    }
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingDevice(null);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(text);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setConfirmConfig({ isOpen: true, type: "create" });
  };

  const executeAction = () => {
    if (confirmConfig.type === "delete" && confirmConfig.id) {
      deleteMutation.mutate(confirmConfig.id);
    } else if (confirmConfig.type === "rotate" && confirmConfig.id) {
      rotateKeyMutation.mutate(confirmConfig.id);
    } else if (confirmConfig.type === "create") {
      createMutation.mutate(formData);
    }
    setConfirmConfig({ isOpen: false, type: null });
  };

  if (isDevicesLoading) {
    return (
      <div className="h-full w-full flex items-center justify-center p-20">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight mb-2">Device Management</h1>
          <p className="text-muted-foreground">Monitor and manage your IoT solar telemetry devices.</p>
        </div>
        <button 
          onClick={() => openModal()}
          className="flex items-center justify-center gap-2 px-6 py-3 bg-primary text-white font-bold rounded-2xl shadow-premium hover:translate-y-[-2px] transition-all"
        >
          <Plus className="w-5 h-5" />
          Register Device
        </button>
      </div>

      <div className="grid grid-cols-1 gap-6">
        <Card className="p-0 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-accent/30 border-b border-border">
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Device Name</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Status</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Device Key</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Last Heartbeat</th>
                  <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {devices?.map((device) => (
                  <tr key={device.id} className="hover:bg-accent/10 transition-colors">
                    <td className="px-6 py-5">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                          <Cpu className="w-5 h-5" />
                        </div>
                        <div>
                          <p className="font-bold">{device.name}</p>
                          <p className="text-xs text-muted-foreground">ID: {device.id.slice(0, 8)}...</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className={cn(
                        "inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border",
                        device.status === "online" 
                          ? "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" 
                          : "bg-red-500/10 text-red-500 border-red-500/20"
                      )}>
                        <Circle className={cn("w-2 h-2 fill-current", device.status === "online" ? "animate-pulse" : "")} />
                        {device.status.toUpperCase()}
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex items-center gap-2">
                        <code className="text-xs bg-accent/50 px-3 py-1.5 rounded-lg border border-border font-mono">
                          {device.device_key.slice(0, 4)}••••••••{device.device_key.slice(-4)}
                        </code>
                        <button 
                          onClick={() => copyToClipboard(device.device_key)}
                          className="p-1.5 rounded-lg hover:bg-accent text-muted-foreground transition-all"
                        >
                          {copiedKey === device.device_key ? <Check className="w-4 h-4 text-emerald-500" /> : <Copy className="w-4 h-4" />}
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-5 text-sm text-muted-foreground font-medium">
                      {device.last_heartbeat ? new Date(device.last_heartbeat).toLocaleString() : "Never"}
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button 
                          onClick={() => setConfirmConfig({ isOpen: true, type: "rotate", id: device.id })}
                          title="Rotate Key"
                          className="p-2 rounded-lg bg-amber-500/10 text-amber-500 hover:bg-amber-500 hover:text-white transition-all"
                        >
                          <RefreshCw className="w-4 h-4" />
                        </button>
                        <button 
                          onClick={() => setConfirmConfig({ isOpen: true, type: "delete", id: device.id })}
                          className="p-2 rounded-lg bg-red-500/10 text-red-500 hover:bg-red-50 hover:text-red-600 transition-all"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {(!devices || devices.length === 0) && (
            <div className="py-20 flex flex-col items-center justify-center text-center">
              <Cpu className="w-16 h-16 text-muted-foreground/30 mb-4" />
              <h3 className="text-xl font-bold text-muted-foreground">No devices registered</h3>
              <p className="text-muted-foreground mt-2 max-w-xs">Register your first IoT device to start streaming telemetry data.</p>
            </div>
          )}
        </Card>
      </div>

      {/* Register Device Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
          <div className="w-full max-w-lg bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden">
            <div className="p-8 border-b border-border flex items-center justify-between">
              <h2 className="text-2xl font-bold">Register Device</h2>
              <button onClick={closeModal} className="p-2 rounded-xl hover:bg-accent transition-all text-muted-foreground">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-8 space-y-6">
              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Device Name</label>
                <input 
                  type="text" 
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  placeholder="e.g. Inverter Unit A"
                  className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                  required
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Associated Solar Profile</label>
                <select 
                  value={formData.solar_profile_id}
                  onChange={(e) => setFormData({...formData, solar_profile_id: e.target.value})}
                  className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none appearance-none"
                  required
                >
                  {profiles?.map(p => (
                    <option key={p.id} value={p.id}>{p.site_name} ({p.capacity_kwp} kWp)</option>
                  ))}
                </select>
              </div>

              <div className="p-4 rounded-2xl bg-amber-500/5 border border-amber-500/10 flex gap-3">
                <Key className="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
                <p className="text-xs text-amber-600 font-medium leading-relaxed">
                  Registering a device will generate a unique <strong>Device Key</strong>. You will need this key to configure your IoT hardware for telemetry ingestion.
                </p>
              </div>

              <div className="flex gap-4 pt-4">
                <button 
                  type="button" 
                  onClick={closeModal}
                  className="flex-1 py-4 font-bold text-muted-foreground bg-accent hover:bg-border transition-all rounded-2xl"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  disabled={createMutation.isPending}
                  className="flex-1 py-4 font-bold text-white bg-primary shadow-premium hover:opacity-90 transition-all rounded-2xl flex items-center justify-center"
                >
                  {createMutation.isPending ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    "Register Device"
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <ConfirmationModal 
        isOpen={confirmConfig.isOpen}
        onClose={() => setConfirmConfig({ isOpen: false, type: null })}
        onConfirm={executeAction}
        title={
          confirmConfig.type === "delete" ? "Delete Device?" :
          confirmConfig.type === "rotate" ? "Rotate Device Key?" : "Register New Device?"
        }
        description={
          confirmConfig.type === "delete" ? "This action will permanently remove the device from our system. Any hardware using this key will stop sending data." :
          confirmConfig.type === "rotate" ? "This will invalidate the current device key. You must update your hardware configuration with the new key immediately." : 
          "Confirm registering this new IoT device. This will generate a unique access key."
        }
        confirmText={
          confirmConfig.type === "delete" ? "Delete Device" :
          confirmConfig.type === "rotate" ? "Rotate Key" : "Register Device"
        }
        variant={confirmConfig.type === "delete" ? "danger" : confirmConfig.type === "rotate" ? "warning" : "primary"}
        isLoading={createMutation.isPending || deleteMutation.isPending || rotateKeyMutation.isPending}
      />
    </div>
  );
}
