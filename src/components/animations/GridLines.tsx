"use client";

import { useEffect, useRef } from "react";

interface GridLinesProps {
  className?: string;
  horizontalLines?: number;
  verticalLines?: number;
  strokeColor?: string;
  strokeOpacity?: number;
  strokeWidth?: number;
  speed?: number;
  opacity?: number;
}

export default function GridLines({
  className = "",
  strokeColor = "#6366F1",
  horizontalLines = 20,
  verticalLines = 20,
  strokeOpacity = 0.2,
  strokeWidth = 0.5,
  speed = 5,
  opacity = 0.2,
}: GridLinesProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const gridContainer = containerRef.current;
    if (!gridContainer) return;

    gridContainer.innerHTML = "";
    const gridSvg = document.createElementNS(
      "http://www.w3.org/2000/svg",
      "svg"
    );
    gridSvg.setAttribute("width", "calc(100% + 100px)");
    gridSvg.setAttribute("height", "calc(100% + 100px)");
    gridSvg.style.position = "absolute";
    gridSvg.style.top = "0";
    gridSvg.style.left = "0";
    gridSvg.style.animation = `move-diagonal ${speed}s linear infinite`;

    if (horizontalLines > 0) {
      const hSpacing = 100 / horizontalLines;
      for (let i = 0; i < horizontalLines; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", "0");
        line.setAttribute("y1", `${i * hSpacing}%`);
        line.setAttribute("x2", "100%");
        line.setAttribute("y2", `${i * hSpacing}%`);
        line.setAttribute("stroke", strokeColor);
        line.setAttribute("stroke-width", String(strokeWidth));
        line.setAttribute("stroke-opacity", String(strokeOpacity));
        line.style.animation = `gridPulse ${3 + (i % 5)}s infinite alternate ease-in-out`;
        line.style.animationDelay = `${i * 0.2}s`;
        gridSvg.appendChild(line);
      }
    }

    if (verticalLines > 0) {
      const vSpacing = 100 / verticalLines;
      for (let i = 0; i < verticalLines; i++) {
        const line = document.createElementNS(
          "http://www.w3.org/2000/svg",
          "line"
        );
        line.setAttribute("x1", `${i * vSpacing}%`);
        line.setAttribute("y1", "0");
        line.setAttribute("x2", `${i * vSpacing}%`);
        line.setAttribute("y2", "100%");
        line.setAttribute("stroke", strokeColor);
        line.setAttribute("stroke-width", String(strokeWidth));
        line.setAttribute("stroke-opacity", String(strokeOpacity));
        line.style.animation = `gridPulse ${3 + (i % 5)}s infinite alternate ease-in-out`;
        line.style.animationDelay = `${i * 0.2}s`;
        gridSvg.appendChild(line);
      }
    }
    gridContainer.appendChild(gridSvg);
  }, [
    horizontalLines,
    verticalLines,
    strokeColor,
    strokeWidth,
    strokeOpacity,
    speed,
  ]);

  return (
    <div
      ref={containerRef}
      className={`absolute inset-0 pointer-events-none overflow-hidden ${className}`}
      style={{ opacity }}
    />
  );
}
