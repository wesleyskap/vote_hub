import { useState } from 'react';
import { fetchGo } from '../api/client';
import { VotePayload, VoteResponse } from '../types';

// useVote gerencia o envio assíncrono do voto para a Ingestion API em Go.
// Exemplo de uso:
//   const { submitVote, loading, success, error } = useVote(1, 2);
//   await submitVote("token_recaptcha", "fingerprint_id");
export const useVote = (paredaoId: number, participantId: number) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);

  const submitVote = async (recaptchaToken: string, fingerprint: string) => {
    setLoading(true);
    setError(null);
    setSuccess(false);
    try {
      const payload: VotePayload = {
        paredao_id: paredaoId,
        participant_id: participantId,
        recaptcha_token: recaptchaToken,
        fingerprint_id: fingerprint,
      };

      await fetchGo<VoteResponse>('/api/v1/votes', {
        method: 'POST',
        body: JSON.stringify(payload),
      });

      setSuccess(true);
    } catch (err: any) {
      setError(err.message || 'Failed to register vote. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return { submitVote, loading, error, success };
};
