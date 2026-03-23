/**
 * Date and time formatting utilities
 * Uses date-fns for consistent formatting across the app
 */

import { format, formatDistance, formatRelative, parseISO } from 'date-fns';

/**
 * Format ISO timestamp to readable date
 * @param isoString - ISO 8601 timestamp from API
 * @returns Formatted date string (e.g., "Mar 23, 2026")
 */
export const formatDate = (isoString: string): string => {
  try {
    return format(parseISO(isoString), 'MMM dd, yyyy');
  } catch (error) {
    console.error('Invalid date:', isoString);
    return 'Invalid date';
  }
};

/**
 * Format ISO timestamp to readable date and time
 * @param isoString - ISO 8601 timestamp from API
 * @returns Formatted datetime string (e.g., "Mar 23, 2026 at 2:45 PM")
 */
export const formatDateTime = (isoString: string): string => {
  try {
    return format(parseISO(isoString), 'MMM dd, yyyy \'at\' h:mm a');
  } catch (error) {
    console.error('Invalid datetime:', isoString);
    return 'Invalid datetime';
  }
};

/**
 * Format ISO timestamp to time only
 * @param isoString - ISO 8601 timestamp from API
 * @returns Formatted time string (e.g., "2:45 PM")
 */
export const formatTime = (isoString: string): string => {
  try {
    return format(parseISO(isoString), 'h:mm a');
  } catch (error) {
    console.error('Invalid time:', isoString);
    return 'Invalid time';
  }
};

/**
 * Format ISO timestamp relative to now
 * @param isoString - ISO 8601 timestamp from API
 * @returns Relative time string (e.g., "2 hours ago", "in 3 days")
 */
export const formatRelativeTime = (isoString: string): string => {
  try {
    return formatDistance(parseISO(isoString), new Date(), { addSuffix: true });
  } catch (error) {
    console.error('Invalid date for relative time:', isoString);
    return 'Unknown time';
  }
};

/**
 * Format ISO timestamp relative to now with more context
 * @param isoString - ISO 8601 timestamp from API
 * @returns Contextual relative string (e.g., "yesterday at 3:00 PM", "tomorrow at 9:00 AM")
 */
export const formatRelativeDateTime = (isoString: string): string => {
  try {
    return formatRelative(parseISO(isoString), new Date());
  } catch (error) {
    console.error('Invalid date for relative datetime:', isoString);
    return 'Unknown time';
  }
};

/**
 * Format duration in milliseconds to readable string
 * @param ms - Duration in milliseconds
 * @returns Formatted duration (e.g., "2h 30m", "45m", "1d 3h")
 */
export const formatDuration = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ${hours % 24}h`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m`;
  }
  return `${seconds}s`;
};

/**
 * Truncate string with ellipsis
 * @param str - String to truncate
 * @param maxLength - Maximum length before truncation
 * @returns Truncated string with ellipsis if needed
 */
export const truncate = (str: string, maxLength: number): string => {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - 3) + '...';
};

/**
 * Capitalize first letter of string
 * @param str - String to capitalize
 * @returns Capitalized string
 */
export const capitalize = (str: string): string => {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
};

/**
 * Convert snake_case to Title Case
 * @param str - Snake case string
 * @returns Title case string (e.g., "charge_nurse" -> "Charge Nurse")
 */
export const snakeToTitle = (str: string): string => {
  return str
    .split('_')
    .map(word => capitalize(word))
    .join(' ');
};
