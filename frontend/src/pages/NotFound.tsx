import { motion } from "framer-motion";
import { Link } from "react-router-dom";
import Mascot from "../components/Mascot";

export default function NotFound() {
  return (
    <motion.main
      className="flex min-h-screen flex-col items-center justify-center px-4 text-center"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
    >
      <Mascot size={160} />
      <h1 className="mt-6 text-6xl font-black">404</h1>
      <p className="mt-2 text-lg text-white/80">
        O Boop procurou, mas não achou essa página.
      </p>
      <Link
        to="/"
        className="mt-6 rounded-2xl bg-white px-8 py-3 font-extrabold text-boop-900 shadow-xl"
      >
        Voltar ao início
      </Link>
    </motion.main>
  );
}
