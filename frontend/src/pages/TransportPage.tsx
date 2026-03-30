/**
 * Transport Page - Transport request management
 * TODO: Implement transport request view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const TransportPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Transport"
        subtitle="Patient transport requests and tracking"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">🚛</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Transport Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display transport requests with status tracking.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Create transport request, assign staff, track progress
        </p>
      </div>
    </Layout>
  );
};

export default TransportPage;
