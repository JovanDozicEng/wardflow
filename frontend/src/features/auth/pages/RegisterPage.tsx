/**
 * RegisterPage component
 * Full registration page with form and navigation
 */

import { Link } from 'react-router-dom';
import { RegisterForm } from '../components/RegisterForm';
import { ROUTES } from '../../../shared/config/routes';

export const RegisterPage = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12">
      <div className="max-w-md w-full space-y-8 p-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-indigo-600 mb-2">WardFlow</h1>
          <h2 className="text-2xl font-bold text-gray-900">Create your account</h2>
          <p className="mt-2 text-sm text-gray-600">
            Join WardFlow for better care coordination
          </p>
        </div>
        
        <div className="mt-8 bg-white shadow-lg rounded-lg p-6">
          <RegisterForm />
          
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Already have an account?{' '}
              <Link
                to={ROUTES.LOGIN}
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Sign in
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RegisterPage;
