import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";

/**
 * GlowingSphere Component
 *
 * Creates a sphere with multiple layers to create a glowing effect.
 * Features a solid core, a bright inner glow, and a soft outer glow.
 *
 * @param position - 3D position of the sphere
 * @param color - Color of the sphere and glow
 * @param size - Base size of the sphere
 * @param intensity - Intensity multiplier for glow effect
 */

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
  intensity = 1.0,
}: GlowingSphereProps) => {
  const meshRef = useRef<THREE.Mesh>(null);
  const frameCount = useRef(0);

  // Create shared geometries to reduce draw calls
  const coreGeometry = useMemo(
    () => new THREE.SphereGeometry(size, 16, 16),
    [size]
  );
  const outerGeometry = useMemo(
    () => new THREE.SphereGeometry(size * 1.2, 12, 12),
    [size]
  );
  const innerGeometry = useMemo(
    () => new THREE.SphereGeometry(size * 0.8, 12, 12),
    [size]
  );

  // Create shared materials
  const coreMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: color,
      }),
    [color]
  );

  const outerMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: color,
        transparent: true,
        opacity: 0.3 * intensity,
        depthWrite: false,
      }),
    [color, intensity]
  );

  const innerMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: "white",
        transparent: true,
        opacity: 0.7 * intensity,
        depthWrite: false,
      }),
    [intensity]
  );

  // Optimize animation to run less frequently
  useFrame(({ clock }) => {
    // Skip frames for better performance
    frameCount.current += 1;
    if (frameCount.current % 2 !== 0) return;

    if (meshRef.current) {
      const t = clock.getElapsedTime();
      // Gentle pulsing animation
      meshRef.current.scale.setScalar(1 + Math.sin(t * 2) * 0.1 * intensity);
    }
  });

  return (
    <group position={position} frustumCulled>
      {/* Core sphere */}
      <mesh ref={meshRef} geometry={coreGeometry} material={coreMaterial} />

      {/* Outer glow */}
      <mesh geometry={outerGeometry} material={outerMaterial} />

      {/* Brightest inner glow */}
      <mesh geometry={innerGeometry} material={innerMaterial} />
    </group>
  );
};

export default GlowingSphere;
