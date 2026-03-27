/**
 * ProtectedRoute component
 * Wrapper for routes that require authentication and/or permissions
 * Redirects to login if not authenticated
 * Shows unauthorized page if missing required permissions
 */

import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { usePermissions } from '../hooks/usePermissions';
import type { Role, Permission } from '../../../shared/types';
import { ROUTES } from '../../../shared/config/routes';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: Role;
  requiredRoles?: Role[];
  requiredPermission?: Permission;
}

export const ProtectedRoute = ({
  children,
  requiredRole,
  requiredRoles,
  requiredPermission,
}: ProtectedRouteProps) => {
  const { isAuthenticated, isLoading, user } = useAuth();
  const { hasPermission } = usePermissions();

  // Show loading spinner while checking auth
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to={ROUTES.LOGIN} replace />;
  }

  // Check for required role(s)
  if (requiredRole && user?.role !== requiredRole) {
    return <Navigate to={ROUTES.UNAUTHORIZED} replace />;
  }

  if (requiredRoles && !requiredRoles.includes(user?.role!)) {
    return <Navigate to={ROUTES.UNAUTHORIZED} replace />;
  }

  // Check for required permission
  if (requiredPermission && !hasPermission(requiredPermission)) {
    return <Navigate to={ROUTES.UNAUTHORIZED} replace />;
  }

  return <>{children}</>;
};
