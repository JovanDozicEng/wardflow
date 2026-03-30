/**
 * IncidentsPage - List all safety incidents
 */

import { useState } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { IncidentList } from '../features/incidents/components/IncidentList';
import { IncidentDetail } from '../features/incidents/components/IncidentDetail';
import type { Incident } from '../features/incidents/types/incident.types';

export const IncidentsPage = () => {
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);

  return (
    <Layout>
      <PageHeader
        title="Safety Incidents"
        subtitle="Quality and safety incident reports"
      />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 px-4 pb-8">
        <IncidentList onSelectIncident={setSelectedIncident} />
        <div>
          {selectedIncident ? (
            <IncidentDetail incident={selectedIncident} />
          ) : (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
              <p className="text-gray-600">Select an incident to view details</p>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};

export default IncidentsPage;
