import { motion } from "framer-motion";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import Confetti from "../components/Confetti";
import Mascot from "../components/Mascot";

// Home é a tela principal após o login: o mascote branco sorridente em
// destaque, dando as boas-vindas.
export default function Home() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [celebrate, setCelebrate] = useState(true);

  // Confete some após alguns segundos.
  useEffect(() => {
    const t = window.setTimeout(() => setCelebrate(false), 2600);
    return () => window.clearTimeout(t);
  }, []);

  async function handleLogout() {
    await logout();
    navigate("/", { replace: true });
  }

  return (
    <motion.main
      className="flex min-h-screen flex-col items-center justify-center px-4 py-10"
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.98 }}
    >
      {celebrate && <Confetti />}

      <motion.div
        initial={{ scale: 0.6, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ type: "spring", stiffness: 140, damping: 12, delay: 0.1 }}
        className="flex justify-center"
      >
        <Mascot size={280} />
      </motion.div>

      <motion.h1
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
        className="mt-6 text-center text-4xl font-black md:text-5xl"
      >
        Olá, {user?.display_name || "amiguinho"}! 🎉
      </motion.h1>

      <motion.p
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.45 }}
        className="mt-3 max-w-md text-center text-lg text-white/80"
      >
        Você entrou com sucesso. Dá um <strong>boop</strong> no nariz dele — pode
        confiar, ele adora.
      </motion.p>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="mt-8 flex items-center gap-3 rounded-2xl bg-white/10 px-5 py-3 text-sm ring-1 ring-white/20"
      >
        <span className="text-white/70">Conectado como</span>
        <span className="font-bold">{user?.email}</span>
      </motion.div>

      <motion.button
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.75 }}
        onClick={handleLogout}
        whileHover={{ scale: 1.05, y: -2 }}
        whileTap={{ scale: 0.96 }}
        className="mt-6 rounded-2xl bg-white px-8 py-3 font-extrabold text-boop-900 shadow-xl"
      >
        Sair
      </motion.button>
    </motion.main>
  );
}
