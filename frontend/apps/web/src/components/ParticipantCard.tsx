import React, { useState } from 'react';
import { Participant } from '@bbb/core';
import { useVote } from '@bbb/core';

interface ParticipantCardProps {
  participant: Participant;
  paredaoId: number;
  onVoteSuccess: () => void;
}

declare global {
  interface Window { grecaptcha: any; }
}

const AVATARS: Record<string, string> = {
  alane: 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=400&h=400&fit=crop&face',
  davi:  'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=400&h=400&fit=crop&face',
};

const getAvatar = (name: string): string => {
  const key = name.toLowerCase();
  return AVATARS[key] ?? `https://ui-avatars.com/api/?name=${encodeURIComponent(name)}&background=7c3aed&color=fff&size=400`;
};

export const ParticipantCard: React.FC<ParticipantCardProps> = ({ participant, paredaoId, onVoteSuccess }) => {
  const { submitVote, loading } = useVote(paredaoId, participant.id);
  const [confirmed, setConfirmed] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);
  const [countdown, setCountdown] = useState(2);

  const handleVote = async () => {
    setLocalError(null);
    try {
      let fp = localStorage.getItem('fingerprint_id');
      if (!fp) {
        fp = 'fp-' + Math.random().toString(36).substring(2, 9) + '-' + Date.now();
        localStorage.setItem('fingerprint_id', fp);
      }

      let token = 'mock_token';
      if (window.grecaptcha) {
        token = await new Promise<string>((resolve) => {
          window.grecaptcha.ready(() => {
            window.grecaptcha
              .execute('6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI', { action: 'vote' })
              .then((t: string) => resolve(t))
              .catch(() => resolve('fallback_token'));
          });
        });
      }

      await submitVote(token, fp);
      setConfirmed(true);

      // Countdown de 2s antes do redirect
      let t = 2;
      const interval = setInterval(() => {
        t--;
        setCountdown(t);
        if (t <= 0) {
          clearInterval(interval);
          onVoteSuccess();
        }
      }, 1000);
    } catch (err: any) {
      setLocalError(err.message || 'Erro ao registrar voto');
    }
  };

  return (
    <div className="card participant-card">
      <div className="participant-img-wrap">
        <img src={getAvatar(participant.name)} alt={participant.name} />
      </div>

      <h3 className="participant-name">{participant.name}</h3>

      {confirmed ? (
        <div className="vote-confirmed">
          <div className="checkmark-circle">✓</div>
          <p className="vote-confirmed-text">Voto registrado!</p>
          <p className="vote-redirect-text">Indo para parciais em {countdown}s...</p>
        </div>
      ) : (
        <>
          <button
            className="btn-primary"
            onClick={handleVote}
            disabled={loading}
            aria-label={`Votar em ${participant.name}`}
          >
            {loading ? 'Enviando...' : 'Votar'}
          </button>
          {localError && (
            <p style={{ color: 'var(--error)', fontSize: '0.85rem', marginTop: '0.75rem', fontWeight: 600 }}>
              {localError}
            </p>
          )}
        </>
      )}
    </div>
  );
};
