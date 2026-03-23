/**
 * LoginPage component
 * TODO: Implement with React Hook Form + Zod validation
 * TODO: Add email/password inputs
 * TODO: Add remember me checkbox
 * TODO: Add link to register page
 * TODO: Show error messages
 * TODO: Redirect to dashboard after login
 */

export const LoginPage = () => {
  // TODO: Use useAuth hook
  // TODO: Setup React Hook Form with Zod schema
  // TODO: Handle form submission
  // TODO: Redirect on successful login
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-gray-900">Sign in to WardFlow</h2>
          <p className="mt-2 text-sm text-gray-600">
            Care coordination made simple
          </p>
        </div>
        
        {/* TODO: Implement LoginForm component here */}
        <div className="mt-8 space-y-6">
          <p className="text-center text-gray-500">Login form coming soon...</p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
