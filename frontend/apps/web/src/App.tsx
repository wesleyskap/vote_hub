import React, { useState } from 'react';
import { useParedao } from '@bbb/core';
import { ParticipantCard } from './components/ParticipantCard';
import { ResultChart } from './components/ResultChart';
import { Dashboard } from './components/Dashboard';

export const App: React.FC = () => {
  const [view, setView] = useState<'voting' | 'results' | 'dashboard'>('voting');
  const { participants, loading, error } = useParedao();

  const activeParedaoId = 1; // ID do Paredão ativo semeado no banco Rails

  return (
    <div className="app-container">
      <header className="header">
        <h1 className="logo">
          👁️ BBB <span style={{ fontSize: '1.2rem', fontWeight: 400, color: 'var(--text-secondary)' }}>Votação</span>
        </h1>
        <nav className="nav-links">
          <button 
            className={`nav-button ${view === 'voting' ? 'active' : ''}`} 
            onClick={() => setView('voting')}
          >
            Votar
          </button>
          <button 
            className={`nav-button ${view === 'results' ? 'active' : ''}`} 
            onClick={() => setView('results')}
          >
            Parciais
          </button>
          <button 
            className={`nav-button ${view === 'dashboard' ? 'active' : ''}`} 
            onClick={() => setView('dashboard')}
          >
            Produção
          </button>
        </nav>
      </header>

      <main style={{ flexGrow: 1, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        {loading && <div className="spinner"></div>}

        {error && (
          <div className="glass-card" style={{ border: '1px solid #ef4444', textAlign: 'center' }}>
            <h3 style={{ color: '#ef4444', marginBottom: '1rem' }}>Erro ao conectar com as APIs do BBB</h3>
            <p style={{ color: 'var(--text-secondary)' }}>
              Verifique se os servidores Rails (port 3001) e Go (port 8080) estão rodando.
            </p>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.8rem' }}>
              Detalhe: {error}
            </p>
          </div>
        )}

        {!loading && !error && (
          <>
            {view === 'voting' && (
              <section className="voting-section">
                <span className="title-badge">Voto Popular</span>
                <h2>Quem você quer eliminar?</h2>
                <p style={{ color: 'var(--text-secondary)' }}>
                  Escolha um candidato abaixo e clique para registrar seu voto.
                </p>
                <div className="voting-grid">
                  {participants.map((p) => (
                    <ParticipantCard 
                      key={p.id} 
                      participant={p} 
                      paredaoId={activeParedaoId} 
                      onVoteSuccess={() => setView('results')}
                    />
                  ))}
                </div>
              </section>
            )}

            {view === 'results' && (
              <ResultChart 
                participants={participants} 
                onBackToVote={() => setView('voting')}
              />
            )}

            {view === 'dashboard' && <Dashboard />}
          </>
        )}
      </main>
    </div>
  );
};
export default App;
