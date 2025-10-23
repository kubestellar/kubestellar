import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import { Text, Billboard } from "@react-three/drei";
import * as THREE from "three";
import { COLORS } from "./colors";

/**
 * LogoElement Component
 *
 * Creates the central logo element with a pulsing sphere and rotating ring.
 * Displays KubeStellar branding in the center.
 *
 * @param animate - Controls whether animation is running
 */

interface LogoElementProps {
  animate?: boolean;
}

const LogoElement = ({ animate = true }: LogoElementProps) => {
  const coreRef = useRef<THREE.Mesh>(null);
  const outerRingRef = useRef<THREE.Mesh>(null);
  const frameCount = useRef(0);

  // Create shared materials for better performance
  const coreMaterial = useMemo(
    () =>
      new THREE.MeshPhongMaterial({
        color: COLORS.highlight,
        emissive: COLORS.highlight,
        emissiveIntensity: 0.5,
        shininess: 100,
        transparent: true,
      }),
    []
  );

  const ringMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: COLORS.primary,
        transparent: true,
        wireframe: true,
      }),
    []
  );

  // Create shared geometries for better performance
  const coreGeometry = useMemo(() => new THREE.SphereGeometry(0.3, 16, 16), []);
  const ringGeometry = useMemo(
    () => new THREE.TorusGeometry(0.6, 0.02, 8, 48),
    []
  );

  // Optimize animation frequency
  useFrame(state => {
    // Skip animation when not animate
    if (!animate) return;

    // Skip frames to improve performance
    frameCount.current += 1;
    if (frameCount.current % 2 !== 0) return;

    const time = state.clock.getElapsedTime();

    // Core pulsing effect
    if (coreRef.current) {
      const scale = 1 + Math.sin(time * 2) * 0.05;
      coreRef.current.scale.setScalar(scale);

      // Update material opacity instead of changing instance
      if (coreRef.current.material) {
        (
          coreRef.current.material as THREE.MeshPhongMaterial
        ).emissiveIntensity = 0.5 + Math.sin(time * 2) * 0.2;
      }
    }

    // Rotate outer ring
    if (outerRingRef.current) {
      outerRingRef.current.rotation.z = time * 0.5;
      outerRingRef.current.rotation.x = Math.sin(time * 0.2) * 0.2;
    }
  });

  return (
    <group>
      {/* Central core sphere */}
      <mesh
        ref={coreRef}
        geometry={coreGeometry}
        material={coreMaterial}
        frustumCulled
      />

      {/* Outer rotating ring */}
      <mesh
        ref={outerRingRef}
        geometry={ringGeometry}
        material={ringMaterial}
        frustumCulled
      />

      {/* KubeStellar Branding */}
      <Billboard position={[0, 1, 0]} frustumCulled>
        <Text
          fontSize={0.2}
          color={COLORS.highlight}
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.01}
          outlineColor={COLORS.background}
          fillOpacity={animate ? 1 : 0}
          characters="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!"
        >
          KubeStellar
        </Text>
        <Text
          position={[0, -0.25, 0]}
          fontSize={0.1}
          color={COLORS.primary}
          anchorX="center"
          anchorY="middle"
          fillOpacity={animate ? 1 : 0}
          characters="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
        >
          Multi-Cluster Management
        </Text>
      </Billboard>
    </group>
  );
};

export default LogoElement;
