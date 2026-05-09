"use client";

import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { adminService } from "@/services/admin";
import { userService } from "@/services/user";
import { Card } from "@/components/ui/card";
import { 
  Users, 
  Shield, 
  Activity, 
  Server, 
  CloudRain, 
  Loader2,
  MoreVertical,
  CheckCircle2,
  XCircle,
  BarChart3,
  Pencil,
  Trash2,
  X
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { useToast } from "@/components/providers/toast-provider";

export default function AdminPage() {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  
  // Modals state
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<any>(null);
  
  // Edit Form state
  const [editForm, setEditForm] = useState({ name: "", email: "" });

  // Queries
  const { data: currentUser } = useQuery({ queryKey: ["me"], queryFn: () => userService.getMe() });
  const { data: users, isLoading: isUsersLoading } = useQuery({ queryKey: ["admin-users"], queryFn: () => adminService.getAllUsers() });
  const { data: stats } = useQuery({ queryKey: ["admin-stats"], queryFn: () => adminService.getSystemStats() });
  const { data: weatherHealth } = useQuery({ queryKey: ["weather-health"], queryFn: () => adminService.getWeatherHealth() });

  // Mutations
  const updateTierMutation = useMutation({
    mutationFn: ({ userId, tier }: { userId: string, tier: string }) => adminService.updateUserTier(userId, tier),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      showToast(`User plan updated to ${variables.tier.toUpperCase()} successfully!`, "success");
    },
    onError: (err: any) => {
      showToast(err.response?.data?.error || "Failed to update user plan.", "error");
    }
  });

  const updateUserMutation = useMutation({
    mutationFn: ({ userId, data }: { userId: string, data: any }) => adminService.updateUser(userId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      showToast("User details updated successfully!", "success");
      setIsEditModalOpen(false);
    },
    onError: (err: any) => {
      showToast(err.response?.data?.error || "Failed to update user details.", "error");
    }
  });

  const deleteUserMutation = useMutation({
    mutationFn: (userId: string) => adminService.deleteUser(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      showToast("User account deleted successfully!", "success");
      setIsDeleteModalOpen(false);
    },
    onError: (err: any) => {
      showToast(err.response?.data?.error || "Failed to delete user.", "error");
    }
  });

  const openEditModal = (user: any) => {
    setSelectedUser(user);
    setEditForm({ name: user.name, email: user.email });
    setIsEditModalOpen(true);
  };

  const openDeleteModal = (user: any) => {
    setSelectedUser(user);
    setIsDeleteModalOpen(true);
  };

  const handleUpdateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUser) return;
    updateUserMutation.mutate({ userId: selectedUser.id, data: editForm });
  };

  if (currentUser?.role !== "admin") {
    return (
      <div className="h-[60vh] flex flex-col items-center justify-center text-center">
        <Shield className="w-16 h-16 text-red-500 mb-4" />
        <h2 className="text-2xl font-bold text-foreground">Access Denied</h2>
        <p className="text-muted-foreground mt-2">You do not have the required permissions to view this page.</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 pb-12">
      <div>
        <h1 className="text-3xl font-bold tracking-tight mb-2">Admin Control Center</h1>
        <p className="text-muted-foreground">Monitor system health and manage user subscriptions.</p>
      </div>

      {/* System Health Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card className="bg-primary/5 border-primary/20">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary">
              <Server className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs font-bold text-muted-foreground uppercase">API Status</p>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                <p className="text-lg font-bold">Operational</p>
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-amber-500/10 flex items-center justify-center text-amber-500">
              <CloudRain className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs font-bold text-muted-foreground uppercase">Weather API</p>
              <p className="text-lg font-bold">{weatherHealth?.status === "healthy" ? "Healthy" : "Check Logs"}</p>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-500">
              <Activity className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs font-bold text-muted-foreground uppercase">Total Forecasts</p>
              <p className="text-lg font-bold">{stats?.total_forecasts_count?.toLocaleString() || 0}</p>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-2xl bg-blue-500/10 flex items-center justify-center text-blue-500">
              <Users className="w-6 h-6" />
            </div>
            <div>
              <p className="text-xs font-bold text-muted-foreground uppercase">Total Users</p>
              <p className="text-lg font-bold">{users?.length || 0}</p>
            </div>
          </div>
        </Card>
      </div>

      {/* User Management Table */}
      <Card className="p-0 overflow-hidden">
        <div className="p-6 border-b border-border flex items-center justify-between bg-accent/20">
          <h2 className="text-xl font-bold flex items-center gap-2">
            <Users className="w-5 h-5" />
            User Management
          </h2>
          <div className="flex gap-2">
            <button className="px-4 py-2 bg-card border border-border rounded-xl text-xs font-bold hover:bg-accent transition-all">Export Users</button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-accent/10 border-b border-border">
                <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">User / Company</th>
                <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Current Plan</th>
                <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Role</th>
                <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground">Joined Date</th>
                <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-muted-foreground text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {users?.map((u: any) => (
                <tr key={u.id} className="hover:bg-accent/5 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-accent flex items-center justify-center font-bold text-primary">
                        {u.name?.charAt(0)}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-bold text-sm">{u.name}</p>
                          {u.email_verified ? (
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                          ) : (
                            <XCircle className="w-3.5 h-3.5 text-amber-500" />
                          )}
                        </div>
                        <p className="text-xs text-muted-foreground">{u.email}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <select 
                      value={u.plan_tier}
                      onChange={(e) => updateTierMutation.mutate({ userId: u.id, tier: e.target.value })}
                      className={cn(
                        "text-xs font-bold px-3 py-1.5 rounded-lg border focus:outline-none transition-all cursor-pointer",
                        u.plan_tier === "enterprise" ? "bg-purple-500/10 text-purple-600 border-purple-200" :
                        u.plan_tier === "pro" ? "bg-blue-500/10 text-blue-600 border-blue-200" :
                        "bg-slate-500/10 text-slate-600 border-slate-200"
                      )}
                    >
                      <option value="free">FREE</option>
                      <option value="pro">PRO</option>
                      <option value="enterprise">ENTERPRISE</option>
                    </select>
                  </td>
                  <td className="px-6 py-4">
                    <span className={cn(
                      "text-[10px] font-bold px-2 py-0.5 rounded-full border",
                      u.role === "admin" ? "bg-red-500/5 text-red-500 border-red-200" : "bg-accent text-muted-foreground"
                    )}>
                      {u.role?.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm font-medium text-muted-foreground">
                    {new Date(u.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <button 
                        onClick={() => openEditModal(u)}
                        className="p-2 rounded-xl hover:bg-primary/10 hover:text-primary text-muted-foreground transition-all group"
                        title="Edit User"
                      >
                        <Pencil className="w-4 h-4 group-hover:scale-110 transition-transform" />
                      </button>
                      <button 
                        onClick={() => openDeleteModal(u)}
                        className="p-2 rounded-xl hover:bg-red-500/10 hover:text-red-500 text-muted-foreground transition-all group"
                        title="Delete User"
                      >
                        <Trash2 className="w-4 h-4 group-hover:scale-110 transition-transform" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Edit Modal */}
      <AnimatePresence>
        {isEditModalOpen && (
          <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
            <motion.div 
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              className="w-full max-w-md bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden"
            >
              <div className="p-8 border-b border-border flex items-center justify-between">
                <h2 className="text-2xl font-bold">Edit User Details</h2>
                <button onClick={() => setIsEditModalOpen(false)} className="p-2 rounded-xl hover:bg-accent transition-all text-muted-foreground">
                  <X className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleUpdateSubmit} className="p-8 space-y-6">
                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Full Name</label>
                  <input 
                    type="text" 
                    value={editForm.name}
                    onChange={(e) => setEditForm({...editForm, name: e.target.value})}
                    className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold"
                    required
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-bold ml-1">Email Address</label>
                  <input 
                    type="email" 
                    value={editForm.email}
                    onChange={(e) => setEditForm({...editForm, email: e.target.value})}
                    className="w-full px-4 py-3 bg-background/50 border border-border rounded-xl focus:ring-2 focus:ring-primary/20 outline-none font-bold"
                    required
                  />
                </div>

                <div className="pt-4 flex gap-3">
                  <button 
                    type="button"
                    onClick={() => setIsEditModalOpen(false)}
                    className="flex-1 px-6 py-3 border border-border font-bold rounded-2xl hover:bg-accent transition-all"
                  >
                    Cancel
                  </button>
                  <button 
                    type="submit"
                    disabled={updateUserMutation.isPending}
                    className="flex-1 px-6 py-3 bg-primary text-white font-bold rounded-2xl shadow-premium hover:opacity-90 transition-all flex items-center justify-center gap-2"
                  >
                    {updateUserMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                    Save Changes
                  </button>
                </div>
              </form>
            </motion.div>
          </div>
        )}
      </AnimatePresence>

      {/* Delete Confirmation Modal */}
      <AnimatePresence>
        {isDeleteModalOpen && (
          <div className="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-background/80 backdrop-blur-sm">
            <motion.div 
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              className="w-full max-w-sm bg-card border border-border rounded-[2.5rem] shadow-2xl overflow-hidden p-8 text-center"
            >
              <div className="w-20 h-20 bg-red-500/10 text-red-500 rounded-full flex items-center justify-center mx-auto mb-6">
                <Trash2 className="w-10 h-10" />
              </div>
              <h2 className="text-2xl font-bold mb-2">Delete User?</h2>
              <p className="text-muted-foreground mb-8">
                Are you sure you want to delete <span className="font-bold text-foreground">{selectedUser?.name}</span>? This action cannot be undone.
              </p>

              <div className="flex gap-3">
                <button 
                  onClick={() => setIsDeleteModalOpen(false)}
                  className="flex-1 px-6 py-3 border border-border font-bold rounded-2xl hover:bg-accent transition-all"
                >
                  Cancel
                </button>
                <button 
                  onClick={() => deleteUserMutation.mutate(selectedUser?.id)}
                  disabled={deleteUserMutation.isPending}
                  className="flex-1 px-6 py-3 bg-red-500 text-white font-bold rounded-2xl shadow-xl shadow-red-500/20 hover:bg-red-600 transition-all flex items-center justify-center gap-2"
                >
                  {deleteUserMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                  Delete
                </button>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
