/**
 * React Router configuration
 * Centralized router setup with route definitions
 */

import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ROUTES } from '../shared/config/routes';

// TODO: Import actual page components when implemented
// import LoginPage from '../features/auth/pages/LoginPage';
// import RegisterPage from '../features/auth/pages/RegisterPage';
// import DashboardPage from '../pages/DashboardPage';
// import NotFoundPage from '../pages/NotFoundPage';
// import UnauthorizedPage from '../pages/UnauthorizedPage';

// Placeholder components (remove when actual pages are imported)
const PlaceholderPage = ({ title }: { title: string }) => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="text-center">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">{title}</h1>
      <p className="text-gray-600">This page is under construction</p>
    </div>
  </div>
);

export const router = createBrowserRouter([
  {
    path: ROUTES.HOME,
    element: <Navigate to={ROUTES.DASHBOARD} replace />,
  },
  {
    path: ROUTES.LOGIN,
    element: <PlaceholderPage title="Login" />,
    // element: <LoginPage />,
  },
  {
    path: ROUTES.REGISTER,
    element: <PlaceholderPage title="Register" />,
    // element: <RegisterPage />,
  },
  {
    path: ROUTES.DASHBOARD,
    element: <PlaceholderPage title="Dashboard" />,
    // element: <DashboardPage />,
  },
  {
    path: ROUTES.NOT_FOUND,
    element: <PlaceholderPage title="404 - Not Found" />,
    // element: <NotFoundPage />,
  },
  {
    path: ROUTES.UNAUTHORIZED,
    element: <PlaceholderPage title="403 - Unauthorized" />,
    // element: <UnauthorizedPage />,
  },
  {
    path: '*',
    element: <Navigate to={ROUTES.NOT_FOUND} replace />,
  },
]);

export default router;
