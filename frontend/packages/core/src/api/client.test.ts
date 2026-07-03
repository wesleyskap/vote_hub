import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fetchRails, fetchGo } from './client';

describe('HTTP Client Abstractions', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('fetchRails should decode JSON on successful response', async () => {
    const mockData = { id: 1, name: 'Alane' };
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => mockData,
    });

    const result = await fetchRails('/api/v1/participants');
    expect(result).toEqual(mockData);
    expect(fetch).toHaveBeenCalledWith('http://localhost:3001/api/v1/participants', expect.any(Object));
  });

  it('fetchGo should handle 202 Accepted status with success output', async () => {
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      status: 202,
    });

    const result = await fetchGo('/api/v1/votes', { method: 'POST' });
    expect(result).toEqual({ success: true });
  });

  it('fetchRails should throw error when request fails', async () => {
    (fetch as any).mockResolvedValueOnce({
      ok: false,
      status: 500,
    });

    await expect(fetchRails('/error')).rejects.toThrow('Rails API call failed: status 500 on path /error');
  });
});
