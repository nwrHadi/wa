export function useApi() {
  const config = useRuntimeConfig();

  const request = async <T>(path: string, init?: RequestInit): Promise<T> => {
    const token = process.client ? localStorage.getItem("wa_token") : "";
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    try {
      return await $fetch<T>(`${config.public.apiBase}${path}`, {
        ...init,
        headers,
      });
    } catch (error: any) {
      // If 401 Unauthorized, clear token and redirect to login
      if (error.status === 401) {
        localStorage.removeItem("wa_token");
        await navigateTo("/login");
      }
      throw error;
    }
  };

  return { request };
}
