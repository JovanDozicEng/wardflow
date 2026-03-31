/**
 * IncidentReviewPage - Page for quality_safety role to review incidents
 */

import { useState } from 'react';
import { IncidentReviewQueue } from '../components/IncidentReviewQueue';
import { IncidentDetail } from '../components/IncidentDetail';
import { StatusUpdateModal } from '../components/StatusUpdateModal';
import { Button } from '@/shared/components/ui/Button';
import { useIncidentActions } from '../hooks/useIncidentActions';
import { usePermissions } from '@/features/auth/hooks/usePermissions';
import type { Incident, UpdateIncidentStatusRequest } from '../types/incident.types';
import { Layout } from '@/shared/components/layout/Layout';
import { PageHeader } from '@/shared/components/layout/PageHeader';

export const IncidentReviewPage = () => {
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const { updateStatus, loading } = useIncidentActions();
  const { hasAnyRole } = usePermissions();

  const canReviewIncidents = hasAnyRole(['quality_safety', 'admin']);

  const handleReview = (incident: Incident) => {
    setSelectedIncident(incident);
  };

  const handleUpdateStatus = () => {
    setShowStatusModal(true);
  };

  const handleStatusSubmit = async (data: UpdateIncidentStatusRequest) => {
    if (!selectedIncident) return;
    try {
      await updateStatus(selectedIncident.id, data);
      setShowStatusModal(false);
      setSelectedIncident((prev) =>
        prev ? { ...prev, status: data.status } : null
      );
      console.log('Status updated successfully');
    } catch (err) {
      console.error('Failed to update status:', err);
      throw err;
    }
  };

  if (!canReviewIncidents) {
    return (
      <Layout>
        <PageHeader title="Review Incidents" subtitle="Quality and safety incident review" />
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
          <h2 className="text-xl font-semibold text-red-900 mb-2">Access Denied</h2>
          <p className="text-red-800">
            You do not have permission to review incidents. This page is restricted to quality_safety and admin roles.
          </p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <PageHeader
        title="Incident Review"
        subtitle="Review and manage reported safety incidents"
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Review Queue</h2>
          <IncidentReviewQueue onReview={handleReview} />
        </div>

        <div>
          {selectedIncident ? (
            <>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">Incident Details</h2>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleUpdateStatus}
                  disabled={loading}
                  className="px-3 py-1.5 text-sm bg-blue-600 hover:bg-blue-700 text-white"
                >
                  Update Status
                </Button>
              </div>
              <IncidentDetail incident={selectedIncident} />
            </>
          ) : (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
              <p className="text-gray-600">Select an incident from the queue to view details</p>
            </div>
          )}
        </div>
      </div>

      <StatusUpdateModal
        isOpen={showStatusModal}
        onClose={() => setShowStatusModal(false)}
        onSubmit={handleStatusSubmit}
        incident={selectedIncident}
        loading={loading}
      />
    </Layout>
  );
};
