import {
  Node,
  NodesResponse,
  HealthResponse,
  CommandResponse,
  PositionRequest,
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
