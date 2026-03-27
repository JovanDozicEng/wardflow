/**
 * React Router configuration
 * Centralized router setup with route definitions
 */

import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ROUTES } from '../shared/config/routes';
import HomePage from '../pages/HomePage';
import LoginPage from '../features/auth/pages/LoginPage';
import RegisterPage from '../features/auth/pages/RegisterPage';
import HuddleDashboard from '../pages/HuddleDashboard';
import TasksPage from '../pages/TasksPage';
import EncountersPage from '../pages/EncountersPage';
import ConsultsPage from '../pages/ConsultsPage';
import BedManagementPage from '../pages/BedManagementPage';
import TransportPage from '../pages/TransportPage';
import DischargePage from '../pages/DischargePage';
import IncidentsPage from '../pages/IncidentsPage';
import NotFoundPage from '../pages/NotFoundPage';
import UnauthorizedPage from '../pages/UnauthorizedPage';
import { ProtectedRoute } from '../features/auth/components/ProtectedRoute';

export const router = createBrowserRouter([
  {
    path: ROUTES.HOME,
    element: <HomePage />,
  },
  {
    path: ROUTES.LOGIN,
    element: <LoginPage />,
  },
  {
    path: ROUTES.REGISTER,
    element: <RegisterPage />,
  },
  // Protected routes
  {
    path: ROUTES.DASHBOARD,
    element: (
      <ProtectedRoute>
        <HuddleDashboard />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.ENCOUNTER_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_care_team">
        <EncountersPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.TASK_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_tasks">
        <TasksPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.CONSULT_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_consults">
        <ConsultsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.BED_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_beds">
        <BedManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.TRANSPORT_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_transport">
        <TransportPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.DISCHARGE_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_care_team">
        <DischargePage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.INCIDENT_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_incidents">
        <IncidentsPage />
      </ProtectedRoute>
    ),
  },
  // Error pages
  {
    path: ROUTES.NOT_FOUND,
    element: <NotFoundPage />,
  },
  {
    path: ROUTES.UNAUTHORIZED,
    element: <UnauthorizedPage />,
  },
  {
    path: '*',
    element: <Navigate to={ROUTES.NOT_FOUND} replace />,
  },
]);

export default router;
