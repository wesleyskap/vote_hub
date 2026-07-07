import React from 'react';
import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom';
import { VotingPage } from './pages/VotingPage';
import { ResultsPage } from './pages/ResultsPage';
import { AdminPage } from './pages/AdminPage';
import './index.css';

const Header: React.FC = () => (
  <header className="site-header">
    <div className="container">
      <NavLink to="/" className="logo">
        👁️ BBB <span style={{ fontWeight: 400, fontSize: '1rem', opacity: 0.6 }}>Votação</span>
      </NavLink>
      <nav className="nav-links">
        <NavLink
          to="/"
          end
          className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
        >
          Votar
        </NavLink>
        <NavLink
          to="/parciais"
          className={({ isActive }) => `nav-link${isActive ? ' active' : ''}`}
        >
          Parciais
        </NavLink>
        {/* /admin intencionalmente sem link no header público */}
      </nav>
    </div>
  </header>
);

export const App: React.FC = () => (
  <BrowserRouter>
    <Header />
    <Routes>
      <Route path="/"         element={<VotingPage />} />
      <Route path="/parciais" element={<ResultsPage />} />
      <Route path="/admin"    element={<AdminPage />} />
    </Routes>
  </BrowserRouter>
);

export default App;
