import { motion } from "framer-motion";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import GoogleButton from "../components/GoogleButton";
import Mascot from "../components/Mascot";

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.12, delayChildren: 0.1 },
  },
};

const item = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: { type: "spring", stiffness: 120, damping: 14 } },
};

// A landing é a porta de entrada (login e cadastro no mesmo lugar): o primeiro
// login com Google já cria a conta.
export default function Landing() {
  const { user, loading } = useAuth();
  const navigate = useNavigate();

  // Já logado? Pula direto para dentro do app.
  useEffect(() => {
    if (!loading && user) {
      navigate(user.onboarded ? "/home" : "/onboarding", { replace: true });
    }
  }, [user, loading, navigate]);

  const devMode = import.meta.env.DEV;

  return (
    <motion.main
      className="flex min-h-screen items-center justify-center px-4 py-10"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
    >
      <motion.div
        variants={container}
        initial="hidden"
        animate="show"
        className="w-full max-w-md rounded-[2rem] bg-white/10 p-8 text-center shadow-2xl backdrop-blur-xl ring-1 ring-white/20"
      >
        <motion.div variants={item} className="flex justify-center">
          <Mascot size={200} waving />
        </motion.div>

        <motion.h1 variants={item} className="mt-4 text-4xl font-black tracking-tight">
          Oi! Eu sou o Boop
        </motion.h1>
        <motion.p variants={item} className="mt-2 text-lg text-white/80">
          Entre ou cadastre-se em um toque para começar.
        </motion.p>

        <motion.div variants={item} className="mt-8">
          <GoogleButton href={api.googleLoginUrl} />
        </motion.div>

        {devMode && (
          <motion.a
            variants={item}
            href={api.devLoginUrl}
            className="mt-3 inline-block text-sm font-semibold text-white/60 underline decoration-dotted underline-offset-4 hover:text-white"
          >
            Entrar em modo dev (sem Google)
          </motion.a>
        )}

        <motion.p variants={item} className="mt-6 text-xs text-white/50">
          Ao continuar, você concorda em ser recebido por um mascote muito fofo.
        </motion.p>
      </motion.div>
    </motion.main>
  );
}
