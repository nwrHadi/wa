export default defineEventHandler((event) => {
  const requestUrl = getRequestURL(event);
  const upstreamPath = requestUrl.pathname.replace(/^\/gateway/, "") || "/";
  const target = `http://127.0.0.1:8090${upstreamPath}${requestUrl.search}`;

  return proxyRequest(event, target);
});