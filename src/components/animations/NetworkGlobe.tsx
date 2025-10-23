"use client";

import { useRef, useMemo, useState, useEffect, useCallback } from "react";
import { useFrame } from "@react-three/fiber";
import { Sphere, Torus } from "@react-three/drei";
import * as THREE from "three";
import { COLORS } from "./globe/colors";
import DataPacket from "./globe/DataPacket";
import LogoElement from "./globe/LogoElement";
import Cluster from "./globe/Cluster";

// Component props interface
interface NetworkGlobeProps {
  isLoaded?: boolean;
}

// Define interfaces for better type safety
interface FlowMaterial extends THREE.Material {
  opacity: number;
  color: THREE.Color;
  dashSize?: number;
  gapSize?: number;
}

interface FlowChild extends THREE.Object3D {
  material?: FlowMaterial;
}

interface CentralNodeChild extends THREE.Object3D {
  material?: THREE.Material & { opacity?: number };
}

/**
 * NetworkGlobe Component
 *
 * A 3D animated globe visualization showing networked clusters with data flows.
 * Features include:
 * - Rotating wireframe globe
 * - Multiple node clusters
 * - Animated data packets
 * - Performance optimizations
 *
 * @param isLoaded - Controls animation start and reveal effect
 */
const NetworkGlobe = ({ isLoaded = true }: NetworkGlobeProps) => {
  const globeRef = useRef<THREE.Mesh>(null);
  const centralNodeRef = useRef<THREE.Group>(null);
  const dataFlowsRef = useRef<THREE.Group>(null);
  const frameCount = useRef(0);

  // Track document visibility to pause rendering when tab/page is not active
  const [isDocumentVisible, setIsDocumentVisible] = useState(true);

  // Animation state for data flows
  const [activeFlows, setActiveFlows] = useState<number[]>([]);
  const [animationProgress, setAnimationProgress] = useState(0);

  // Handle visibility change to pause rendering
  useEffect(() => {
    const handleVisibilityChange = () => {
      setIsDocumentVisible(!document.hidden);
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  // Create cluster configurations for KubeStellar
  const clusters = useMemo(
    () => [
      {
        name: "Edge Clusters",
        position: [0, 3, 0] as [number, number, number],
        nodeCount: 6,
        radius: 0.8,
        color: COLORS.primary,
        description: "Edge computing nodes",
      },
      {
        name: "Workload Clusters",
        position: [3, 0, 0] as [number, number, number],
        nodeCount: 8,
        radius: 1,
        color: COLORS.aiInference,
        description: "Application workload clusters",
      },
      {
        name: "Control Plane",
        position: [0, -3, 0] as [number, number, number],
        nodeCount: 5,
        radius: 0.7,
        color: COLORS.aiTraining,
        description: "Control plane clusters",
      },
      {
        name: "Service Mesh",
        position: [-3, 0, 0] as [number, number, number],
        nodeCount: 7,
        radius: 0.9,
        color: COLORS.accent2,
        description: "Service mesh nodes",
      },
      {
        name: "Compute",
        position: [2, 2, -2] as [number, number, number],
        nodeCount: 4,
        radius: 0.6,
        color: COLORS.success,
        description: "Compute resources",
      },
    ],
    []
  );

  // Generate data flow paths
  const dataFlows = useMemo(() => {
    const flows: {
      path: [number, number, number][];
      id: number;
      type: string;
    }[] = [];
    const centralPos: [number, number, number] = [0, 0, 0];

    // Connect central node to each cluster
    clusters.forEach((cluster, clusterIdx) => {
      flows.push({
        path: [centralPos, cluster.position],
        id: clusterIdx,
        type: "control",
      });
    });

    // Add cross-cluster connections
    flows.push({
      path: [clusters[2].position, clusters[1].position],
      id: clusters.length + 1,
      type: "model",
    });

    flows.push({
      path: [clusters[0].position, clusters[1].position],
      id: clusters.length + 2,
      type: "inference",
    });

    // Add fewer cross-cluster connections
    const maxConnections = 3;
    let connectionCount = 0;

    for (
      let i = 0;
      i < clusters.length && connectionCount < maxConnections;
      i++
    ) {
      for (
        let j = i + 1;
        j < clusters.length && connectionCount < maxConnections;
        j++
      ) {
        if (Math.random() > 0.5) {
          flows.push({
            path: [clusters[i].position, clusters[j].position],
            id: clusters.length + i * 10 + j,
            type: "data",
          });
          connectionCount++;
        }
      }
    }

    return flows;
  }, [clusters]);

  // Shared materials for better performance
  const globeMaterial = useMemo(
    () =>
      new THREE.MeshPhongMaterial({
        color: COLORS.primary,
        transparent: true,
        wireframe: true,
        depthWrite: false,
      }),
    []
  );

  const gridMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: COLORS.primary,
        transparent: true,
        depthWrite: false,
      }),
    []
  );

  // Animate data flows
  useEffect(() => {
    if (!isLoaded || !isDocumentVisible) return;

    const interval = setInterval(() => {
      const numFlows = Math.min(Math.floor(dataFlows.length / 3), 3);
      const randomFlows = Array.from({ length: numFlows }, () =>
        Math.floor(Math.random() * dataFlows.length)
      );
      setActiveFlows(randomFlows);
    }, 4000);

    return () => clearInterval(interval);
  }, [dataFlows.length, isLoaded, isDocumentVisible]);

  // Optimize update function
  const updateOpacity = useCallback(
    (material: FlowMaterial, targetOpacity: number, type: string) => {
      if (!material) return;

      material.opacity = THREE.MathUtils.lerp(
        material.opacity,
        targetOpacity,
        0.1
      );

      if (type === "model") {
        material.color.set(COLORS.aiTraining);
      } else if (type === "inference") {
        material.color.set(COLORS.aiInference);
      } else if (type === "control") {
        material.color.set(COLORS.secondary);
      } else {
        material.color.set(COLORS.success);
      }

      if (material.dashSize !== undefined) {
        material.dashSize = type === "control" ? 0.1 : 0.05;
      }
      if (material.gapSize !== undefined) {
        material.gapSize = type === "control" ? 0.05 : 0.1;
      }
    },
    []
  );

  // Animation frame updates with performance optimizations
  useFrame(state => {
    if (!isDocumentVisible) return;

    frameCount.current += 1;

    if (isLoaded && animationProgress < 1) {
      setAnimationProgress(Math.min(animationProgress + 0.01, 1));
    }

    if (globeRef.current && frameCount.current % 2 === 0) {
      globeRef.current.rotation.y = state.clock.getElapsedTime() * 0.05;
      globeRef.current.rotation.x =
        Math.sin(state.clock.getElapsedTime() * 0.2) * 0.02;

      if (globeMaterial.opacity !== 0.08 * animationProgress) {
        globeMaterial.opacity = 0.08 * animationProgress;
      }

      const scale = isLoaded ? 1 * animationProgress : 0.5;
      globeRef.current.scale.setScalar(scale);
    }

    if (centralNodeRef.current && frameCount.current % 3 === 0) {
      centralNodeRef.current.rotation.y = state.clock.getElapsedTime() * 0.2;
      centralNodeRef.current.scale.setScalar(
        (1 + Math.sin(state.clock.getElapsedTime() * 1.5) * 0.05) *
          animationProgress
      );

      centralNodeRef.current.children.forEach((child: CentralNodeChild) => {
        if (child.material && typeof child.material.opacity !== "undefined") {
          child.material.opacity = THREE.MathUtils.lerp(
            child.material.opacity || 0,
            animationProgress,
            0.1
          );
        }
      });
    }

    if (dataFlowsRef.current && frameCount.current % 3 === 0) {
      dataFlowsRef.current.children.forEach((flow: FlowChild, i) => {
        if (flow.material) {
          const flowData = dataFlows[i];
          const flowType = flowData?.type || "data";

          const targetOpacity = activeFlows.includes(i)
            ? 0.8 * animationProgress
            : 0.1 * animationProgress;

          updateOpacity(flow.material, targetOpacity, flowType);
        }
      });
    }
  });

  return (
    <group>
      {/* Main globe - represents the global network */}
      <Sphere ref={globeRef} args={[3.5, 32, 32]} frustumCulled>
        <primitive object={globeMaterial} />
      </Sphere>

      {/* Grid lines for the globe */}
      <group rotation={[0, 0, 0]}>
        {Array.from({ length: 4 }).map((_, idx) => (
          <Torus
            key={`h-${idx}`}
            args={[3.5, 0.01, 8, 50]}
            rotation={[0, 0, (Math.PI * idx) / 4]}
            frustumCulled
          >
            <primitive
              object={gridMaterial.clone()}
              attach="material"
              opacity={0.1 * animationProgress}
            />
          </Torus>
        ))}
        {Array.from({ length: 4 }).map((_, idx) => (
          <Torus
            key={`v-${idx}`}
            args={[3.5, 0.01, 8, 50]}
            rotation={[Math.PI / 2, (Math.PI * idx) / 4, 0]}
            frustumCulled
          >
            <primitive
              object={gridMaterial.clone()}
              attach="material"
              opacity={0.1 * animationProgress}
            />
          </Torus>
        ))}
      </group>

      {/* Central hub */}
      <group ref={centralNodeRef}>
        <LogoElement animate={isLoaded && isDocumentVisible} />
      </group>

      {/* Clusters of nodes */}
      {clusters.map((cluster, idx) => (
        <Cluster key={idx} {...cluster} isActive={isDocumentVisible} />
      ))}

      {/* Data flow connections */}
      <group ref={dataFlowsRef}>
        {dataFlows.map((flow, idx) => (
          <group key={idx} frustumCulled>
            {activeFlows.includes(idx) && isDocumentVisible && (
              <DataPacket
                path={flow.path}
                speed={flow.type === "control" ? 1.5 : 1}
                color={
                  flow.type === "model"
                    ? COLORS.aiTraining
                    : flow.type === "inference"
                      ? COLORS.aiInference
                      : flow.type === "control"
                        ? COLORS.secondary
                        : COLORS.success
                }
                isActive={isDocumentVisible}
              />
            )}
          </group>
        ))}
      </group>
    </group>
  );
};

export default NetworkGlobe;
