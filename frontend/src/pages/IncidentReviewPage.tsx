/**
 * Incident Review Page - Quality/Safety review workflow (restricted access)
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const IncidentReviewPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Review Incidents"
        subtitle="Quality and safety incident review (Quality/Safety & Admin only)"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">🔍</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Incident Review
        </h2>
        <p className="text-gray-600 mb-4">
          Review and resolve submitted incident reports.
        </p>
        <p className="text-sm text-gray-500">
          Feature available via the Governance &amp; Safety branch integration.
        </p>
      </div>
    </Layout>
  );
};

export default IncidentReviewPage;
