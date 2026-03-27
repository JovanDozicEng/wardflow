/**
 * Sidebar component - Left navigation menu with role-aware links
 * Displays navigation based on user permissions
 * Highlights active routes and disables unavailable modules
 */

import { Link, useLocation } from 'react-router-dom';
import { ROUTES } from '../../config/routes';
import { usePermissions } from '../../../features/auth/hooks/usePermissions';

interface NavLinkProps {
  to: string;
  children: React.ReactNode;
  isActive: boolean;
  disabled?: boolean;
}

const NavLink = ({ to, children, isActive, disabled = false }: NavLinkProps) => {
  const baseClass = 'block px-4 py-2 text-sm font-medium rounded-lg transition-colors';
  
  if (disabled) {
    return (
      <div className={`${baseClass} text-gray-400 cursor-not-allowed bg-gray-50`}>
        {children}
        <span className="text-xs ml-2">(Coming Soon)</span>
      </div>
    );
  }
  
  return (
    <Link
      to={to}
      className={`${baseClass} ${
        isActive
          ? 'bg-blue-100 text-blue-700'
          : 'text-gray-700 hover:bg-gray-100'
      }`}
    >
      {children}
    </Link>
  );
};

export const Sidebar = () => {
  const location = useLocation();
  const { hasAnyRole } = usePermissions();
  
  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };
  
  // Role-based access control
  const canManageConsults = hasAnyRole(['provider', 'consult', 'admin']);
  const canAccessIncidentReview = hasAnyRole(['quality_safety', 'admin']);
  
  return (
    <aside className="bg-white border-r border-gray-200 w-64 min-h-screen">
      <nav className="p-4 space-y-1">
        {/* Dashboard */}
        <NavLink to={ROUTES.DASHBOARD} isActive={isActive(ROUTES.DASHBOARD)}>
          📊 Dashboard
        </NavLink>
        
        {/* Governance & Safety Section */}
        <div className="pt-4 border-t border-gray-200 mt-4">
          <p className="px-4 text-xs font-semibold text-gray-500 uppercase mb-2">
            Governance & Safety
          </p>
          
          {/* Consults - All can view, limited roles can manage */}
          <NavLink to={ROUTES.CONSULT_LIST} isActive={isActive(ROUTES.CONSULT_LIST)}>
            <span>🏥 Consults</span>
            {!canManageConsults && (
              <span className="text-xs text-gray-400 ml-2">(View Only)</span>
            )}
          </NavLink>
          
          {/* Exceptions - All authenticated users can access */}
          <NavLink to={ROUTES.EXCEPTION_LIST} isActive={isActive(ROUTES.EXCEPTION_LIST)}>
            ⚠️ Exceptions
          </NavLink>
          
          {/* Incident Reporting - All users can report */}
          <NavLink to={ROUTES.INCIDENT_REPORT} isActive={isActive(ROUTES.INCIDENT_REPORT)}>
            📝 Report Incident
          </NavLink>
          
          {/* Incident Review - Quality/Safety and Admin only */}
          {canAccessIncidentReview && (
            <NavLink to={ROUTES.INCIDENT_REVIEW} isActive={isActive(ROUTES.INCIDENT_REVIEW)}>
              🔍 Review Incidents
            </NavLink>
          )}
        </div>
        
        {/* Care Coordination Section - Future modules */}
        <div className="pt-4 border-t border-gray-200 mt-4">
          <p className="px-4 text-xs font-semibold text-gray-500 uppercase mb-2">
            Care Coordination
          </p>
          
          <NavLink to={ROUTES.ENCOUNTER_LIST} isActive={false} disabled>
            👥 Encounters
          </NavLink>
          
          <NavLink to={ROUTES.TASK_LIST} isActive={false} disabled>
            📋 Tasks
          </NavLink>
        </div>
        
        {/* Operations Section - Future modules */}
        <div className="pt-4 border-t border-gray-200 mt-4">
          <p className="px-4 text-xs font-semibold text-gray-500 uppercase mb-2">
            Operations
          </p>
          
          <NavLink to={ROUTES.BED_LIST} isActive={false} disabled>
            🛏️ Bed Management
          </NavLink>
          
          <NavLink to={ROUTES.TRANSPORT_LIST} isActive={false} disabled>
            🚑 Transport
          </NavLink>
          
          <NavLink to={ROUTES.DISCHARGE_LIST} isActive={false} disabled>
            🏠 Discharge Planning
          </NavLink>
        </div>
      </nav>
    </aside>
  );
};
