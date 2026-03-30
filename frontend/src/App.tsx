/**
 * App - Root application component
 * Sets up router and global providers
 */

import { useEffect } from 'react';
import { RouterProvider } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { ErrorBoundary } from './shared/components/feedback';
import { router } from './lib/router';
import { useAuthStore } from './features/auth/store/authStore';

function App() {
  const { loadUser, isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      loadUser();
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <ErrorBoundary>
      <RouterProvider router={router} />
      <Toaster position="top-right" />
    </ErrorBoundary>
  );
}

export default App;
