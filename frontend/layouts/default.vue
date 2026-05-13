<template>
  <div class="flex min-h-screen bg-[#F5F5F7] text-[#1a1c1c]">
    <!-- Mobile overlay backdrop -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 bg-[rgba(47,49,49,0.4)] backdrop-blur-sm z-40 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- Sidebar -->
    <aside
      :class="[
        'fixed left-0 top-0 h-screen w-[280px] bg-[rgba(249,249,249,0.8)] backdrop-blur-3xl border-r border-white/20 shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex flex-col p-gutter z-50 transition-transform duration-300',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full',
        'lg:translate-x-0'
      ]"
    >
      <!-- Brand -->
      <div class="mb-xl flex items-center gap-md">
        <div class="w-10 h-10 bg-[#0070eb] rounded-xl flex items-center justify-center shadow-md shrink-0">
          <span class="material-symbols-outlined text-white text-[22px]" style="font-variation-settings:'FILL' 1">hub</span>
        </div>
        <div>
          <h1 class="text-[#0058bc] font-bold tracking-tight text-[20px] leading-7">WA Control</h1>
          <p class="text-[#5f5e60] text-[11px] font-semibold uppercase tracking-widest opacity-70">Enterprise Management</p>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 flex flex-col gap-1">
        <NuxtLink
          to="/dashboard"
          :class="[
            'flex items-center gap-md px-md py-sm rounded-lg text-[13px] font-medium transition-all duration-200',
            isActive('/dashboard')
              ? 'bg-[#0058bc] text-white shadow-md scale-[0.98]'
              : 'text-[#5f5e60] hover:bg-[rgba(226,226,226,0.5)]'
          ]"
          @click="sidebarOpen = false"
        >
          <span class="material-symbols-outlined text-[22px]">dashboard</span>
          Dashboard
        </NuxtLink>
        <NuxtLink
          to="/devices"
          :class="[
            'flex items-center gap-md px-md py-sm rounded-lg text-[13px] font-medium transition-all duration-200',
            isActive('/devices')
              ? 'bg-[#0058bc] text-white shadow-md scale-[0.98]'
              : 'text-[#5f5e60] hover:bg-[rgba(226,226,226,0.5)]'
          ]"
          @click="sidebarOpen = false"
        >
          <span class="material-symbols-outlined text-[22px]">smartphone</span>
          Devices
        </NuxtLink>
        <NuxtLink
          to="/logs"
          :class="[
            'flex items-center gap-md px-md py-sm rounded-lg text-[13px] font-medium transition-all duration-200',
            isActive('/logs')
              ? 'bg-[#0058bc] text-white shadow-md scale-[0.98]'
              : 'text-[#5f5e60] hover:bg-[rgba(226,226,226,0.5)]'
          ]"
          @click="sidebarOpen = false"
        >
          <span class="material-symbols-outlined text-[22px]">history</span>
          Message Logs
        </NuxtLink>
        <NuxtLink
          to="/webhook-settings"
          :class="[
            'flex items-center gap-md px-md py-sm rounded-lg text-[13px] font-medium transition-all duration-200',
            isActive('/webhook-settings')
              ? 'bg-[#0058bc] text-white shadow-md scale-[0.98]'
              : 'text-[#5f5e60] hover:bg-[rgba(226,226,226,0.5)]'
          ]"
          @click="sidebarOpen = false"
        >
          <span class="material-symbols-outlined text-[22px]">webhook</span>
          Webhook Settings
        </NuxtLink>
      </nav>

      <!-- Bottom CTA -->
      <div class="mt-auto pt-gutter border-t border-[rgba(193,198,215,0.2)]">
        <NuxtLink
          to="/devices"
          class="w-full py-md px-md bg-[#0070eb] text-white rounded-xl font-semibold shadow-sm hover:opacity-90 transition-opacity flex items-center justify-center gap-sm text-[13px]"
          @click="sidebarOpen = false"
        >
          <span class="material-symbols-outlined text-[20px]">add</span>
          Add Device
        </NuxtLink>
      </div>
    </aside>

    <!-- Main content area -->
    <div class="lg:ml-[280px] flex flex-col min-h-screen flex-1 w-full">
      <!-- Top navigation bar -->
      <header class="sticky top-0 h-16 flex justify-between items-center px-4 sm:px-gutter bg-[rgba(249,249,249,0.6)] backdrop-blur-md border-b border-[rgba(193,198,215,0.1)] z-40">
        <div class="flex items-center gap-3">
          <!-- Hamburger menu button (mobile only) -->
          <button
            class="lg:hidden p-2 -ml-2 text-[#414755] hover:text-[#0058bc] transition-colors rounded-lg hover:bg-[#eeeeee]"
            @click="sidebarOpen = !sidebarOpen"
          >
            <span class="material-symbols-outlined text-[24px]">{{ sidebarOpen ? 'close' : 'menu' }}</span>
          </button>
          <div class="relative w-full max-w-[200px] sm:max-w-md">
            <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-[#717786] text-[20px]">search</span>
            <input
              class="w-full bg-[#f3f3f4] border-none rounded-full pl-10 pr-4 py-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-[rgba(0,88,188,0.2)] transition-all"
              placeholder="Search..."
              type="text"
            />
          </div>
        </div>
        <div class="flex items-center gap-2 sm:gap-md ml-2 sm:ml-lg">
          <button class="p-1.5 sm:p-sm text-[#414755] hover:text-[#0058bc] transition-colors">
            <span class="material-symbols-outlined text-[20px]">notifications</span>
          </button>
          <button class="p-1.5 sm:p-sm text-[#414755] hover:text-[#0058bc] transition-colors hidden sm:block">
            <span class="material-symbols-outlined text-[20px]">help_outline</span>
          </button>
          <div class="h-8 w-px bg-[rgba(193,198,215,0.3)] hidden sm:block"></div>
          <div class="flex items-center gap-1 sm:gap-sm cursor-pointer hover:opacity-80 transition-opacity" @click="logout">
            <div class="w-8 h-8 rounded-full bg-[#0058bc] flex items-center justify-center text-white font-bold text-[13px] shrink-0">A</div>
            <span class="text-[13px] font-semibold text-[#1a1c1c] hidden sm:inline">Admin</span>
            <span class="material-symbols-outlined text-[#717786] text-[18px]">logout</span>
          </div>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 p-4 sm:p-container-margin max-w-[1440px]">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
const sidebarOpen = ref(false);

// Close sidebar on route change (mobile)
watch(() => route.path, () => {
  sidebarOpen.value = false;
});

const isActive = (path: string) => route.path === path || route.path.startsWith(path + "/");

const logout = () => {
  localStorage.removeItem("wa_token");
  navigateTo("/login");
};
</script>
