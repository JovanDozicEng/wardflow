/**
 * CareTeamPage - Care team management for a specific encounter
 */

import { useParams, Link } from 'react-router-dom';
import { Layout } from '@/shared/components/layout/Layout';
import { PageHeader } from '@/shared/components/layout/PageHeader';
import { CareTeamList } from '../components/CareTeamList';
import { usePermissions } from '@/features/auth/hooks/usePermissions';
import { Permission } from '@/shared/types';
import { ROUTES } from '@/shared/config/routes';

export const CareTeamPage = () => {
  const { id: encounterId } = useParams<{ id: string }>();
  const { hasPermission } = usePermissions();

  const canAssign = hasPermission(Permission.ASSIGN_CARE_TEAM);
  const canTransfer = hasPermission(Permission.TRANSFER_CARE_TEAM);

  if (!encounterId) {
    return (
      <Layout>
        <div className="max-w-4xl mx-auto px-4 py-8">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <h2 className="text-xl font-semibold text-red-900">Missing Encounter ID</h2>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="mb-4">
          <Link
            to={ROUTES.ENCOUNTER_LIST}
            className="text-sm text-blue-600 hover:underline"
          >
            ← Back to Encounters
          </Link>
        </div>
        <PageHeader
          title="Care Team"
          subtitle={`Assignments and handoffs for encounter ${encounterId}`}
        />
        <div className="mt-6">
          <CareTeamList
            encounterId={encounterId}
            canAssign={canAssign}
            canTransfer={canTransfer}
          />
        </div>
      </div>
    </Layout>
  );
};
