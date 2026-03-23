/**
 * usePermissions hook
 * Convenience hook for checking user permissions
 */

import { useAuthStore } from '../store/authStore';
import { Role, Permission, ROLE_PERMISSIONS } from '../../../shared/types';

export const usePermissions = () => {
  const { user, hasRole, hasUnitAccess, hasDepartmentAccess } = useAuthStore();

  /**
   * Check if user has a specific permission
   * Note: This is a client-side hint only. Backend must enforce permissions.
   */
  const hasPermission = (permission: Permission): boolean => {
    if (!user) return false;
    
    const rolePermissions = ROLE_PERMISSIONS[user.role];
    return rolePermissions.includes(permission);
  };

  /**
   * Check if user has any of the provided roles
   */
  const hasAnyRole = (roles: Role[]): boolean => {
    if (!user) return false;
    return roles.includes(user.role);
  };

  /**
   * Check if user has all of the provided roles
   * (Usually just one role, but keeping for flexibility)
   */
  const hasAllRoles = (roles: Role[]): boolean => {
    if (!user) return false;
    return roles.every((role) => user.role === role);
  };

  return {
    user,
    hasRole,
    hasPermission,
    hasAnyRole,
    hasAllRoles,
    hasUnitAccess,
    hasDepartmentAccess,
  };
};
