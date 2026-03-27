/**
 * cn (classNames) utility
 * Simple utility for conditionally joining CSS class names
 * Lightweight alternative to clsx/classnames
 */

type ClassValue = string | number | boolean | undefined | null;
type ClassArray = ClassValue[];
type ClassObject = Record<string, boolean | undefined | null>;

/**
 * Combines class names conditionally
 * @param classes - Classes to combine (strings, objects, arrays)
 * @returns Combined class string
 * 
 * @example
 * cn('base', isActive && 'active', { disabled: isDisabled })
 * // Returns: 'base active' if isActive=true, isDisabled=false
 */
export function cn(...classes: (ClassValue | ClassObject | ClassArray)[]): string {
  const result: string[] = [];

  for (const cls of classes) {
    if (!cls) continue;

    if (typeof cls === 'string' || typeof cls === 'number') {
      result.push(String(cls));
    } else if (Array.isArray(cls)) {
      const nested = cn(...cls);
      if (nested) result.push(nested);
    } else if (typeof cls === 'object') {
      for (const [key, value] of Object.entries(cls)) {
        if (value) result.push(key);
      }
    }
  }

  return result.join(' ');
}
