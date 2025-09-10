import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "space-dark": "#0a0a0a",
      },
      fontFamily: {
        inter: ["var(--font-inter)"],
        mono: ["var(--font-jetbrains-mono)"],
      },
      animation: {
        "spin-slow": "spin 8s linear infinite",
        float: "float 6s ease-in-out infinite",
        "pulse-slow": "pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        twinkle: "twinkle 3s infinite alternate",
        gradient: "gradient 15s ease infinite",
        typing: "typing 2s steps(40, end) forwards",
        blink: "blink 1s infinite",
        "fade-in-up": "fade-in-up 0.8s ease-out forwards",
        "slide-in-left": "slide-in-left 0.8s ease-out forwards",
        "text-reveal":
          "text-reveal 1s cubic-bezier(0.25, 0.46, 0.45, 0.94) forwards",
        "status-glow": "status-glow 2s ease-in-out infinite",
        "status-pulse": "status-pulse 2s ease-in-out infinite",
        "command-glow": "command-glow 3s ease-in-out infinite",
        "btn-float": "btn-float 3s ease-in-out infinite",
        "stat-float": "stat-float 3s ease-in-out infinite",
        "float-particle": "float-particle 4s ease-in infinite",
        "shooting-star": "shootingStar 15s infinite linear",
        "nebula-float": "nebula-float 60s infinite alternate ease-in-out",
        "grid-pulse": "gridPulse 4s infinite alternate ease-in-out",
      },
      keyframes: {
        float: {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-20px)" },
        },
        pulse: {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.7" },
        },
        gradient: {
          "0%": { backgroundPosition: "0% 50%" },
          "50%": { backgroundPosition: "100% 50%" },
          "100%": { backgroundPosition: "0% 50%" },
        },
        twinkle: {
          "0%": { opacity: "0.2", transform: "scale(0.8)" },
          "100%": { opacity: "1", transform: "scale(1)" },
        },
        "fade-in-up": {
          from: { opacity: "0", transform: "translateY(30px)" },
          to: { opacity: "1", transform: "translateY(0)" },
        },
        "slide-in-left": {
          from: { opacity: "0", transform: "translateX(-30px)" },
          to: { opacity: "1", transform: "translateX(0)" },
        },
        typing: {
          from: { width: "0" },
          to: { width: "100%" },
        },
        blink: {
          "0%, 50%": { opacity: "1" },
          "51%, 100%": { opacity: "0" },
        },
        "text-reveal": {
          "0%": { opacity: "0", transform: "translateY(20px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "status-glow": {
          "0%, 100%": { boxShadow: "0 0 20px rgba(16, 185, 129, 0.3)" },
          "50%": { boxShadow: "0 0 40px rgba(16, 185, 129, 0.6)" },
        },
        "status-pulse": {
          "0%, 100%": { transform: "scale(1)", opacity: "1" },
          "50%": { transform: "scale(1.2)", opacity: "0.7" },
        },
        "command-glow": {
          "0%, 100%": { boxShadow: "0 0 30px rgba(59, 130, 246, 0.2)" },
          "50%": { boxShadow: "0 0 50px rgba(59, 130, 246, 0.4)" },
        },
        "btn-float": {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-3px)" },
        },
        "stat-float": {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-10px)" },
        },
        "float-particle": {
          "0%": { transform: "translate(0, 0)", opacity: "0" },
          "20%": { opacity: "1" },
          "80%": { opacity: "1" },
          "100%": { transform: "translate(20px, -15px)", opacity: "0" },
        },
        shootingStar: {
          "0%": {
            opacity: "0",
            transform: "translateX(0) translateY(0) scale(0)",
          },
          "2%": {
            opacity: "1",
            transform: "translateX(0) translateY(0) scale(1)",
          },
          "7%": { opacity: "1" },
          "9%": {
            opacity: "0",
            transform: "translateX(300px) translateY(-100px) scale(0)",
          },
          "100%": {
            opacity: "0",
            transform: "translateX(300px) translateY(-100px) scale(0)",
          },
        },
        "nebula-float": {
          "0%": { transform: "translate(0, 0) scale(1)" },
          "50%": { transform: "translate(20px, -10px) scale(1.1)" },
          "100%": { transform: "translate(0, 0) scale(1)" },
        },
        gridPulse: {
          "0%": { strokeOpacity: "0.1" },
          "100%": { strokeOpacity: "0.4" },
        },
      },
      backdropBlur: {
        xs: "2px",
      },
      perspective: {
        "1000": "1000px",
        "500": "500px",
      },
    },
  },
  plugins: [],
};

export default config;
