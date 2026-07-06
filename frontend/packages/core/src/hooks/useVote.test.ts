// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useVote } from './useVote';
import { fetchGo } from '../api/client';

// Mock do cliente HTTP fetchGo
vi.mock('../api/client', () => ({
  fetchGo: vi.fn(),
}));

describe('useVote hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('deve iniciar com estados padrao de votacao', () => {
    const { result } = renderHook(() => useVote(1, 2));

    expect(result.current.loading).toBe(false);
    expect(result.current.success).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('deve registrar o voto com sucesso e mudar estados', async () => {
    // Simula resposta aceita da API Go
    vi.mocked(fetchGo).mockResolvedValueOnce({ success: true });

    const { result } = renderHook(() => useVote(1, 2));

    // Executa submissão do voto
    await act(async () => {
      await result.current.submitVote('token-recaptcha-123', 'fingerprint-abc');
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.success).toBe(true);
    expect(result.current.error).toBe(null);
    expect(fetchGo).toHaveBeenCalledWith('/api/v1/votes', {
      method: 'POST',
      body: JSON.stringify({
        paredao_id: 1,
        participant_id: 2,
        recaptcha_token: 'token-recaptcha-123',
        fingerprint_id: 'fingerprint-abc',
      }),
    });
  });

  it('deve capturar falhas da API e definir estado de erro', async () => {
    // Simula erro da API Go
    vi.mocked(fetchGo).mockRejectedValueOnce(new Error('Rate limit exceeded'));

    const { result } = renderHook(() => useVote(1, 2));

    await act(async () => {
      await result.current.submitVote('token-recaptcha-123', 'fingerprint-abc');
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.success).toBe(false);
    expect(result.current.error).toBe('Rate limit exceeded');
  });
});
