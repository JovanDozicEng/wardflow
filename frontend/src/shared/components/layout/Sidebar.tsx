/**
 * Sidebar component - Role-aware left navigation menu
 * Filters navigation items based on user permissions
 * Disables modules not available to the current user
 */

import { useLocation, Link } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  UserCheck,
  ListTodo,
  MessageSquare,
  Bed,
  Truck,
  ClipboardCheck,
  ShieldAlert,
  FileWarning,
  Search,
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
  permission?: Permission;
  badge?: number;
  comingSoon?: boolean;
  /** Shown alongside label, not as access control */
  hint?: string;
}

interface NavSection {
  title?: string;
  items: NavItem[];
}

export const Sidebar = () => {
  const location = useLocation();
  const { hasPermission, hasAnyRole } = usePermissions();

  const canManageConsults = hasAnyRole(['provider', 'consult', 'admin']);
  const canReviewIncidents = hasAnyRole(['quality_safety', 'admin']);

  const navigationSections: NavSection[] = [
    // Core — no section label
    {
      items: [
        {
          label: 'Dashboard',
          path: ROUTES.DASHBOARD,
          icon: LayoutDashboard,
        },
      ],
    },

    // Clinical Care (Developer A scope)
    {
      title: 'Clinical Care',
      items: [
        {
          label: 'Patients',
          path: ROUTES.PATIENT_LIST,
          icon: UserCheck,
          permission: Permission.VIEW_CARE_TEAM,
        },
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

    // Governance & Safety (mirrors feature/governance-safety sidebar)
    {
      title: 'Governance & Safety',
      items: [
        {
          label: 'Consults',
          path: ROUTES.CONSULT_LIST,
          icon: MessageSquare,
          permission: Permission.VIEW_CONSULTS,
          hint: canManageConsults ? undefined : 'View Only',
        },
        {
          label: 'Exceptions',
          path: ROUTES.EXCEPTION_LIST,
          icon: FileWarning,
          // All authenticated users can access exceptions
        },
        {
          label: 'Report Incident',
          path: ROUTES.INCIDENT_REPORT,
          icon: ShieldAlert,
          permission: Permission.CREATE_INCIDENT,
        },
        // Incident review — only shown for quality_safety + admin roles
        ...(canReviewIncidents
          ? [
              {
                label: 'Review Incidents',
                path: ROUTES.INCIDENT_REVIEW,
                icon: Search,
                permission: Permission.REVIEW_INCIDENT as Permission,
              },
            ]
          : []),
      ],
    },

    // Operations (coming soon — Developer B scope)
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
          label: 'Discharge Planning',
          path: ROUTES.DISCHARGE_LIST,
          icon: ClipboardCheck,
          permission: Permission.VIEW_CARE_TEAM,
          comingSoon: true,
        },
      ],
    },
  ];

  const allNavigationItems = navigationSections.flatMap((s) => s.items);

  const hasAccess = (item: NavItem): boolean => {
    if (!item.permission) return true;
    return hasPermission(item.permission);
  };

  const isActive = (path: string): boolean => {
    if (path === ROUTES.DASHBOARD) {
      return location.pathname === path || location.pathname === ROUTES.HOME;
    }
    return location.pathname.startsWith(path);
  };

  return (
    <aside className="bg-gray-50 border-r border-gray-200 w-64 min-h-screen flex flex-col">
      <nav className="flex-1 p-4">
        {navigationSections.map((section, sectionIdx) => (
          <div key={sectionIdx} className={sectionIdx > 0 ? 'mt-4 pt-4 border-t border-gray-200' : ''}>
            {section.title && (
              <p className="px-4 mb-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                {section.title}
              </p>
            )}

            <div className="space-y-1">
              {section.items.map((item) => {
                const Icon = item.icon;
                const active = isActive(item.path);
                const allowed = hasAccess(item);

                // Coming soon — not yet built
                if (item.comingSoon) {
                  return (
                    <div
                      key={item.path}
                      className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium text-gray-400 rounded-lg cursor-not-allowed opacity-40"
                      title="Coming soon"
                    >
                      <Icon className="w-5 h-5 flex-shrink-0" />
                      <span className="flex-1">{item.label}</span>
                      <span className="text-xs">(Soon)</span>
                    </div>
                  );
                }

                // Permission denied
                if (!allowed) {
                  return (
                    <div
                      key={item.path}
                      className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium text-gray-400 rounded-lg cursor-not-allowed opacity-50"
                      title="You don't have permission to access this module"
                    >
                      <Icon className="w-5 h-5 flex-shrink-0" />
                      <span className="flex-1">{item.label}</span>
                      <span className="text-xs">🔒</span>
                    </div>
                  );
                }

                // Accessible link
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
                    {item.hint && (
                      <span className="text-xs text-gray-400">{item.hint}</span>
                    )}
                    {item.badge !== undefined && (
                      <span
                        className={cn(
                          'px-2 py-0.5 text-xs font-semibold rounded-full',
                          active ? 'bg-blue-200 text-blue-800' : 'bg-gray-200 text-gray-700'
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

      {/* Footer: accessible module count */}
      <div className="p-4 border-t border-gray-200">
        <div className="px-4 py-2 bg-gray-100 rounded-lg">
          <p className="text-xs font-semibold text-gray-500 uppercase mb-1">Your Access</p>
          <p className="text-sm text-gray-700">
            {allNavigationItems.filter((item) => hasAccess(item) && !item.comingSoon).length} of{' '}
            {allNavigationItems.filter((item) => !item.comingSoon).length} modules
          </p>
        </div>
      </div>
    </aside>
  );
};

