/**
 * ConsultsPage - Main page for consult management
 */

import { useState } from 'react';
import { ConsultInbox } from '../components/ConsultInbox';
import { ConsultForm } from '../components/ConsultForm';
import { DeclineModal } from '../components/DeclineModal';
import { RedirectModal } from '../components/RedirectModal';
import { Button } from '@/shared/components/ui/Button';
import { useConsultActions } from '../hooks/useConsultActions';
import { usePermissions } from '@/features/auth/hooks/usePermissions';
import { Layout } from '@/shared/components/layout/Layout';

export const ConsultsPage = () => {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [declineConsultId, setDeclineConsultId] = useState<string | null>(null);
  const [redirectConsultId, setRedirectConsultId] = useState<string | null>(null);
  
  const {
    createConsult,
    acceptConsult,
    declineConsult,
    redirectConsult,
    completeConsult,
    loading,
  } = useConsultActions();
  
  const { hasAnyRole } = usePermissions();
  const canCreateConsults = hasAnyRole(['provider', 'consult', 'admin']);

  const handleAccept = async (id: string) => {
    try {
      await acceptConsult(id);
      // TODO: Show success toast
      console.log('Consult accepted successfully');
    } catch (err) {
      // TODO: Show error toast
      console.error('Failed to accept consult:', err);
    }
  };

  const handleDecline = (id: string) => {
    setDeclineConsultId(id);
  };

  const handleDeclineSubmit = async (data: { reason: string }) => {
    if (!declineConsultId) return;
    try {
      await declineConsult(declineConsultId, data);
      setDeclineConsultId(null);
      // TODO: Show success toast
      console.log('Consult declined successfully');
    } catch (err) {
      // TODO: Show error toast
      console.error('Failed to decline consult:', err);
      throw err;
    }
  };

  const handleRedirect = (id: string) => {
    setRedirectConsultId(id);
  };

  const handleRedirectSubmit = async (data: { targetService: string; reason: string }) => {
    if (!redirectConsultId) return;
    try {
      await redirectConsult(redirectConsultId, data);
      setRedirectConsultId(null);
      // TODO: Show success toast
      console.log('Consult redirected successfully');
    } catch (err) {
      // TODO: Show error toast
      console.error('Failed to redirect consult:', err);
      throw err;
    }
  };

  const handleComplete = async (id: string) => {
    try {
      await completeConsult(id);
      // TODO: Show success toast
      console.log('Consult completed successfully');
    } catch (err) {
      // TODO: Show error toast
      console.error('Failed to complete consult:', err);
    }
  };

  return (
    <Layout>
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Consults</h1>
          <p className="text-gray-600 mt-1">Manage consultation requests</p>
        </div>
        {canCreateConsults && (
          <Button
            variant="primary"
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white"
          >
            + New Consult
          </Button>
        )}
      </div>

      {/* Consult Inbox */}
      <ConsultInbox
        onAccept={handleAccept}
        onDecline={handleDecline}
        onRedirect={handleRedirect}
        onComplete={handleComplete}
        actionLoading={loading}
      />

      {/* Create Modal */}
      <ConsultForm
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={async (data) => {
          await createConsult(data);
        }}
        loading={loading}
      />

      {/* Decline Modal */}
      <DeclineModal
        isOpen={declineConsultId !== null}
        onClose={() => setDeclineConsultId(null)}
        onSubmit={handleDeclineSubmit}
        loading={loading}
      />

      {/* Redirect Modal */}
      <RedirectModal
        isOpen={redirectConsultId !== null}
        onClose={() => setRedirectConsultId(null)}
        onSubmit={handleRedirectSubmit}
        loading={loading}
      />
    </div>
    </Layout>
  );
};
