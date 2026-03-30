/**
 * Axios configuration and API client setup
 * Central place for HTTP requests with interceptors
 */

import axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios';

// API base URL from environment or default
// All routes are under /api/v1 (auth: /api/v1/auth/*, resources: /api/v1/*)
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

// Create axios instance with defaults
export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: Add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // TODO: Get token from auth store
    const token = localStorage.getItem('auth_token');
    
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor: Handle errors globally
api.interceptors.response.use(
  (response) => {
    // Return data directly for successful responses
    return response;
  },
  (error: AxiosError) => {
    // TODO: Add toast notifications for errors
    
    if (error.response) {
      // Server responded with error status
      const status = error.response.status;
      
      switch (status) {
        case 401:
          // Unauthorized - clear token and redirect to login
          // TODO: Integrate with auth store
          localStorage.removeItem('auth_token');
          window.location.href = '/login';
          break;
        case 403:
          // Forbidden - user doesn't have permission
          console.error('Access forbidden:', error.response.data);
          break;
        case 404:
          // Not found
          console.error('Resource not found:', error.config?.url);
          break;
        case 500:
          // Server error
          console.error('Server error:', error.response.data);
          break;
        default:
          console.error('API error:', error.response.data);
      }
    } else if (error.request) {
      // Request made but no response received
      console.error('No response from server:', error.message);
    } else {
      // Error in request setup
      console.error('Request setup error:', error.message);
    }
    
    return Promise.reject(error);
  }
);

// Helper functions for common request patterns
export const apiHelpers = {
  // GET request with type safety
  get: <T>(url: string, params?: Record<string, any>) => {
    return api.get<T>(url, { params });
  },
  
  // POST request with type safety
  post: <T>(url: string, data?: any) => {
    return api.post<T>(url, data);
  },
  
  // PUT request with type safety
  put: <T>(url: string, data?: any) => {
    return api.put<T>(url, data);
  },
  
  // PATCH request with type safety
  patch: <T>(url: string, data?: any) => {
    return api.patch<T>(url, data);
  },
  
  // DELETE request with type safety
  delete: <T>(url: string) => {
    return api.delete<T>(url);
  },
};

export default api;
