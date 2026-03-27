/**
 * ProtectedRoute component
 * Wrapper for routes that require authentication
 * TODO: Check if user is authenticated
 * TODO: Optionally check for required roles
 * TODO: Redirect to login if not authenticated
 * TODO: Show unauthorized page if missing required role
 */

import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import type { Role } from '../../../shared/types';
import { ROUTES } from '../../../shared/config/routes';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: Role;
  requiredRoles?: Role[];
}

export const ProtectedRoute = ({
  children,
  requiredRole,
  requiredRoles,
}: ProtectedRouteProps) => {
  const { isAuthenticated, isLoading, user } = useAuth();

  // TODO: Show loading spinner while checking auth
  if (isLoading) {
    return <div>Loading...</div>;
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

  return <>{children}</>;
};
