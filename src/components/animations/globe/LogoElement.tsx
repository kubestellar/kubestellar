import { useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import GlowingSphere from "./GlowingSphere";

// KubeStellar theme colors
const COLORS = {
  primary: "#1a90ff", // Main blue color
  secondary: "#6236FF", // Purple accent
  highlight: "#00C2FF", // Bright blue for highlights
  success: "#00E396", // Green for active connections
  background: "#0a0f1c", // Dark background
  accent1: "#FF5E84", // Accent color for special elements
  accent2: "#FFD166", // Secondary accent for highlights
};

// KubeStellar logo element
interface LogoElementProps {
  position?: [number, number, number];
  rotation?: [number, number, number];
  scale?: number;
}

const LogoElement = ({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
  scale = 1,
}: LogoElementProps) => {
  const groupRef = useRef<THREE.Group>(null);
  const ringRef1 = useRef<THREE.Mesh>(null);
  const ringRef2 = useRef<THREE.Mesh>(null);
  const ringRef3 = useRef<THREE.Mesh>(null);

  useFrame(state => {
    const t = state.clock.getElapsedTime();

    if (groupRef.current) {
      groupRef.current.rotation.y = t * 0.2;
      groupRef.current.rotation.z = Math.sin(t * 0.5) * 0.1;
    }

    // Animate rings independently
    if (ringRef1.current) {
      ringRef1.current.rotation.x = t * 0.5;
      ringRef1.current.rotation.z = t * 0.2;
    }

    if (ringRef2.current) {
      ringRef2.current.rotation.x = -t * 0.3;
      ringRef2.current.rotation.y = t * 0.4;
    }

    if (ringRef3.current) {
      ringRef3.current.rotation.y = t * 0.2;
      ringRef3.current.rotation.z = -t * 0.3;
    }
  });

  return (
    <group ref={groupRef} position={position} rotation={rotation} scale={scale}>
      {/* Central glowing sphere */}
      <GlowingSphere
        position={[0, 0, 0]}
        color={COLORS.secondary}
        size={0.25}
      />

      {/* Orbital rings representing the stellar aspect */}
      <mesh ref={ringRef1}>
        <torusGeometry args={[0.6, 0.02, 16, 100]} />
        <meshPhongMaterial
          color={COLORS.primary}
          emissive={COLORS.primary}
          emissiveIntensity={0.5}
        />
      </mesh>

      <mesh ref={ringRef2}>
        <torusGeometry args={[0.7, 0.02, 16, 100]} />
        <meshPhongMaterial
          color={COLORS.highlight}
          emissive={COLORS.highlight}
          emissiveIntensity={0.5}
        />
      </mesh>

      <mesh ref={ringRef3}>
        <torusGeometry args={[0.5, 0.02, 16, 100]} />
        <meshPhongMaterial
          color={COLORS.accent1}
          emissive={COLORS.accent1}
          emissiveIntensity={0.5}
        />
      </mesh>

      {/* Small orbiting particles */}
      {Array.from({ length: 8 }).map((_, i) => {
        const angle = (i / 8) * Math.PI * 2;
        const radius = 0.8;
        const x = Math.cos(angle) * radius;
        const y = Math.sin(angle) * radius;
        const color = i % 2 === 0 ? COLORS.highlight : COLORS.accent2;

        return (
          <mesh key={i} position={[x, y, 0]}>
            <sphereGeometry args={[0.03, 8, 8]} />
            <meshBasicMaterial color={color} />
          </mesh>
        );
      })}
    </group>
  );
};

export default LogoElement;
