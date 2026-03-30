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
import PatientsPage from '../pages/PatientsPage';
import EncountersPage from '../pages/EncountersPage';
import EncounterDetailPage from '../pages/EncounterDetailPage';
import BedManagementPage from '../pages/BedManagementPage';
import TransportPage from '../pages/TransportPage';
import DischargePage from '../pages/DischargePage';
import IncidentsPage from '../pages/IncidentsPage';
import NotFoundPage from '../pages/NotFoundPage';
import UnauthorizedPage from '../pages/UnauthorizedPage';
import { ProtectedRoute } from '../features/auth/components/ProtectedRoute';

// Governance & Safety feature pages (fully implemented)
import { ConsultsPage } from '../features/consults';
import { ExceptionsPage } from '../features/exceptions';
import { IncidentReportPage, IncidentReviewPage } from '../features/incidents';

// Clinical Care feature pages
import { CareTeamPage } from '../features/care-team/pages/CareTeamPage';

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

  // Dashboard
  {
    path: ROUTES.DASHBOARD,
    element: (
      <ProtectedRoute>
        <HuddleDashboard />
      </ProtectedRoute>
    ),
  },

  // Clinical Care (Developer A)
  {
    path: ROUTES.PATIENT_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_care_team">
        <PatientsPage />
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
    path: ROUTES.ENCOUNTER_DETAIL,
    element: (
      <ProtectedRoute requiredPermission="view_care_team">
        <EncounterDetailPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.ENCOUNTER_CARE_TEAM,
    element: (
      <ProtectedRoute requiredPermission="view_care_team">
        <CareTeamPage />
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

  // Governance & Safety (feature/governance-safety)
  {
    path: ROUTES.CONSULT_LIST,
    element: (
      <ProtectedRoute requiredPermission="view_consults">
        <ConsultsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.EXCEPTION_LIST,
    element: (
      <ProtectedRoute>
        <ExceptionsPage />
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
  {
    path: ROUTES.INCIDENT_REPORT,
    element: (
      <ProtectedRoute requiredPermission="create_incident">
        <IncidentReportPage />
      </ProtectedRoute>
    ),
  },
  {
    path: ROUTES.INCIDENT_REVIEW,
    element: (
      <ProtectedRoute requiredPermission="review_incident">
        <IncidentReviewPage />
      </ProtectedRoute>
    ),
  },

  // Operations (Developer B - placeholder)
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
