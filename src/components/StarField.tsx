'use client';

import { useEffect, useRef } from 'react';

interface StarFieldProps {
  className?: string;
  density?: 'low' | 'medium' | 'high';
  showComets?: boolean;
  cometCount?: number;
}

export default function StarField({ 
  className = '', 
  density = 'medium',
  showComets = true,
  cometCount = 3
}: StarFieldProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const container = containerRef.current;
    
    // Clear existing stars and comets
    container.innerHTML = '';

    // Calculate star count based on density
    const densityMap = {
      low: 200,
      medium: 300,
      high: 600
    };
    
    const starCount = densityMap[density];

    // Create stars with varying sizes and positions
    for (let i = 0; i < starCount; i++) {
      const star = document.createElement('div');
      star.className = 'star absolute';
      
      // Random size distribution (more small stars than large ones)
      const sizeRand = Math.random();
      if (sizeRand < 0.6) {
        star.classList.add('star-tiny');
      } else if (sizeRand < 0.85) {
        star.classList.add('star-small');
      } else if (sizeRand < 0.95) {
        star.classList.add('star-medium');
      } else {
        star.classList.add('star-large');
      }

      // Random position
      star.style.left = `${Math.random() * 100}%`;
      star.style.top = `${Math.random() * 100}%`;
      
      // Random animation delay for natural twinkling (0 to 5 seconds)
      const randomDelay = Math.random() * 5;
      star.style.animationDelay = `${randomDelay}s`;
      
      // Random animation duration for more variation (1.5 to 4 seconds)
      const randomDuration = 1.5 + Math.random() * 2.5;
      star.style.animationDuration = `${randomDuration}s`;
      
      container.appendChild(star);
    }

    // Create comets if enabled
    if (showComets) {
      for (let i = 0; i < cometCount; i++) {
        const comet = document.createElement('div');
        comet.className = 'comet absolute';
        
        // Random comet size
        const cometSizeRand = Math.random();
        if (cometSizeRand < 0.5) {
          comet.classList.add('comet-small');
        } else if (cometSizeRand < 0.8) {
          comet.classList.add('comet-medium');
        } else {
          comet.classList.add('comet-large');
        }

        // Start position (off-screen top-left area)
        comet.style.left = `${Math.random() * 20 - 10}%`;
        comet.style.top = `${Math.random() * 20 - 10}%`;
        
        // Random animation delay to spread out comet appearances
        comet.style.animationDelay = `${Math.random() * 6}s`;
        
        container.appendChild(comet);
      }
    }

  }, [density, showComets, cometCount]);

  return (
    <div 
      ref={containerRef}
      className={`absolute inset-0 pointer-events-none overflow-hidden ${className}`}
      style={{ zIndex: 1 }}
    />
  );
}