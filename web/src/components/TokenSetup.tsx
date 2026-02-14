import { useState, useEffect } from 'react';
import { Blinds, AlertCircle, ArrowRight } from 'lucide-react';
import * as api from '../services/api';

interface TokenSetupProps {
  onComplete: () => void;
}

export function TokenSetup({ onComplete }: TokenSetupProps) {
  const [host, setHost] = useState('');
  const [port, setPort] = useState('51200');
  const [password, setPassword] = useState('');
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
      await api.updateConfig({
        klf200: {
          host,
          port: parseInt(port, 10),
          password,
          reconnect_interval: '30s',
          refresh_interval: '300s',
        },
      });
      onComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Konfiguration fehlgeschlagen');
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
              <Blinds size={32} className="text-white" />
            </div>
            <h1 className="text-2xl font-bold text-white mb-2">Loxone2Velux</h1>
            <p className="text-gray-400">KLF-200 Erstkonfiguration</p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="host" className="block text-sm font-medium text-gray-300 mb-2">
                IP-Adresse / Hostname
              </label>
              <input
                type="text"
                id="host"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="192.168.1.100"
                className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-velux-blue focus:border-transparent"
                required
              />
            </div>

            <div>
              <label htmlFor="port" className="block text-sm font-medium text-gray-300 mb-2">
                Port
              </label>
              <input
                type="number"
                id="port"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                placeholder="51200"
                className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-velux-blue focus:border-transparent"
                required
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-2">
                Passwort
              </label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="KLF-200 Passwort"
                className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-velux-blue focus:border-transparent"
                required
              />
            </div>

            {error && (
              <div className="flex items-center gap-2 p-3 bg-red-900/30 border border-red-500/50 rounded-lg">
                <AlertCircle size={18} className="text-red-400 flex-shrink-0" />
                <p className="text-sm text-red-400">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !host || !password}
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
