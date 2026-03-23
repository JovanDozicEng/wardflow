/**
 * Dashboard Page - Main dashboard after login
 */

import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { useAuth } from '../features/auth/hooks/useAuth';
import { ROLE_LABELS } from '../shared/utils/constants';

export const DashboardPage = () => {
  const { user } = useAuth();

  return (
    <Layout>
      <PageHeader
        title="Dashboard"
        subtitle={`Welcome back, ${user?.name || 'User'}!`}
      />
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Quick stats */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Your Role</p>
              <p className="text-2xl font-semibold text-gray-900 mt-1">
                {user?.role ? ROLE_LABELS[user.role] : 'N/A'}
              </p>
            </div>
            <div className="text-4xl">👤</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Email</p>
              <p className="text-lg font-semibold text-gray-900 mt-1 truncate">
                {user?.email || 'N/A'}
              </p>
            </div>
            <div className="text-4xl">📧</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Status</p>
              <p className="text-2xl font-semibold text-green-600 mt-1">
                {user?.isActive ? 'Active' : 'Inactive'}
              </p>
            </div>
            <div className="text-4xl">✅</div>
          </div>
        </div>
      </div>

      {/* Welcome message */}
      <div className="mt-8 bg-indigo-50 border border-indigo-200 rounded-lg p-6">
        <h2 className="text-xl font-semibold text-indigo-900 mb-2">
          Welcome to WardFlow!
        </h2>
        <p className="text-indigo-700">
          You are successfully logged in. This dashboard will show care coordination metrics,
          active tasks, and patient flow status once the modules are fully implemented.
        </p>
      </div>

      {/* Module placeholders */}
      <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            🏥 Active Encounters
          </h3>
          <p className="text-gray-600">
            View and manage active patient encounters
          </p>
          <p className="text-sm text-gray-500 mt-2">Coming soon...</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            📋 My Tasks
          </h3>
          <p className="text-gray-600">
            Tasks assigned to you or your team
          </p>
          <p className="text-sm text-gray-500 mt-2">Coming soon...</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            👥 Care Teams
          </h3>
          <p className="text-gray-600">
            Manage care team assignments
          </p>
          <p className="text-sm text-gray-500 mt-2">Coming soon...</p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            🚑 Consults
          </h3>
          <p className="text-gray-600">
            Inter-department consultation requests
          </p>
          <p className="text-sm text-gray-500 mt-2">Coming soon...</p>
        </div>
      </div>
    </Layout>
  );
};

export default DashboardPage;
