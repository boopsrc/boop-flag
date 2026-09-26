import { motion } from "framer-motion";
import { useMemo } from "react";

// Confetti dispara uma chuva rápida de pedacinhos coloridos — usado na
// celebração de boas-vindas na home.
export default function Confetti({ count = 40 }: { count?: number }) {
  const pieces = useMemo(() => {
    const colors = ["#7442ff", "#b7a4ff", "#ff6fa5", "#ffd0e2", "#ffffff"];
    return Array.from({ length: count }, (_, i) => ({
      id: i,
      left: Math.random() * 100,
      color: colors[i % colors.length],
      delay: Math.random() * 0.4,
      duration: 1.6 + Math.random() * 1.2,
      rotate: Math.random() * 360,
      size: 8 + Math.random() * 8,
    }));
  }, [count]);

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden">
      {pieces.map((p) => (
        <motion.div
          key={p.id}
          className="absolute top-0 rounded-sm"
          style={{
            left: `${p.left}%`,
            width: p.size,
            height: p.size * 0.6,
            backgroundColor: p.color,
          }}
          initial={{ y: -40, opacity: 1, rotate: p.rotate }}
          animate={{ y: "110vh", opacity: [1, 1, 0], rotate: p.rotate + 240 }}
          transition={{ duration: p.duration, delay: p.delay, ease: "easeIn" }}
        />
      ))}
    </div>
  );
}
