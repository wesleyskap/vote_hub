import { useState, useEffect } from 'react';
import { fetchRails } from '../api/client';
import { Participant } from '../types';

// useParedao busca os participantes do paredão ativo na Main API do Rails.
// Exemplo de uso:
//   const { participants, loading, error } = useParedao();
export const useParedao = () => {
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchParticipants = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchRails<Participant[]>('/api/v1/participants');
      setParticipants(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load active participants');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchParticipants();
  }, []);

  return { participants, loading, error, refetch: fetchParticipants };
};
