import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-4xl mx-auto">
          {/* Header */}
          <header className="text-center mb-12">
            <h1 className="text-5xl font-bold text-gray-900 dark:text-white mb-4">
              WardFlow
            </h1>
            <p className="text-xl text-gray-600 dark:text-gray-300">
              Inpatient/ED Care Coordination System
            </p>
          </header>

          {/* Main Card */}
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 mb-8">
            <div className="text-center">
              <div className="mb-8">
                <span className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-300 text-4xl font-bold mb-4">
                  {count}
                </span>
              </div>
              
              <button
                onClick={() => setCount((count) => count + 1)}
                className="px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200 transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
              >
                Click to Increment
              </button>
            </div>
          </div>

          {/* Info Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-3">
                React 18+
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Modern functional components with hooks
              </p>
            </div>
            
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-3">
                Tailwind CSS
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Utility-first CSS framework for rapid UI development
              </p>
            </div>
            
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-3">
                TypeScript
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Type-safe code with excellent IDE support
              </p>
            </div>
            
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-900 dark:text-white mb-3">
                Vite
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Lightning-fast development server and build tool
              </p>
            </div>
          </div>

          {/* Footer */}
          <footer className="text-center mt-12 text-gray-500 dark:text-gray-400">
            <p>Edit <code className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm">src/App.tsx</code> to get started</p>
          </footer>
        </div>
      </div>
    </div>
  )
}

export default App
