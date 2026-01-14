import { useRef, useMemo, useState, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import { Sphere, Line, Text, Torus, Billboard } from "@react-three/drei";
import * as THREE from "three";
import { useTranslations } from "next-intl";
import { COLORS } from "./colors";
import DataPacket from "./DataPacket";
import LogoElement from "./LogoElement";
import Cluster from "./Cluster";

// Add this interface for the component props
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

// Update the main component to accept props
const NetworkGlobe = ({ isLoaded = true }: NetworkGlobeProps) => {
  const t = useTranslations("networkGlobe");
  const globeRef = useRef<THREE.Mesh>(null);
  const gridLinesRef = useRef<THREE.Group>(null);
  const centralNodeRef = useRef<THREE.Group>(null);
  const dataFlowsRef = useRef<THREE.Group>(null);
  const rotatingContentRef = useRef<THREE.Group>(null);

  // Animation state for data flows
  const [activeFlows, setActiveFlows] = useState<number[]>([]);
  const [animationProgress, setAnimationProgress] = useState(0);

  // Create cluster configurations with KubeStellar-related names and descriptions
  const clusters = useMemo(
    () => [
      {
        name: t("clusters.kubeflexCore.name"),
        position: [0, 3, 0] as [number, number, number],
        nodeCount: 6,
        radius: 0.8,
        color: COLORS.primary,
        description: t("clusters.kubeflexCore.description"),
      },
      {
        name: t("clusters.edgeClusters.name"),
        position: [3, 0, 0] as [number, number, number],
        nodeCount: 8,
        radius: 1,
        color: COLORS.highlight,
        description: t("clusters.edgeClusters.description"),
      },
      {
        name: t("clusters.productionCluster.name"),
        position: [0, -3, 0] as [number, number, number],
        nodeCount: 5,
        radius: 0.7,
        color: COLORS.success,
        description: t("clusters.productionCluster.description"),
      },
      {
        name: t("clusters.devTestCluster.name"),
        position: [-3, 0, 0] as [number, number, number],
        nodeCount: 7,
        radius: 0.9,
        color: COLORS.accent2,
        description: t("clusters.devTestCluster.description"),
      },
      {
        name: t("clusters.multiCloudHub.name"),
        position: [2, 2, -2] as [number, number, number],
        nodeCount: 4,
        radius: 0.6,
        color: COLORS.accent1,
        description: t("clusters.multiCloudHub.description"),
      },
    ],
    [t]
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

    // Add some cross-cluster connections with specific types
    // Production to Edge (workload distribution)
    flows.push({
      path: [clusters[2].position, clusters[1].position],
      id: clusters.length + 1,
      type: "workload",
    });

    // KubeFlex to Edge (control commands)
    flows.push({
      path: [clusters[0].position, clusters[1].position],
      id: clusters.length + 2,
      type: "control",
    });

    // Dev/Test to Production (deployment pipeline)
    flows.push({
      path: [clusters[3].position, clusters[2].position],
      id: clusters.length + 3,
      type: "deploy",
    });

    // Add some other cross-cluster connections
    for (let i = 0; i < clusters.length; i++) {
      for (let j = i + 1; j < clusters.length; j++) {
        if (Math.random() > 0.7) {
          flows.push({
            path: [clusters[i].position, clusters[j].position],
            id: clusters.length + i * 10 + j,
            type: "data",
          });
        }
      }
    }

    return flows;
  }, [clusters]);

  // Animate data flows - only start when loaded
  useEffect(() => {
    if (!isLoaded) return;

    const interval = setInterval(() => {
      const randomFlows = Array.from(
        { length: Math.floor(dataFlows.length / 2) },
        () => Math.floor(Math.random() * dataFlows.length)
      );
      setActiveFlows(randomFlows);
    }, 4000);

    return () => clearInterval(interval);
  }, [dataFlows.length, isLoaded]);

  // Animation frame updates with progressive reveal
  useFrame(state => {
    const time = state.clock.getElapsedTime();

    // Update animation progress for reveal effect
    if (isLoaded && animationProgress < 1) {
      setAnimationProgress(Math.min(animationProgress + 0.01, 1));
    }

    // Rotate the globe and grid lines together with slower speed to match clusters
    if (globeRef.current) {
      // Slower Y-axis rotation to match cluster speed
      globeRef.current.rotation.y = time * 0.1; // Reduced from 0.3 to 0.1

      // Subtle X-axis tilt for dynamic movement
      globeRef.current.rotation.x = Math.sin(time * 0.15) * 0.08; // Reduced speed and amplitude

      // Optional: Add slight Z-axis rotation for even more dynamic movement
      globeRef.current.rotation.z = Math.cos(time * 0.08) * 0.03; // Reduced speed and amplitude

      // Fixed scale - no zoom effect
      const scale = isLoaded ? 1 * animationProgress : 0.5;
      globeRef.current.scale.setScalar(scale);
    }

    // Rotate grid lines to match globe rotation with same slow speed
    if (gridLinesRef.current) {
      gridLinesRef.current.rotation.y = time * 0.1; // Match globe speed
      gridLinesRef.current.rotation.x = Math.sin(time * 0.15) * 0.08;
      gridLinesRef.current.rotation.z = Math.cos(time * 0.08) * 0.03;
    }

    // Rotate clusters and data flows to match globe rotation
    if (rotatingContentRef.current) {
      rotatingContentRef.current.rotation.y = time * 0.1; // Match globe speed
      rotatingContentRef.current.rotation.x = Math.sin(time * 0.15) * 0.08;
      rotatingContentRef.current.rotation.z = Math.cos(time * 0.08) * 0.03;
    }

    // Animate central node with slower rotation to match globe
    if (centralNodeRef.current) {
      centralNodeRef.current.rotation.y = time * 0.15; // Reduced from 0.4 to 0.15
      centralNodeRef.current.rotation.x = Math.sin(time * 0.2) * 0.05; // Reduced amplitude
      centralNodeRef.current.scale.setScalar(
        (1 + Math.sin(time * 1.5) * 0.05) * animationProgress
      ); // Reduced from 0.08 to 0.05

      // Fade in the central node
      centralNodeRef.current.children.forEach((child: CentralNodeChild) => {
        if (child.material && typeof child.material.opacity !== "undefined") {
          child.material.opacity = Math.min(
            child.material.opacity + 0.01,
            animationProgress
          );
        }
      });
    }

    // Animate data flows
    if (dataFlowsRef.current) {
      dataFlowsRef.current.children.forEach((flow: FlowChild, i) => {
        if (flow.material) {
          const flowData = dataFlows[i];
          const flowType = flowData?.type || "data";

          if (activeFlows.includes(i)) {
            flow.material.opacity = Math.min(
              flow.material.opacity + 0.05,
              0.8 * animationProgress
            );

            // Set color based on flow type
            if (flowType === "workload") {
              flow.material.color.set(COLORS.success);
            } else if (flowType === "deploy") {
              flow.material.color.set(COLORS.accent1);
            } else if (flowType === "control") {
              flow.material.color.set(COLORS.secondary);
            } else {
              flow.material.color.set(COLORS.highlight);
            }

            if (flow.material.dashSize !== undefined) {
              flow.material.dashSize = 0.1;
            }
            if (flow.material.gapSize !== undefined) {
              flow.material.gapSize = 0.05;
            }
          } else {
            flow.material.opacity = Math.max(
              flow.material.opacity - 0.02,
              0.1 * animationProgress
            );
            flow.material.color.set(COLORS.primary);

            if (flow.material.dashSize !== undefined) {
              flow.material.dashSize = 0.05;
            }
            if (flow.material.gapSize !== undefined) {
              flow.material.gapSize = 0.1;
            }
          }
        }
      });
    }
  });

  return (
    <group>
      {/* Main globe - represents the global network */}
      <Sphere ref={globeRef} args={[3.5, 64, 64]}>
        <meshPhongMaterial
          color={COLORS.primary}
          transparent
          opacity={0.15 * animationProgress} // Increased from 0.08 to 0.15 for more opacity
          wireframe
        />
      </Sphere>

      {/* Grid lines for the globe */}
      <group ref={gridLinesRef} rotation={[0, 0, 0]}>
        {Array.from({ length: 8 }).map((_, idx) => (
          <Torus
            key={idx}
            args={[3.5, 0.01, 16, 100]}
            rotation={[0, 0, (Math.PI * idx) / 8]}
          >
            <meshBasicMaterial
              color={COLORS.primary}
              transparent
              opacity={0.18 * animationProgress} // Increased from 0.1 to 0.18
            />
          </Torus>
        ))}
        {Array.from({ length: 8 }).map((_, idx) => (
          <Torus
            key={idx + 8}
            args={[3.5, 0.01, 16, 100]}
            rotation={[Math.PI / 2, (Math.PI * idx) / 8, 0]}
          >
            <meshBasicMaterial
              color={COLORS.primary}
              transparent
              opacity={0.18 * animationProgress} // Increased from 0.1 to 0.18
            />
          </Torus>
        ))}
      </group>

      {/* Central KubeStellar control plane */}
      <group ref={centralNodeRef}>
        <LogoElement position={[0, 0, 0]} rotation={[0, 0, 0]} scale={1} />

        <Billboard position={[0, 1, 0]}>
          <Text
            fontSize={0.2}
            color={COLORS.highlight}
            anchorX="center"
            anchorY="middle"
            outlineWidth={0.01}
            outlineColor={COLORS.background}
            fillOpacity={animationProgress}
          >
            {t("kubestellar")}
          </Text>
          <Text
            position={[0, -0.25, 0]}
            fontSize={0.1}
            color={COLORS.primary}
            anchorX="center"
            anchorY="middle"
            fillOpacity={animationProgress}
          >
            {t("controlPlane")}
          </Text>
        </Billboard>
      </group>

      <group ref={rotatingContentRef}>
        {/* Clusters with staggered appearance */}
        {clusters.map((cluster, idx) => (
          <group
            key={idx}
            scale={animationProgress > idx * 0.15 ? animationProgress : 0}
            position={[
              cluster.position[0] * animationProgress,
              cluster.position[1] * animationProgress,
              cluster.position[2] * animationProgress,
            ]}
          >
            <Cluster
              position={[0, 0, 0]}
              name={cluster.name}
              nodeCount={cluster.nodeCount}
              radius={cluster.radius}
              color={cluster.color}
              description={cluster.description}
            />
          </group>
        ))}

        {/* Data flow connections */}
        <group ref={dataFlowsRef}>
          {dataFlows.map((flow, idx) => (
            <Line
              key={idx}
              points={flow.path}
              color={
                activeFlows.includes(idx)
                  ? flow.type === "workload"
                    ? COLORS.success
                    : flow.type === "deploy"
                      ? COLORS.accent1
                      : flow.type === "control"
                        ? COLORS.secondary
                        : COLORS.highlight
                  : COLORS.primary
              }
              lineWidth={1.5}
              transparent
              opacity={
                (activeFlows.includes(idx) ? 0.8 : 0.1) * animationProgress
              }
              dashed
              dashSize={0.1}
              gapSize={0.1}
            />
          ))}
        </group>

        {/* Data packets traveling along active connections */}
        {isLoaded &&
          animationProgress > 0.7 &&
          dataFlows.map(
            (flow, idx) =>
              activeFlows.includes(idx) && (
                <DataPacket
                  key={idx}
                  path={flow.path}
                  speed={1 + Math.random()}
                  color={
                    flow.type === "workload"
                      ? COLORS.success
                      : flow.type === "deploy"
                        ? COLORS.accent1
                        : flow.type === "control"
                          ? COLORS.secondary
                          : idx % 2 === 0
                            ? COLORS.highlight
                            : COLORS.primary
                  }
                />
              )
          )}
      </group>
    </group>
  );
};

export default NetworkGlobe;
