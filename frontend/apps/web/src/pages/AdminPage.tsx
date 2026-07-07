import React, { useState, useEffect, useRef } from 'react';
import { fetchRails } from '@bbb/core';

interface AdminStats {
  total_votes: number;
  votes_by_participant: { [name: string]: number };
  votes_by_hour: Array<{ hour: string; votes: number }>;
}

export const AdminPage: React.FC = () => {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<string>('');
  const timerRef = useRef<any>(null);

  const fetchStats = async () => {
    try {
      const data = await fetchRails<AdminStats>('/admin/v1/stats');
      setStats(data);
      setError(null);
      setLastUpdate(new Date().toLocaleTimeString('pt-BR'));
    } catch (err: any) {
      setError(err.message || 'Erro ao carregar métricas');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    timerRef.current = setInterval(fetchStats, 5000);
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, []);

  return (
    <div className="page-wrapper">
      <main className="page-main">
        <div className="container">

          {/* Header */}
          <div style={{ marginBottom: '2rem' }}>
            <div className="admin-warning">⚠ Painel Interno — Não Compartilhar</div>
            <h1 className="section-title" style={{ fontSize: '2rem' }}>Controle de Produção</h1>
            <p className="section-subtitle">
              Métricas em tempo real da votação. Atualização automática a cada 5s.
              {lastUpdate && (
                <span style={{ marginLeft: '0.5rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                  (Última: {lastUpdate})
                </span>
              )}
            </p>
          </div>

          {loading && <div className="spinner" />}
          {error && <div className="error-box">{error}</div>}

          {!loading && stats && (
            <>
              {/* Cards de Métricas */}
              <div className="stat-grid">
                <div className="stat-card">
                  <p className="stat-label">Total de Votos</p>
                  <p className="stat-value accent">{stats.total_votes.toLocaleString('pt-BR')}</p>
                </div>
                {Object.entries(stats.votes_by_participant).map(([name, votes]) => (
                  <div key={name} className="stat-card">
                    <p className="stat-label">Votos — {name}</p>
                    <p className="stat-value">{Number(votes).toLocaleString('pt-BR')}</p>
                  </div>
                ))}
              </div>

              {/* Histórico por Hora */}
              <div className="card" style={{ marginTop: '0' }}>
                <h3 style={{ marginBottom: '1.5rem', fontSize: '1.1rem', color: 'var(--text-primary)' }}>
                  Evolução por Hora (últimas 24h)
                </h3>
                {stats.votes_by_hour.length === 0 ? (
                  <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
                    Aguardando consolidação horária do Worker.
                  </p>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    {stats.votes_by_hour.map((item) => {
                      const max = Math.max(...stats.votes_by_hour.map((h) => h.votes));
                      const pct = max > 0 ? (item.votes / max) * 100 : 0;
                      return (
                        <div key={item.hour} className="hour-bar-row">
                          <span className="hour-label">{item.hour}</span>
                          <div className="hour-track">
                            <div className="hour-fill" style={{ width: `${pct}%` }} />
                          </div>
                          <span className="hour-count">{Number(item.votes).toLocaleString('pt-BR')}</span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  );
};
