import { motion, useAnimationControls } from "framer-motion";
import { useEffect, useRef, useState } from "react";

interface MascotProps {
  size?: number;
  // waving faz o mascote acenar (usado na landing).
  waving?: boolean;
}

// Mascot é o personagem da marca: um blob branco arredondado, sempre sorrindo.
// Ele flutua suavemente, pisca de tempos em tempos e reage a um "boop" no nariz
// (clique) — uma piscadela ao nome do repositório, boop-flag.
export default function Mascot({ size = 220, waving = false }: MascotProps) {
  const [blink, setBlink] = useState(false);
  const [booped, setBooped] = useState(false);
  const [hearts, setHearts] = useState<number[]>([]);
  const heartId = useRef(0);
  const body = useAnimationControls();

  // Piscar periódico, com intervalo levemente aleatório para parecer natural.
  useEffect(() => {
    let timeout: number;
    const scheduleBlink = () => {
      timeout = window.setTimeout(() => {
        setBlink(true);
        window.setTimeout(() => setBlink(false), 140);
        scheduleBlink();
      }, 2200 + Math.random() * 2600);
    };
    scheduleBlink();
    return () => window.clearTimeout(timeout);
  }, []);

  async function boop() {
    setBooped(true);
    const id = heartId.current++;
    setHearts((h) => [...h, id]);
    window.setTimeout(() => {
      setHearts((h) => h.filter((x) => x !== id));
    }, 900);
    // Squash & stretch: reação fofa ao toque.
    await body.start({
      scale: [1, 0.86, 1.06, 1],
      transition: { duration: 0.5, ease: "easeOut" },
    });
    setBooped(false);
  }

  const eyeScaleY = blink ? 0.1 : 1;

  return (
    <div className="relative select-none" style={{ width: size, height: size }}>
      {/* Coraçõezinhos que sobem ao dar boop */}
      {hearts.map((id) => (
        <motion.div
          key={id}
          className="pointer-events-none absolute left-1/2 top-6 text-3xl"
          initial={{ opacity: 0, y: 0, x: "-50%", scale: 0.4 }}
          animate={{ opacity: [0, 1, 0], y: -90, scale: 1.1 }}
          transition={{ duration: 0.9, ease: "easeOut" }}
        >
          💜
        </motion.div>
      ))}

      {/* Flutuação contínua no eixo Y */}
      <motion.div
        animate={{ y: [0, -12, 0] }}
        transition={{ duration: 3.4, repeat: Infinity, ease: "easeInOut" }}
        style={{ width: size, height: size }}
      >
        <motion.div animate={body} style={{ width: "100%", height: "100%" }}>
          <svg
            viewBox="0 0 200 200"
            width="100%"
            height="100%"
            role="img"
            aria-label="Mascote branco sorridente"
          >
            <defs>
              <radialGradient id="bodyShade" cx="50%" cy="38%" r="70%">
                <stop offset="0%" stopColor="#ffffff" />
                <stop offset="100%" stopColor="#eef0ff" />
              </radialGradient>
              <filter id="softShadow" x="-30%" y="-30%" width="160%" height="160%">
                <feDropShadow
                  dx="0"
                  dy="10"
                  stdDeviation="12"
                  floodColor="#1a0a52"
                  floodOpacity="0.35"
                />
              </filter>
            </defs>

            {/* Corpo branco arredondado */}
            <path
              d="M100 18 C150 18 176 54 176 104 C176 154 148 184 100 184 C52 184 24 154 24 104 C24 54 50 18 100 18 Z"
              fill="url(#bodyShade)"
              filter="url(#softShadow)"
            />

            {/* Bochechas rosadas */}
            <ellipse cx="60" cy="118" rx="13" ry="9" fill="#ffb3d1" opacity="0.7" />
            <ellipse cx="140" cy="118" rx="13" ry="9" fill="#ffb3d1" opacity="0.7" />

            {/* Olhos (piscam via scaleY) */}
            <g fill="#2b1b66">
              <motion.ellipse
                cx="74"
                cy="96"
                rx="9"
                ry="12"
                animate={{ scaleY: eyeScaleY }}
                style={{ originY: "96px", transformBox: "fill-box" } as never}
                transition={{ duration: 0.1 }}
              />
              <motion.ellipse
                cx="126"
                cy="96"
                rx="9"
                ry="12"
                animate={{ scaleY: eyeScaleY }}
                style={{ originY: "96px", transformBox: "fill-box" } as never}
                transition={{ duration: 0.1 }}
              />
            </g>
            {/* Brilho dos olhos */}
            <circle cx="77" cy="91" r="3" fill="#fff" />
            <circle cx="129" cy="91" r="3" fill="#fff" />

            {/* Sorriso */}
            <path
              d="M78 126 Q100 148 122 126"
              stroke="#2b1b66"
              strokeWidth="5"
              strokeLinecap="round"
              fill="none"
            />

            {/* Nariz — a zona de boop. Clicável e com destaque ao toque. */}
            <motion.circle
              cx="100"
              cy="112"
              r="7"
              fill={booped ? "#ff6fa5" : "#ffd0e2"}
              stroke="#ff9ec4"
              strokeWidth="2"
              onClick={boop}
              whileHover={{ scale: 1.25 }}
              whileTap={{ scale: 0.8 }}
              style={{ cursor: "pointer", originX: "100px", originY: "112px", transformBox: "fill-box" } as never}
            />
          </svg>
        </motion.div>
      </motion.div>

      {/* Bracinho que acena (opcional) */}
      {waving && (
        <motion.div
          className="absolute text-4xl"
          style={{ right: size * 0.05, top: size * 0.42 }}
          animate={{ rotate: [0, 20, -8, 20, 0] }}
          transition={{ duration: 1.6, repeat: Infinity, repeatDelay: 1.2 }}
        >
          👋
        </motion.div>
      )}
    </div>
  );
}
