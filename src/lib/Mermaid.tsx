"use client";

import { useEffect, useRef } from "react";
import mermaid from "mermaid";

mermaid.initialize({ startOnLoad: false, theme: "neutral" });

type MermaidProps = {
  children: string;
};

export const MermaidComponent = ({ children }: MermaidProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const chart = children;

  useEffect(() => {
    if (ref.current && chart) {
      ref.current.innerHTML = "";
      mermaid.run({
        nodes: [ref.current],
      });
    }
  }, [chart]);

  if (!chart) {
    return null;
  }

  return (
    <div className="mermaid" ref={ref}>
      {chart}
    </div>
  );
};
