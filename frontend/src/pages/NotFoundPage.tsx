/**
 * NotFoundPage - 404 error page
 */

import { Link } from 'react-router-dom';
import { ROUTES } from '../shared/config/routes';

export const NotFoundPage = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="text-center">
        <div className="text-9xl mb-4">🔍</div>
        <h1 className="text-6xl font-bold text-gray-900 mb-4">404</h1>
        <p className="text-xl text-gray-600 mb-8">
          Oops! The page you're looking for doesn't exist.
        </p>
        <Link
          to={ROUTES.DASHBOARD}
          className="inline-flex items-center px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors"
        >
          ← Back to Dashboard
        </Link>
      </div>
    </div>
  );
};

export default NotFoundPage;
