import { useRef, useState, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";

/**
 * DataPacket Component
 *
 * Creates an animated sphere that travels along a path between two points.
 * Includes a particle trail effect.
 *
 * @param path - Array of 3D positions defining the path [start, end]
 * @param speed - Speed multiplier for animation
 * @param color - Color of the data packet
 * @param size - Size of the packet sphere
 * @param isActive - Controls whether animation is running
 */

interface DataPacketProps {
  path: [number, number, number][];
  speed?: number;
  color?: string;
  size?: number;
  isActive?: boolean;
}

const DataPacket = ({
  path,
  speed = 1,
  color = "#00E396",
  size = 0.08,
  isActive = true,
}: DataPacketProps) => {
  const ref = useRef<THREE.Mesh>(null);
  const [progress, setProgress] = useState(0);
  const trailRef = useRef<THREE.Points>(null);

  // Generate trail points - reduced particle count
  const trailPositions = useMemo(() => {
    const count = 10;
    const positions = new Float32Array(count * 3);
    return positions;
  }, []);

  // Shared geometry for all packets
  const packetGeometry = useMemo(() => {
    // Use lower poly count (8x8 instead of 16x16)
    return new THREE.SphereGeometry(size, 8, 8);
  }, [size]);

  // Create material once
  const packetMaterial = useMemo(() => {
    return new THREE.MeshBasicMaterial({
      color: color,
      transparent: true,
      opacity: 0.9,
    });
  }, [color]);

  // Optimize frame updates - only update every other frame
  useFrame(state => {
    // Skip animation entirely when not active
    if (!isActive) return;

    // Skip every other frame for better performance
    if (Math.floor(state.clock.getElapsedTime() * 30) % 2 !== 0) return;

    // Update packet position
    setProgress(prev => (prev >= 1 ? 0 : prev + 0.005 * speed));

    // Update trail positions
    if (trailRef.current && ref.current && path.length >= 2) {
      const positions = trailRef.current.geometry.attributes.position
        .array as Float32Array;
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
      start[2] + (end[2] - start[2]) * progress,
    ] as [number, number, number];
  }, [path, progress]);

  return (
    <group>
      {/* Simple trail with optimized settings */}
      <points ref={trailRef} frustumCulled>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            args={[trailPositions, 3, false]}
          />
        </bufferGeometry>
        <pointsMaterial
          color={color}
          size={size * 0.6}
          transparent
          opacity={0.5}
          sizeAttenuation
          depthWrite={false}
        />
      </points>

      {/* Main data packet - using shared geometry and materials */}
      <mesh
        ref={ref}
        position={position}
        frustumCulled
        geometry={packetGeometry}
        material={packetMaterial}
      />
    </group>
  );
};

export default DataPacket;
