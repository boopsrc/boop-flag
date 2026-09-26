// Cliente HTTP fino sobre fetch. Todas as chamadas incluem o cookie de sessão
// (credentials: "include") e passam pelo proxy /api do Vite em dev.

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  onboarded: boolean;
}

export class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options.headers },
    ...options,
  });

  if (!res.ok) {
    let code = "error";
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) {
        code = body.error.code ?? code;
        message = body.error.message ?? message;
      }
    } catch {
      // resposta sem corpo JSON; mantém o statusText
    }
    throw new ApiError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  me: () => request<User>("/me"),
  updateMe: (displayName: string) =>
    request<User>("/me", {
      method: "PATCH",
      body: JSON.stringify({ display_name: displayName }),
    }),
  logout: () => request<{ status: string }>("/auth/logout", { method: "POST" }),
  googleLoginUrl: "/api/v1/auth/google/login",
  devLoginUrl: "/api/v1/auth/dev/login",
};
