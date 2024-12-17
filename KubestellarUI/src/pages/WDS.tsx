import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface WorkloadInfo {
  name: string;
  kind: string;  // 'Deployment' or 'Service'
  namespace: string;
  creationTime: string;
}

const WDS = () => {
  const [workloads, setWorkloads] = useState<WorkloadInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    axios.get('http://localhost:4000/api/wds/workloads')
      .then(response => {
        console.log('Response data:', response.data);
        if (Array.isArray(response.data)) {
          setWorkloads(response.data);
        } else {
          console.error('Invalid data format received:', response.data);
          setError('Invalid data format received from server');
        }
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching WDS information:', error);
        setError('Error fetching WDS information');
        setLoading(false);
      });
  }, []);

  if (loading) return <p className="text-center p-4">Loading WDS information...</p>;
  if (error) return <p className="text-center p-4 text-error">{error}</p>;
  if (!workloads.length) return <p className="text-center p-4">No workloads found</p>;

  return (
    <div className="w-full max-w-7xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">WDS Workloads ({workloads.length})</h1>
      <div className="grid gap-4">
        {workloads.map((workload) => (
          <div key={`${workload.kind}-${workload.namespace}-${workload.name}`} className="card bg-base-200 shadow-xl">
            <div className="card-body">
              <h2 className="card-title">{workload.name}</h2>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <h3 className="font-semibold mb-2">Kind</h3>
                  <span className="badge badge-primary">{workload.kind}</span>
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Namespace</h3>
                  <span className="badge badge-secondary">{workload.namespace}</span>
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Creation Time</h3>
                  <p>{new Date(workload.creationTime).toLocaleString()}</p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default WDS;
