"use client";

import { useEffect } from "react";
import Link from "next/link";
import Navigation from "@/components/Navigation";
import Footer from "@/components/Footer";
import StarField from "@/components/StarField";
import { handbookCards, HandbookCard } from "./handbook";

interface HandbookCardComponentProps {
  card: HandbookCard;
}

function HandbookCardComponent({ card }: HandbookCardComponentProps) {
  return (
    <Link href={card.link}>
      <div className="relative group bg-slate-800/50 border border-slate-700 rounded-xl p-8 h-72 overflow-hidden transition-all duration-300 cursor-pointer hover:shadow-2xl hover:shadow-purple-500/30">
        <div className="transition-all duration-300 group-hover:-translate-y-2 h-full flex flex-col">
          <div className={`w-12 h-12 ${card.bgColor} rounded-lg flex items-center justify-center mb-4`}>
            <svg className={`w-6 h-6 ${card.iconColor}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d={card.iconPath}></path>
            </svg>
          </div>
          <h3 className="text-2xl font-bold text-white mb-4">{card.title}</h3>
          <p className="text-gray-300 leading-relaxed flex-grow">{card.description}</p>
        </div>
        <div className="absolute bottom-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
          <span className="learn-more-enhanced">Learn More</span>
        </div>
      </div>
    </Link>
  );
}

export default function CommunityHandbook() {
  useEffect(() => {
    // Create grid lines with subtle animation
    const createGrid = (gridContainer: HTMLElement | null) => {
      if (!gridContainer) return;
      
      // Clear existing content
      gridContainer.innerHTML = '';
      
      const gridSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      gridSvg.setAttribute('width', '100%');
      gridSvg.setAttribute('height', '100%');
      gridSvg.style.position = 'absolute';
      gridSvg.style.top = '0';
      gridSvg.style.left = '0';
      
      // Create horizontal grid lines with animation
      for (let i = 0; i < 20; i++) {
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', '0');
        line.setAttribute('y1', `${i * 5}%`);
        line.setAttribute('x2', '100%');
        line.setAttribute('y2', `${i * 5}%`);
        line.setAttribute('stroke', '#6366F1');
        line.setAttribute('stroke-width', '0.5');
        line.setAttribute('stroke-opacity', '0.2');
        
        const animate = document.createElementNS('http://www.w3.org/2000/svg', 'animate');
        animate.setAttribute('attributeName', 'stroke-opacity');
        animate.setAttribute('values', '0.1;0.4;0.1');
        animate.setAttribute('dur', `${2 + i * 0.1}s`);
        animate.setAttribute('repeatCount', 'indefinite');
        line.appendChild(animate);
        
        gridSvg.appendChild(line);
      }
      
      // Create vertical grid lines with animation
      for (let i = 0; i < 16; i++) {
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', `${i * 6.25}%`);
        line.setAttribute('y1', '0');
        line.setAttribute('x2', `${i * 6.25}%`);
        line.setAttribute('y2', '100%');
        line.setAttribute('stroke', '#6366F1');
        line.setAttribute('stroke-width', '0.5');
        line.setAttribute('stroke-opacity', '0.2');
        
        const animate = document.createElementNS('http://www.w3.org/2000/svg', 'animate');
        animate.setAttribute('attributeName', 'stroke-opacity');
        animate.setAttribute('values', '0.1;0.4;0.1');
        animate.setAttribute('dur', `${2.5 + i * 0.1}s`);
        animate.setAttribute('repeatCount', 'indefinite');
        line.appendChild(animate);
        
        gridSvg.appendChild(line);
      }
      
      gridContainer.appendChild(gridSvg);
    };

    // Interactive grid canvas initialization
    const initGridCanvas = () => {
      const canvas = document.getElementById('grid-canvas-handbook') as HTMLCanvasElement;
      if (!canvas) return;

      const ctx = canvas.getContext('2d');
      if (!ctx) return;

      // Set canvas size
      const resizeCanvas = () => {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
      };

      resizeCanvas();
      window.addEventListener('resize', resizeCanvas);

      let mouseX = 0;
      let mouseY = 0;
      let isMouseOver = false;

      canvas.addEventListener('mousemove', (e) => {
        const rect = canvas.getBoundingClientRect();
        mouseX = e.clientX - rect.left;
        mouseY = e.clientY - rect.top;
        isMouseOver = true;
      });

      canvas.addEventListener('mouseleave', () => {
        isMouseOver = false;
      });

      const animate = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        if (isMouseOver) {
          // Draw interactive grid around mouse
          const gridSize = 50;
          const startX = Math.floor(mouseX / gridSize) * gridSize;
          const startY = Math.floor(mouseY / gridSize) * gridSize;

          ctx.strokeStyle = 'rgba(124, 58, 237, 0.3)';
          ctx.lineWidth = 1;

          for (let x = startX - gridSize * 3; x <= startX + gridSize * 3; x += gridSize) {
            if (x >= 0 && x <= canvas.width) {
              ctx.beginPath();
              ctx.moveTo(x, startY - gridSize * 3);
              ctx.lineTo(x, startY + gridSize * 3);
              ctx.stroke();
            }
          }

          for (let y = startY - gridSize * 3; y <= startY + gridSize * 3; y += gridSize) {
            if (y >= 0 && y <= canvas.height) {
              ctx.beginPath();
              ctx.moveTo(startX - gridSize * 3, y);
              ctx.lineTo(startX + gridSize * 3, y);
              ctx.stroke();
            }
          }
        }

        requestAnimationFrame(animate);
      };

      animate();
    };

    // Initialize grid and canvas
    createGrid(document.getElementById('grid-lines-handbook'));
    initGridCanvas();

    // Back to top functionality
    const backToTopButton = document.getElementById('back-to-top');
    if (backToTopButton) {
      const handleScroll = () => {
        if (window.scrollY > 300) {
          backToTopButton.classList.remove('opacity-0', 'translate-y-10');
          backToTopButton.classList.add('opacity-100', 'translate-y-0');
        } else {
          backToTopButton.classList.add('opacity-0', 'translate-y-10');
          backToTopButton.classList.remove('opacity-100', 'translate-y-0');
        }
      };

      window.addEventListener('scroll', handleScroll);
      
      backToTopButton.addEventListener('click', () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
      });

      return () => {
        window.removeEventListener('scroll', handleScroll);
      };
    }
  }, []);

  return (
    <div className="bg-slate-900 text-white overflow-x-hidden dark">
      <Navigation />
      
      <main className="pt-24 relative overflow-hidden bg-slate-900 text-white">
        {/* Dark base background */}
        <div className="absolute inset-0 bg-[#0a0a0a]"></div>
        
        {/* Starfield background */}
        <div className="absolute inset-0 overflow-hidden">
          <StarField density="high" showComets={true} cometCount={5} />
        </div>
        
        {/* Grid lines background */}
        <div id="grid-lines-handbook" className="absolute inset-0 opacity-20"></div>
        
        {/* Interactive Grid Network */}
        <div className="absolute inset-0">
          <canvas id="grid-canvas-handbook" className="w-full h-full opacity-20"></canvas>
        </div>
        
        {/* Floating Data Particles */}
        <div className="absolute inset-0">
          <div className="data-particle" style={{ '--delay': '0s' } as React.CSSProperties}></div>
          <div className="data-particle" style={{ '--delay': '1s' } as React.CSSProperties}></div>
          <div className="data-particle" style={{ '--delay': '2s' } as React.CSSProperties}></div>
          <div className="data-particle" style={{ '--delay': '3s' } as React.CSSProperties}></div>
          <div className="data-particle" style={{ '--delay': '4s' } as React.CSSProperties}></div>
          <div className="data-particle" style={{ '--delay': '5s' } as React.CSSProperties}></div>
        </div>
        
        <div className="relative py-16">
          <div className="container mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
            <h1 className="text-6xl font-bold text-center mb-16 text-shadow-lg">
              <span className="text-gradient animated-gradient bg-gradient-to-r from-purple-600 via-blue-500 to-purple-600">Contribute</span>{' '}
              <span className="text-gradient animated-gradient bg-gradient-to-r from-cyan-400 via-emerald-500 to-blue-500">Handbook</span>
            </h1>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {handbookCards.map((card) => (
                <HandbookCardComponent key={card.id} card={card} />
              ))}
            </div>
          </div>
        </div>
      </main>

      <Footer />

      {/* Floating back to top button */}
      <button 
        id="back-to-top" 
        className="fixed bottom-8 right-8 p-2 rounded-full bg-blue-600 text-white shadow-lg z-50 transition-all duration-300 opacity-0 translate-y-10"
      >
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
        </svg>
      </button>
    </div>
  );
}
