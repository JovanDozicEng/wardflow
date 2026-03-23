/**
 * App - Root application component
 * Sets up router and global providers
 */

import { RouterProvider } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { ErrorBoundary } from './shared/components/feedback';
import { router } from './lib/router';

function App() {
  // TODO: Load user on app mount
  // const { loadUser } = useAuth();
  // useEffect(() => {
  //   loadUser();
  // }, []);

  return (
    <ErrorBoundary>
      <RouterProvider router={router} />
      <Toaster position="top-right" />
    </ErrorBoundary>
  );
}

export default App;
