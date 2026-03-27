/**
 * Header component - Top navigation bar
 * Shows user info and logout button when authenticated
 */

import { useAuth } from '../../../features/auth/hooks/useAuth';
import { ROLE_LABELS } from '../../utils/constants';

export const Header = () => {
  const { user, isAuthenticated, logout } = useAuth();
  
  return (
    <header className="bg-white border-b border-gray-200 h-16 flex items-center px-4 lg:px-6">
      <div className="flex items-center justify-between w-full">
        {/* Logo */}
        <div className="flex items-center space-x-4">
          <h1 className="text-xl font-bold text-indigo-600">WardFlow</h1>
        </div>
        
        {/* User menu */}
        {isAuthenticated && user && (
          <div className="flex items-center space-x-4">
            <div className="text-sm text-gray-700">
              <span className="font-medium">{user.name}</span>
              <span className="text-gray-500 ml-2">
                {ROLE_LABELS[user.role]}
              </span>
            </div>
            <button
              onClick={logout}
              className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            >
              Logout
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
