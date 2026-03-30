/**
 * Route path constants
 * Centralized route definitions for type-safe navigation
 */

export const ROUTES = {
  // Public routes
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  
  // Protected routes
  DASHBOARD: '/dashboard',
  
  // Patient routes
  PATIENT_LIST: '/patients',
  
  // Encounter routes
  ENCOUNTER_LIST: '/encounters',
  ENCOUNTER_DETAIL: '/encounters/:id',
  ENCOUNTER_CARE_TEAM: '/encounters/:id/care-team',
  ENCOUNTER_FLOW: '/encounters/:id/flow',
  ENCOUNTER_TASKS: '/encounters/:id/tasks',
  
  // Task routes
  TASK_LIST: '/tasks',
  TASK_DETAIL: '/tasks/:id',
  
  // Consult routes
  CONSULT_LIST: '/consults',
  CONSULT_DETAIL: '/consults/:id',

  // Exception routes
  EXCEPTION_LIST: '/exceptions',
  EXCEPTION_DETAIL: '/exceptions/:id',

  // Incident routes
  INCIDENT_LIST: '/incidents',
  INCIDENT_REPORT: '/incidents/report',
  INCIDENT_REVIEW: '/incidents/review',
  INCIDENT_DETAIL: '/incidents/:id',

  // Bed management routes
  BED_LIST: '/beds',
  BED_DETAIL: '/beds/:id',
  
  // Transport routes
  TRANSPORT_LIST: '/transport',
  TRANSPORT_DETAIL: '/transport/:id',
  
  // Discharge routes
  DISCHARGE_LIST: '/discharge',
  DISCHARGE_DETAIL: '/discharge/:id',
  
  // Admin routes
  DEPARTMENT_LIST: '/admin/departments',
  UNIT_LIST: '/admin/units',
  
  // Error routes
  NOT_FOUND: '/404',
  UNAUTHORIZED: '/unauthorized',
} as const;

/**
 * Helper to generate route with parameters
 * @param route - Route template with :param
 * @param params - Object with parameter values
 * @returns Populated route string
 */
export const buildRoute = (route: string, params: Record<string, string>): string => {
  let result = route;
  Object.entries(params).forEach(([key, value]) => {
    result = result.replace(`:${key}`, value);
  });
  return result;
};

/**
 * Navigation helpers
 */
export const navHelpers = {
  toEncounterCareTeam: (encounterId: string) => 
    buildRoute(ROUTES.ENCOUNTER_CARE_TEAM, { id: encounterId }),
  
  toEncounterFlow: (encounterId: string) => 
    buildRoute(ROUTES.ENCOUNTER_FLOW, { id: encounterId }),
  
  toEncounterTasks: (encounterId: string) => 
    buildRoute(ROUTES.ENCOUNTER_TASKS, { id: encounterId }),
  
  toTaskDetail: (taskId: string) => 
    buildRoute(ROUTES.TASK_DETAIL, { id: taskId }),
  
  toConsultDetail: (consultId: string) => 
    buildRoute(ROUTES.CONSULT_DETAIL, { id: consultId }),
  
  toExceptionDetail: (exceptionId: string) =>
    buildRoute(ROUTES.EXCEPTION_DETAIL, { id: exceptionId }),
  
  toIncidentDetail: (incidentId: string) =>
    buildRoute(ROUTES.INCIDENT_DETAIL, { id: incidentId }),
};
