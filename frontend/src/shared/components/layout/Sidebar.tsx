/**
 * Sidebar component - Role-aware left navigation menu
 * Filters navigation items based on user permissions
 * Disables modules not available to the current user
 */

import { useLocation, Link } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  ListTodo,
  MessageSquare,
  Bed,
  Truck,
  ClipboardCheck,
  AlertTriangle,
  Activity,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { usePermissions } from '../../../features/auth/hooks/usePermissions';
import { Permission } from '../../types';
import { ROUTES } from '../../config/routes';
import { cn } from '../../utils/cn';

interface NavItem {
  label: string;
  path: string;
  icon: LucideIcon;
  permission?: Permission; // If undefined, available to all authenticated users
  badge?: number;
  comingSoon?: boolean; // Feature not yet built
}

interface NavSection {
  title?: string; // If undefined, no section header shown
  items: NavItem[];
}

export const Sidebar = () => {
  const location = useLocation();
  const { hasPermission } = usePermissions();

  // Define navigation structure with required permissions (grouped by section)
  const navigationSections: NavSection[] = [
    // Core section (no label)
    {
      items: [
        {
          label: 'Dashboard',
          path: ROUTES.DASHBOARD,
          icon: LayoutDashboard,
          // No permission required - all authenticated users can view dashboard
        },
      ],
    },
    // Clinical Care section
    {
      title: 'Clinical Care',
      items: [
        {
          label: 'Encounters',
          path: ROUTES.ENCOUNTER_LIST,
          icon: Users,
          permission: Permission.VIEW_CARE_TEAM,
        },
        {
          label: 'Tasks',
          path: ROUTES.TASK_LIST,
          icon: ListTodo,
          permission: Permission.VIEW_TASKS,
        },
      ],
    },
    // Governance & Safety section
    {
      title: 'Governance & Safety',
      items: [
        {
          label: 'Consults',
          path: ROUTES.CONSULT_LIST,
          icon: MessageSquare,
          permission: Permission.VIEW_CONSULTS,
        },
        {
          label: 'Incidents',
          path: ROUTES.INCIDENT_LIST,
          icon: AlertTriangle,
          permission: Permission.VIEW_INCIDENTS,
        },
      ],
    },
    // Operations section (coming soon)
    {
      title: 'Operations',
      items: [
        {
          label: 'Bed Management',
          path: ROUTES.BED_LIST,
          icon: Bed,
          permission: Permission.VIEW_BEDS,
          comingSoon: true,
        },
        {
          label: 'Transport',
          path: ROUTES.TRANSPORT_LIST,
          icon: Truck,
          permission: Permission.VIEW_TRANSPORT,
          comingSoon: true,
        },
        {
          label: 'Discharge',
          path: ROUTES.DISCHARGE_LIST,
          icon: ClipboardCheck,
          permission: Permission.VIEW_CARE_TEAM,
          comingSoon: true,
        },
      ],
    },
  ];

  // Flatten all items for the "Your Access" count
  const allNavigationItems = navigationSections.flatMap((section) => section.items);

  /**
   * Check if user has access to a navigation item
   */
  const hasAccess = (item: NavItem): boolean => {
    if (!item.permission) return true; // No permission required
    return hasPermission(item.permission);
  };

  /**
   * Check if current route is active
   */
  const isActive = (path: string): boolean => {
    // Exact match for dashboard
    if (path === ROUTES.DASHBOARD) {
      return location.pathname === path || location.pathname === ROUTES.HOME;
    }
    // Prefix match for other routes (to highlight parent when on detail pages)
    return location.pathname.startsWith(path);
  };

  return (
    <aside className="bg-gray-50 border-r border-gray-200 w-64 min-h-screen flex flex-col">
      <nav className="flex-1 p-4 space-y-1">
        {navigationSections.map((section, sectionIdx) => (
          <div key={sectionIdx}>
            {/* Section header */}
            {section.title && (
              <p className="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                {section.title}
              </p>
            )}

            {/* Section items */}
            <div className={cn('space-y-1', section.title && 'mb-4')}>
              {section.items.map((item) => {
                const Icon = item.icon;
                const active = isActive(item.path);
                const allowed = hasAccess(item);

                // Coming soon items (not built yet)
                if (item.comingSoon) {
                  return (
                    <div
                      key={item.path}
                      className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium text-gray-400 rounded-lg cursor-not-allowed opacity-40"
                      title="Coming soon"
                    >
                      <Icon className="w-5 h-5 flex-shrink-0" />
                      <span className="flex-1">{item.label}</span>
                      <span className="text-xs text-gray-400">(Soon)</span>
                    </div>
                  );
                }

                // If user doesn't have access, render as locked
                if (!allowed) {
                  return (
                    <div
                      key={item.path}
                      className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium text-gray-400 rounded-lg cursor-not-allowed opacity-50"
                      title="You don't have permission to access this module"
                    >
                      <Icon className="w-5 h-5 flex-shrink-0" />
                      <span className="flex-1">{item.label}</span>
                      <span className="text-xs text-gray-400">🔒</span>
                    </div>
                  );
                }

                // Render as enabled link
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={cn(
                      'flex items-center gap-3 px-4 py-2.5 text-sm font-medium rounded-lg transition-colors',
                      active
                        ? 'bg-blue-100 text-blue-700 hover:bg-blue-200'
                        : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                    )}
                  >
                    <Icon className="w-5 h-5 flex-shrink-0" />
                    <span className="flex-1">{item.label}</span>
                    {item.badge !== undefined && (
                      <span
                        className={cn(
                          'px-2 py-0.5 text-xs font-semibold rounded-full',
                          active
                            ? 'bg-blue-200 text-blue-800'
                            : 'bg-gray-200 text-gray-700'
                        )}
                      >
                        {item.badge}
                      </span>
                    )}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* Footer with user role info */}
      <div className="p-4 border-t border-gray-200">
        <div className="px-4 py-2 bg-gray-100 rounded-lg">
          <p className="text-xs font-semibold text-gray-500 uppercase mb-1">
            Your Access
          </p>
          <p className="text-sm text-gray-700">
            {allNavigationItems.filter((item) => hasAccess(item) && !item.comingSoon).length} of{' '}
            {allNavigationItems.filter((item) => !item.comingSoon).length} modules
          </p>
        </div>
      </div>
    </aside>
  );
};
