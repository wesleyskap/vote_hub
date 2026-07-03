const getEnv = (key: string, fallback: string): string => {
  if (typeof import.meta !== 'undefined' && (import.meta as any).env && (import.meta as any).env[key]) {
    return (import.meta as any).env[key];
  }
  if (typeof globalThis !== 'undefined' && (globalThis as any).process && (globalThis as any).process.env && (globalThis as any).process.env[key]) {
    return (globalThis as any).process.env[key];
  }
  return fallback;
};

export const RAILS_API_URL = getEnv('VITE_RAILS_API_URL', 'http://localhost:3001');
export const GO_API_URL = getEnv('VITE_GO_API_URL', 'http://localhost:8080');

// fetchRails executa chamadas HTTP de leitura para a Main API em Rails.
// Exemplo de uso:
//   const data = await fetchRails<Participant[]>('/api/v1/participants')
export const fetchRails = async <T>(path: string, options?: RequestInit): Promise<T> => {
  const url = `${RAILS_API_URL}${path}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`Rails API call failed: status ${response.status} on path ${path}`);
  }

  return response.json() as Promise<T>;
};

// fetchGo executa chamadas HTTP de escrita (Ingestão) para a API em Go.
// Exemplo de uso:
//   await fetchGo('/api/v1/votes', { method: 'POST', body: JSON.stringify(payload) })
export const fetchGo = async <T>(path: string, options?: RequestInit): Promise<T> => {
  const url = `${GO_API_URL}${path}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({}));
    const msg = errorBody.error || `Go API call failed: status ${response.status}`;
    throw new Error(msg);
  }

  if (response.status === 202) {
    return { success: true } as unknown as T;
  }

  return response.json() as Promise<T>;
};
