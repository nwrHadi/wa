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
    console.log("[Auth] No token found, redirecting to login");
    return navigateTo("/login");
  }

  // Do not block route navigation with network calls.
  // Token validity is enforced by API responses in useApi().
  return;
});
