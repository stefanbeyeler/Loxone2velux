import { useState, useEffect } from 'react';
import { Dashboard } from './components/Dashboard';
import { TokenSetup } from './components/TokenSetup';
import * as api from './services/api';
import { RefreshCw } from 'lucide-react';

function App() {
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      // Check if we have a stored token and if it's valid
      const token = localStorage.getItem('api_token');
      if (!token) {
        setAuthenticated(false);
        setLoading(false);
        return;
      }

      // Try to fetch nodes to verify token
      await api.getNodes();
      setAuthenticated(true);
    } catch (error) {
      console.error('Auth check failed:', error);
      setAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <RefreshCw className="animate-spin text-velux-blue" size={48} />
      </div>
    );
  }

  if (!authenticated) {
    return <TokenSetup onComplete={() => setAuthenticated(true)} />;
  }

  return <Dashboard onLogout={() => setAuthenticated(false)} />;
}

export default App;
