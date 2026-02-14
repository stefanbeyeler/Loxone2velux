import { useState, useEffect } from 'react';
import { Key, AlertCircle, ArrowRight } from 'lucide-react';
import * as api from '../services/api';

interface TokenSetupProps {
  onComplete: () => void;
}

export function TokenSetup({ onComplete }: TokenSetupProps) {
  const [token, setToken] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [version, setVersion] = useState('');

  useEffect(() => {
    api.getHealth().then(h => setVersion(h.version)).catch(() => {});
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      // Store token temporarily
      localStorage.setItem('api_token', token);

      // Test the token
      await api.getNodes();

      // Success
      onComplete();
    } catch (err) {
      localStorage.removeItem('api_token');
      setError(err instanceof Error ? err.message : 'Verbindung fehlgeschlagen');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        <div className="bg-gray-800 rounded-2xl p-8 shadow-xl">
          {/* Header */}
          <div className="text-center mb-8">
            <div className="w-16 h-16 bg-velux-blue rounded-2xl flex items-center justify-center mx-auto mb-4">
              <Key size={32} className="text-white" />
            </div>
            <h1 className="text-2xl font-bold text-white mb-2">Loxone2Velux</h1>
            <p className="text-gray-400">Gateway Authentifizierung</p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="token" className="block text-sm font-medium text-gray-300 mb-2">
                API Token
              </label>
              <input
                type="password"
                id="token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="Token aus config.yaml"
                className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-velux-blue focus:border-transparent"
                required
              />
              <p className="mt-2 text-xs text-gray-500">
                Das Token findest du in der config.yaml unter server.api_token
              </p>
            </div>

            {error && (
              <div className="flex items-center gap-2 p-3 bg-red-900/30 border border-red-500/50 rounded-lg">
                <AlertCircle size={18} className="text-red-400 flex-shrink-0" />
                <p className="text-sm text-red-400">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !token}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-velux-blue hover:bg-velux-dark text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <>
                  <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Verbinden...
                </>
              ) : (
                <>
                  Verbinden
                  <ArrowRight size={18} />
                </>
              )}
            </button>
          </form>
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-gray-600 mt-6">
          Loxone2Velux Gateway {version && `v${version}`} &copy; 2026
        </p>
      </div>
    </div>
  );
}
