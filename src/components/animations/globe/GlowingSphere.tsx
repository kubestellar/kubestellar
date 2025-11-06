import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

// Glowing sphere effect
interface GlowingSphereProps {
  position?: [number, number, number];
  color: string;
  size?: number;
  intensity?: number;
}

const GlowingSphere = ({ 
  position = [0, 0, 0], 
  color, 
  size = 0.3, 
  intensity = 1.0 
}: GlowingSphereProps) => {
  const meshRef = useRef<THREE.Mesh>(null);
  
  useFrame(({ clock }) => {
    if (meshRef.current) {
      const t = clock.getElapsedTime();
      meshRef.current.scale.setScalar(1 + Math.sin(t * 2) * 0.1 * intensity);
    }
  });
  
  return (
    <group position={position}>
      {/* Core sphere */}
      <mesh ref={meshRef}>
        <sphereGeometry args={[size, 32, 32]} />
        <meshBasicMaterial color={color} />
      </mesh>
      
      {/* Outer glow */}
      <mesh>
        <sphereGeometry args={[size * 1.2, 32, 32]} />
        <meshBasicMaterial 
          color={color} 
          transparent 
          opacity={0.3 * intensity} 
          depthWrite={false}
        />
      </mesh>
      
      {/* Brightest inner glow */}
      <mesh>
        <sphereGeometry args={[size * 0.8, 32, 32]} />
        <meshBasicMaterial 
          color="white" 
          transparent 
          opacity={0.7 * intensity} 
          depthWrite={false}
        />
      </mesh>
    </group>
  );
};

export default GlowingSphere;