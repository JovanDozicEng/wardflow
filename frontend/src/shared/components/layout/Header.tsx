/**
 * Header component - Top navigation bar
 * TODO: Implement with logo, user menu, notifications
 * TODO: Add mobile hamburger menu toggle
 * TODO: Show user name and role
 * TODO: Add logout button in dropdown
 */

export const Header = () => {
  // TODO: Get user from auth store
  // TODO: Implement user dropdown menu
  // TODO: Add notifications icon
  // TODO: Add logout handler
  
  return (
    <header className="bg-white border-b border-gray-200 h-16 flex items-center px-4 lg:px-6">
      <div className="flex items-center justify-between w-full">
        {/* Logo */}
        <div className="flex items-center space-x-4">
          <h1 className="text-xl font-bold text-indigo-600">WardFlow</h1>
        </div>
        
        {/* User menu */}
        <div className="flex items-center space-x-4">
          <div className="text-sm text-gray-700">
            {/* TODO: Display user name and role */}
            <span className="font-medium">User Name</span>
            <span className="text-gray-500 ml-2">Role</span>
          </div>
          {/* TODO: Add dropdown menu with logout */}
        </div>
      </div>
    </header>
  );
};
