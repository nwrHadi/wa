<template>
  <div class="min-h-screen flex items-center justify-center p-gutter bg-[#F5F5F7]"
    style="background-image: radial-gradient(at 0% 0%, rgba(0, 88, 188, 0.05) 0px, transparent 50%), radial-gradient(at 100% 100%, rgba(0, 88, 188, 0.03) 0px, transparent 50%)">
    <main class="w-full max-w-[440px] flex flex-col items-center">
      <!-- Logo Branding -->
      <div class="mb-xl text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[#0058bc] shadow-lg mb-md">
          <span class="material-symbols-outlined text-white text-[36px]" style="font-variation-settings:'FILL' 1">hub</span>
        </div>
        <h1 class="text-[#1a1c1c] font-bold tracking-tight text-[36px] leading-[44px] sm:text-[48px] sm:leading-[56px]" style="letter-spacing:-0.02em">WA Control</h1>
        <p class="text-[#5f5e60] text-[14px] sm:text-[15px] mt-xs">Enterprise Management</p>
      </div>

      <!-- Login Card -->
      <div class="glass-card w-full rounded-[28px] p-xl">
        <div class="mb-lg">
          <h2 class="text-[#1a1c1c] font-semibold text-[22px] leading-[30px] sm:text-[24px] sm:leading-[32px]">Welcome back</h2>
          <p class="text-[#414755] text-[13px] mt-xs">Sign in to manage your automated flows.</p>
        </div>

        <form @submit.prevent="onSubmit" class="flex flex-col gap-y-md">
          <!-- Username Input -->
          <div class="flex flex-col gap-xs">
            <label class="text-[13px] font-medium text-[#1a1c1c] ml-unit" for="username">Username</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-md flex items-center pointer-events-none">
                <span class="material-symbols-outlined text-[#717786] text-[20px]">person</span>
              </div>
              <input
                id="username"
                v-model="username"
                type="text"
                required
                placeholder="Enter your username"
                class="w-full bg-[#EDEDF0] border-none rounded-xl py-3 pl-11 pr-md text-[15px] text-[#1a1c1c] transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white"
              />
            </div>
          </div>

          <!-- Password Input -->
          <div class="flex flex-col gap-xs">
            <label class="text-[13px] font-medium text-[#1a1c1c] ml-unit" for="password">Password</label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-md flex items-center pointer-events-none">
                <span class="material-symbols-outlined text-[#717786] text-[20px]">lock</span>
              </div>
              <input
                id="password"
                v-model="password"
                type="password"
                required
                placeholder="••••••••"
                class="w-full bg-[#EDEDF0] border-none rounded-xl py-3 pl-11 pr-md text-[15px] text-[#1a1c1c] transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white"
              />
            </div>
          </div>

          <!-- Submit -->
          <button
            :disabled="loading"
            type="submit"
            class="mt-md w-full bg-[#0058bc] text-white font-semibold py-4 rounded-xl shadow-md hover:bg-[#004493] active:scale-[0.98] transition-all duration-300 flex items-center justify-center gap-x-sm text-[17px] disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {{ loading ? 'Signing in...' : 'Sign In' }}
            <span v-if="!loading" class="material-symbols-outlined">arrow_forward</span>
          </button>
        </form>

        <!-- Error message -->
        <p v-if="error" class="mt-md text-[#ba1a1a] text-[13px] flex items-center gap-xs">
          <span class="material-symbols-outlined text-[16px]">error</span>
          {{ error }}
        </p>
      </div>

      <!-- Footer -->
      <footer class="mt-xl text-center">
        <p class="text-[#5f5e60] text-[15px]">WhatsApp Gateway Platform</p>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false });

const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

const { request } = useApi();

const onSubmit = async () => {
  loading.value = true;
  error.value = "";

  try {
    console.log("Sending login request...", { username: username.value });
    const result = await request<{ token: string }>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({
        username: username.value,
        password: password.value,
      }),
    });

    if (!result || !result.token) {
      console.error("Invalid response structure:", result);
      error.value = "Login response invalid. Please try again.";
      return;
    }

    console.log("Login successful, token received");
    localStorage.setItem("wa_token", result.token);
    console.log("Token stored, navigating to dashboard...");
    
    // Use navigateTo and wait for it to complete
    await navigateTo("/dashboard");
    console.log("Navigation completed");
  } catch (err: any) {
    console.error("Login error:", err);
    error.value = err?.data?.message || "Login failed. Check your credentials.";
  } finally {
    loading.value = false;
  }
};
</script>
