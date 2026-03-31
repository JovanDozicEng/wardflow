/**
 * IncidentReportPage - Page for reporting new incidents
 */

import { IncidentForm } from '../components/IncidentForm';
import { useIncidentActions } from '../hooks/useIncidentActions';
import type { CreateIncidentRequest } from '../types/incident.types';
import { Layout } from '@/shared/components/layout/Layout';
import { PageHeader } from '@/shared/components/layout/PageHeader';

export const IncidentReportPage = () => {
  const { createIncident, loading } = useIncidentActions();

  const handleSubmit = async (data: CreateIncidentRequest) => {
    try {
      await createIncident(data);
      console.log('Incident reported successfully');
    } catch (err) {
      console.error('Failed to report incident:', err);
      throw err;
    }
  };

  return (
    <Layout>
      <PageHeader
        title="Report Incident"
        subtitle="Submit a safety incident or event for review"
      />

      <div className="max-w-2xl">
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <h3 className="font-medium text-blue-900 mb-2">Incident Reporting Guidelines</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• Report all safety events, near misses, and adverse incidents</li>
            <li>• Provide as much detail as possible about the event</li>
            <li>• Include the event time as accurately as you can</li>
            <li>• All reports are confidential and used for quality improvement</li>
          </ul>
        </div>

        <IncidentForm onSubmit={handleSubmit} loading={loading} />
      </div>
    </Layout>
  );
};
