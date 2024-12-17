import React, { useEffect, useState } from 'react';
import axios from 'axios';

interface ContextInfo {
  name: string;
  cluster: string;
}

const K8sInfo = () => {
  const [contexts, setContexts] = useState<ContextInfo[]>([]);
  const [clusters, setClusters] = useState<string[]>([]);
  const [currentContext, setCurrentContext] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    axios.get('http://localhost:4000/api/clusters')
      .then(response => {
        setContexts(response.data.contexts);
        setClusters(response.data.clusters);
        setCurrentContext(response.data.currentContext);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching Kubernetes information:', error);
        setError('Error fetching Kubernetes information');
        setLoading(false);
      });
  }, []);

  if (loading) return <p>Loading Kubernetes information...</p>;
  if (error) return <p>{error}</p>;

  return (
    <div className="w-full max-w-full p-4">
      <div className="grid grid-cols-3">
      <div>
        <h2 className="text-2xl font-bold mb-6">Kubernetes Clusters ({clusters.length})</h2>
        <ul>
          {clusters.map(cluster => (
            <li key={cluster}>{cluster}</li>
          ))}
        </ul>
      </div>

      <div>
        <h2 className="text-2xl font-bold mb-6">Kubernetes Contexts ({contexts.length})</h2>
        <ul>
          {contexts.map(ctx => (
            <li key={ctx.name}>
              {ctx.name} {ctx.name === currentContext && '(Current)'} 
              <span style={{color: '#666'}}> → cluster: {ctx.cluster}</span>
            </li>
          ))}
        </ul>
      </div>
      <div>
        <h2 className="text-2xl font-bold mb-6">Current Context</h2>
        <p>{currentContext}</p>
      </div>
    </div>
    </div>
  );
};

export default K8sInfo;
