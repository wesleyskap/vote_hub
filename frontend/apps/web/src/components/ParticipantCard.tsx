import React, { useState } from 'react';
import { Participant } from '@bbb/core';
import { useVote } from '@bbb/core';

interface ParticipantCardProps {
  participant: Participant;
  paredaoId: number;
  onVoteSuccess: () => void;
}

declare global {
  interface Window {
    grecaptcha: any;
  }
}

export const ParticipantCard: React.FC<ParticipantCardProps> = ({ participant, paredaoId, onVoteSuccess }) => {
  const { submitVote, loading, error } = useVote(paredaoId, participant.id);
  const [localError, setLocalError] = useState<string | null>(null);

  const handleVote = async () => {
    setLocalError(null);
    try {
      // 1. Obtém ou cria o Fingerprint local persistente
      let fp = localStorage.getItem('fingerprint_id');
      if (!fp) {
        fp = 'fp-' + Math.random().toString(36).substring(2, 9) + '-' + Date.now();
        localStorage.setItem('fingerprint_id', fp);
      }

      // 2. Executa o reCAPTCHA Invisível v3
      let token = 'mock_token';
      if (window.grecaptcha) {
        token = await new Promise<string>((resolve) => {
          window.grecaptcha.ready(() => {
            window.grecaptcha.execute('6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI', { action: 'vote' })
              .then((t: string) => resolve(t))
              .catch(() => resolve('fallback_token'));
          });
        });
      }

      // 3. Envia o voto à API de Ingestão em Go
      await submitVote(token, fp);
      onVoteSuccess();
    } catch (err: any) {
      setLocalError(err.message || 'Erro de validação anti-bot');
    }
  };

  // Portraits reais de alta qualidade do Unsplash para visual premium
  const avatar = participant.name.toLowerCase() === 'alane'
    ? 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=400&h=400&fit=crop'
    : 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=400&h=400&fit=crop';

  return (
    <div className="glass-card participant-card">
      <div className="image-container">
        <img src={avatar} alt={participant.name} className="participant-avatar" />
      </div>
      <h3 className="participant-name">{participant.name}</h3>
      <button className="glow-button" onClick={handleVote} disabled={loading}>
        {loading ? 'Enviando voto...' : 'Votar'}
      </button>
      {(error || localError) && (
        <p style={{ color: '#ef4444', fontSize: '0.85rem', marginTop: '0.8rem', fontWeight: 600 }}>
          {error || localError}
        </p>
      )}
    </div>
  );
};
