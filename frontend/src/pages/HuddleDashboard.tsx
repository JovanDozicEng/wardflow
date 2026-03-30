/**
 * HuddleDashboard Page
 * Main dashboard page showing daily huddle metrics
 */

import { useState, useEffect, useCallback } from 'react';
import { RefreshCw, Filter } from 'lucide-react';
import { dashboardService } from '../features/dashboard/services/dashboardService';
import type { HuddleMetrics, DashboardFilterParams } from '../features/dashboard/types';
import { CensusCard } from '../features/dashboard/components/CensusCard';
import { FlowDistributionCard } from '../features/dashboard/components/FlowDistributionCard';
import { TaskMetricsCard } from '../features/dashboard/components/TaskMetricsCard';
import { RiskIndicatorsCard } from '../features/dashboard/components/RiskIndicatorsCard';
import { DrillDownList } from '../features/dashboard/components/DrillDownList';
import { Button } from '../shared/components/ui/Button';
import { Card } from '../shared/components/ui/Card';
import { Layout } from '../shared/components/layout/Layout';
import { UnitAutocomplete } from '../shared/components/ui/UnitAutocomplete';
import { DepartmentAutocomplete } from '../shared/components/ui/DepartmentAutocomplete';

export const HuddleDashboard = () => {
  const [metrics, setMetrics] = useState<HuddleMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<DashboardFilterParams>({});
  const [showFilters, setShowFilters] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchMetrics = useCallback(async () => {
    try {
      setError(null);
      const data = await dashboardService.getHuddleMetrics(filters);
      setMetrics(data);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || 'Failed to fetch dashboard metrics');
      console.error('Dashboard fetch error:', err);
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  // Auto-refresh every 2 minutes
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      fetchMetrics();
    }, 120000); // 2 minutes

    return () => clearInterval(interval);
  }, [autoRefresh, fetchMetrics]);

  const handleRefresh = () => {
    setIsLoading(true);
    fetchMetrics();
  };

  if (error) {
    return (
      <Layout>
        <Card padding="md">
          <div className="text-center py-12">
            <p className="text-red-600 font-medium mb-2">Error Loading Dashboard</p>
            <p className="text-gray-600 text-sm mb-4">{error}</p>
            <Button onClick={handleRefresh}>Try Again</Button>
          </div>
        </Card>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Daily Huddle Dashboard</h1>
          {metrics && (
            <p className="text-sm text-gray-500 mt-1">
              Last updated: {new Date(metrics.generatedAt).toLocaleTimeString()}
            </p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded"
            />
            Auto-refresh
          </label>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2"
          >
            <Filter className="w-4 h-4" />
            Filters
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleRefresh}
            disabled={isLoading}
            className="flex items-center gap-2"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      {showFilters && (
        <Card padding="md">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <UnitAutocomplete
              label="Unit"
              placeholder="Filter by unit"
              value={filters.unitId || ''}
              onChange={(id) => setFilters({ ...filters, unitId: id || undefined })}
            />
            <DepartmentAutocomplete
              label="Department"
              placeholder="Filter by department"
              value={filters.departmentId || ''}
              onChange={(id) => setFilters({ ...filters, departmentId: id || undefined })}
            />
          </div>
        </Card>
      )}

      {/* Metrics Grid */}
      {isLoading && !metrics ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} padding="md">
              <div className="animate-pulse space-y-3">
                <div className="h-4 bg-gray-200 rounded w-1/2" />
                <div className="h-8 bg-gray-200 rounded w-3/4" />
              </div>
            </Card>
          ))}
        </div>
      ) : metrics ? (
        <>
          {/* Top Row: Census & Flow */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <CensusCard census={metrics.census} />
            <FlowDistributionCard distribution={metrics.flowDistribution} />
          </div>

          {/* Middle Row: Tasks & Risk */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <TaskMetricsCard metrics={metrics.taskMetrics} />
            <RiskIndicatorsCard indicators={metrics.riskIndicators} />
          </div>

          {/* Drill-Down Lists */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <DrillDownList
              title="Overdue Tasks"
              items={metrics.overdueTasks}
              type="task"
              emptyMessage="No overdue tasks"
            />
            <DrillDownList
              title="Long Stay Patients"
              items={metrics.longStayPatients}
              type="encounter"
              emptyMessage="No long stay patients"
            />
            <DrillDownList
              title="Pending Discharges"
              items={metrics.pendingDischarges}
              type="encounter"
              emptyMessage="No pending discharges"
            />
          </div>
        </>
      ) : null}
      </div>
    </Layout>
  );
};

export default HuddleDashboard;
