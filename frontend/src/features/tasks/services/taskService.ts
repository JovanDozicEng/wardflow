/**
 * Task Board API service
 * Handles all clinical task management API calls
 */

import api from '../../../shared/utils/api';
import type {
  Task,
  TaskAssignmentEvent,
  CreateTaskRequest,
  UpdateTaskRequest,
  AssignTaskRequest,
  CompleteTaskRequest,
  ListTasksResponse,
  TaskFilterParams,
} from '../types';

export const taskService = {
  /**
   * List tasks with filtering
   * @param filters - Optional filter parameters
   */
  listTasks: async (filters?: TaskFilterParams): Promise<ListTasksResponse> => {
    const response = await api.get<ListTasksResponse>('/tasks', {
      params: filters,
    });
    return response.data;
  },

  /**
   * Get a single task by ID
   * @param taskId - The task ID
   */
  getTask: async (taskId: string): Promise<Task> => {
    const response = await api.get<Task>(`/tasks/${taskId}`);
    return response.data;
  },

  /**
   * Create a new task
   * @param data - Task creation request
   */
  createTask: async (data: CreateTaskRequest): Promise<Task> => {
    const response = await api.post<Task>('/tasks', data);
    return response.data;
  },

  /**
   * Update an existing task
   * @param taskId - The task ID
   * @param data - Fields to update
   */
  updateTask: async (taskId: string, data: UpdateTaskRequest): Promise<Task> => {
    const response = await api.patch<Task>(`/tasks/${taskId}`, data);
    return response.data;
  },

  /**
   * Assign or reassign a task
   * @param taskId - The task ID
   * @param data - Assignment request (toOwnerId can be null to unassign)
   */
  assignTask: async (taskId: string, data: AssignTaskRequest): Promise<Task> => {
    const response = await api.post<Task>(`/tasks/${taskId}/assign`, data);
    return response.data;
  },

  /**
   * Complete a task
   * @param taskId - The task ID
   * @param data - Optional completion note
   */
  completeTask: async (taskId: string, data?: CompleteTaskRequest): Promise<Task> => {
    const response = await api.post<Task>(`/tasks/${taskId}/complete`, data || {});
    return response.data;
  },

  /**
   * Get task assignment history
   * @param taskId - The task ID
   */
  getTaskHistory: async (taskId: string): Promise<TaskAssignmentEvent[]> => {
    const response = await api.get<{ data: TaskAssignmentEvent[] }>(`/tasks/${taskId}/history`);
    return response.data.data;
  },
};
