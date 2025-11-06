"use client";
import { useEffect, useState } from "react";
import ShaderBackground from "./shader-background";
import StarField from "../StarField";
import GridLines from "../GridLines";

interface LoaderProps {
  isLoading?: boolean;
  text?: string;
  className?: string;
}

const Loader = ({
  isLoading = true,
  text = "Loading...",
  className = "",
}: LoaderProps) => {
  const [dots, setDots] = useState("");

  useEffect(() => {
    if (!isLoading) return;

    const interval = setInterval(() => {
      setDots(prev => (prev.length >= 3 ? "" : prev + "."));
    }, 500);

    return () => clearInterval(interval);
  }, [isLoading]);

  if (!isLoading) return null;

  return (
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center ${className}`}
    >
      <ShaderBackground />
      <StarField
        className="z-20"
        density="medium"
        showComets={true}
        cometCount={2}
      />
      <GridLines
        className="z-10"
        strokeColor="#667eea"
        strokeOpacity={0.3}
        horizontalLines={15}
        verticalLines={15}
        speed={8}
        opacity={0.4}
      />
      <div className="relative z-30 text-center">
        <div className="text-white text-xl font-medium mb-4">
          {text}
          {dots}
        </div>
        <div className="w-8 h-8 border-2 border-white border-t-transparent rounded-full animate-spin mx-auto"></div>
      </div>
    </div>
  );
};

export default Loader;
