/**
 * Tasks Page - Task board view
 * Displays kanban-style task board for clinical tasks
 */

import { useState, useEffect, useCallback } from 'react';
import { Layout } from '../shared/components/layout/Layout';
import { PageHeader } from '../shared/components/layout/PageHeader';
import { TaskBoard } from '../features/tasks/components/TaskBoard';
import { CreateTaskForm } from '../features/tasks/components/CreateTaskForm';
import { taskService } from '../features/tasks/services/taskService';
import type { Task, TaskFilterParams, CreateTaskRequest } from '../features/tasks/types';

export const TasksPage = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);

  const fetchTasks = useCallback(async (filters?: TaskFilterParams) => {
    try {
      setIsLoading(true);
      const response = await taskService.listTasks(filters);
      setTasks(response.data);
    } catch (error) {
      console.error('Failed to fetch tasks:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleTaskClick = (task: Task) => {
    console.log('Task clicked:', task);
    // TODO: Navigate to task detail page or open modal
  };

  const handleCreateTask = () => {
    setShowCreateForm(true);
  };

  const handleSubmitTask = async (data: CreateTaskRequest) => {
    await taskService.createTask(data);
    setShowCreateForm(false);
    fetchTasks(); // Refresh tasks
  };

  return (
    <Layout>
      <PageHeader
        title="Clinical Tasks"
        subtitle="Manage and track clinical tasks across your units"
      />
      
      <TaskBoard
        tasks={tasks}
        onTaskClick={handleTaskClick}
        onCreateTask={handleCreateTask}
        onFilterChange={fetchTasks}
        isLoading={isLoading}
      />

      <CreateTaskForm
        isOpen={showCreateForm}
        onClose={() => setShowCreateForm(false)}
        onSubmit={handleSubmitTask}
      />
    </Layout>
  );
};

export default TasksPage;
