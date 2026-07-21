import { useState } from 'react';
import { Rocket, Server, Loader2, Plus, ExternalLink } from 'lucide-react';
import './App.css';

// 1. Data Contract for our Sandbox environments
interface Sandbox {
  id: string;
  url: string;
  status: string;
}

function App() {
  // 2. State: Tracks the list of active sandboxes and the loading status
  const [sandboxes, setSandboxes] = useState<Sandbox[]>([]);
  const [isProvisioning, setIsProvisioning] = useState(false);

  // 3. Logic: Triggers the Go Orchestrator API
  const handleLaunch = async () => {
    setIsProvisioning(true);
    try {
      // POST to our Go Microservice on Port 8081
      const response = await fetch('http://localhost:8081/provision', {
        method: 'POST',
      });

      if (!response.ok) throw new Error('Orchestration engine failed');

      const data = await response.json();
      
      const newSandbox: Sandbox = {
        id: data.id,
        url: data.url, // URL provided dynamically by the Gateway
        status: 'Running'
      };

      // Add the new sandbox to the top of the list
      setSandboxes(prev => [newSandbox, ...prev]);
    } catch (err) {
      console.error(err);
      alert("System Error: Go Orchestrator (8081) is offline or CORS is not enabled.");
    } finally {
      setIsProvisioning(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', padding: '3rem', maxWidth: '1000px', margin: '0 auto' }}>
      
      {/* HEADER: Title and Action Button */}
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4rem' }}>
        <div>
          <h1 style={{ fontSize: '2.5rem', margin: 0, fontWeight: 800, letterSpacing: '-0.05em' }}>
            AETHER<span style={{ color: '#3b82f6' }}>.</span>
          </h1>
          <p style={{ color: '#64748b', fontWeight: 500 }}>Enterprise Environments-as-a-Service</p>
        </div>
        
        <button 
          onClick={handleLaunch} 
          disabled={isProvisioning}
          style={{
            display: 'flex', alignItems: 'center', gap: '8px',
            backgroundColor: isProvisioning ? '#1e293b' : '#3b82f6',
            color: 'white', padding: '12px 24px', borderRadius: '10px',
            border: 'none', fontWeight: 600, cursor: isProvisioning ? 'not-allowed' : 'pointer',
            transition: 'all 0.2s ease'
          }}
        >
          {isProvisioning ? <Loader2 className="spinner" size={20} /> : <Plus size={20} />}
          {isProvisioning ? 'Provisioning...' : 'Launch Sandbox'}
        </button>
      </header>

      {/* MAIN CONTENT: The Workload Grid */}
      <main>
        <h2 style={{ fontSize: '0.75rem', textTransform: 'uppercase', color: '#64748b', letterSpacing: '0.15em', marginBottom: '1.5rem' }}>
          Active Workloads ({sandboxes.length})
        </h2>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1.5rem' }}>
          
          {/* Loop through all active sandboxes */}
          {sandboxes.map(sb => (
            <div key={sb.id} style={{ backgroundColor: '#1e293b', padding: '1.5rem', borderRadius: '16px', border: '1px solid #334155' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#10b981', fontSize: '0.75rem', fontWeight: 700, marginBottom: '1rem' }}>
                <Server size={14} /> ONLINE
              </div>
              <h3 style={{ margin: '0 0 1.5rem 0', fontFamily: 'monospace', color: '#f8fafc' }}>{sb.id}</h3>
              <a 
                href={sb.url} 
                target="_blank" 
                rel="noreferrer"
                style={{ 
                  display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
                  width: '100%', padding: '12px', backgroundColor: '#334155', color: 'white',
                  textDecoration: 'none', borderRadius: '8px', fontSize: '0.875rem', fontWeight: 600,
                  transition: 'background 0.2s'
                }}
              >
                Open Environment <ExternalLink size={14} />
              </a>
            </div>
          ))}
          
          {/* EMPTY STATE: Shown only when no sandboxes exist */}
          {sandboxes.length === 0 && !isProvisioning && (
            <div style={{ 
              gridColumn: '1/-1', padding: '5rem', textAlign: 'center', 
              border: '2px dashed #1e293b', borderRadius: '24px', color: '#475569' 
            }}>
               <Rocket size={48} style={{ marginBottom: '1.5rem', opacity: 0.3, display: 'inline-block' }} />
               <p style={{ fontSize: '1.1rem', margin: 0 }}>No active sandboxes found.</p>
               <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>Launch a new environment to see it appear here.</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default App;