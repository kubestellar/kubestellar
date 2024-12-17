import { useState, useEffect } from 'react';
import axios from 'axios';
import { BindingPolicy } from '../types/BindingPolicy';

const BindingPolicies = () => {
  const [policies, setPolicies] = useState<BindingPolicy[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    axios.get('http://localhost:4000/api/binding-policies')
      .then(response => {
        setPolicies(response.data);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching binding policies:', error);
        setError('Binding policies are a work in progress!');
        setLoading(false);
      });
  }, []);

  if (loading) return <p className="text-center p-4">Loading binding policies...</p>;
  if (error) return <p className="text-center p-4 text-error">{error}</p>;
  if (!policies.length) return <p className="text-center p-4">No binding policies found</p>;

  return (
    <div className="w-full max-w-7xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Binding Policies ({policies.length})</h1>
      <div className="grid gap-4">
        {policies.map((policy) => (
          <div key={policy.name} className="card bg-base-200 shadow-xl">
            <div className="card-body">
              <h2 className="card-title">{policy.name}</h2>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="font-semibold mb-2">Namespace</h3>
                  <p>{policy.namespace}</p>
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Cluster Name</h3>
                  <p>{policy.clusterName}</p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default BindingPolicies;