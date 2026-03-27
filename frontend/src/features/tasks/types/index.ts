/**
 * Task Board types
 * Matches backend API for clinical task management
 */

// Task status enum (matches backend TaskStatus)
export type TaskStatus = 'open' | 'in_progress' | 'completed' | 'cancelled' | 'escalated';

// Task priority enum (matches backend TaskPriority)
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent';

// Scope type enum (matches backend ScopeType)
export type ScopeType = 'encounter' | 'patient' | 'unit';

// Task entity (matches backend Task)
export interface Task {
  id: string;
  scopeType: ScopeType;
  scopeId: string;
  title: string;
  details: string | null;
  status: TaskStatus;
  priority: TaskPriority;
  ownerId: string | null;
  ownerName?: string; // Populated when withOwnerDetails=true
  ownerEmail?: string; // Populated when withOwnerDetails=true
  slaDueAt: string | null; // ISO timestamp
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  completedBy: string | null;
  completedAt: string | null;
}

// Task assignment event (immutable history)
export interface TaskAssignmentEvent {
  id: string;
  taskId: string;
  fromOwnerId: string | null;
  toOwnerId: string | null;
  assignedAt: string; // ISO timestamp
  assignedBy: string;
  reason: string | null;
  createdAt: string;
}

// Request to create a task (matches backend CreateTaskRequest)
export interface CreateTaskRequest {
  scopeType: ScopeType;
  scopeId: string;
  title: string;
  details?: string;
  priority?: TaskPriority; // Defaults to 'medium'
  slaDueAt?: string; // ISO timestamp
}

// Request to update a task (matches backend UpdateTaskRequest)
export interface UpdateTaskRequest {
  status?: TaskStatus;
  priority?: TaskPriority;
  title?: string;
  details?: string;
  slaDueAt?: string;
}

// Request to assign/reassign a task (matches backend AssignTaskRequest)
export interface AssignTaskRequest {
  toOwnerId: string | null; // null to unassign
  reason?: string;
}

// Request to complete a task (matches backend CompleteTaskRequest)
export interface CompleteTaskRequest {
  completionNote?: string;
}

// List tasks response (matches backend ListTasksResponse)
export interface ListTasksResponse {
  data: Task[];
  total: number;
  limit: number;
  offset: number;
}

// Task list filter parameters
export interface TaskFilterParams {
  scopeType?: ScopeType;
  scopeId?: string;
  status?: TaskStatus;
  ownerId?: string;
  priority?: TaskPriority;
  overdue?: boolean;
  withOwnerDetails?: boolean;
  limit?: number;
  offset?: number;
}

// Helper function to check if task is overdue
export const isTaskOverdue = (task: Task): boolean => {
  if (!task.slaDueAt) return false;
  if (task.status === 'completed' || task.status === 'cancelled') return false;
  return new Date(task.slaDueAt) < new Date();
};

// Task status display labels
export const TaskStatusLabels: Record<TaskStatus, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
  escalated: 'Escalated',
};

// Task priority display labels
export const TaskPriorityLabels: Record<TaskPriority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
};

// Task status colors for UI
export const TaskStatusColors: Record<TaskStatus, string> = {
  open: 'blue',
  in_progress: 'yellow',
  completed: 'green',
  cancelled: 'gray',
  escalated: 'red',
};

// Task priority colors for UI
export const TaskPriorityColors: Record<TaskPriority, string> = {
  low: 'gray',
  medium: 'blue',
  high: 'orange',
  urgent: 'red',
};

// Scope type display labels
export const ScopeTypeLabels: Record<ScopeType, string> = {
  encounter: 'Encounter',
  patient: 'Patient',
  unit: 'Unit',
};
