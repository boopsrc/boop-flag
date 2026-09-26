import { motion } from "framer-motion";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, ApiError } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import Mascot from "../components/Mascot";

// Onboarding é o passo de "cadastro" após o primeiro login: o usuário confirma
// como quer ser chamado antes de entrar na home.
export default function Onboarding() {
  const { user, setUser } = useAuth();
  const navigate = useNavigate();
  const [name, setName] = useState(user?.display_name ?? "");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSaving(true);
    try {
      const updated = await api.updateMe(name);
      setUser(updated);
      navigate("/home", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Algo deu errado");
      setSaving(false);
    }
  }

  return (
    <motion.main
      className="flex min-h-screen items-center justify-center px-4 py-10"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
    >
      <div className="w-full max-w-md rounded-[2rem] bg-white/10 p-8 text-center shadow-2xl backdrop-blur-xl ring-1 ring-white/20">
        <div className="flex justify-center">
          <Mascot size={160} />
        </div>
        <h1 className="mt-4 text-3xl font-black">Como podemos te chamar?</h1>
        <p className="mt-2 text-white/80">O Boop quer saber seu nome. 💜</p>

        <form onSubmit={submit} className="mt-6 space-y-4">
          <input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={60}
            placeholder="Seu nome"
            className="w-full rounded-2xl border-none bg-white/90 px-5 py-4 text-center text-lg font-bold text-boop-900 outline-none ring-2 ring-transparent transition focus:ring-boop-300"
          />
          {error && (
            <motion.p
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-sm font-semibold text-pink-200"
            >
              {error}
            </motion.p>
          )}
          <motion.button
            type="submit"
            disabled={saving}
            whileHover={{ scale: 1.04, y: -2 }}
            whileTap={{ scale: 0.97 }}
            className="w-full rounded-2xl bg-white px-6 py-4 text-lg font-extrabold text-boop-900 shadow-xl disabled:opacity-60"
          >
            {saving ? "Salvando…" : "Vamos lá!"}
          </motion.button>
        </form>
      </div>
    </motion.main>
  );
}
