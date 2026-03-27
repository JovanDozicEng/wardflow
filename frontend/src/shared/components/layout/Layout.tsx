/**
 * Layout component - Main application layout wrapper
 * TODO: Implement with Header and Sidebar
 * TODO: Add responsive mobile layout
 * TODO: Add sidebar toggle for mobile
 */

import type { ReactNode } from 'react';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

interface LayoutProps {
  children: ReactNode;
}

export const Layout = ({ children }: LayoutProps) => {
  // TODO: Add sidebar collapse state
  // TODO: Add mobile menu toggle
  
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          {children}
        </main>
      </div>
    </div>
  );
};
