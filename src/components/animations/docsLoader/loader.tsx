"use client";
import { useEffect, useState } from "react";
import Image from "next/image";
import StarField from "../StarField";
import GridLines from "../GridLines";

interface DocsLoaderProps {
  isLoading?: boolean;
  text?: string;
  className?: string;
}

const DocsLoader = ({
  isLoading = true,
  text = "Loading Documentation...",
  className = "",
}: DocsLoaderProps) => {
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
      {/* Dark Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#0a0a0a] via-[#111111] to-[#0a0a0a]" />

      {/* StarField Background */}
      <StarField
        className="z-10"
        density="medium"
        showComets={true}
        cometCount={2}
      />

      {/* GridLines Background */}
      <GridLines
        className="z-10"
        strokeColor="#667eea"
        strokeOpacity={0.2}
        horizontalLines={20}
        verticalLines={20}
        speed={12}
        opacity={0.3}
      />

      {/* Main Content */}
      <div className="relative z-30 text-center">
        {/* KubeStellar Logo/Icon */}
        <div className="mb-8 relative">
          {/* Outer rotating ring */}
          <div className="w-32 h-32 mx-auto relative">
            <div className="absolute inset-0 border-2 border-transparent bg-gradient-to-r from-[#667eea] via-[#764ba2] to-[#f093fb] rounded-full animate-spin opacity-80">
              <div className="absolute inset-1 bg-[#0a0a0a] rounded-full"></div>
            </div>

            {/* Inner pulsing circle */}
            <div className="absolute inset-4 bg-gradient-to-br from-[#667eea]/20 to-[#764ba2]/20 rounded-full animate-pulse blur-sm"></div>

            {/* Favicon container */}
            <div className="absolute inset-6 bg-gradient-to-br from-[#667eea]/10 to-[#764ba2]/10 rounded-full flex items-center justify-center backdrop-blur-sm border border-[#667eea]/30">
              <Image
                src="/favicon.ico"
                alt="KubeStellar"
                width={58}
                height={58}
                className="object-contain filter brightness-110 drop-shadow-lg"
              />
            </div>
          </div>
        </div>

        {/* Loading Text */}
        <div className="mb-6">
          <h2 className="text-2xl font-semibold mb-2">
            <span className="bg-gradient-to-r from-[#667eea] via-[#764ba2] to-[#f093fb] bg-clip-text text-transparent">
              {text.replace("...", "")}
            </span>
            <span className="text-[#667eea] text-xl">{dots}</span>
          </h2>
          <p className="text-gray-400 text-sm">
            Please wait while we load the content
          </p>
        </div>

        {/* Progress Bar */}
        <div className="w-64 h-1 bg-gray-800 rounded-full mx-auto mb-6 overflow-hidden">
          <div className="h-full bg-gradient-to-r from-[#667eea] to-[#764ba2] rounded-full animate-pulse"></div>
        </div>

        {/* Floating particles */}
        <div className="absolute -top-10 -left-10 w-20 h-20 bg-gradient-to-r from-[#667eea]/10 to-[#764ba2]/10 rounded-full blur-xl animate-float"></div>
        <div
          className="absolute -bottom-10 -right-10 w-16 h-16 bg-gradient-to-r from-[#f093fb]/10 to-[#667eea]/10 rounded-full blur-xl animate-float"
          style={{ animationDelay: "1s" }}
        ></div>
      </div>
    </div>
  );
};

export default DocsLoader;
