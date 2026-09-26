import { motion } from "framer-motion";

// AnimatedBackground desenha blobs de gradiente que se movem lentamente atrás
// do conteúdo, dando vida à tela sem competir com a leitura.
export default function AnimatedBackground() {
  const blobs = [
    { color: "#7442ff", size: 460, x: "-10%", y: "-15%", delay: 0 },
    { color: "#b7a4ff", size: 380, x: "70%", y: "10%", delay: 1.5 },
    { color: "#ff6fa5", size: 320, x: "20%", y: "70%", delay: 3 },
  ];

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-boop-900 via-boop-800 to-boop-700" />
      {blobs.map((b, i) => (
        <motion.div
          key={i}
          className="absolute rounded-full blur-3xl"
          style={{
            width: b.size,
            height: b.size,
            left: b.x,
            top: b.y,
            backgroundColor: b.color,
            opacity: 0.35,
          }}
          animate={{
            x: [0, 40, -30, 0],
            y: [0, -30, 40, 0],
            scale: [1, 1.1, 0.95, 1],
          }}
          transition={{
            duration: 18,
            repeat: Infinity,
            ease: "easeInOut",
            delay: b.delay,
          }}
        />
      ))}
    </div>
  );
}
