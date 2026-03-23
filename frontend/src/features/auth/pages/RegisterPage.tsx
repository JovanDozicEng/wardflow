/**
 * RegisterPage component
 * TODO: Implement with React Hook Form + Zod validation
 * TODO: Add email, password, name, role inputs
 * TODO: Add password confirmation
 * TODO: Add link to login page
 * TODO: Show success message
 * TODO: Redirect to dashboard after registration
 */

export const RegisterPage = () => {
  // TODO: Use useAuth hook
  // TODO: Setup React Hook Form with Zod schema
  // TODO: Handle form submission
  // TODO: Redirect on successful registration
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-gray-900">Create your account</h2>
          <p className="mt-2 text-sm text-gray-600">
            Join WardFlow for better care coordination
          </p>
        </div>
        
        {/* TODO: Implement RegisterForm component here */}
        <div className="mt-8 space-y-6">
          <p className="text-center text-gray-500">Registration form coming soon...</p>
        </div>
      </div>
    </div>
  );
};

export default RegisterPage;
