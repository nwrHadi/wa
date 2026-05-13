<template>
  <div>
    <!-- Page Header -->
    <header class="mb-xl">
      <div class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
        <div>
          <h2 class="text-[#1a1c1c] font-semibold text-[26px] leading-[34px] sm:text-[32px] sm:leading-[40px]" style="letter-spacing:-0.01em">Overview</h2>
          <p class="text-[#5f5e60] text-[14px] sm:text-[15px] mt-xs">Real-time status of your WhatsApp bot infrastructure.</p>
        </div>
        <button
          @click="refresh"
          class="flex items-center gap-sm px-lg py-sm bg-white border border-[rgba(193,198,215,0.3)] rounded-xl text-[13px] font-medium text-[#1a1c1c] hover:bg-[#f3f3f4] transition-colors shadow-sm self-start sm:self-auto"
        >
          <span class="material-symbols-outlined text-[18px]">refresh</span>
          Refresh
        </button>
      </div>
      <p v-if="apiState !== 'unknown'" :class="apiState === 'online' ? 'text-green-600' : 'text-[#ba1a1a]'" class="text-[11px] font-semibold uppercase tracking-widest mt-sm flex items-center gap-xs">
        <span class="w-2 h-2 rounded-full inline-block" :class="apiState === 'online' ? 'bg-green-500' : 'bg-[#ba1a1a]'"></span>
        API {{ apiState }}
      </p>
    </header>

    <!-- KPI Stat Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-gutter mb-xl">
      <!-- Connected Devices -->
      <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex flex-col gap-sm">
        <div class="flex justify-between items-start">
          <div class="p-sm bg-[rgba(0,88,188,0.1)] text-[#0058bc] rounded-xl">
            <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">sensors</span>
          </div>
          <span class="px-2 py-0.5 bg-green-100 text-green-700 rounded-full text-[11px] font-semibold flex items-center gap-xs">
            <span class="material-symbols-outlined text-[14px]">trending_up</span> Live
          </span>
        </div>
        <div>
          <p class="text-[#5f5e60] text-[13px] font-medium">Connected Devices</p>
          <h3 class="text-[#1a1c1c] font-semibold mt-xs text-[20px] leading-7 sm:text-[24px] sm:leading-[32px]">{{ summary.connectedDevices }}</h3>
        </div>
      </div>

      <!-- Messages Sent -->
      <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex flex-col gap-sm">
        <div class="flex justify-between items-start">
          <div class="p-sm bg-[rgba(0,88,188,0.1)] text-[#0058bc] rounded-xl">
            <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">send</span>
          </div>
          <span class="px-2 py-0.5 bg-green-100 text-green-700 rounded-full text-[11px] font-semibold flex items-center gap-xs">
            <span class="material-symbols-outlined text-[14px]">trending_up</span>
          </span>
        </div>
        <div>
          <p class="text-[#5f5e60] text-[13px] font-medium">Messages Sent</p>
          <h3 class="text-[#1a1c1c] font-semibold mt-xs text-[20px] leading-7 sm:text-[24px] sm:leading-[32px]">{{ summary.sent }}</h3>
        </div>
      </div>

      <!-- Failed -->
      <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex flex-col gap-sm">
        <div class="flex justify-between items-start">
          <div class="p-sm bg-[rgba(186,26,26,0.1)] text-[#ba1a1a] rounded-xl">
            <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">error</span>
          </div>
          <span class="px-2 py-0.5 bg-red-100 text-red-700 rounded-full text-[11px] font-semibold flex items-center gap-xs">
            <span class="material-symbols-outlined text-[14px]">trending_down</span>
          </span>
        </div>
        <div>
          <p class="text-[#5f5e60] text-[13px] font-medium">Failed</p>
          <h3 class="text-[#1a1c1c] font-semibold mt-xs text-[20px] leading-7 sm:text-[24px] sm:leading-[32px]">{{ summary.failed }}</h3>
        </div>
      </div>

      <!-- Processing -->
      <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex flex-col gap-sm">
        <div class="flex justify-between items-start">
          <div class="p-sm bg-[rgba(115,117,118,0.1)] text-[#5a5c5e] rounded-xl">
            <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">sync</span>
          </div>
          <span class="px-2 py-0.5 bg-[#e8e8e8] text-[#5f5e60] rounded-full text-[11px] font-semibold">
            Steady
          </span>
        </div>
        <div>
          <p class="text-[#5f5e60] text-[13px] font-medium">Processing</p>
          <h3 class="text-[#1a1c1c] font-semibold mt-xs text-[20px] leading-7 sm:text-[24px] sm:leading-[32px]">{{ summary.processing }}</h3>
        </div>
      </div>
    </div>

    <!-- Visualization Area -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-gutter mb-xl">
      <!-- Activity Chart -->
      <div class="lg:col-span-2 bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
        <div class="flex justify-between items-center mb-lg">
          <h4 class="text-[#1a1c1c] font-semibold text-[18px] leading-6 sm:text-[20px] sm:leading-[28px]">Message Activity</h4>
          <div class="flex gap-sm">
            <span class="px-sm py-unit text-[11px] font-bold uppercase tracking-widest bg-[#e2e2e2] text-[#1a1c1c] rounded-lg">Live</span>
          </div>
        </div>
        <div class="h-64 w-full relative flex items-end gap-2">
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[40%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[60%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[55%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[80%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[70%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[90%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[75%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.3)] hover:bg-[rgba(0,88,188,0.5)] transition-colors rounded-t-sm h-[85%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[60%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[65%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.4)] hover:bg-[rgba(0,88,188,0.6)] transition-colors rounded-t-sm h-[95%]"></div>
          <div class="flex-1 bg-[rgba(0,88,188,0.2)] hover:bg-[rgba(0,88,188,0.4)] transition-colors rounded-t-sm h-[50%]"></div>
          <div class="absolute inset-0 flex flex-col justify-between pointer-events-none">
            <div class="border-t border-dashed border-[rgba(193,198,215,0.3)] w-full"></div>
            <div class="border-t border-dashed border-[rgba(193,198,215,0.3)] w-full"></div>
            <div class="border-t border-dashed border-[rgba(193,198,215,0.3)] w-full"></div>
          </div>
        </div>
      </div>

      <!-- Status / Webhook Events -->
      <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
        <h4 class="text-[#1a1c1c] font-semibold mb-lg text-[18px] leading-6 sm:text-[20px] sm:leading-[28px]">Webhook Event Scope</h4>
        <div class="space-y-md">
          <div v-for="event in webhookEvents" :key="event" class="flex items-center gap-md">
            <div class="w-8 h-8 bg-[rgba(0,88,188,0.1)] rounded-lg flex items-center justify-center">
              <span class="material-symbols-outlined text-[#0058bc] text-[16px]">webhook</span>
            </div>
            <span class="text-[13px] font-medium text-[#1a1c1c]">{{ event }}</span>
          </div>
        </div>
        <div class="mt-lg pt-lg border-t border-[rgba(193,198,215,0.1)]">
          <div class="flex items-center gap-md">
            <div class="w-10 h-10 bg-[#eeeeee] rounded-lg flex items-center justify-center">
              <span class="material-symbols-outlined text-[#0058bc]">public</span>
            </div>
            <div>
              <p class="text-[13px] font-semibold text-[#1a1c1c]">API Status</p>
              <p class="text-[11px] text-[#5f5e60]">{{ summary.updatedAt ? 'Updated ' + summary.updatedAt : 'Awaiting data' }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Activity Table -->
    <div class="bg-white rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] overflow-hidden">
      <div class="p-lg border-b border-[rgba(193,198,215,0.1)] flex justify-between items-center">
        <h4 class="text-[#1a1c1c] font-semibold text-[18px] leading-6 sm:text-[20px] sm:leading-[28px]">System Status</h4>
        <NuxtLink to="/logs" class="text-[#0058bc] text-[13px] font-medium hover:underline">View All Logs</NuxtLink>
      </div>
      <div class="p-lg">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-gutter">
          <div class="flex items-center gap-md p-md bg-[#f3f3f4] rounded-xl">
            <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
              <span class="material-symbols-outlined text-green-600" style="font-variation-settings:'FILL' 1">check_circle</span>
            </div>
            <div>
              <p class="text-[13px] font-semibold text-[#1a1c1c]">API Online</p>
              <p class="text-[11px] text-[#5f5e60]">All systems operational</p>
            </div>
          </div>
          <div class="flex items-center gap-md p-md bg-[#f3f3f4] rounded-xl">
            <div class="w-10 h-10 bg-[rgba(0,88,188,0.1)] rounded-lg flex items-center justify-center">
              <span class="material-symbols-outlined text-[#0058bc]" style="font-variation-settings:'FILL' 1">smartphone</span>
            </div>
            <div>
              <p class="text-[13px] font-semibold text-[#1a1c1c]">{{ summary.connectedDevices }} Devices</p>
              <p class="text-[11px] text-[#5f5e60]">Connected instances</p>
            </div>
          </div>
          <div class="flex items-center gap-md p-md bg-[#f3f3f4] rounded-xl">
            <div class="w-10 h-10 bg-[rgba(0,88,188,0.1)] rounded-lg flex items-center justify-center">
              <span class="material-symbols-outlined text-[#0058bc]" style="font-variation-settings:'FILL' 1">send</span>
            </div>
            <div>
              <p class="text-[13px] font-semibold text-[#1a1c1c]">{{ summary.sent }} Sent</p>
              <p class="text-[11px] text-[#5f5e60]">Total messages delivered</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: "default" });

const { request } = useApi();
const apiState = ref("unknown");
const summary = ref({
  connectedDevices: 0,
  processing: 0,
  sent: 0,
  failed: 0,
  updatedAt: "",
});

const webhookEvents = [
  "message.received",
  "message.sent",
  "message.failed",
  "device.connected",
  "device.disconnected",
  "qr.updated",
];

const refresh = async () => {
  try {
    const data = await request<typeof summary.value>("/api/v1/dashboard/summary");
    summary.value = data;
    apiState.value = "online";
  } catch {
    apiState.value = "offline";
  }
};

onMounted(refresh);
</script>
