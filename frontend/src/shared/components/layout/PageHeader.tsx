/**
 * PageHeader component - Page title with breadcrumbs and actions
 * TODO: Implement breadcrumb navigation
 * TODO: Add action buttons slot
 * TODO: Add subtitle/description
 */

import { Link } from 'react-router-dom';
import type { ReactNode } from 'react';

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  breadcrumbs?: { label: string; href?: string }[];
  actions?: ReactNode;
}

export const PageHeader = ({
  title,
  subtitle,
  breadcrumbs,
  actions,
}: PageHeaderProps) => {
  // TODO: Implement breadcrumb navigation
  // TODO: Style breadcrumbs with separators
  
  return (
    <div className="mb-6">
      {breadcrumbs && breadcrumbs.length > 0 && (
        <nav className="mb-2 flex items-center space-x-2 text-sm text-gray-500">
          {/* TODO: Render breadcrumbs with separators */}
          {breadcrumbs.map((crumb, index) => (
            <span key={index}>
              {crumb.href ? (
                <Link to={crumb.href} className="hover:text-gray-700">
                  {crumb.label}
                </Link>
              ) : (
                <span>{crumb.label}</span>
              )}
              {index < breadcrumbs.length - 1 && <span className="mx-2">/</span>}
            </span>
          ))}
        </nav>
      )}
      
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
          {subtitle && (
            <p className="mt-1 text-sm text-gray-600">{subtitle}</p>
          )}
        </div>
        
        {actions && (
          <div className="flex items-center space-x-2">
            {actions}
          </div>
        )}
      </div>
    </div>
  );
};
