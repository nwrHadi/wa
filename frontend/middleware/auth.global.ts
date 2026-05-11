export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === "/") {
    return navigateTo("/login");
  }
  if (to.path === "/login") {
    return;
  }
  if (process.server) {
    return;
  }
  
  const token = localStorage.getItem("wa_token");
  if (!token) {
    return navigateTo("/login");
  }

  // Validate token by making a test request
  try {
    const config = useRuntimeConfig();
    const headers: HeadersInit = {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    };
    
    const response = await $fetch(`${config.public.apiBase}/api/v1/devices`, {
      method: "GET",
      headers,
      retry: 0,
    });
    
    // Token is valid, continue
    return;
  } catch (error: any) {
    // If 401 or any error, token is invalid
    localStorage.removeItem("wa_token");
    return navigateTo("/login");
  }
});
