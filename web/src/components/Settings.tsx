import { useState, useEffect } from 'react';
import { GatewayConfig } from '../types';
import * as api from '../services/api';
import {
  Save,
  RefreshCw,
  Server,
  Shield,
  FileText,
  Eye,
  EyeOff,
  CheckCircle,
  AlertCircle,
} from 'lucide-react';

interface SettingsProps {
  onConfigChange?: () => void;
}

export function Settings({ onConfigChange }: SettingsProps) {
  const [config, setConfig] = useState<GatewayConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [reconnecting, setReconnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showToken, setShowToken] = useState(false);

  // Form state
  const [formData, setFormData] = useState({
    klf200Host: '',
    klf200Port: 51200,
    klf200Password: '',
    apiToken: '',
    logLevel: 'info',
  });

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    setLoading(true);
    setError(null);
    try {
      const cfg = await api.getConfig();
      setConfig(cfg);
      setFormData({
        klf200Host: cfg.klf200.host,
        klf200Port: cfg.klf200.port,
        klf200Password: cfg.klf200.password,
        apiToken: cfg.server.api_token,
        logLevel: cfg.logging.level,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler beim Laden der Konfiguration');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    setSuccess(null);

    try {
      const updatedConfig: Partial<GatewayConfig> = {
        klf200: {
          host: formData.klf200Host,
          port: formData.klf200Port,
          password: formData.klf200Password,
          reconnect_interval: config?.klf200.reconnect_interval || '30s',
          refresh_interval: config?.klf200.refresh_interval || '5m',
        },
        server: {
          host: config?.server.host || '0.0.0.0',
          port: config?.server.port || 8080,
          api_token: formData.apiToken,
        },
        logging: {
          level: formData.logLevel,
          format: config?.logging.format || 'console',
        },
      };

      const newConfig = await api.updateConfig(updatedConfig);
      setConfig(newConfig);

      // Update local storage token if changed
      if (formData.apiToken) {
        localStorage.setItem('api_token', formData.apiToken);
      } else {
        localStorage.removeItem('api_token');
      }

      setSuccess('Konfiguration gespeichert');
      onConfigChange?.();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler beim Speichern');
    } finally {
      setSaving(false);
    }
  };

  const handleReconnect = async () => {
    setReconnecting(true);
    setError(null);
    setSuccess(null);

    try {
      const result = await api.reconnectGateway();
      if (result.success) {
        setSuccess('Verbindung zum KLF-200 wurde neu hergestellt');
      } else {
        setError(result.message || 'Verbindung fehlgeschlagen');
      }
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler beim Neuverbinden');
    } finally {
      setReconnecting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <RefreshCw className="animate-spin text-velux-blue" size={32} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-gray-800 rounded-xl p-6">
        <h2 className="text-xl font-bold text-white mb-2">Einstellungen</h2>
        <p className="text-gray-400">
          Konfiguration des Loxone2Velux Gateways
        </p>
      </div>

      {/* Status Messages */}
      {error && (
        <div className="bg-red-900/30 border border-red-700 rounded-lg p-4 flex items-center gap-3">
          <AlertCircle className="text-red-400 flex-shrink-0" size={20} />
          <p className="text-red-400">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-900/30 border border-green-700 rounded-lg p-4 flex items-center gap-3">
          <CheckCircle className="text-green-400 flex-shrink-0" size={20} />
          <p className="text-green-400">{success}</p>
        </div>
      )}

      {/* KLF-200 Settings */}
      <div className="bg-gray-800 rounded-xl p-6">
        <div className="flex items-center gap-3 mb-4">
          <Server className="text-velux-blue" size={24} />
          <h3 className="text-lg font-semibold text-white">KLF-200 Verbindung</h3>
        </div>

        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                IP-Adresse / Hostname
              </label>
              <input
                type="text"
                value={formData.klf200Host}
                onChange={(e) => setFormData({ ...formData, klf200Host: e.target.value })}
                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:border-velux-blue focus:outline-none"
                placeholder="192.168.1.100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Port
              </label>
              <input
                type="number"
                value={formData.klf200Port}
                onChange={(e) => setFormData({ ...formData, klf200Port: parseInt(e.target.value) || 51200 })}
                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:border-velux-blue focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">
              Passwort
            </label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                value={formData.klf200Password}
                onChange={(e) => setFormData({ ...formData, klf200Password: e.target.value })}
                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:border-velux-blue focus:outline-none pr-12"
                placeholder="KLF-200 Passwort"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
              >
                {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              Das Passwort befindet sich auf der Rückseite des KLF-200
            </p>
          </div>

          <button
            onClick={handleReconnect}
            disabled={reconnecting}
            className="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw size={18} className={reconnecting ? 'animate-spin' : ''} />
            {reconnecting ? 'Verbinde...' : 'Neu verbinden'}
          </button>
        </div>
      </div>

      {/* Security Settings */}
      <div className="bg-gray-800 rounded-xl p-6">
        <div className="flex items-center gap-3 mb-4">
          <Shield className="text-velux-blue" size={24} />
          <h3 className="text-lg font-semibold text-white">Sicherheit</h3>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-2">
              API Token (optional)
            </label>
            <div className="relative">
              <input
                type={showToken ? 'text' : 'password'}
                value={formData.apiToken}
                onChange={(e) => setFormData({ ...formData, apiToken: e.target.value })}
                className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:border-velux-blue focus:outline-none pr-12 font-mono"
                placeholder="Leer lassen um Auth zu deaktivieren"
              />
              <button
                type="button"
                onClick={() => setShowToken(!showToken)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
              >
                {showToken ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              {formData.apiToken
                ? 'Authentifizierung ist aktiviert'
                : 'Authentifizierung ist deaktiviert (API ist offen zugänglich)'}
            </p>
          </div>
        </div>
      </div>

      {/* Logging Settings */}
      <div className="bg-gray-800 rounded-xl p-6">
        <div className="flex items-center gap-3 mb-4">
          <FileText className="text-velux-blue" size={24} />
          <h3 className="text-lg font-semibold text-white">Logging</h3>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-400 mb-2">
            Log-Level
          </label>
          <select
            value={formData.logLevel}
            onChange={(e) => setFormData({ ...formData, logLevel: e.target.value })}
            className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:border-velux-blue focus:outline-none"
          >
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warn">Warnung</option>
            <option value="error">Fehler</option>
          </select>
        </div>
      </div>

      {/* Save Button */}
      <div className="flex justify-end">
        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 px-6 py-3 bg-velux-blue hover:bg-velux-blue/80 text-white rounded-lg transition-colors disabled:opacity-50 font-medium"
        >
          <Save size={20} />
          {saving ? 'Speichern...' : 'Einstellungen speichern'}
        </button>
      </div>

      {/* Current Config Info */}
      {config && (
        <div className="bg-gray-800/50 rounded-xl p-4 text-xs text-gray-500">
          <p>Aktuelle Konfiguration geladen von: config.yaml</p>
          <p>Server läuft auf: {config.server.host}:{config.server.port}</p>
        </div>
      )}
    </div>
  );
}
