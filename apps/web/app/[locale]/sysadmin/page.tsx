"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useSysadminStore } from "@/features/sysadmin/stores/use-sysadmin-store";
import { waitingListService } from "@/features/waiting-list/services/waiting-list-service";
import type { WaitingListEntry } from "@/features/waiting-list/types";
import { toast } from "sonner";
import {
  LogOut,
  Users,
  Shield,
  Trash2,
  RefreshCw,
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  TrendingUp,
  Inbox,
  UserCheck,
  MapPin,
  Mail,
  Building,
  Phone,
  CheckCircle,
  FileText,
  AlertCircle,
  Loader2
} from "lucide-react";
import { format } from "date-fns";

export default function SysadminDashboard() {
  const { admin, logout } = useSysadminStore();
  const [entries, setEntries] = useState<WaitingListEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);

  // Statistics state
  const [stats, setStats] = useState({
    total: 0,
    suppliers: 0,
    buyers: 0,
    pending: 0,
  });

  const fetchEntries = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await waitingListService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setEntries(data.items);
      setTotal(data.total);

      // Compute statistics based on list parameters or calculate local stats for the UI
      // In production, we'd fetch this from a dashboard stats API, but here we can calculate from total & status filter or list
      // Let's call list with no filter to compute correct global counts
      const allData = await waitingListService.list({ page: 1, limit: 1000 });
      const supplierCount = allData.items.filter(i => i.company_type === "supplier").length;
      const buyerCount = allData.items.filter(i => i.company_type === "buyer").length;
      const pendingCount = allData.items.filter(i => i.status === "pending").length;

      setStats({
        total: allData.total,
        suppliers: supplierCount,
        buyers: buyerCount,
        pending: pendingCount,
      });
    } catch (error) {
      toast.error("Failed to load waiting list entries.");
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter]);

  useEffect(() => {
    fetchEntries();
  }, [fetchEntries]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await waitingListService.updateStatus(id, newStatus);
      toast.success(`Entry marked as ${newStatus}`);
      fetchEntries();
    } catch (error) {
      toast.error("Failed to update status.");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this waitlist entry?")) return;
    try {
      await waitingListService.delete(id);
      toast.success("Entry deleted successfully.");
      fetchEntries();
    } catch (error) {
      toast.error("Failed to delete entry.");
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-900/50">Approved</span>;
      case "contacted":
        return <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-blue-100 text-blue-800 dark:bg-blue-950/40 dark:text-blue-400 border border-blue-200 dark:border-blue-900/50">Contacted</span>;
      case "rejected":
        return <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-100 text-rose-800 dark:bg-rose-950/40 dark:text-rose-400 border border-rose-200 dark:border-rose-900/50">Rejected</span>;
      default:
        return <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200 dark:border-amber-900/50">Pending</span>;
    }
  };

  const getCompanyTypeBadge = (type: string) => {
    switch (type) {
      case "supplier":
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-violet-100 text-violet-800 dark:bg-violet-950/30 dark:text-violet-400">Supplier</span>;
      case "buyer":
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-sky-100 text-sky-800 dark:bg-sky-950/30 dark:text-sky-400 font-medium">Buyer</span>;
      default:
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-neutral-100 text-neutral-800 dark:bg-neutral-800 dark:text-neutral-300">Other</span>;
    }
  };

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-50">
      {/* Top Navbar */}
      <header className="sticky top-0 z-40 border-b border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600 text-white font-bold text-base shadow-sm">
              IS
            </span>
            <div>
              <span className="font-bold text-base tracking-tight block">IndoSupplier</span>
              <span className="text-[10px] text-neutral-500 font-semibold tracking-wider uppercase block">System Admin Panel</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
              <span className="text-sm font-semibold block">{admin?.name}</span>
              <span className="text-xs text-neutral-500 block capitalize">{admin?.role.replace("_", " ")}</span>
            </div>
            <button
              onClick={logout}
              className="inline-flex items-center gap-1.5 rounded-lg border border-neutral-200 dark:border-neutral-850 px-3.5 py-2 text-xs font-semibold text-red-600 dark:text-red-400 hover:bg-red-550/10 dark:hover:bg-red-950/20 hover:border-red-300 transition-all cursor-pointer"
            >
              <LogOut className="h-3.5 w-3.5" />
              <span>Sign Out</span>
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-10 space-y-8">
        {/* Welcome */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-extrabold tracking-tight">Waiting List Dashboard</h1>
            <p className="text-neutral-500 text-sm mt-1">
              Review and manage early access registrations for IndoSupplier SAAS.
            </p>
          </div>
          <button
            onClick={fetchEntries}
            className="self-start sm:self-center inline-flex items-center gap-1.5 rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 px-3.5 py-2.5 text-xs font-semibold shadow-sm hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            Refresh
          </button>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Card 1: Total */}
          <div className="rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 p-6 shadow-sm flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-neutral-500 tracking-wider uppercase block">Total Registrations</span>
              <span className="text-3xl font-extrabold block">{stats.total}</span>
            </div>
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400">
              <Users className="h-6 w-6" />
            </span>
          </div>

          {/* Card 2: Suppliers */}
          <div className="rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 p-6 shadow-sm flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-neutral-500 tracking-wider uppercase block">Suppliers</span>
              <span className="text-3xl font-extrabold block">{stats.suppliers}</span>
            </div>
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-violet-50 dark:bg-violet-950/30 text-violet-600 dark:text-violet-400">
              <Building className="h-6 w-6" />
            </span>
          </div>

          {/* Card 3: Buyers */}
          <div className="rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 p-6 shadow-sm flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-neutral-500 tracking-wider uppercase block">Buyers</span>
              <span className="text-3xl font-extrabold block">{stats.buyers}</span>
            </div>
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-sky-50 dark:bg-sky-950/30 text-sky-600 dark:text-sky-400">
              <UserCheck className="h-6 w-6" />
            </span>
          </div>

          {/* Card 4: Pending */}
          <div className="rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 p-6 shadow-sm flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-neutral-500 tracking-wider uppercase block">Pending Review</span>
              <span className="text-3xl font-extrabold block">{stats.pending}</span>
            </div>
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-amber-50 dark:bg-amber-950/30 text-amber-600 dark:text-amber-400">
              <Inbox className="h-6 w-6" />
            </span>
          </div>
        </div>

        {/* Filter and Content Card */}
        <div className="rounded-xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 shadow-sm overflow-hidden">
          {/* Filtering Header */}
          <div className="p-5 border-b border-neutral-200 dark:border-neutral-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-neutral-50/50 dark:bg-neutral-900/50">
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-neutral-500" />
              <span className="text-sm font-semibold">Filter Registrants</span>
            </div>

            <div className="flex items-center gap-3">
              <select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPage(1);
                }}
                className="rounded-lg border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-950 px-3 py-2 text-sm outline-none focus:border-blue-500"
              >
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="contacted">Contacted</option>
                <option value="approved">Approved</option>
                <option value="rejected">Rejected</option>
              </select>
            </div>
          </div>

          {/* Table Content */}
          <div className="overflow-x-auto">
            {isLoading ? (
              <div className="py-24 text-center">
                <Loader2 className="h-8 w-8 animate-spin text-blue-600 mx-auto mb-3" />
                <span className="text-sm font-semibold text-neutral-500">Fetching records...</span>
              </div>
            ) : entries.length === 0 ? (
              <div className="py-24 text-center space-y-2">
                <Inbox className="h-12 w-12 text-neutral-300 dark:text-neutral-700 mx-auto" />
                <h3 className="font-bold text-lg text-neutral-900 dark:text-white">No registrations found</h3>
                <p className="text-neutral-500 text-sm max-w-sm mx-auto">
                  Try changing your status filter or wait for new waitlist submissions.
                </p>
              </div>
            ) : (
              <table className="w-full text-left border-collapse text-sm">
                <thead>
                  <tr className="border-b border-neutral-200 dark:border-neutral-800 text-neutral-500 dark:text-neutral-400 font-semibold bg-neutral-50/30 dark:bg-neutral-900/30">
                    <th className="p-4 pl-6">Company / Name</th>
                    <th className="p-4">Contact</th>
                    <th className="p-4">Business Type</th>
                    <th className="p-4">Status</th>
                    <th className="p-4">Signup Date</th>
                    <th className="p-4 pr-6 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-200 dark:divide-neutral-800">
                  {entries.map((entry) => (
                    <tr key={entry.id} className="hover:bg-neutral-50/50 dark:hover:bg-neutral-900/30 transition-colors">
                      {/* Name / Company */}
                      <td className="p-4 pl-6">
                        <div>
                          <span className="font-bold text-neutral-900 dark:text-white block">{entry.company_name}</span>
                          <span className="text-xs text-neutral-500 block mt-0.5">{entry.name}</span>
                        </div>
                      </td>

                      {/* Contact */}
                      <td className="p-4">
                        <div className="space-y-1">
                          <span className="flex items-center gap-1.5 text-xs text-neutral-600 dark:text-neutral-400">
                            <Mail className="h-3 w-3 text-neutral-400" />
                            {entry.email}
                          </span>
                          {entry.phone && (
                            <span className="flex items-center gap-1.5 text-xs text-neutral-600 dark:text-neutral-400">
                              <Phone className="h-3 w-3 text-neutral-400" />
                              {entry.phone}
                            </span>
                          )}
                        </div>
                      </td>

                      {/* Business Type */}
                      <td className="p-4">{getCompanyTypeBadge(entry.company_type)}</td>

                      {/* Status */}
                      <td className="p-4">{getStatusBadge(entry.status)}</td>

                      {/* Date */}
                      <td className="p-4 text-xs text-neutral-500">
                        {format(new Date(entry.created_at), "MMM d, yyyy HH:mm")}
                      </td>

                      {/* Actions */}
                      <td className="p-4 pr-6 text-right">
                        <div className="inline-flex items-center gap-2">
                          <select
                            value={entry.status}
                            onChange={(e) => handleUpdateStatus(entry.id, e.target.value)}
                            className="rounded border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-950 px-2 py-1 text-xs outline-none focus:border-blue-500"
                          >
                            <option value="pending">Pending</option>
                            <option value="contacted">Contacted</option>
                            <option value="approved">Approved</option>
                            <option value="rejected">Rejected</option>
                          </select>

                          <button
                            onClick={() => handleDelete(entry.id)}
                            className="rounded p-1 text-neutral-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20 transition-all cursor-pointer"
                            title="Delete entry"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {/* Pagination Footer */}
          {entries.length > 0 && (
            <div className="p-4 border-t border-neutral-200 dark:border-neutral-800 flex items-center justify-between bg-neutral-50/30 dark:bg-neutral-900/30">
              <span className="text-xs text-neutral-500 font-medium">
                Showing page <span className="font-semibold text-neutral-900 dark:text-white">{page}</span> of{" "}
                <span className="font-semibold text-neutral-900 dark:text-white">{totalPages}</span> ({total} entries)
              </span>

              <div className="inline-flex items-center gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="rounded-lg border border-neutral-250 dark:border-neutral-700 px-3 py-1.5 text-xs font-semibold hover:bg-neutral-50 dark:hover:bg-neutral-800 disabled:opacity-50 disabled:pointer-events-none cursor-pointer"
                >
                  <ChevronLeft className="h-4 w-4" />
                </button>
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="rounded-lg border border-neutral-250 dark:border-neutral-700 px-3 py-1.5 text-xs font-semibold hover:bg-neutral-50 dark:hover:bg-neutral-800 disabled:opacity-50 disabled:pointer-events-none cursor-pointer"
                >
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
