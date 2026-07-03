import { useState, useEffect, useRef } from 'react';
import { fetchRails } from '../api/client';
import { ParedaoResults } from '../types';

// useResult encapsula o polling de segundo em segundo para recuperar o resultado
// parcial do paredão ativo a partir da Main API do Rails.
// Exemplo de uso:
//   const { results, startPolling, stopPolling } = useResult(2000);
export const useResult = (pollingIntervalMs = 2000) => {
  const [results, setResults] = useState<ParedaoResults | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isPolling, setIsPolling] = useState<boolean>(false);
  const timerRef = useRef<any>(null);

  const fetchResults = async () => {
    try {
      const data = await fetchRails<ParedaoResults>('/api/v1/results/current');
      setResults(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load paredao parciais');
    } finally {
      setLoading(false);
    }
  };

  const startPolling = () => {
    if (timerRef.current) return;
    setIsPolling(true);
    fetchResults(); // Busca imediata inicial
    timerRef.current = setInterval(fetchResults, pollingIntervalMs);
  };

  const stopPolling = () => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    setIsPolling(false);
  };

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  return { results, loading, error, isPolling, startPolling, stopPolling, refetch: fetchResults };
};
