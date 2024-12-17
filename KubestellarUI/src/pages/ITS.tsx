import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface ManagedClusterInfo {
  name: string;
  labels: { [key: string]: string };
  creationTime: string;
}

const ITS = () => {
  const [clusters, setClusters] = useState<ManagedClusterInfo[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [sortOption, setSortOption] = useState<string>('name');

  const fetchClusters = () => {
    setLoading(true);
    axios.get('http://localhost:4000/api/clusters')
      .then(response => {
        const itsData: ManagedClusterInfo[] = response.data.itsData || [];

        if (Array.isArray(itsData)) {
          setClusters(itsData);
        } else {
          setError('Invalid data format received from server');
        }
        setLoading(false);
      })
      .catch(() => {
        setError('Error fetching ITS information');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  const filteredClusters = clusters.filter(cluster =>
    cluster.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const sortedClusters = [...filteredClusters].sort((a, b) => {
    if (sortOption === 'name') {
      return a.name.localeCompare(b.name);
    } else {
      return new Date(a.creationTime).getTime() - new Date(b.creationTime).getTime();
    }
  });

  const totalClusters = clusters.length;

  if (loading) return <p className="text-center p-4">Loading ITS information...</p>;
  if (error) return <p className="text-center p-4 text-error">{error}</p>;
  if (!filteredClusters.length) return <p className="text-center p-4">No clusters found</p>;

  return (
    <div className="w-full max-w-7xl mx-auto p-4">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Managed Clusters ({totalClusters})</h1>
      </div>
      <div className="flex justify-end mb-4">
        <button className="btn btn-secondary" onClick={fetchClusters}>Refresh</button>
      </div>
      <div className="flex justify-between mb-4">
        <input
          type="text"
          placeholder="Search clusters..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="input input-bordered w-full mr-4"
        />
        <select
          value={sortOption}
          onChange={(e) => setSortOption(e.target.value)}
          className="select select-bordered w-1/4"
        >
          <option value="name">Sort by Name</option>
          <option value="creationTime">Sort by Creation Time</option>
        </select>
      </div>
      <div className="grid gap-4">
        {sortedClusters.map((cluster) => (
          <div key={cluster.name} className="card bg-base-200 shadow-xl">
            <div className="card-body">
              <h2 className="card-title">{cluster.name}</h2>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h3 className="font-semibold mb-2">Labels</h3>
                  {Object.keys(cluster.labels).length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(cluster.labels).map(([key, value]) => (
                        <span 
                          key={`${key}-${value}`} 
                          className="badge badge-primary"
                          title={`${key}: ${value}`}
                        >
                          {key}={value}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <p>No labels available.</p>
                  )}
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Creation Time</h3>
                  <p>{new Date(cluster.creationTime).toLocaleString()}</p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ITS;
