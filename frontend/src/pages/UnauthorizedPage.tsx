/**
 * UnauthorizedPage - 403 error page
 */

import { Link } from 'react-router-dom';
import { ROUTES } from '../shared/config/routes';
import { useAuth } from '../features/auth/hooks/useAuth';

export const UnauthorizedPage = () => {
  const { user } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="text-center">
        <div className="text-9xl mb-4">🔒</div>
        <h1 className="text-6xl font-bold text-gray-900 mb-4">403</h1>
        <p className="text-xl text-gray-600 mb-4">
          Access Denied
        </p>
        <p className="text-gray-500 mb-8">
          You don't have permission to access this page.
          {user && (
            <span className="block mt-2">
              Your role: <strong>{user.role}</strong>
            </span>
          )}
        </p>
        <div className="flex gap-4 justify-center">
          <Link
            to={ROUTES.DASHBOARD}
            className="inline-flex items-center px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors"
          >
            ← Back to Dashboard
          </Link>
          {!user && (
            <Link
              to={ROUTES.LOGIN}
              className="inline-flex items-center px-6 py-3 bg-gray-200 text-gray-900 font-medium rounded-lg hover:bg-gray-300 transition-colors"
            >
              Login
            </Link>
          )}
        </div>
      </div>
    </div>
  );
};

export default UnauthorizedPage;
