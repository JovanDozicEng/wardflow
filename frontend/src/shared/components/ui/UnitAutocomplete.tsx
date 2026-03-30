/**
 * UnitAutocomplete
 * Reusable unit search/select component.
 * Fetches units on focus, filters client-side by typing.
 */

import { useState, useRef, useEffect, useCallback } from 'react';
import { listUnits } from '../../../features/units/services/unitService';
import type { Unit } from '../../../features/units/types';

interface Props {
  value: string; // selected unit ID
  onChange: (id: string) => void;
  disabled?: boolean;
  required?: boolean;
  label?: string;
  placeholder?: string;
  error?: string;
  departmentId?: string; // optional filter: only show units in this department
}

export const UnitAutocomplete = ({
  value,
  onChange,
  disabled = false,
  required = false,
  label,
  placeholder = 'Search units…',
  error,
  departmentId,
}: Props) => {
  const [query, setQuery] = useState('');
  const [allUnits, setAllUnits] = useState<Unit[]>([]);
  const [filtered, setFiltered] = useState<Unit[]>([]);
  const [selected, setSelected] = useState<Unit | null>(null);
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

  const fetchUnits = useCallback(async () => {
    if (fetched) return;
    setLoading(true);
    try {
      const data = await listUnits(undefined, departmentId);
      setAllUnits(data);
      setFiltered(data.slice(0, 20));
      setFetched(true);

      // If a value is already set, find the matching unit
      if (value && !selected) {
        const match = data.find((u) => u.id === value);
        if (match) setSelected(match);
      }
    } catch {
      // silent — user can still type manually
    } finally {
      setLoading(false);
    }
  }, [fetched, value, selected, departmentId]);

  // Filter units whenever query changes
  useEffect(() => {
    if (!query.trim()) {
      setFiltered(allUnits.slice(0, 20));
      return;
    }
    const q = query.toLowerCase();
    setFiltered(
      allUnits
        .filter(
          (u) =>
            u.name.toLowerCase().includes(q) ||
            u.code.toLowerCase().includes(q) ||
            u.id.toLowerCase().includes(q)
        )
        .slice(0, 20)
    );
  }, [query, allUnits]);

  const handleFocus = () => {
    fetchUnits();
    setOpen(true);
  };

  const handleSelect = (unit: Unit) => {
    setSelected(unit);
    onChange(unit.id);
    setQuery('');
    setOpen(false);
  };

  const handleClear = () => {
    setSelected(null);
    onChange('');
    setQuery('');
    setFetched(false); // refetch next time
    setAllUnits([]);
  };

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
            <span className="text-gray-700 truncate font-medium">{selected.name}</span>
            <span className="inline-flex px-1.5 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600 shrink-0">
              {selected.code}
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
            <div className="px-3 py-2 text-sm text-gray-500">Loading units…</div>
          )}

          {!loading && filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">No units found</div>
          )}

          {!loading &&
            filtered.map((unit) => (
              <button
                key={unit.id}
                type="button"
                onMouseDown={() => handleSelect(unit)}
                className="w-full text-left px-3 py-2 hover:bg-blue-50 flex items-center gap-3 border-b border-gray-50 last:border-0"
              >
                <span className="text-sm text-gray-700 truncate flex-1 font-medium">
                  {unit.name}
                </span>
                <span className="inline-flex px-1.5 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600 shrink-0">
                  {unit.code}
                </span>
              </button>
            ))}
        </div>
      )}
    </div>
  );
};
