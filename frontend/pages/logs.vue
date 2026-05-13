<template>
  <div>
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:justify-between sm:items-end gap-4 mb-xl">
      <div>
        <div class="flex items-center gap-sm text-[#5f5e60] mb-sm">
          <span class="text-[13px] font-medium">Audit Trail</span>
          <span class="material-symbols-outlined text-[16px]">chevron_right</span>
          <span class="text-[13px] font-semibold text-[#1a1c1c]">Message Logs</span>
        </div>
        <h2 class="text-[#1a1c1c] font-semibold text-[26px] leading-[34px] sm:text-[32px] sm:leading-[40px]" style="letter-spacing:-0.01em">Message Delivery Logs</h2>
        <p class="text-[#5f5e60] text-[14px] sm:text-[15px] mt-xs">Track every outgoing message and its delivery status.</p>
      </div>
      <div class="flex gap-2 sm:gap-md self-start sm:self-auto">
        <button class="flex items-center gap-sm px-3 sm:px-md py-sm bg-white border border-[rgba(193,198,215,0.2)] rounded-xl text-[12px] sm:text-[13px] font-medium text-[#414755] hover:bg-[#f3f3f4] transition-colors shadow-sm whitespace-nowrap">
          <span class="material-symbols-outlined text-[16px] sm:text-[18px]">download</span>
          Export Logs
        </button>
        <button @click="load" class="flex items-center gap-sm px-3 sm:px-md py-sm bg-[#0058bc] text-white rounded-xl text-[12px] sm:text-[13px] font-semibold hover:opacity-90 transition-opacity shadow-md whitespace-nowrap">
          <span class="material-symbols-outlined text-[16px] sm:text-[18px]">refresh</span>
          Refresh
        </button>
      </div>
    </div>

    <!-- Stats Bar -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-gutter mb-xl">
      <div class="bg-white p-md rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex items-center gap-md">
        <div class="w-10 h-10 bg-[rgba(0,88,188,0.1)] rounded-xl flex items-center justify-center text-[#0058bc]">
          <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">send</span>
        </div>
        <div>
          <p class="text-[13px] text-[#5f5e60]">Total Records</p>
          <p class="text-[#1a1c1c] font-semibold text-[20px]">{{ items.length }}</p>
        </div>
      </div>
      <div class="bg-white p-md rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex items-center gap-md">
        <div class="w-10 h-10 bg-green-100 rounded-xl flex items-center justify-center text-green-600">
          <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">check_circle</span>
        </div>
        <div>
          <p class="text-[13px] text-[#5f5e60]">Delivered</p>
          <p class="text-[#1a1c1c] font-semibold text-[20px]">{{ sentCount }}</p>
        </div>
      </div>
      <div class="bg-white p-md rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] flex items-center gap-md">
        <div class="w-10 h-10 bg-[rgba(186,26,26,0.1)] rounded-xl flex items-center justify-center text-[#ba1a1a]">
          <span class="material-symbols-outlined" style="font-variation-settings:'FILL' 1">error</span>
        </div>
        <div>
          <p class="text-[13px] text-[#5f5e60]">Failed</p>
          <p class="text-[#1a1c1c] font-semibold text-[20px]">{{ failedCount }}</p>
        </div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="flex flex-wrap gap-md mb-lg items-center">
      <div class="relative">
        <span class="absolute inset-y-0 left-3 flex items-center pointer-events-none">
          <span class="material-symbols-outlined text-[#717786] text-[18px]">search</span>
        </span>
        <input
          v-model="searchQuery"
          placeholder="Search recipient..."
          class="bg-white border border-[rgba(193,198,215,0.3)] rounded-xl py-2 pl-10 pr-md text-[13px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] w-60"
        />
      </div>
      <select
        v-model="statusFilter"
        class="bg-white border border-[rgba(193,198,215,0.3)] rounded-xl py-2 px-md text-[13px] focus:outline-none focus:ring-2 focus:ring-[#0058bc]"
      >
        <option value="">All Statuses</option>
        <option value="sent">Sent</option>
        <option value="failed">Failed</option>
        <option value="pending">Pending</option>
      </select>
    </div>

    <!-- Logs Table -->
    <div class="bg-white rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] overflow-hidden mb-xl">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b border-[rgba(193,198,215,0.1)]">
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Timestamp</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Recipient</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Status</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">ID</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in filteredItems"
              :key="item.id"
              class="border-b border-[rgba(193,198,215,0.05)] hover:bg-[rgba(0,88,188,0.02)] transition-colors"
            >
              <td class="px-lg py-md text-[13px] text-[#5f5e60] whitespace-nowrap">
                {{ item.createdAt ? new Date(item.createdAt).toLocaleString() : '-' }}
              </td>
              <td class="px-lg py-md text-[13px] font-medium text-[#1a1c1c]">{{ item.to }}</td>
              <td class="px-lg py-md">
                <span
                  v-if="item.status === 'sent'"
                  class="inline-flex items-center gap-1 px-sm py-0.5 bg-green-100 text-green-700 rounded-full text-[11px] font-bold uppercase tracking-wider"
                >
                  <span class="material-symbols-outlined text-[12px]" style="font-variation-settings:'FILL' 1">check_circle</span> Sent
                </span>
                <span
                  v-else-if="item.status === 'failed'"
                  class="inline-flex items-center gap-1 px-sm py-0.5 bg-red-100 text-[#ba1a1a] rounded-full text-[11px] font-bold uppercase tracking-wider"
                >
                  <span class="material-symbols-outlined text-[12px]" style="font-variation-settings:'FILL' 1">error</span> Failed
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1 px-sm py-0.5 bg-yellow-100 text-yellow-700 rounded-full text-[11px] font-bold uppercase tracking-wider"
                >
                  <span class="material-symbols-outlined text-[12px]">schedule</span> {{ item.status || 'Pending' }}
                </span>
              </td>
              <td class="px-lg py-md text-[13px] text-[#5f5e60] font-mono">#{{ item.id }}</td>
            </tr>
            <tr v-if="filteredItems.length === 0">
              <td colspan="4" class="px-lg py-xl text-center text-[#5f5e60] text-[13px]">
                <div class="flex flex-col items-center gap-md">
                  <span class="material-symbols-outlined text-[#c1c6d7] text-[48px]">inbox</span>
                  No message logs found.
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

definePageMeta({ layout: "default" });

const { request } = useApi();
const items = ref<Array<{ id: number; to: string; status: string; createdAt: string }>>([]);
const searchQuery = ref("");
const statusFilter = ref("");

const filteredItems = computed(() => {
  return items.value.filter(item => {
    const matchSearch = !searchQuery.value || item.to.toLowerCase().includes(searchQuery.value.toLowerCase());
    const matchStatus = !statusFilter.value || item.status === statusFilter.value;
    return matchSearch && matchStatus;
  });
});

const sentCount = computed(() => items.value.filter(i => i.status === "sent").length);
const failedCount = computed(() => items.value.filter(i => i.status === "failed").length);

const load = async () => {
  const result = await request<{ items: Array<{ id: number; to: string; status: string; createdAt: string }> }>("/api/v1/logs/messages");
  items.value = result.items || [];
};

onMounted(load);
</script>

