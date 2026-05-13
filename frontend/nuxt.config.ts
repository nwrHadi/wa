export default defineNuxtConfig({
  devtools: { enabled: false },
  css: ["~/assets/css/main.css"],
  vite: {
    server: {
      allowedHosts: ["wa.aracel.my.id"],
    },
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "https://api.aracel.my.id/wa",
      wsBase: process.env.NUXT_PUBLIC_WS_BASE || "wss://api.aracel.my.id/wa",
      gatewayBase: process.env.NUXT_PUBLIC_GATEWAY_BASE || "/gateway",
    },
  },
  app: {
    head: {
      title: "WA Control",
      htmlAttrs: { class: "light" },
      meta: [{ name: "viewport", content: "width=device-width, initial-scale=1" }],
      link: [
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap",
        },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap",
        },
      ],
      script: [
        {
          src: "https://cdn.tailwindcss.com?plugins=forms,container-queries",
        },
        {
          innerHTML: `tailwind.config={darkMode:"class",theme:{extend:{colors:{"on-primary":"#ffffff","on-primary-fixed":"#001a41","inverse-on-surface":"#f0f1f1","on-error":"#ffffff","surface-bright":"#f9f9f9","surface-variant":"#e2e2e2","primary":"#0058bc","secondary":"#5f5e60","surface":"#f9f9f9","inverse-primary":"#adc6ff","surface-container-high":"#e8e8e8","on-tertiary":"#ffffff","on-primary-fixed-variant":"#004493","on-tertiary-fixed-variant":"#454749","on-tertiary-fixed":"#1a1c1d","secondary-fixed":"#e4e2e4","surface-container-highest":"#e2e2e2","inverse-surface":"#2f3131","primary-fixed":"#d8e2ff","on-tertiary-container":"#fcfcfe","surface-tint":"#005bc1","tertiary-fixed-dim":"#c6c6c8","tertiary":"#5a5c5e","on-secondary-container":"#636264","surface-dim":"#dadada","on-secondary-fixed":"#1b1b1d","tertiary-container":"#737576","on-surface-variant":"#414755","on-secondary-fixed-variant":"#474649","outline":"#717786","secondary-fixed-dim":"#c8c6c8","on-secondary":"#ffffff","on-surface":"#1a1c1c","on-background":"#1a1c1c","surface-container":"#eeeeee","primary-container":"#0070eb","error-container":"#ffdad6","surface-container-lowest":"#ffffff","error":"#ba1a1a","secondary-container":"#e2dfe1","surface-container-low":"#f3f3f4","on-primary-container":"#fefcff","tertiary-fixed":"#e2e2e4","outline-variant":"#c1c6d7","primary-fixed-dim":"#adc6ff","background":"#f9f9f9","on-error-container":"#93000a"},borderRadius:{DEFAULT:"0.25rem",lg:"0.5rem",xl:"0.75rem",full:"9999px"},spacing:{sm:"8px",gutter:"20px","container-margin":"32px",lg:"24px",xs:"4px",unit:"4px",md:"16px",xl:"40px"},fontSize:{"label-md":["13px",{lineHeight:"18px",letterSpacing:"0.01em",fontWeight:"500"}],"body-md":["15px",{lineHeight:"22px",fontWeight:"400"}],"headline-lg":["32px",{lineHeight:"40px",letterSpacing:"-0.01em",fontWeight:"600"}],"body-lg":["17px",{lineHeight:"24px",fontWeight:"400"}],"headline-md":["24px",{lineHeight:"32px",letterSpacing:"-0.01em",fontWeight:"600"}],"display-lg":["48px",{lineHeight:"56px",letterSpacing:"-0.02em",fontWeight:"700"}],"label-sm":["11px",{lineHeight:"16px",letterSpacing:"0.02em",fontWeight:"600"}],"title-lg":["20px",{lineHeight:"28px",fontWeight:"600"}]}}}}`,
          type: "text/javascript",
        },
      ],
    },
  },
});
