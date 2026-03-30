/**
 * Encounters Page - Encounter list/management
 * TODO: Implement encounter list view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const EncountersPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Encounters"
        subtitle="Active patient encounters and care coordination"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">🏥</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Encounters Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display active patient encounters with filtering and search.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Encounter list, filters, care team assignment, flow tracking
        </p>
      </div>
    </Layout>
  );
};

export default EncountersPage;
