import React, { useState, useEffect, useRef } from 'react';
import { fetchRails } from '@bbb/core';

interface AdminStats {
  total_votes: number;
  votes_by_participant: { [name: string]: number };
  votes_by_hour: Array<{ hour: string; votes: number }>;
}

export const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const timerRef = useRef<any>(null);

  const fetchStats = async () => {
    try {
      const data = await fetchRails<AdminStats>('/admin/v1/stats');
      setStats(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Erro ao carregar métricas administrativas');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    timerRef.current = setInterval(fetchStats, 5000); // Polling a cada 5 segundos para a produção
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  return (
    <div style={{ width: '100%' }}>
      <div className="results-header" style={{ marginBottom: '3rem' }}>
        <span className="title-badge">Painel de Produção</span>
        <h2>Métricas de Controle Interno</h2>
      </div>

      {loading && (
        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <div className="spinner"></div>
        </div>
      )}
      
      {error && (
        <p style={{ color: '#ef4444', textAlign: 'center', fontWeight: 600 }}>
          {error}
        </p>
      )}

      {!loading && stats && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
          
          {/* Cards de Métricas Consolidadas */}
          <div className="dashboard-grid">
            <div className="glass-card stat-card">
              <p className="stat-label">Total Geral de Votos</p>
              <p className="stat-value" style={{ color: 'var(--secondary-color)' }}>
                {stats.total_votes.toLocaleString()}
              </p>
            </div>
            
            {Object.entries(stats.votes_by_participant).map(([name, votes]) => (
              <div key={name} className="glass-card stat-card">
                <p className="stat-label">Votos brutos - {name}</p>
                <p className="stat-value">{votes.toLocaleString()}</p>
              </div>
            ))}
          </div>

          {/* Histórico Consolidado por Hora */}
          <div className="glass-card">
            <h3 style={{ marginBottom: '1.5rem', fontSize: '1.2rem' }}>Curva de Crescimento (Últimas 24 Horas)</h3>
            {stats.votes_by_hour.length === 0 ? (
              <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
                Aguardando consolidação de dados. O processamento horários do Worker populá esta seção.
              </p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                {stats.votes_by_hour.map((item) => {
                  const maxHourVotes = Math.max(...stats.votes_by_hour.map(h => h.votes));
                  const pct = maxHourVotes > 0 ? (item.votes / maxHourVotes) * 100 : 0;
                  return (
                    <div key={item.hour} style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                      <span style={{ minWidth: '130px', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                        {item.hour}
                      </span>
                      <div className="progress-bar-bg" style={{ flexGrow: 1, height: '10px' }}>
                        <div 
                          className="progress-bar-fill" 
                          style={{ 
                            width: `${pct}%`, 
                            background: 'linear-gradient(90deg, var(--primary-color), var(--primary-color))',
                            boxShadow: 'none'
                          }}
                        ></div>
                      </div>
                      <span style={{ fontSize: '0.85rem', fontWeight: 'bold', minWidth: '80px', textAlign: 'right' }}>
                        {item.votes.toLocaleString()}
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
