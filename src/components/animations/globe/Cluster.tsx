import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import {
  Sphere,
  Text,
  Billboard,
  Instances,
  Instance,
} from "@react-three/drei";
import * as THREE from "three";

// Theme colors
const COLORS = {
  primary: "#1a90ff",
  secondary: "#6236FF",
  highlight: "#00C2FF",
  success: "#00E396",
  background: "#0a0f1c",
};

/**
 * Cluster Component
 *
 * Creates a cluster of nodes arranged in a sphere, with connections between them.
 * Nodes randomly activate to show activity.
 *
 * @param position - 3D position of the cluster center
 * @param name - Display name of the cluster
 * @param nodeCount - Number of nodes in the cluster
 * @param radius - Radius of the cluster
 * @param color - Color theme for the cluster
 * @param description - Optional description shown on hover
 * @param isActive - Controls whether animation is running
 */

interface ClusterProps {
  position?: [number, number, number];
  name: string;
  nodeCount: number;
  radius: number;
  color: string;
  description?: string;
  isActive?: boolean;
}

const Cluster = ({
  position = [0, 0, 0],
  name,
  nodeCount,
  radius,
  color,
  description,
  isActive = true,
}: ClusterProps) => {
  const clusterRef = useRef<THREE.Group>(null);
  const [activeNodes, setActiveNodes] = useState<number[]>([]);
  const [hovered, setHovered] = useState(false);
  const frameCount = useRef(0);

  // Generate nodes in a sphere distribution
  const nodes = useMemo(() => {
    return Array.from({ length: nodeCount }, (_, i) => {
      const phi = Math.acos(-1 + (2 * i) / nodeCount);
      const theta = Math.sqrt(nodeCount * Math.PI) * phi;

      return [
        radius * Math.cos(theta) * Math.sin(phi),
        radius * Math.sin(theta) * Math.sin(phi),
        radius * Math.cos(phi),
      ] as [number, number, number];
    });
  }, [nodeCount, radius]);

  // Generate connection lines between nodes (reduced for performance)
  const connections = useMemo(() => {
    const lines: [[number, number, number], [number, number, number]][] = [];

    for (let i = 0; i < nodes.length; i += 2) {
      for (let j = i + 1; j < nodes.length; j += 3) {
        lines.push([nodes[i], nodes[j]]);
      }
    }

    return lines;
  }, [nodes]);

  // Randomly activate nodes periodically
  useEffect(() => {
    if (!isActive) return;

    const interval = setInterval(() => {
      const randomNodes = Array.from(
        { length: Math.min(3, Math.floor(nodeCount / 4)) },
        () => Math.floor(Math.random() * nodeCount)
      );
      setActiveNodes(randomNodes);
    }, 4000);

    return () => clearInterval(interval);
  }, [nodeCount, isActive]);

  // Animation updates
  useFrame(state => {
    if (!isActive) return;

    frameCount.current += 1;

    // Update rotation every 2 frames
    if (frameCount.current % 2 === 0 && clusterRef.current) {
      clusterRef.current.rotation.y = state.clock.getElapsedTime() * 0.1;
      clusterRef.current.rotation.x =
        Math.sin(state.clock.getElapsedTime() * 0.2) * 0.05;
    }

    // Scale effect on hover
    if (clusterRef.current) {
      const targetScale = hovered ? 1.05 : 1;
      clusterRef.current.scale.lerp(
        new THREE.Vector3(targetScale, targetScale, targetScale),
        0.1
      );
    }
  });

  // Shared materials
  const nodeMaterial = useMemo(
    () =>
      new THREE.MeshPhongMaterial({
        color: color,
        emissive: color,
        emissiveIntensity: 0.3,
        shininess: 50,
      }),
    [color]
  );

  const lineMaterial = useMemo(
    () =>
      new THREE.LineBasicMaterial({
        color: color,
        transparent: true,
        opacity: 0.2,
      }),
    [color]
  );

  return (
    <group
      position={position}
      onPointerOver={() => setHovered(true)}
      onPointerOut={() => setHovered(false)}
    >
      {/* Cluster boundary sphere */}
      <Sphere args={[radius * 1.2, 16, 16]} frustumCulled>
        <meshPhongMaterial
          color={color}
          transparent
          opacity={hovered ? 0.25 : 0.15}
          wireframe
          emissive={color}
          emissiveIntensity={hovered ? 0.3 : 0.1}
          depthWrite={false}
        />
      </Sphere>

      {/* Cluster name */}
      <Billboard position={[0, radius * 1.4, 0]} frustumCulled>
        <Text
          fontSize={0.18}
          color={color}
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.01}
          outlineColor={COLORS.background}
        >
          {name}
        </Text>

        {/* Description (only shown when hovered) */}
        {description && hovered && (
          <Text
            position={[0, 0.2, 0]}
            fontSize={0.1}
            color="white"
            anchorX="center"
            anchorY="middle"
            outlineWidth={0.005}
            outlineColor={COLORS.background}
            maxWidth={2}
            textAlign="center"
          >
            {description}
          </Text>
        )}
      </Billboard>

      {/* Node group */}
      <group ref={clusterRef}>
        {/* Connection lines */}
        {connections.map((points, i) => (
          <primitive
            key={i}
            object={
              new THREE.Line(
                new THREE.BufferGeometry().setFromPoints([
                  new THREE.Vector3(...points[0]),
                  new THREE.Vector3(...points[1]),
                ]),
                lineMaterial
              )
            }
            frustumCulled
          />
        ))}

        {/* Instanced nodes for performance */}
        <Instances limit={nodeCount} range={nodeCount}>
          <sphereGeometry args={[0.08, 8, 8]} />
          <primitive object={nodeMaterial} />

          {nodes.map((nodePos, idx) => (
            <Instance
              key={idx}
              position={nodePos}
              scale={activeNodes.includes(idx) ? 1.2 : 1}
              color={activeNodes.includes(idx) ? COLORS.success : color}
            />
          ))}
        </Instances>
      </group>
    </group>
  );
};

export default Cluster;
