/**
 * Sidebar component - Left navigation menu
 * TODO: Implement role-aware navigation
 * TODO: Add active route highlighting
 * TODO: Add collapse/expand functionality
 * TODO: Show only links user has permission to access
 */

import { ROUTES } from '../../config/routes';

export const Sidebar = () => {
  // TODO: Get user permissions from auth store
  // TODO: Filter navigation items based on permissions
  // TODO: Highlight active route
  // TODO: Add icons using lucide-react
  
  return (
    <aside className="bg-gray-50 border-r border-gray-200 w-64 min-h-screen">
      <nav className="p-4 space-y-2">
        {/* Dashboard */}
        <a
          href={ROUTES.DASHBOARD}
          className="block px-4 py-2 text-sm font-medium text-gray-900 rounded-lg hover:bg-gray-100"
        >
          Dashboard
        </a>
        
        {/* TODO: Add navigation links for:
         * - Encounters
         * - Tasks
         * - Consults
         * - Bed Management
         * - Transport
         * - Discharge
         * - Incidents
         * 
         * Show/hide based on user role and permissions
         */}
        
        <div className="pt-4 border-t border-gray-200 mt-4">
          <p className="px-4 text-xs font-semibold text-gray-500 uppercase">
            Modules
          </p>
          {/* Module links */}
        </div>
      </nav>
    </aside>
  );
};
