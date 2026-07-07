import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useResult, useParedao } from '@bbb/core';

export const ResultsPage: React.FC = () => {
  const { results, loading, error, startPolling, stopPolling } = useResult(3000);
  const { participants } = useParedao();
  const navigate = useNavigate();

  useEffect(() => {
    startPolling();
    return () => stopPolling();
  }, []);

  const getParticipant = (idStr: string) =>
    participants.find((p) => p.id.toString() === idStr);

  // Imagens por nome (fallback para avatar genérico)
  const getAvatar = (name: string) => {
    const lower = name.toLowerCase();
    if (lower.includes('alane')) return 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=80&h=80&fit=crop&face';
    if (lower.includes('davi'))  return 'https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=80&h=80&fit=crop&face';
    return `https://ui-avatars.com/api/?name=${encodeURIComponent(name)}&background=7c3aed&color=fff&size=80`;
  };

  return (
    <div className="page-wrapper">
      <main className="page-main">
        <div className="container">
          <div className="results-card">
            {/* Header */}
            <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
              <span className="section-badge live">Ao Vivo</span>
              <h1 className="section-title" style={{ fontSize: '2rem' }}>Resultado Parcial</h1>
              {!loading && results && (
                <p className="total-count">
                  {results.total_votes.toLocaleString('pt-BR')}
                  <span style={{ fontSize: '1rem', fontWeight: 400, color: 'var(--text-secondary)', marginLeft: '0.5rem' }}>
                    votos
                  </span>
                </p>
              )}
            </div>

            {loading && <div className="spinner" />}
            {error && <div className="error-box">{error}</div>}

            {!loading && results && (
              <div className="result-rows">
                {Object.entries(results.percentages)
                  .sort(([, a], [, b]) => Number(b) - Number(a))
                  .map(([pid, pct]) => {
                    const p = getParticipant(pid);
                    const name = p?.name ?? `Participante ${pid}`;
                    const votes = Math.round((Number(pct) / 100) * results.total_votes);
                    return (
                      <div key={pid}>
                        <div className="result-row-header">
                          <img
                            src={getAvatar(name)}
                            alt={name}
                            className="result-avatar"
                          />
                          <span className="result-name">{name}</span>
                          <span className="result-count">{Number(votes).toLocaleString('pt-BR')} votos</span>
                          <span className="result-pct">{pct}%</span>
                        </div>
                        <div className="progress-track">
                          <div className="progress-fill" style={{ width: `${pct}%` }} />
                        </div>
                      </div>
                    );
                  })}
              </div>
            )}

            {/* CTA */}
            <div style={{ textAlign: 'center', marginTop: '2.5rem' }}>
              <button className="btn-secondary" onClick={() => navigate('/')}>
                ← Votar Novamente
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};
