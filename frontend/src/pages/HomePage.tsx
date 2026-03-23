/**
 * Home Page - Landing page with login/register options
 */

import { Link } from 'react-router-dom';
import { ROUTES } from '../shared/config/routes';
import { useAuth } from '../features/auth/hooks/useAuth';

export const HomePage = () => {
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    // Redirect to dashboard if already logged in
    window.location.href = ROUTES.DASHBOARD;
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-4xl mx-auto">
          {/* Header */}
          <header className="text-center mb-12">
            <h1 className="text-6xl font-bold text-indigo-600 dark:text-indigo-400 mb-4">
              WardFlow
            </h1>
            <p className="text-2xl text-gray-700 dark:text-gray-300 mb-2">
              Inpatient/ED Care Coordination System
            </p>
            <p className="text-lg text-gray-600 dark:text-gray-400">
              Streamline handoffs, track patient flow, and improve care coordination
            </p>
          </header>

          {/* Call to Action */}
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 mb-8">
            <div className="text-center">
              <h2 className="text-3xl font-semibold text-gray-900 dark:text-white mb-6">
                Get Started
              </h2>
              <p className="text-gray-600 dark:text-gray-300 mb-8">
                Sign in to manage care teams, coordinate tasks, and track patient flow
              </p>
              
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Link
                  to={ROUTES.LOGIN}
                  className="px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200 transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                >
                  Sign In
                </Link>
                <Link
                  to={ROUTES.REGISTER}
                  className="px-8 py-4 bg-white hover:bg-gray-50 text-indigo-600 font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200 border-2 border-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                >
                  Create Account
                </Link>
              </div>
            </div>
          </div>

          {/* Features Grid */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <div className="text-indigo-600 dark:text-indigo-400 text-3xl mb-3">👥</div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Care Team Management
              </h3>
              <p className="text-gray-600 dark:text-gray-300">
                Clear role assignments and structured handoffs
              </p>
            </div>
            
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <div className="text-indigo-600 dark:text-indigo-400 text-3xl mb-3">📋</div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Task Coordination
              </h3>
              <p className="text-gray-600 dark:text-gray-300">
                Track and manage clinical tasks with SLA monitoring
              </p>
            </div>
            
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <div className="text-indigo-600 dark:text-indigo-400 text-3xl mb-3">🏥</div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                Patient Flow
              </h3>
              <p className="text-gray-600 dark:text-gray-300">
                Real-time visibility into patient journey
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage;
