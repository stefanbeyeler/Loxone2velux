import {
  Node,
  NodesResponse,
  HealthResponse,
  CommandResponse,
  PositionRequest,
  SensorStatus,
  GatewayConfig,
} from '../types';

// Get API token from localStorage (no prompt - handled by App.tsx)
function getApiToken(): string {
  return localStorage.getItem('api_token') || '';
}

// Clear stored token (for logout/reset)
export function clearApiToken(): void {
  localStorage.removeItem('api_token');
}

// Base fetch with token
async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const token = getApiToken();

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options?.headers,
    },
  });

  if (response.status === 401) {
    clearApiToken();
    throw new Error('Unauthorized - Token ungültig');
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'API request failed');
  }

  return response.json();
}

// Health check (no auth required)
export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch('/health');
  if (!response.ok) {
    throw new Error('Health check failed');
  }
  return response.json();
}

// Get all nodes
export async function getNodes(): Promise<NodesResponse> {
  return fetchJSON<NodesResponse>('/api/nodes');
}

// Get single node
export async function getNode(id: number): Promise<Node> {
  return fetchJSON<Node>(`/api/nodes/${id}`);
}

// Set node position (0-100%)
export async function setNodePosition(id: number, position: number): Promise<CommandResponse> {
  const body: PositionRequest = { position };
  return fetchJSON<CommandResponse>(`/api/nodes/${id}/position`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// Open node fully
export async function openNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`/api/nodes/${id}/open`, {
    method: 'POST',
  });
}

// Close node fully
export async function closeNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`/api/nodes/${id}/close`, {
    method: 'POST',
  });
}

// Stop node movement
export async function stopNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`/api/nodes/${id}/stop`, {
    method: 'POST',
  });
}

// Get sensor status (rain, wind, etc.)
export async function getSensorStatus(): Promise<SensorStatus> {
  return fetchJSON<SensorStatus>('/api/sensors');
}

// Refresh sensor status from KLF-200
export async function refreshSensorStatus(): Promise<SensorStatus> {
  return fetchJSON<SensorStatus>('/api/sensors/refresh', {
    method: 'POST',
  });
}

// Configuration API

// Get current configuration
export async function getConfig(): Promise<GatewayConfig> {
  return fetchJSON<GatewayConfig>('/api/config');
}

// Update configuration
export async function updateConfig(config: Partial<GatewayConfig>): Promise<GatewayConfig> {
  return fetchJSON<GatewayConfig>('/api/config', {
    method: 'PUT',
    body: JSON.stringify(config),
  });
}

// Reconnect to KLF-200
export async function reconnectGateway(): Promise<{ success: boolean; message: string }> {
  return fetchJSON<{ success: boolean; message: string }>('/api/reconnect', {
    method: 'POST',
  });
}
