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
      console.log("[API] Request:", {
        url: `${config.public.apiBase}${path}`,
        method: init?.method || "GET",
      });

      const response = await $fetch<T>(`${config.public.apiBase}${path}`, {
        ...init,
        headers,
        retry: 0,
        timeout: 10000, // 10 second timeout
      });

      console.log("[API] Response OK for", path);
      return response;
    } catch (error: any) {
      console.error("[API] Error:", {
        path,
        status: error?.status,
        statusCode: error?.statusCode,
        message: error?.message,
      });

      // If 401 Unauthorized, clear token and redirect to login
      if (error?.status === 401 || error?.statusCode === 401) {
        console.log("[API] Token invalid (401), clearing and redirecting...");
        localStorage.removeItem("wa_token");
        await navigateTo("/login");
      }
      throw error;
    }
  };

  return { request };
}
