"use client";

import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { solarService, SolarProfile } from "@/services/solar";
import { Card } from "@/components/ui/card";
import { 
  Plus, 
  MapPin, 
  Trash2, 
  Edit2, 
  Search, 
  Loader2,
  AlertTriangle,
  Activity,
  X
} from "lucide-react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

import ConfirmationModal from "@/components/ui/confirmation-modal";
import { usePlan } from "@/components/providers/plan-provider";
import { forecastService } from "@/services/forecast";

export default function SolarProfilesPage() {
  const queryClient = useQueryClient();
  const { checkAccess, userTier } = usePlan();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState<SolarProfile | null>(null);

  // Manual Data Entry State
  const [isManualModalOpen, setIsManualModalOpen] = useState(false);
  const [manualData, setManualData] = useState({
    profileId: "",
    date: new Date().toISOString().split("T")[0],
    actualKwh: 0,
  });
  
  // Confirmation Modal State
  const [confirmConfig, setConfirmConfig] = useState<{
    isOpen: boolean;
    type: "create" | "update" | "delete" | "manual" | null;
    id?: string;
  }>({ isOpen: false, type: null });

  // Mutations
  const recordActualMutation = useMutation({
    mutationFn: (data: any) => forecastService.recordActual(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["forecast-summary"] });
      queryClient.invalidateQueries({ queryKey: ["actual-history"] });
      setIsManualModalOpen(false);
      alert("Actual data recorded successfully!");
    }
  });

  // Form State
  const [formData, setFormData] = useState({
    site_name: "",
    capacity_kwp: 0,
    lat: 0,
    lng: 0,
    tilt: 30,
    azimuth: 180
  });

  // Queries
  const { data: profiles, isLoading } = useQuery({
    queryKey: ["solar-profiles"],
    queryFn: () => solarService.getProfiles(),
  });

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: any) => solarService.createProfile(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["solar-profiles"] });
      closeModal();
    }
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => solarService.deleteProfile(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["solar-profiles"] });
    }
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string, data: any }) => solarService.updateProfile(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["solar-profiles"] });
      closeModal();
    }
  });

  const openModal = (profile?: SolarProfile) => {
    if (!profile) {
      // Logic for site limits
      const currentCount = profiles?.length || 0;
      if (userTier === "free" && currentCount >= 1) {
        checkAccess(
          "pro", 
          "Add More Sites", 
          "The Free plan is limited to 1 solar site. Upgrade to Pro to manage up to 10 sites with advanced analytics."
        );
        return;
      }
      if (userTier === "pro" && currentCount >= 5) {
        checkAccess(
          "enterprise", 
          "Enterprise Scaling", 
          "You've reached the 5-site limit on Pro. Upgrade to Enterprise for unlimited solar installation sites and multi-user management."
        );
        return;
      }

      setEditingProfile(null);
      setFormData({
        site_name: "",
        capacity_kwp: 0,
        lat: 0,
        lng: 0,
        tilt: 30,
        azimuth: 180
      });
    } else {
      setEditingProfile(profile);
      setFormData({
        site_name: profile.site_name,
        capacity_kwp: profile.capacity_kwp,
        lat: profile.lat,
        lng: profile.lng,
        tilt: profile.tilt,
        azimuth: profile.azimuth
      });
    }
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingProfile(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setConfirmConfig({ 
      isOpen: true, 
      type: editingProfile ? "update" : "create" 
    });
  };

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setConfirmConfig({ isOpen: true, type: "manual" });
  };

  const executeAction = () => {
    if (confirmConfig.type === "delete" && confirmConfig.id) {
      deleteMutation.mutate(confirmConfig.id);
    } else if (confirmConfig.type === "manual") {
      recordActualMutation.mutate({
        solar_profile_id: manualData.profileId,
        date: manualData.date,
        actual_kwh: Number(manualData.actualKwh),
        source: "manual"
      });
    } else {
      const payload = {
        ...formData,
        capacity_kwp: Number(formData.capacity_kwp),
        lat: Number(formData.lat),
        lng: Number(formData.lng),
        tilt: Number(formData.tilt),
        azimuth: Number(formData.azimuth),
      };

      if (editingProfile) {
        updateMutation.mutate({ id: editingProfile.id, data: payload });
      } else {
        createMutation.mutate(payload);
      }
    }
    setConfirmConfig({ isOpen: false, type: null });
  };

  if (isLoading) {
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
          <h1 className="text-3xl font-bold tracking-tight mb-2">Solar Profiles</h1>
          <p className="text-muted-foreground">Manage your solar installation sites and capacities.</p>
        </div>
        <button 
          onClick={() => openModal()}
          className="flex items-center justify-center gap-2 px-6 py-3 bg-primary text-white font-bold rounded-2xl shadow-premium hover:translate-y-[-2px] transition-all"
        >
          <Plus className="w-5 h-5" />
          Add New Site
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {profiles?.map((profile) => (
          <Card key={profile.id} className="group relative hover:border-primary transition-all duration-300">
            <div className="absolute top-4 right-4 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button 
                onClick={() => openModal(profile)}
                className="p-2 rounded-lg bg-accent/80 text-muted-foreground hover:text-primary transition-all"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              <button 
                onClick={() => setConfirmConfig({ isOpen: true, type: "delete", id: profile.id })}
                className="p-2 rounded-lg bg-red-500/10 text-red-500 hover:bg-red-50 hover:text-red-600 transition-all"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>

            <div className="flex items-center gap-4 mb-6">
              <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary">
                <MapPin className="w-6 h-6" />
              </div>
              <div>
                <h3 className="font-bold text-lg">{profile.site_name}</h3>
                <p className="text-xs text-muted-foreground font-semibold uppercase tracking-wider">
                  Active Site
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 rounded-2xl bg-accent/30 border border-border/50 text-center">
                <p className="text-[10px] font-bold text-muted-foreground uppercase mb-1">Capacity</p>
                <p className="text-xl font-bold">{profile.capacity_kwp} <span className="text-sm font-normal text-muted-foreground">kWp</span></p>
              </div>
              <div className="p-4 rounded-2xl bg-accent/30 border border-border/50 text-center">
                <p className="text-[10px] font-bold text-muted-foreground uppercase mb-1">Tilt / Azimuth</p>
                <p className="text-xl font-bold">{profile.tilt}° <span className="text-sm text-muted-foreground">/</span> {profile.azimuth}°</p>
              </div>
            </div>

            <div className="mt-6 flex items-center gap-2 text-xs text-muted-foreground font-medium bg-background/50 p-3 rounded-xl border border-border/50">
              <AlertTriangle className="w-3.5 h-3.5 text-amber-500" />
              Lat: {profile.lat}, Lng: {profile.lng}
            </div>

            <button 
              onClick={() => {
                setManualData({ ...manualData, profileId: profile.id });
                setIsManualModalOpen(true);
              }}
              className="mt-6 w-full py-3 bg-accent/50 hover:bg-primary hover:text-white transition-all rounded-xl font-bold text-xs flex items-center justify-center gap-2"
            >
              <Activity className="w-4 h-4" />
              Record Actual Data
            </button>
          </Card>
        ))}

        {(!profiles || profiles.length === 0) && (
          <div className="col-span-full py-20 flex flex-col items-center justify-center text-center bg-card/50 rounded-[2.5rem] border border-dashed border-border">
            <MapPin className="w-16 h-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-xl font-bold text-muted-foreground">No sites found</h3>
            <p className="text-muted-foreground mt-2 max-w-xs">Start by adding your first solar installation site to begin monitoring.</p>
          </div>
        )}
      </div>

      {/* Manual Data Entry Modal */}
      {isManualModalOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
          <motion.div 
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            className="w-full max-w-md bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden"
          >
            <div className="p-8 border-b border-border flex items-center justify-between">
              <h2 className="text-2xl font-bold text-foreground">Record Production</h2>
              <button onClick={() => setIsManualModalOpen(false)} className="p-2 rounded-xl hover:bg-accent transition-all text-muted-foreground">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleManualSubmit} className="p-8 space-y-6">
              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Production Date</label>
                <input 
                  type="date" 
                  value={manualData.date}
                  max={new Date().toISOString().split("T")[0]}
                  onChange={(e) => setManualData({...manualData, date: e.target.value})}
                  className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold"
                  required
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Actual Production (kWh)</label>
                <div className="relative">
                  <input 
                    type="number" 
                    step="0.01"
                    value={manualData.actualKwh}
                    onChange={(e) => setManualData({...manualData, actualKwh: Number(e.target.value)})}
                    placeholder="0.00"
                    className="w-full pl-4 pr-16 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold text-lg"
                    required
                  />
                  <div className="absolute right-4 top-1/2 -translate-y-1/2 text-xs font-black text-muted-foreground uppercase tracking-widest bg-accent px-2 py-1 rounded-lg">
                    kWh
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-2xl bg-primary/5 border border-primary/10 flex gap-3">
                <Activity className="w-5 h-5 text-primary shrink-0 mt-0.5" />
                <p className="text-[10px] text-primary font-medium leading-relaxed uppercase tracking-wider">
                  Manual entries will appear as "Manual" in reports and will overwrite any existing IoT telemetry for this date.
                </p>
              </div>

              <div className="flex gap-4 pt-2">
                <button 
                  type="button" 
                  onClick={() => setIsManualModalOpen(false)}
                  className="flex-1 py-4 font-bold text-muted-foreground bg-accent hover:bg-border transition-all rounded-2xl"
                >
                  Cancel
                </button>
                <button 
                  type="submit"
                  disabled={recordActualMutation.isPending}
                  className="flex-1 py-4 font-bold text-white bg-primary shadow-premium hover:opacity-90 transition-all rounded-2xl flex items-center justify-center gap-2"
                >
                  {recordActualMutation.isPending ? <Loader2 className="w-5 h-5 animate-spin" /> : "Save Record"}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}

      {/* Modal Form */}
      {isModalOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
          <motion.div 
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            className="w-full max-w-lg bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden"
          >
            <div className="p-8 border-b border-border flex items-center justify-between">
              <h2 className="text-2xl font-bold">{editingProfile ? "Edit Site" : "Add New Site"}</h2>
              <button onClick={closeModal} className="p-2 rounded-xl hover:bg-accent transition-all text-muted-foreground">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-8 space-y-6">
              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Site Name</label>
                <input 
                  type="text" 
                  value={formData.site_name}
                  onChange={(e) => setFormData({...formData, site_name: e.target.value})}
                  placeholder="e.g. Roof Top Office"
                  className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Latitude</label>
                  <div className="relative group/field">
                    <input 
                      type="number" 
                      step="0.000001"
                      value={formData.lat}
                      onChange={(e) => setFormData({...formData, lat: parseFloat(e.target.value)})}
                      className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none pr-12"
                      required
                    />
                    <button 
                      type="button"
                      onClick={() => {
                        if (navigator.geolocation) {
                          navigator.geolocation.getCurrentPosition((pos) => {
                            setFormData(prev => ({ ...prev, lat: pos.coords.latitude, lng: pos.coords.longitude }));
                          });
                        }
                      }}
                      title="Get current location"
                      className="absolute right-2 top-1/2 -translate-y-1/2 p-2 rounded-lg bg-primary/10 text-primary hover:bg-primary hover:text-white transition-all shadow-sm"
                    >
                      <MapPin className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Longitude</label>
                  <input 
                    type="number" 
                    step="0.000001"
                    value={formData.lng}
                    onChange={(e) => setFormData({...formData, lng: parseFloat(e.target.value)})}
                    className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-bold ml-1">Capacity (kWp)</label>
                <input 
                  type="number" 
                  step="0.1"
                  value={formData.capacity_kwp}
                  onChange={(e) => setFormData({...formData, capacity_kwp: parseFloat(e.target.value)})}
                  className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Panel Tilt (°)</label>
                  <input 
                    type="number" 
                    value={formData.tilt}
                    onChange={(e) => setFormData({...formData, tilt: parseInt(e.target.value)})}
                    className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Azimuth (°)</label>
                  <input 
                    type="number" 
                    value={formData.azimuth}
                    onChange={(e) => setFormData({...formData, azimuth: parseInt(e.target.value)})}
                    className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none"
                  />
                </div>
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
                  disabled={createMutation.isPending || updateMutation.isPending}
                  className="flex-1 py-4 font-bold text-white bg-primary shadow-premium hover:opacity-90 transition-all rounded-2xl flex items-center justify-center"
                >
                  {(createMutation.isPending || updateMutation.isPending) ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    editingProfile ? "Save Changes" : "Create Site"
                  )}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}

      <ConfirmationModal 
        isOpen={confirmConfig.isOpen}
        onClose={() => setConfirmConfig({ isOpen: false, type: null })}
        onConfirm={executeAction}
        title={
          confirmConfig.type === "delete" ? "Delete Solar Profile?" :
          confirmConfig.type === "create" ? "Create New Profile?" : "Save Changes?"
        }
        description={
          confirmConfig.type === "delete" ? "This action cannot be undone. All associated device data will remain but the profile link will be severed." :
          confirmConfig.type === "create" ? "Are you sure you want to add this new solar installation site?" : "Confirm saving the updated details for this solar site."
        }
        confirmText={
          confirmConfig.type === "delete" ? "Delete Profile" :
          confirmConfig.type === "create" ? "Create Site" : "Save Changes"
        }
        variant={confirmConfig.type === "delete" ? "danger" : "primary"}
        isLoading={createMutation.isPending || updateMutation.isPending || deleteMutation.isPending}
      />
    </div>
  );
}
