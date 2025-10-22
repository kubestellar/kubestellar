import { useState, useEffect } from "react";

interface StatCardProps {
  icon: React.ReactNode;
  value: number;
  suffix: string;
  title: string;
  color: "blue" | "purple" | "emerald";
  animationDelay: string;
}

const StatCard = ({
  icon,
  value,
  suffix,
  title,
  color,
  animationDelay,
}: StatCardProps) => {
  const [count, setCount] = useState(0);

  const colorStyles = {
    blue: {
      gradient: "from-blue-500/20 to-blue-600/20",
      border: "border-blue-500/30",
      text: "text-blue-400",
      glow: "rgba(59, 130, 246, 0.2)",
    },
    purple: {
      gradient: "from-purple-500/20 to-purple-600/20",
      border: "border-purple-500/30",
      text: "text-purple-400",
      glow: "rgba(147, 51, 234, 0.2)",
    },
    emerald: {
      gradient: "from-emerald-500/20 to-emerald-600/20",
      border: "border-emerald-500/30",
      text: "text-emerald-400",
      glow: "rgba(16, 185, 129, 0.2)",
    },
  };

  const styles = colorStyles[color] || colorStyles.blue;

  useEffect(() => {
    const duration = 2000;
    const totalSteps = 50;
    const stepValue = value / totalSteps;
    const intervalTime = duration / totalSteps;

    let currentStep = 0;
    const counterInterval = setInterval(() => {
      currentStep++;
      if (currentStep > totalSteps) {
        setCount(value);
        clearInterval(counterInterval);
      } else {
        setCount(Math.round(stepValue * currentStep));
      }
    }, intervalTime);

    return () => clearInterval(counterInterval);
  }, [value]);

  return (
    <div
      className={`stat-card bg-gradient-to-br ${styles.gradient} backdrop-blur-md rounded-2xl p-4 border ${styles.border} relative group overflow-hidden animate-stat-float`}
      style={
        {
          animationDelay,
          "--glow-color": styles.glow,
        } as React.CSSProperties
      }
    >
      <div className="stat-glow"></div>
      <div className="relative z-10">
        <div className="flex items-center space-x-3 mb-2">
          <div className="stat-icon">{icon}</div>
          <p className={`text-3xl font-black ${styles.text} counter`}>
            {count}
          </p>
          <span className={`${styles.text} text-lg font-bold`}>{suffix}</span>
        </div>
        <p className="text-sm text-gray-300 font-semibold">{title}</p>
      </div>
    </div>
  );
};

export default StatCard;
