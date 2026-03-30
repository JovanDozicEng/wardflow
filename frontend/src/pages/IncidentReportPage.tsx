/**
 * Incident Report Page - Submit new quality/safety incidents
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const IncidentReportPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Report Incident"
        subtitle="Submit a new quality or safety incident"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">📝</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Incident Reporting
        </h2>
        <p className="text-gray-600 mb-4">
          Report new quality and safety incidents for review and tracking.
        </p>
        <p className="text-sm text-gray-500">
          Feature available via the Governance &amp; Safety branch integration.
        </p>
      </div>
    </Layout>
  );
};

export default IncidentReportPage;
