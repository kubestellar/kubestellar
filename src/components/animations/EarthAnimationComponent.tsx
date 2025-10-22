import { OrbitControls, useGLTF, Html, useProgress } from "@react-three/drei";
import { Canvas } from "@react-three/fiber";
import { Suspense } from "react";

// Loader component for showing progress
const CanvasLoader = () => {
  const { progress } = useProgress();

  return (
    <Html center>
      <div className="canvas-loader">
        <p
          style={{
            fontSize: 14,
            color: "#f1f1f1",
            fontWeight: 800,
            marginTop: 40,
            textAlign: "center",
          }}
        >
          {progress.toFixed(2)}%
        </p>
      </div>
    </Html>
  );
};

// Props interface for the Earth component
interface EarthProps {
  scale?: number;
  position?: [number, number, number];
  rotation?: [number, number, number];
}

// Earth/Planet 3D model component
const Earth: React.FC<EarthProps> = ({ scale = 2.5, position = [0, 0, 0], rotation = [0, 0, 0] }) => {
  const earth = useGLTF("./planet/scene.gltf");

  return (
    <primitive
      object={earth.scene}
      scale={scale}
      position={position}
      rotation={rotation}
    />
  );
};

// Props interface for the main component
interface EarthAnimationProps {
  /** Width of the canvas container (default: "100%") */
  width?: string | number;
  /** Height of the canvas container (default: "400px") */
  height?: string | number;
  /** Scale of the planet model (default: 2.5) */
  scale?: number;
  /** Position of the planet [x, y, z] (default: [0, 0, 0]) */
  position?: [number, number, number];
  /** Initial rotation of the planet [x, y, z] (default: [0, 0, 0]) */
  rotation?: [number, number, number];
  /** Enable auto-rotation (default: true) */
  autoRotate?: boolean;
  /** Auto-rotation speed (default: 1) */
  autoRotateSpeed?: number;
  /** Enable zoom controls (default: false) */
  enableZoom?: boolean;
  /** Camera field of view (default: 45) */
  fov?: number;
  /** Camera position [x, y, z] (default: [-4, 3, 6]) */
  cameraPosition?: [number, number, number];
  /** Custom loading component */
  loader?: React.ComponentType;
  /** Custom CSS class for container */
  className?: string;
  /** Custom inline styles for container */
  style?: React.CSSProperties;
}

// Main Earth Animation Component
const EarthAnimation: React.FC<EarthAnimationProps> = ({
  width = "100%",
  height = "400px",
  scale = 2.5,
  position = [0, 0, 0],
  rotation = [0, 0, 0],
  autoRotate = true,
  autoRotateSpeed = 1,
  enableZoom = false,
  fov = 45,
  cameraPosition = [-4, 3, 6],
  loader: CustomLoader = CanvasLoader,
  className = "",
  style = {},
}) => {
  const containerStyle: React.CSSProperties = {
    width,
    height,
    ...style,
  };

  return (
    <div className={className} style={containerStyle}>
      <Canvas
        shadows
        frameloop="demand"
        gl={{ preserveDrawingBuffer: true }}
        camera={{
          fov,
          near: 0.1,
          far: 200,
          position: cameraPosition,
        }}
      >
        <Suspense fallback={<CustomLoader />}>
          <OrbitControls
            autoRotate={autoRotate}
            autoRotateSpeed={autoRotateSpeed}
            enableZoom={enableZoom}
            maxPolarAngle={Math.PI / 2}
            minPolarAngle={Math.PI / 2}
          />

          <Earth scale={scale} position={position} rotation={rotation} />
        </Suspense>
      </Canvas>
    </div>
  );
};

// Export for use in other components
export default EarthAnimation;

// Named export for the individual components if needed
export { Earth, CanvasLoader };
