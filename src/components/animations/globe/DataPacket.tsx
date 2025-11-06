import { useRef, useState, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

// Data packet that travels along connections
interface DataPacketProps {
  path: [number, number, number][];
  speed?: number;
  color?: string;
  size?: number;
}

const DataPacket = ({ path, speed = 1, color = "#00E396", size = 0.08 }: DataPacketProps) => {
  const ref = useRef<THREE.Mesh>(null);
  const [progress, setProgress] = useState(0);
  const trailRef = useRef<THREE.Points>(null);
  
  // Generate trail points
  const trailPositions = useMemo(() => {
    const count = 20;
    const positions = new Float32Array(count * 3);
    return positions;
  }, []);
  
  useFrame(() => {
    // Update packet position
    setProgress((prev) => (prev >= 1 ? 0 : prev + 0.005 * speed));
    
    // Update trail positions
    if (trailRef.current && ref.current && path.length >= 2) {
      const positions = trailRef.current.geometry.attributes.position.array as Float32Array;
      const start = path[0];
      const end = path[1];
      
      // Current position
      const x = start[0] + (end[0] - start[0]) * progress;
      const y = start[1] + (end[1] - start[1]) * progress;
      const z = start[2] + (end[2] - start[2]) * progress;
      
      // Shift all positions forward
      for (let i = positions.length - 3; i >= 3; i -= 3) {
        positions[i] = positions[i - 3];
        positions[i + 1] = positions[i - 2];
        positions[i + 2] = positions[i - 1];
      }
      
      // Set the first position to current position
      positions[0] = x;
      positions[1] = y;
      positions[2] = z;
      
      // Update the buffer attribute
      trailRef.current.geometry.attributes.position.needsUpdate = true;
    }
  });
  
  // Calculate position along the path
  const position = useMemo(() => {
    if (path.length < 2) return [0, 0, 0] as [number, number, number];
    
    const start = path[0];
    const end = path[1];
    
    return [
      start[0] + (end[0] - start[0]) * progress,
      start[1] + (end[1] - start[1]) * progress,
      start[2] + (end[2] - start[2]) * progress
    ] as [number, number, number];
  }, [path, progress]);
  
  return (
    <group>
      {/* Simple trail */}
      <points ref={trailRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            args={[trailPositions, 3, false]}
          />
        </bufferGeometry>
        <pointsMaterial 
          color={color} 
          size={size * 0.8} 
          transparent 
          opacity={0.6}
          sizeAttenuation
        />
      </points>
      
      {/* Main data packet */}
      <mesh ref={ref} position={position}>
        <sphereGeometry args={[size, 16, 16]} />
        <meshBasicMaterial color={color} />
      </mesh>
    </group>
  );
};

export default DataPacket;