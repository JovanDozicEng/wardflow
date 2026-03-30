/**
 * EncounterAutocomplete
 * Reusable encounter search/select component.
 * Fetches encounters on focus, filters client-side by typing.
 */

import { useState, useRef, useEffect, useCallback } from 'react';
import api from '../../utils/api';

interface Encounter {
  id: string;
  patientId: string;
  unitId: string;
  departmentId: string;
  status: string;
  startedAt: string;
}

interface Props {
  value: string;
  onChange: (id: string) => void;
  disabled?: boolean;
  required?: boolean;
  label?: string;
  placeholder?: string;
  error?: string;
}

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-100 text-green-700',
  discharged: 'bg-gray-100 text-gray-600',
  cancelled: 'bg-red-100 text-red-600',
};

export const EncounterAutocomplete = ({
  value,
  onChange,
  disabled = false,
  required = false,
  label,
  placeholder = 'Search encounters…',
  error,
}: Props) => {
  const [query, setQuery] = useState('');
  const [allEncounters, setAllEncounters] = useState<Encounter[]>([]);
  const [filtered, setFiltered] = useState<Encounter[]>([]);
  const [selected, setSelected] = useState<Encounter | null>(null);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fetched, setFetched] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const fetchEncounters = useCallback(async () => {
    if (fetched) return;
    setLoading(true);
    try {
      const res = await api.get<{ data: Encounter[] }>('/encounters', {
        params: { limit: 100, status: 'active' },
      });
      const data = Array.isArray(res.data?.data) ? res.data.data : [];
      setAllEncounters(data);
      setFiltered(data.slice(0, 20));
      setFetched(true);

      // If a value is already set, find the matching encounter
      if (value && !selected) {
        const match = data.find((e) => e.id === value);
        if (match) setSelected(match);
      }
    } catch {
      // silent — user can still type manually
    } finally {
      setLoading(false);
    }
  }, [fetched, value, selected]);

  // Filter encounters whenever query changes
  useEffect(() => {
    if (!query.trim()) {
      setFiltered(allEncounters.slice(0, 20));
      return;
    }
    const q = query.toLowerCase();
    setFiltered(
      allEncounters
        .filter(
          (e) =>
            e.id.toLowerCase().includes(q) ||
            e.unitId.toLowerCase().includes(q) ||
            e.patientId.toLowerCase().includes(q)
        )
        .slice(0, 20)
    );
  }, [query, allEncounters]);

  const handleFocus = () => {
    fetchEncounters();
    setOpen(true);
  };

  const handleSelect = (enc: Encounter) => {
    setSelected(enc);
    onChange(enc.id);
    setQuery('');
    setOpen(false);
  };

  const handleClear = () => {
    setSelected(null);
    onChange('');
    setQuery('');
    setFetched(false); // refetch next time
    setAllEncounters([]);
  };

  const shortId = (id: string) => id.slice(0, 8) + '…';

  return (
    <div ref={containerRef} className="relative">
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">
          {label} {required && <span className="text-red-500">*</span>}
        </label>
      )}

      {selected ? (
        // Selected chip
        <div
          className={`flex items-center justify-between px-3 py-2 border rounded-lg text-sm ${
            error ? 'border-red-500' : 'border-blue-400 bg-blue-50'
          }`}
        >
          <div className="flex items-center gap-2 min-w-0">
            <span className="font-mono text-xs text-gray-500 shrink-0">{shortId(selected.id)}</span>
            <span className="text-gray-700 truncate">Unit: {selected.unitId}</span>
            <span
              className={`inline-flex px-1.5 py-0.5 rounded text-xs font-medium capitalize shrink-0 ${
                STATUS_COLORS[selected.status] ?? 'bg-gray-100 text-gray-600'
              }`}
            >
              {selected.status}
            </span>
          </div>
          {!disabled && (
            <button
              type="button"
              onClick={handleClear}
              className="ml-2 text-gray-400 hover:text-gray-600 shrink-0 text-xs underline"
            >
              Change
            </button>
          )}
        </div>
      ) : (
        // Search input
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={handleFocus}
          disabled={disabled}
          placeholder={placeholder}
          autoComplete="off"
          className={`w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed ${
            error ? 'border-red-500' : 'border-gray-300'
          }`}
        />
      )}

      {error && <p className="mt-1 text-sm text-red-600">{error}</p>}

      {/* Dropdown */}
      {open && !selected && (
        <div className="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-64 overflow-y-auto">
          {loading && (
            <div className="px-3 py-2 text-sm text-gray-500">Loading encounters…</div>
          )}

          {!loading && filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">No encounters found</div>
          )}

          {!loading &&
            filtered.map((enc) => (
              <button
                key={enc.id}
                type="button"
                onMouseDown={() => handleSelect(enc)}
                className="w-full text-left px-3 py-2 hover:bg-blue-50 flex items-center gap-3 border-b border-gray-50 last:border-0"
              >
                <span className="font-mono text-xs text-gray-400 shrink-0 w-20 truncate">
                  {enc.id.slice(0, 8)}…
                </span>
                <span className="text-sm text-gray-700 truncate flex-1">
                  Unit: {enc.unitId}
                </span>
                <span
                  className={`inline-flex px-1.5 py-0.5 rounded text-xs font-medium capitalize shrink-0 ${
                    STATUS_COLORS[enc.status] ?? 'bg-gray-100 text-gray-600'
                  }`}
                >
                  {enc.status}
                </span>
              </button>
            ))}
        </div>
      )}
    </div>
  );
};
