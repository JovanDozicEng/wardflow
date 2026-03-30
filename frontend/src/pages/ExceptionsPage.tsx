/**
 * Exceptions Page - Exception event management
 * Handles draft → finalized → corrected exception workflow
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';

export const ExceptionsPage = () => {
  return (
    <Layout>
      <PageHeader
        title="Exceptions"
        subtitle="Track and manage clinical exception events"
      />
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <div className="text-6xl mb-4">⚠️</div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Exceptions Module
        </h2>
        <p className="text-gray-600 mb-4">
          Create and manage exception events with draft, finalized, and correction workflows.
        </p>
        <p className="text-sm text-gray-500">
          Feature available via the Governance &amp; Safety branch integration.
        </p>
      </div>
    </Layout>
  );
};

export default ExceptionsPage;
