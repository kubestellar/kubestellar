"use client";

import { Canvas } from "@react-three/fiber";
import { OrbitControls } from "@react-three/drei";
import { Suspense, useState, useEffect } from "react";
import NetworkGlobe from "./NetworkGlobe";

/**
 * GlobeAnimation Component - Wrapper for NetworkGlobe
 *
 * Main component to use in your hero section
 * Handles Canvas setup and loading states
 */

interface GlobeAnimationProps {
  /** Width of the canvas container (default: "100%") */
  width?: string | number;
  /** Height of the canvas container (default: "500px") */
  height?: string | number;
  /** Custom CSS class for container */
  className?: string;
  /** Custom inline styles for container */
  style?: React.CSSProperties;
  /** Show loading indicator (default: true) */
  showLoader?: boolean;
  /** Enable orbit controls for user interaction (default: true) */
  enableControls?: boolean;
  /** Enable pan controls (default: false) */
  enablePan?: boolean;
  /** Auto-rotate the globe (default: false, will use internal rotation if false) */
  autoRotate?: boolean;
  /** Auto-rotate speed (default: 0.5) */
  autoRotateSpeed?: number;
  /** Camera field of view (default: 50) */
  fov?: number;
  /** Camera position [x, y, z] (default: [0, 0, 10]) */
  cameraPosition?: [number, number, number];
}

const GlobeAnimation: React.FC<GlobeAnimationProps> = ({
  width = "100%",
  height = "500px",
  className = "",
  style = {},
  showLoader = true,
  enableControls = true,
  enablePan = false,
  autoRotate = false,
  autoRotateSpeed = 0.5,
  fov = 50,
  cameraPosition = [0, 0, 10],
}) => {
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    // Simulate loading and then set to loaded
    const timer = setTimeout(() => {
      setIsLoaded(true);
    }, 500);

    return () => clearTimeout(timer);
  }, []);

  const containerStyle: React.CSSProperties = {
    width,
    height,
    position: "relative",
    ...style,
  };

  return (
    <div className={className} style={containerStyle}>
      {/* Loading indicator */}
      {showLoader && !isLoaded && (
        <div
          style={{
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            color: "white",
            fontSize: "16px",
            zIndex: 10,
          }}
        >
          Loading Globe...
        </div>
      )}

      {/* Canvas with 3D Globe */}
      <Canvas
        camera={{ position: cameraPosition, fov: fov }}
        style={{ background: "transparent" }}
      >
        {/* Lighting for the scene */}
        <ambientLight intensity={0.5} />
        <pointLight position={[10, 10, 10]} />

        {/* Orbit Controls for user interaction */}
        {enableControls && (
          <OrbitControls
            enablePan={enablePan}
            enableZoom={false}
            enableRotate={true}
            autoRotate={autoRotate}
            autoRotateSpeed={autoRotateSpeed}
            // Smooth damping for better feel
            enableDamping={true}
            dampingFactor={0.05}
            // Prevent flipping
            maxPolarAngle={Math.PI}
            minPolarAngle={0}
          />
        )}

        {/* Globe Component with Suspense for loading */}
        <Suspense fallback={null}>
          <NetworkGlobe isLoaded={isLoaded} />
        </Suspense>
      </Canvas>
    </div>
  );
};

export default GlobeAnimation;
