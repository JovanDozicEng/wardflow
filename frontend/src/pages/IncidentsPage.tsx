/**
 * Incidents Page - Quality/Safety incident logging
 * TODO: Implement incident reporting view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const IncidentsPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Safety Incidents"
        subtitle="Quality and safety incident reporting"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">⚠️</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Safety Incidents Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display incident reports with review workflows.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Incident reporting, review process, resolution tracking
        </p>
      </div>
    </Layout>
  );
};

export default IncidentsPage;
