/**
 * Consults Page - Consult request management
 * TODO: Implement consult list view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const ConsultsPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Consults"
        subtitle="Inter-department consultation requests"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">💬</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Consults Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display consultation requests between departments.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Create consult, accept/decline, track status
        </p>
      </div>
    </Layout>
  );
};

export default ConsultsPage;
