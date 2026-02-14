import { useState, useEffect, useCallback } from 'react';
import { NodeList } from './NodeList';
import { LoxoneGuide } from './LoxoneGuide';
import { SensorCard } from './SensorCard';
import { Settings } from './Settings';
import * as api from '../services/api';
import { Node, HealthResponse, SensorStatus } from '../types';
import {
  Blinds,
  BookOpen,
  RefreshCw,
  Wifi,
  WifiOff,
  Menu,
  X,
  LogOut,
  Server,
  Settings as SettingsIcon,
} from 'lucide-react';


type Tab = 'devices' | 'guide' | 'settings';

interface DashboardProps {
  onLogout?: () => void;
}

export function Dashboard({ onLogout }: DashboardProps) {
  const [activeTab, setActiveTab] = useState<Tab>('devices');
  const [menuOpen, setMenuOpen] = useState(false);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [sensors, setSensors] = useState<SensorStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [sensorsRefreshing, setSensorsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const [healthData, nodesData, sensorData] = await Promise.all([
        api.getHealth(),
        api.getNodes().catch(() => ({ nodes: [], count: 0 })),
        api.getSensorStatus().catch(() => null),
      ]);

      setHealth(healthData);
      setNodes(nodesData.nodes || []);
      if (sensorData) setSensors(sensorData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler beim Laden');
    } finally {
      setLoading(false);
    }
  }, []);

  const handleRefreshSensors = async () => {
    setSensorsRefreshing(true);
    try {
      const sensorData = await api.refreshSensorStatus();
      setSensors(sensorData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sensor-Abfrage fehlgeschlagen');
    } finally {
      setSensorsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Refresh every 10s
    return () => clearInterval(interval);
  }, [fetchData]);

  const handleSetPosition = async (id: number, position: number) => {
    try {
      await api.setNodePosition(id, position);
      // Optimistic update
      setNodes(nodes.map(n =>
        n.id === id ? { ...n, position_percent: position } : n
      ));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler');
    }
  };

  const handleOpen = async (id: number) => {
    try {
      await api.openNode(id);
      setNodes(nodes.map(n =>
        n.id === id ? { ...n, position_percent: 0 } : n
      ));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler');
    }
  };

  const handleClose = async (id: number) => {
    try {
      await api.closeNode(id);
      setNodes(nodes.map(n =>
        n.id === id ? { ...n, position_percent: 100 } : n
      ));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler');
    }
  };

  const handleStop = async (id: number) => {
    try {
      await api.stopNode(id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Fehler');
    }
  };

  const handleLogout = () => {
    api.clearApiToken();
    onLogout?.();
  };

  const isConnected = health?.connected ?? false;

  const tabs = [
    { id: 'devices' as Tab, label: 'Geräte', icon: Blinds, count: nodes.length },
    { id: 'guide' as Tab, label: 'Loxone Anleitung', icon: BookOpen },
    { id: 'settings' as Tab, label: 'Einstellungen', icon: SettingsIcon },
  ];

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700 sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-velux-blue rounded-lg flex items-center justify-center">
                <Blinds size={24} className="text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">Loxone2Velux</h1>
                <p className="text-xs text-gray-400">Gateway Service</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              {/* Gateway Status */}
              <div className="hidden md:flex items-center gap-2 bg-gray-700/50 px-3 py-1.5 rounded-lg">
                <Server size={16} className="text-velux-blue" />
                <span className="text-sm text-gray-300">KLF-200</span>
              </div>

              {/* Connection Status */}
              <div className="hidden sm:flex items-center gap-2">
                {isConnected ? (
                  <Wifi size={18} className="text-green-500" />
                ) : (
                  <WifiOff size={18} className="text-red-500" />
                )}
                <span className="text-sm text-gray-400">
                  {isConnected ? 'Verbunden' : 'Getrennt'}
                </span>
              </div>

              {/* Refresh Button */}
              <button
                onClick={fetchData}
                disabled={loading}
                className="p-2 text-gray-400 hover:text-white transition-colors"
                title="Aktualisieren"
              >
                <RefreshCw size={20} className={loading ? 'animate-spin' : ''} />
              </button>

              {/* Logout Button - only show if auth is enabled */}
              {onLogout && (
                <button
                  onClick={handleLogout}
                  className="hidden sm:flex p-2 text-gray-400 hover:text-red-400 transition-colors"
                  title="Abmelden"
                >
                  <LogOut size={20} />
                </button>
              )}

              {/* Mobile Menu Toggle */}
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="sm:hidden p-2 text-gray-400"
              >
                {menuOpen ? <X size={24} /> : <Menu size={24} />}
              </button>
            </div>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden sm:flex gap-1 mt-4">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`
                  flex items-center gap-2 px-4 py-2 rounded-lg transition-colors
                  ${activeTab === tab.id
                    ? 'bg-velux-blue text-white'
                    : 'text-gray-400 hover:text-white hover:bg-gray-700'
                  }
                `}
              >
                <tab.icon size={18} />
                {tab.label}
                {tab.count !== undefined && (
                  <span className={`
                    text-xs px-2 py-0.5 rounded-full
                    ${activeTab === tab.id ? 'bg-white/20' : 'bg-gray-700'}
                  `}>
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </nav>

          {/* Mobile Navigation */}
          {menuOpen && (
            <nav className="sm:hidden flex flex-col gap-1 mt-4 pb-2">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => {
                    setActiveTab(tab.id);
                    setMenuOpen(false);
                  }}
                  className={`
                    flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
                    ${activeTab === tab.id
                      ? 'bg-velux-blue text-white'
                      : 'text-gray-400 hover:bg-gray-700'
                    }
                  `}
                >
                  <tab.icon size={20} />
                  {tab.label}
                  {tab.count !== undefined && (
                    <span className="text-xs bg-gray-700 px-2 py-0.5 rounded-full ml-auto">
                      {tab.count}
                    </span>
                  )}
                </button>
              ))}
              {onLogout && (
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-3 px-4 py-3 rounded-lg text-red-400 hover:bg-gray-700"
                >
                  <LogOut size={20} />
                  Abmelden
                </button>
              )}
            </nav>
          )}
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="bg-red-900/30 border-b border-red-500 px-4 py-3">
          <div className="max-w-6xl mx-auto">
            <p className="text-red-400 text-sm">{error}</p>
          </div>
        </div>
      )}

      {/* Main Content */}
      <main className="max-w-6xl mx-auto px-4 py-6">
        {activeTab === 'devices' && (
          <div className="space-y-6">
            {/* Sensor Status */}
            <SensorCard
              sensors={sensors}
              onRefresh={handleRefreshSensors}
              isRefreshing={sensorsRefreshing}
            />

            {/* Node List */}
            <NodeList
              nodes={nodes}
              loading={loading}
              onSetPosition={handleSetPosition}
              onOpen={handleOpen}
              onClose={handleClose}
              onStop={handleStop}
              onRefresh={fetchData}
            />
          </div>
        )}

        {activeTab === 'guide' && <LoxoneGuide />}

        {activeTab === 'settings' && <Settings onConfigChange={fetchData} />}
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-800 py-4 mt-8">
        <div className="max-w-6xl mx-auto px-4 text-center text-xs text-gray-500 space-y-1">
          <div>
            Loxone2Velux Gateway {health?.version ? `v${health.version}` : ''}
          </div>
          <div className="text-gray-600">
            &copy; 2026 Stefan Beyeler
          </div>
        </div>
      </footer>
    </div>
  );
}
