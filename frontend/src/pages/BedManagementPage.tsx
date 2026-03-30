/**
 * Bed Management Page
 * TODO: Implement bed management view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const BedManagementPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Bed Management"
        subtitle="Unit bed status and availability tracking"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">🛏️</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Bed Management Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display bed status, availability, and capabilities across units.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Bed status tracking, capacity management, assignment
        </p>
      </div>
    </Layout>
  );
};

export default BedManagementPage;
