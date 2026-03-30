/**
 * IncidentForm - Form to report a new incident
 */

import { useState } from 'react';
import { Input } from '@/shared/components/ui/Input';
import { Button } from '@/shared/components/ui/Button';
import { Select } from '@/shared/components/ui/Select';
import { Card } from '@/shared/components/ui/Card';
import type { CreateIncidentRequest } from '../types/incident.types';

interface IncidentFormProps {
  onSubmit: (data: CreateIncidentRequest) => Promise<void>;
  loading?: boolean;
}

export const IncidentForm = ({ onSubmit, loading = false }: IncidentFormProps) => {
  const [formData, setFormData] = useState<CreateIncidentRequest>({
    encounterId: '',
    type: '',
    severity: '',
    harmIndicators: {},
    eventTime: '',
  });
  const [harmIndicatorsJson, setHarmIndicatorsJson] = useState('{}');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);

    // Validate required fields
    if (!formData.type.trim()) {
      setError('Incident type is required');
      return;
    }
    if (!formData.eventTime) {
      setError('Event time is required');
      return;
    }

    // Parse harm indicators JSON
    let parsedHarmIndicators: Record<string, any> = {};
    if (harmIndicatorsJson.trim()) {
      try {
        parsedHarmIndicators = JSON.parse(harmIndicatorsJson);
      } catch (err) {
        setError('Invalid JSON format for harm indicators');
        return;
      }
    }

    // Normalize datetime-local value ("YYYY-MM-DDTHH:mm") to RFC3339 ("YYYY-MM-DDTHH:mm:00Z")
    const normalizeEventTime = (dt: string): string => {
      if (!dt) return dt;
      // Already has seconds/timezone — leave as-is
      if (/T\d{2}:\d{2}:\d{2}/.test(dt)) return dt.endsWith('Z') ? dt : dt + 'Z';
      return dt + ':00Z';
    };

    const submitData: CreateIncidentRequest = {
      ...formData,
      encounterId: formData.encounterId?.trim() || undefined,
      severity: formData.severity || undefined,
      harmIndicators: Object.keys(parsedHarmIndicators).length > 0 ? parsedHarmIndicators : undefined,
      eventTime: normalizeEventTime(formData.eventTime),
    };

    try {
      await onSubmit(submitData);
      // Reset form on success
      setFormData({
        encounterId: '',
        type: '',
        severity: '',
        harmIndicators: {},
        eventTime: '',
      });
      setHarmIndicatorsJson('{}');
      setSuccess(true);
      // Auto-hide success message after 3 seconds
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: any) {
      setError(err.message || 'Failed to report incident');
    }
  };

  return (
    <Card className="p-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800">
            {error}
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded p-3 text-sm text-green-800">
            Incident reported successfully!
          </div>
        )}

        <div>
          <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-1">
            Incident Type *
          </label>
          <Input
            id="type"
            value={formData.type}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, type: e.target.value })}
            placeholder="e.g., medication_error, fall, pressure_injury"
            disabled={loading}
            required
          />
        </div>

        <div>
          <label htmlFor="encounterId" className="block text-sm font-medium text-gray-700 mb-1">
            Encounter ID (Optional)
          </label>
          <Input
            id="encounterId"
            value={formData.encounterId || ''}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, encounterId: e.target.value })}
            placeholder="Associated encounter ID"
            disabled={loading}
          />
        </div>

        <div>
          <label htmlFor="severity" className="block text-sm font-medium text-gray-700 mb-1">
            Severity (Optional)
          </label>
          <Select
            id="severity"
            value={formData.severity || ''}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setFormData({ ...formData, severity: e.target.value })}
            disabled={loading}
            placeholder="Select severity"
            options={[
              { value: 'minor', label: 'Minor' },
              { value: 'moderate', label: 'Moderate' },
              { value: 'severe', label: 'Severe' },
              { value: 'critical', label: 'Critical' },
            ]}
          />
        </div>

        <div>
          <label htmlFor="eventTime" className="block text-sm font-medium text-gray-700 mb-1">
            Event Time *
          </label>
          <Input
            id="eventTime"
            type="datetime-local"
            value={formData.eventTime}
            onChange={(e) => setFormData({ ...formData, eventTime: e.target.value })}
            disabled={loading}
            required
          />
        </div>

        <div>
          <label htmlFor="harmIndicators" className="block text-sm font-medium text-gray-700 mb-1">
            Harm Indicators (JSON, Optional)
          </label>
          <textarea
            id="harmIndicators"
            value={harmIndicatorsJson}
            onChange={(e) => setHarmIndicatorsJson(e.target.value)}
            placeholder='{"patient_harm": false, "witness_present": true}'
            rows={4}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed font-mono text-sm"
          />
          <p className="text-xs text-gray-500 mt-1">Optional: Structured data about harm indicators</p>
        </div>

        <Button
          type="submit"
          variant="primary"
          disabled={loading}
          isLoading={loading}
          className="w-full px-4 py-2 bg-red-600 hover:bg-red-700 text-white"
        >
          {loading ? 'Reporting...' : 'Report Incident'}
        </Button>
      </form>
    </Card>
  );
};
