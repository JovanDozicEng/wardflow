/**
 * StaffManagementPage - Admin-only staff assignment management
 * Allows admins to view all staff and manage their unit/department assignments and roles.
 */

import { useEffect, useState, useCallback, useRef } from 'react';
import { Navigate } from 'react-router-dom';
import { Users, Search, Pencil, X, Check, ChevronLeft, ChevronRight } from 'lucide-react';
import toast from 'react-hot-toast';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { Spinner } from '../shared/components/ui/Spinner';
import { Button } from '../shared/components/ui/Button';
import { Modal } from '../shared/components/ui/Modal';
import { useAuthStore } from '../features/auth/store/authStore';
import { staffService, type StaffProfile } from '../features/admin/services/staffService';
import { listUnits, listDepartments } from '../features/units/services/unitService';
import type { Unit, Department } from '../features/units/types';
import { Role } from '../shared/types/common.types';

import { ROUTES } from '../shared/config/routes';

const ROLE_LABELS: Record<string, string> = {
  nurse: 'Nurse',
  provider: 'Provider',
  charge_nurse: 'Charge Nurse',
  operations: 'Operations',
  consult: 'Consult',
  transport: 'Transport',
  quality_safety: 'Quality & Safety',
  admin: 'Admin',
};

const ROLE_COLORS: Record<string, string> = {
  nurse: 'bg-blue-100 text-blue-800',
  provider: 'bg-green-100 text-green-800',
  charge_nurse: 'bg-purple-100 text-purple-800',
  operations: 'bg-orange-100 text-orange-800',
  consult: 'bg-teal-100 text-teal-800',
  transport: 'bg-yellow-100 text-yellow-800',
  quality_safety: 'bg-red-100 text-red-800',
  admin: 'bg-gray-800 text-white',
};

const PAGE_SIZE = 20;

// --- Multi-select pill component ---
interface MultiSelectProps {
  label: string;
  placeholder: string;
  selectedIds: string[];
  options: { id: string; label: string }[];
  onSearch: (q: string) => void;
  onChange: (ids: string[]) => void;
}

function MultiSelect({ label, placeholder, selectedIds, options, onSearch, onChange }: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState('');
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => onSearch(q), 300);
    return () => clearTimeout(timer);
  }, [q, onSearch]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const toggle = (id: string) => {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((x) => x !== id));
    } else {
      onChange([...selectedIds, id]);
    }
  };

  const selectedOptions = options.filter((o) => selectedIds.includes(o.id));
  const unselectedOptions = options.filter((o) => !selectedIds.includes(o.id));

  return (
    <div className="space-y-1" ref={ref}>
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <div
        className="min-h-[42px] w-full border border-gray-300 rounded-lg px-3 py-2 flex flex-wrap gap-1 cursor-text bg-white focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-blue-500"
        onClick={() => setOpen(true)}
      >
        {selectedIds.map((id) => {
          const opt = options.find((o) => o.id === id);
          return (
            <span key={id} className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-100 text-blue-800 rounded text-xs font-medium">
              {opt?.label ?? id}
              <button
                type="button"
                onClick={(e) => { e.stopPropagation(); toggle(id); }}
                className="hover:text-blue-600"
                aria-label={`Remove ${opt?.label ?? id}`}
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          );
        })}
        <input
          type="text"
          value={q}
          onChange={(e) => { setQ(e.target.value); setOpen(true); }}
          placeholder={selectedIds.length === 0 ? placeholder : ''}
          className="flex-1 min-w-[120px] outline-none text-sm bg-transparent"
        />
      </div>
      {open && (
        <div className="absolute z-50 mt-1 w-full max-w-sm bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto">
          {[...selectedOptions, ...unselectedOptions].map((opt) => (
            <button
              key={opt.id}
              type="button"
              onClick={() => toggle(opt.id)}
              className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 text-left"
            >
              {selectedIds.includes(opt.id) ? (
                <Check className="w-4 h-4 text-blue-600 flex-shrink-0" />
              ) : (
                <span className="w-4 h-4 flex-shrink-0" />
              )}
              {opt.label}
            </button>
          ))}
          {options.length === 0 && (
            <p className="px-3 py-2 text-sm text-gray-500">No results</p>
          )}
        </div>
      )}
    </div>
  );
}

// --- Edit Staff Modal ---
interface EditModalProps {
  staff: StaffProfile;
  units: Unit[];
  departments: Department[];
  onUnitSearch: (q: string) => void;
  onDeptSearch: (q: string) => void;
  onClose: () => void;
  onSaved: (updated: StaffProfile) => void;
}

function EditStaffModal({ staff, units, departments, onUnitSearch, onDeptSearch, onClose, onSaved }: EditModalProps) {
  const [role, setRole] = useState(staff.role);
  const [isActive, setIsActive] = useState(staff.isActive);
  const [unitIds, setUnitIds] = useState<string[]>(staff.unitIds ?? []);
  const [deptIds, setDeptIds] = useState<string[]>(staff.departmentIds ?? []);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      const updated = await staffService.updateStaff(staff.id, {
        role,
        isActive,
        unitIds,
        departmentIds: deptIds,
      });
      toast.success('Staff member updated');
      onSaved(updated);
    } catch {
      toast.error('Failed to update staff member');
    } finally {
      setSaving(false);
    }
  };

  const unitOptions = units.map((u) => ({ id: u.id, label: `${u.name} (${u.code})` }));
  const deptOptions = departments.map((d) => ({ id: d.id, label: `${d.name} (${d.code})` }));

  return (
    <Modal isOpen onClose={onClose} title={`Edit — ${staff.name}`} size="lg">
      <div className="space-y-5">
        <div className="flex items-start gap-4">
          <div className="flex-1">
            <p className="text-sm text-gray-500">{staff.email}</p>
          </div>
        </div>

        {/* Role */}
        <div className="space-y-1">
          <label className="block text-sm font-medium text-gray-700">Role</label>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as typeof role)}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          >
            {Object.values(Role).map((r) => (
              <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
            ))}
          </select>
        </div>

        {/* Active status */}
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
              aria-label="Active"
              className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
            />
            <span className="text-sm font-medium text-gray-700">Active</span>
          </label>
          <span className="text-xs text-gray-500">(inactive users cannot log in)</span>
        </div>

        {/* Unit assignment */}
        <div className="relative">
          <MultiSelect
            label="Assigned Units"
            placeholder="Search units…"
            selectedIds={unitIds}
            options={unitOptions}
            onSearch={onUnitSearch}
            onChange={setUnitIds}
          />
        </div>

        {/* Department assignment */}
        <div className="relative">
          <MultiSelect
            label="Assigned Departments"
            placeholder="Search departments…"
            selectedIds={deptIds}
            options={deptOptions}
            onSearch={onDeptSearch}
            onChange={setDeptIds}
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button variant="secondary" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? 'Saving…' : 'Save Changes'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}

// --- Main page ---
export default function StaffManagementPage() {
  const { user } = useAuthStore();
  if (user?.role !== 'admin') return <Navigate to={ROUTES.UNAUTHORIZED} replace />;

  const [staff, setStaff] = useState<StaffProfile[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [offset, setOffset] = useState(0);
  const [editing, setEditing] = useState<StaffProfile | null>(null);

  const [units, setUnits] = useState<Unit[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);

  const fetchStaff = useCallback(async () => {
    setLoading(true);
    try {
      const res = await staffService.listStaff({ q: search, role: roleFilter || undefined, limit: PAGE_SIZE, offset });
      setStaff(res.data);
      setTotal(res.total);
    } catch {
      toast.error('Failed to load staff');
    } finally {
      setLoading(false);
    }
  }, [search, roleFilter, offset]);

  useEffect(() => { fetchStaff(); }, [fetchStaff]);

  // Pre-load all units and departments for the edit modal
  const loadUnits = useCallback(async (q: string) => {
    const res = await listUnits(q);
    setUnits(res);
  }, []);

  const loadDepts = useCallback(async (q: string) => {
    const res = await listDepartments(q);
    setDepartments(res);
  }, []);

  useEffect(() => { loadUnits(''); loadDepts(''); }, [loadUnits, loadDepts]);

  const handleSaved = (updated: StaffProfile) => {
    setStaff((prev) => prev.map((s) => (s.id === updated.id ? updated : s)));
    setEditing(null);
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;

  return (
    <Layout>
      <PageHeader
        title="Staff Management"
        subtitle={`${total} staff member${total !== 1 ? 's' : ''}`}
      />

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
          <input
            type="text"
            placeholder="Search by name or email…"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setOffset(0); }}
            className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>
        <select
          value={roleFilter}
          onChange={(e) => { setRoleFilter(e.target.value); setOffset(0); }}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
        >
          <option value="">All Roles</option>
          {Object.values(Role).map((r) => (
            <option key={r} value={r}>{ROLE_LABELS[r] ?? r}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex justify-center py-16"><Spinner /></div>
      ) : staff.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <Users className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">No staff members found.</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 font-semibold text-gray-700">Name</th>
                <th className="text-left px-4 py-3 font-semibold text-gray-700">Role</th>
                <th className="text-left px-4 py-3 font-semibold text-gray-700 hidden md:table-cell">Units</th>
                <th className="text-left px-4 py-3 font-semibold text-gray-700 hidden lg:table-cell">Departments</th>
                <th className="text-center px-4 py-3 font-semibold text-gray-700">Status</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {staff.map((s) => (
                <tr key={s.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <p className="font-medium text-gray-900">{s.name}</p>
                    <p className="text-xs text-gray-500">{s.email}</p>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${ROLE_COLORS[s.role] ?? 'bg-gray-100 text-gray-700'}`}>
                      {ROLE_LABELS[s.role] ?? s.role}
                    </span>
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell">
                    <div className="flex flex-wrap gap-1">
                      {(s.unitIds ?? []).length === 0 ? (
                        <span className="text-xs text-gray-400">—</span>
                      ) : (
                        (s.unitIds ?? []).map((id) => {
                          const u = units.find((x) => x.id === id);
                          return (
                            <span key={id} className="inline-flex px-1.5 py-0.5 bg-blue-50 text-blue-700 rounded text-xs">
                              {u?.code ?? id.slice(0, 8)}
                            </span>
                          );
                        })
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 hidden lg:table-cell">
                    <div className="flex flex-wrap gap-1">
                      {(s.departmentIds ?? []).length === 0 ? (
                        <span className="text-xs text-gray-400">—</span>
                      ) : (
                        (s.departmentIds ?? []).map((id) => {
                          const d = departments.find((x) => x.id === id);
                          return (
                            <span key={id} className="inline-flex px-1.5 py-0.5 bg-purple-50 text-purple-700 rounded text-xs">
                              {d?.code ?? id.slice(0, 8)}
                            </span>
                          );
                        })
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${s.isActive ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                      {s.isActive ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => setEditing(s)}
                      className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                      aria-label={`Edit ${s.name}`}
                    >
                      <Pencil className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
              <p className="text-sm text-gray-600">
                Page {currentPage} of {totalPages} · {total} total
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                  disabled={offset === 0}
                  className="p-1.5 rounded border border-gray-300 disabled:opacity-40 hover:bg-gray-100"
                  aria-label="Previous page"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  disabled={offset + PAGE_SIZE >= total}
                  className="p-1.5 rounded border border-gray-300 disabled:opacity-40 hover:bg-gray-100"
                  aria-label="Next page"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Edit modal */}
      {editing && (
        <EditStaffModal
          staff={editing}
          units={units}
          departments={departments}
          onUnitSearch={loadUnits}
          onDeptSearch={loadDepts}
          onClose={() => setEditing(null)}
          onSaved={handleSaved}
        />
      )}
    </Layout>
  );
}
