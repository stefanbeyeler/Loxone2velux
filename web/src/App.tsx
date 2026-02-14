import { useState, useEffect } from 'react';
import { Dashboard } from './components/Dashboard';
import { TokenSetup } from './components/TokenSetup';
import * as api from './services/api';
import { RefreshCw } from 'lucide-react';

function App() {
  const [configured, setConfigured] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkSetup();
  }, []);

  const checkSetup = async () => {
    try {
      const health = await api.getHealth();
      // If KLF-200 is connected or config has a host, consider it configured
      if (health.connected) {
        setConfigured(true);
      } else {
        // Check if KLF200 host is set in config
        try {
          const config = await api.getConfig();
          setConfigured(config.klf200.host !== '' && config.klf200.password !== '');
        } catch {
          setConfigured(false);
        }
      }
    } catch {
      setConfigured(false);
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

  if (!configured) {
    return <TokenSetup onComplete={() => setConfigured(true)} />;
  }

  return <Dashboard />;
}

export default App;
