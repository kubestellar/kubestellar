import React from 'react';

const GlobeLoader = () => {
  return (
    <div className="flex flex-col items-center justify-center space-y-4">
      {/* KubeStellar Logo Animation */}
      <div className="relative w-16 h-16">
        {/* Central core */}
        <div className="absolute inset-4 bg-gradient-to-r from-blue-400 to-purple-500 rounded-full animate-pulse"></div>
        
        {/* Rotating rings */}
        <div className="absolute inset-0 border-2 border-blue-400/30 rounded-full animate-spin"></div>
        <div className="absolute inset-2 border-2 border-purple-400/30 rounded-full animate-spin" style={{ animationDirection: 'reverse', animationDuration: '3s' }}></div>
        
        {/* Orbiting dots */}
        <div className="absolute inset-0 animate-spin" style={{ animationDuration: '2s' }}>
          <div className="absolute -top-1 left-1/2 w-2 h-2 bg-cyan-400 rounded-full transform -translate-x-1/2"></div>
          <div className="absolute top-1/2 -right-1 w-2 h-2 bg-yellow-400 rounded-full transform -translate-y-1/2"></div>
          <div className="absolute -bottom-1 left-1/2 w-2 h-2 bg-green-400 rounded-full transform -translate-x-1/2"></div>
          <div className="absolute top-1/2 -left-1 w-2 h-2 bg-pink-400 rounded-full transform -translate-y-1/2"></div>
        </div>
      </div>
      
      {/* Loading text */}
      <div className="text-center">
        <div className="text-blue-400 font-semibold text-lg">KubeStellar</div>
        <div className="text-gray-400 text-sm">Initializing clusters...</div>
        
        {/* Progress dots */}
        <div className="flex space-x-1 mt-2 justify-center">
          <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce"></div>
          <div className="w-2 h-2 bg-purple-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
          <div className="w-2 h-2 bg-cyan-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
        </div>
      </div>
    </div>
  );
};

export default GlobeLoader;