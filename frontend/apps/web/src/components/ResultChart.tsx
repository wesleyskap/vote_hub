import React, { useEffect } from 'react';
import { useResult, Participant } from '@bbb/core';

interface ResultChartProps {
  participants: Participant[];
  onBackToVote: () => void;
}

export const ResultChart: React.FC<ResultChartProps> = ({ participants, onBackToVote }) => {
  const { results, loading, error, startPolling, stopPolling } = useResult(2000); // Polling a cada 2 segundos

  useEffect(() => {
    startPolling();
    return () => stopPolling();
  }, []);

  const getParticipantName = (idStr: string) => {
    const p = participants.find((part) => part.id.toString() === idStr);
    return p ? p.name : `Participante ${idStr}`;
  };

  return (
    <div className="glass-card results-container">
      <div className="results-header">
        <span className="title-badge pulse">Parciais ao Vivo</span>
        <h2>Resultado Parcial do Paredão</h2>
        {!loading && results && (
          <p className="total-votes-count">
            {results.total_votes.toLocaleString()} <span style={{ fontSize: '1rem', color: 'var(--text-secondary)' }}>votos computados</span>
          </p>
        )}
      </div>

      {loading && (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '2rem' }}>
          <div className="spinner"></div>
        </div>
      )}

      {error && (
        <p style={{ color: '#ef4444', textAlign: 'center', fontWeight: 600 }}>
          Erro ao carregar dados do Rails: {error}
        </p>
      )}

      {!loading && results && (
        <div style={{ marginTop: '2rem' }}>
          {Object.entries(results.percentages).map(([pid, percentage]) => {
            const name = getParticipantName(pid);
            return (
              <div key={pid} className="result-row">
                <div className="result-info">
                  <span>{name}</span>
                  <span className="result-percentage">{percentage}%</span>
                </div>
                <div className="progress-bar-bg">
                  <div className="progress-bar-fill" style={{ width: `${percentage}%` }}></div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      <button 
        className="glow-button" 
        onClick={onBackToVote} 
        style={{ 
          marginTop: '2rem', 
          background: 'rgba(255,255,255,0.05)', 
          border: '1px solid var(--card-border)', 
          boxShadow: 'none' 
        }}
      >
        Votar Novamente
      </button>
    </div>
  );
};
