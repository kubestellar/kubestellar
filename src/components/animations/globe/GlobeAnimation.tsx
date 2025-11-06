"use client";

import { Suspense, useState, useEffect } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, PerspectiveCamera } from '@react-three/drei';
import NetworkGlobe from './NetworkGlobe';
import GlobeLoader from './GlobeLoader';

interface GlobeAnimationProps {
  width?: string;
  height?: string;
  className?: string;
  showLoader?: boolean;
  enableControls?: boolean;
  enablePan?: boolean;
  autoRotate?: boolean;
  style?: React.CSSProperties;
}

const GlobeAnimation = ({
  width = "100%",
  height = "600px",
  className = "",
  showLoader = true,
  enableControls = false,
  enablePan = false,
  autoRotate = false,
  style = {}
}: GlobeAnimationProps) => {
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    // Simulate loading delay to show progressive animation
    const timer = setTimeout(() => {
      setIsLoaded(true);
    }, 1000);

    return () => clearTimeout(timer);
  }, []);

  return (
    <div 
      className={`relative ${className}`} 
      style={{ width, height, ...style }}
    >
      {/* Loader */}
      {showLoader && !isLoaded && (
        <div className="absolute inset-0 flex items-center justify-center bg-transparent z-10">
          <GlobeLoader />
        </div>
      )}

      {/* Three.js Canvas */}
      <Canvas
        className="w-full h-full"
        style={{ background: 'transparent' }}
      >
        {/* Camera */}
        <PerspectiveCamera
          makeDefault
          position={[0, 0, 10]}
          fov={50}
          near={0.1}
          far={1000}
        />

        {/* Lighting */}
        <ambientLight intensity={0.4} />
        <pointLight position={[10, 10, 10]} intensity={1} />
        <pointLight position={[-10, -10, -10]} intensity={0.5} />

        {/* Controls - allow full 360-degree rotation */}
        {enableControls && (
          <OrbitControls
            enableZoom={false} // Disable zoom as requested
            enablePan={enablePan}
            enableRotate={true}
            autoRotate={autoRotate}
            autoRotateSpeed={0.3} // Reduced from 1.0 to 0.3 to match slower globe rotation
            maxPolarAngle={Math.PI * 0.8} // Allow more vertical rotation
            minPolarAngle={Math.PI * 0.2} // Allow more vertical rotation
            // Remove azimuth limits for full 360-degree horizontal rotation
            maxAzimuthAngle={Infinity}
            minAzimuthAngle={-Infinity}
          />
        )}

        {/* Globe Animation */}
        <Suspense fallback={null}>
          <NetworkGlobe isLoaded={isLoaded} />
        </Suspense>
      </Canvas>
    </div>
  );
};

export default GlobeAnimation;