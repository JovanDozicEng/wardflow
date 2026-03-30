/**
 * Discharge Page - Discharge planning and checklists
 * TODO: Implement discharge checklist view
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const DischargePage = () => {
  return (
    <Layout>
      <PageHeader
        title="Discharge Planning"
        subtitle="Discharge checklists and coordination"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">📋</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Discharge Planning Module
        </h2>
        <p className="text-gray-600 mb-4">
          This page will display discharge checklists and planning workflows.
        </p>
        <p className="text-sm text-gray-500">
          Coming soon: Discharge checklists, required items, completion tracking
        </p>
      </div>
    </Layout>
  );
};

export default DischargePage;
