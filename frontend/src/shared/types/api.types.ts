/**
 * API-related types for HTTP requests and responses
 */

// Standard API response wrapper
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
}

// Standard error response from backend
export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: FieldError[];
    correlationId?: string;
  };
}

// Field-level validation errors
export interface FieldError {
  field: string;
  issue: string;
}

// Paginated response
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// Cursor-based pagination (preferred for future)
export interface CursorPaginatedResponse<T> {
  data: T[];
  pageInfo: {
    hasNextPage: boolean;
    hasPreviousPage: boolean;
    startCursor?: string;
    endCursor?: string;
  };
}

// Common query parameters
export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface CursorPaginationParams {
  limit?: number;
  cursor?: string;
}

// Filtering parameters (common across modules)
export interface FilterParams {
  unitId?: string;
  departmentId?: string;
  encounterId?: string;
  status?: string;
  startDate?: string; // ISO timestamp
  endDate?: string; // ISO timestamp
}
