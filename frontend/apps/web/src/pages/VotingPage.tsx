import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useParedao } from '@bbb/core';
import { ParticipantCard } from '../components/ParticipantCard';

export const VotingPage: React.FC = () => {
  const { participants, loading, error } = useParedao();
  const navigate = useNavigate();
  const PAREDAO_ID = 1;

  const handleVoteSuccess = () => {
    // Redireciona para parciais após 2s (feito dentro do ParticipantCard)
    setTimeout(() => navigate('/parciais'), 2200);
  };

  return (
    <div className="page-wrapper">
      <main className="page-main">
        <div className="container">
          {/* Hero */}
          <div className="voting-hero">
            <span className="section-badge">Paredão Ativo</span>
            <h1 className="section-title">Quem você quer eliminar?</h1>
            <p className="section-subtitle">
              Escolha um participante abaixo. Seu voto é anônimo e processado em tempo real.
            </p>
          </div>

          {/* Estados */}
          {loading && <div className="spinner" />}

          {error && (
            <div className="error-box">
              <p>Erro ao conectar com o servidor.</p>
              <p style={{ fontSize: '0.85rem', marginTop: '0.4rem', opacity: 0.7 }}>{error}</p>
            </div>
          )}

          {!loading && !error && (
            <div className="voting-grid">
              {participants.map((p) => (
                <ParticipantCard
                  key={p.id}
                  participant={p}
                  paredaoId={PAREDAO_ID}
                  onVoteSuccess={handleVoteSuccess}
                />
              ))}
            </div>
          )}

          {/* Link para parciais */}
          {!loading && !error && (
            <div style={{ textAlign: 'center', marginTop: '2.5rem' }}>
              <button className="btn-secondary" onClick={() => navigate('/parciais')}>
                Ver Parciais
              </button>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};
