import { useState } from 'react';
import { Rocket, Server, Loader2, Plus, ExternalLink, Trash2 } from 'lucide-react';
import './App.css';

interface Sandbox {
  id: string;
  url: string;
}

function App() {
  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [isProvisioning, setIsProvisioning] = useState(false);

  const handleLaunch = async () => {
    setIsProvisioning(true);
    try {
      const response = await fetch('http://localhost:8081/provision', { method: 'POST' });
      const data = await response.json();
      setSandboxes(prev => [{ id: data.id, url: data.url }, ...prev]);
    } catch (err) {
      alert("Orchestrator offline.");
    } finally {
      setIsProvisioning(false);
    }
  };

  // NEW: Logical deletion from both Backend and UI
  const handleTerminate = async (id: string) => {
    try {
      const response = await fetch(`http://localhost:8081/terminate?id=${id}`, { method: 'DELETE' });
      if (response.ok) {
        setSandboxes(prev => prev.filter(sb => sb.id !== id));
      }
    } catch (err) {
      alert("Failed to terminate sandbox.");
    }
  };

  return (
    <div style={{ minHeight: '100vh', padding: '3rem', maxWidth: '1000px', margin: '0 auto' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4rem' }}>
        <div>
          <h1 style={{ fontSize: '2.5rem', margin: 0, fontWeight: 800 }}>AETHER<span style={{ color: '#3b82f6' }}>.</span></h1>
          <p style={{ color: '#64748b' }}>Environments-as-a-Service</p>
        </div>
        <button onClick={handleLaunch} disabled={isProvisioning} style={{
            display: 'flex', alignItems: 'center', gap: '8px',
            backgroundColor: isProvisioning ? '#1e293b' : '#3b82f6',
            color: 'white', padding: '12px 24px', borderRadius: '10px',
            border: 'none', fontWeight: 600, cursor: isProvisioning ? 'not-allowed' : 'pointer'
        }}>
          {isProvisioning ? <Loader2 className="spinner" size={20} /> : <Plus size={20} />}
          {isProvisioning ? 'Provisioning...' : 'Launch Sandbox'}
        </button>
      </header>

      <main>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1.5rem' }}>
          {sandboxes.map(sb => (
            <div key={sb.id} style={{ backgroundColor: '#1e293b', padding: '1.5rem', borderRadius: '16px', border: '1px solid #334155' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '1rem' }}>
                <div style={{ color: '#10b981', fontSize: '0.75rem', fontWeight: 700 }}><Server size={14} /> ONLINE</div>
                {/* KILL BUTTON */}
                <button onClick={() => handleTerminate(sb.id)} style={{ background: 'none', border: 'none', color: '#ef4444', cursor: 'pointer' }}>
                   <Trash2 size={16} />
                </button>
              </div>
              <h3 style={{ margin: '0 0 1.5rem 0', fontFamily: 'monospace' }}>{sb.id}</h3>
              <a href={sb.url} target="_blank" rel="noreferrer" style={{ 
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
                  width: '100%', padding: '10px', backgroundColor: '#334155', color: 'white',
                  textDecoration: 'none', borderRadius: '8px', fontSize: '0.875rem'
              }}>Open Environment <ExternalLink size={14} /></a>
            </div>
          ))}
          
          {sandboxes.length === 0 && !isProvisioning && (
            <div style={{ gridColumn: '1/-1', padding: '5rem', textAlign: 'center', border: '2px dashed #1e293b', borderRadius: '24px', color: '#475569' }}>
               <Rocket size={48} style={{ marginBottom: '1.5rem', opacity: 0.3, display: 'inline-block' }} />
               <p>No active sandboxes. Launch one to begin.</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default App;