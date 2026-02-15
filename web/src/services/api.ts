import {
  Node,
  NodesResponse,
  HealthResponse,
  CommandResponse,
  PositionRequest,
  SensorStatus,
  GatewayConfig,
} from '../types';

// Base fetch helper — uses relative URLs; the Go server injects a <base> tag
// for HA Ingress so that relative paths resolve correctly.
async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'API request failed');
  }

  // Detect HTML response (e.g. SPA fallback or wrong proxy route)
  const ct = response.headers.get('content-type') || '';
  if (!ct.includes('json')) {
    throw new Error(`Expected JSON from ${response.url} but got ${ct}`);
  }

  return response.json();
}

// Health check (no auth required)
export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch('health');
  if (!response.ok) {
    throw new Error(`Health ${response.status} from ${response.url}`);
  }
  const ct = response.headers.get('content-type') || '';
  if (!ct.includes('json')) {
    throw new Error(`Health: got ${ct} from ${response.url} (base: ${document.baseURI})`);
  }
  return response.json();
}

// Get all nodes
export async function getNodes(): Promise<NodesResponse> {
  return fetchJSON<NodesResponse>('api/nodes');
}

// Get single node
export async function getNode(id: number): Promise<Node> {
  return fetchJSON<Node>(`api/nodes/${id}`);
}

// Set node position (0-100%)
export async function setNodePosition(id: number, position: number): Promise<CommandResponse> {
  const body: PositionRequest = { position };
  return fetchJSON<CommandResponse>(`api/nodes/${id}/position`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// Open node fully
export async function openNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`api/nodes/${id}/open`, {
    method: 'POST',
  });
}

// Close node fully
export async function closeNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`api/nodes/${id}/close`, {
    method: 'POST',
  });
}

// Stop node movement
export async function stopNode(id: number): Promise<CommandResponse> {
  return fetchJSON<CommandResponse>(`api/nodes/${id}/stop`, {
    method: 'POST',
  });
}

// Get sensor status (rain, wind, etc.)
export async function getSensorStatus(): Promise<SensorStatus> {
  return fetchJSON<SensorStatus>('api/sensors');
}

// Refresh sensor status from KLF-200
export async function refreshSensorStatus(): Promise<SensorStatus> {
  return fetchJSON<SensorStatus>('api/sensors/refresh', {
    method: 'POST',
  });
}

// Configuration API

// Get current configuration
export async function getConfig(): Promise<GatewayConfig> {
  return fetchJSON<GatewayConfig>('api/config');
}

// Update configuration
export async function updateConfig(config: Partial<GatewayConfig>): Promise<GatewayConfig> {
  return fetchJSON<GatewayConfig>('api/config', {
    method: 'POST',
    body: JSON.stringify(config),
  });
}

// Reconnect to KLF-200
export async function reconnectGateway(): Promise<{ success: boolean; message: string }> {
  return fetchJSON<{ success: boolean; message: string }>('api/reconnect', {
    method: 'POST',
  });
}
