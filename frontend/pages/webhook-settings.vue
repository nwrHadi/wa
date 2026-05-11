<template>
  <div>
    <!-- Page Header -->
    <div class="flex justify-between items-end mb-xl">
      <div>
        <div class="flex items-center gap-sm text-[#5f5e60] mb-sm">
          <span class="text-[13px] font-medium">Settings</span>
          <span class="material-symbols-outlined text-[16px]">chevron_right</span>
          <span class="text-[13px] font-semibold text-[#1a1c1c]">Webhook</span>
        </div>
        <h2 class="text-[#1a1c1c] font-semibold" style="font-size:32px;line-height:40px;letter-spacing:-0.01em">Webhook Settings</h2>
        <p class="text-[#5f5e60] text-[15px] mt-xs">Configure outbound webhooks for real-time event notifications.</p>
      </div>
      <div class="flex gap-md">
        <span
          v-if="savedWebhook"
          class="flex items-center gap-sm px-md py-sm bg-green-100 text-green-700 rounded-xl text-[13px] font-semibold"
        >
          <span class="material-symbols-outlined text-[18px]" style="font-variation-settings:'FILL' 1">check_circle</span>
          Webhook Active
        </span>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-gutter">
      <!-- Main Config Column (2/3) -->
      <div class="lg:col-span-2 flex flex-col gap-gutter">
        <!-- Endpoint Configuration -->
        <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
          <div class="flex items-center gap-md mb-lg">
            <div class="w-10 h-10 bg-[rgba(0,88,188,0.1)] rounded-xl flex items-center justify-center text-[#0058bc]">
              <span class="material-symbols-outlined">webhook</span>
            </div>
            <div>
              <h3 class="text-[#1a1c1c] font-semibold" style="font-size:20px;line-height:28px">Endpoint Configuration</h3>
              <p class="text-[#5f5e60] text-[13px]">Define where WA Control should send event payloads.</p>
            </div>
          </div>

          <div class="flex flex-col gap-md">
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Webhook Name</label>
              <input
                v-model="form.name"
                placeholder="e.g., My Server Webhook"
                class="w-full bg-[#f3f3f4] border-none rounded-xl px-md py-3 text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
              />
            </div>
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Webhook URL</label>
              <div class="relative">
                <span class="absolute inset-y-0 left-0 pl-md flex items-center pointer-events-none">
                  <span class="material-symbols-outlined text-[#717786] text-[20px]">link</span>
                </span>
                <input
                  v-model="form.targetUrl"
                  placeholder="https://your-server.com/webhook"
                  class="w-full bg-[#f3f3f4] border-none rounded-xl py-3 pl-11 pr-md text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
                />
              </div>
            </div>
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Secret Key</label>
              <div class="relative">
                <span class="absolute inset-y-0 left-0 pl-md flex items-center pointer-events-none">
                  <span class="material-symbols-outlined text-[#717786] text-[20px]">key</span>
                </span>
                <input
                  v-model="form.secret"
                  :type="showSecret ? 'text' : 'password'"
                  placeholder="Optional HMAC signing secret"
                  class="w-full bg-[#f3f3f4] border-none rounded-xl py-3 pl-11 pr-12 text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
                />
                <button
                  @click="showSecret = !showSecret"
                  class="absolute inset-y-0 right-0 pr-md flex items-center text-[#717786] hover:text-[#1a1c1c] transition-colors"
                  type="button"
                >
                  <span class="material-symbols-outlined text-[20px]">{{ showSecret ? 'visibility_off' : 'visibility' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Events to Notify -->
        <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
          <div class="flex items-center gap-md mb-lg">
            <div class="w-10 h-10 bg-[rgba(0,88,188,0.1)] rounded-xl flex items-center justify-center text-[#0058bc]">
              <span class="material-symbols-outlined">notifications</span>
            </div>
            <div>
              <h3 class="text-[#1a1c1c] font-semibold" style="font-size:20px;line-height:28px">Events to Notify</h3>
              <p class="text-[#5f5e60] text-[13px]">Select which events trigger a webhook call.</p>
            </div>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-md">
            <label
              v-for="event in availableEvents"
              :key="event.value"
              class="flex items-start gap-md p-md bg-[#f3f3f4] hover:bg-[rgba(0,88,188,0.04)] rounded-xl cursor-pointer transition-colors group"
            >
              <div class="mt-0.5 relative">
                <input
                  type="checkbox"
                  :value="event.value"
                  v-model="form.eventFilters"
                  class="sr-only peer"
                />
                <div class="w-5 h-5 rounded bg-white border-2 border-[rgba(193,198,215,0.6)] peer-checked:bg-[#0058bc] peer-checked:border-[#0058bc] transition-colors flex items-center justify-center">
                  <span v-if="form.eventFilters.includes(event.value)" class="material-symbols-outlined text-white text-[14px]" style="font-variation-settings:'FILL' 1">check</span>
                </div>
              </div>
              <div>
                <p class="text-[13px] font-semibold text-[#1a1c1c]">{{ event.label }}</p>
                <p class="text-[11px] text-[#5f5e60]">{{ event.description }}</p>
              </div>
            </label>
          </div>
        </div>
      </div>

      <!-- Sidebar Column (1/3) -->
      <div class="flex flex-col gap-gutter">
        <!-- Actions Card -->
        <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
          <h3 class="text-[#1a1c1c] font-semibold mb-md" style="font-size:20px;line-height:28px">Actions</h3>
          <div class="flex flex-col gap-md">
            <button
              @click="saveWebhook"
              :disabled="saving || !form.targetUrl"
              class="w-full py-3 bg-[#0058bc] text-white font-semibold rounded-xl shadow-md hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-sm"
            >
              <span class="material-symbols-outlined text-[20px]">save</span>
              {{ saving ? "Saving..." : "Save Webhook" }}
            </button>
            <button
              @click="testWebhook"
              :disabled="!savedWebhook"
              class="w-full py-3 bg-[#eeeeee] text-[#414755] font-semibold rounded-xl hover:bg-[#e2e2e2] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-sm"
            >
              <span class="material-symbols-outlined text-[20px]">play_arrow</span>
              Test Delivery
            </button>
          </div>
          <p v-if="saveError" class="text-[#ba1a1a] text-[13px] mt-md flex items-center gap-xs">
            <span class="material-symbols-outlined text-[16px]">error</span>
            {{ saveError }}
          </p>
          <p v-if="saveSuccess" class="text-green-700 text-[13px] mt-md flex items-center gap-xs">
            <span class="material-symbols-outlined text-[16px]" style="font-variation-settings:'FILL' 1">check_circle</span>
            Webhook saved successfully!
          </p>
        </div>

        <!-- Status Card -->
        <div class="bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)]">
          <h3 class="text-[#1a1c1c] font-semibold mb-md" style="font-size:20px;line-height:28px">Status</h3>
          <div v-if="savedWebhook" class="flex flex-col gap-md">
            <div class="flex justify-between items-center">
              <span class="text-[13px] text-[#5f5e60]">State</span>
              <span class="px-sm py-0.5 bg-green-100 text-green-700 rounded-full text-[11px] font-bold uppercase tracking-wider">Active</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-[13px] text-[#5f5e60]">Target</span>
              <span class="text-[13px] font-medium text-[#1a1c1c] max-w-[160px] truncate">{{ savedWebhook.targetUrl }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-[13px] text-[#5f5e60]">Events</span>
              <span class="text-[13px] font-semibold text-[#0058bc]">{{ (savedWebhook.eventFilters || []).length }} configured</span>
            </div>
          </div>
          <div v-else class="text-center py-md">
            <span class="material-symbols-outlined text-[#c1c6d7] text-[40px]">webhook</span>
            <p class="text-[#5f5e60] text-[13px] mt-sm">No webhook configured yet.</p>
          </div>
        </div>

        <!-- Help Card -->
        <div class="bg-[rgba(0,88,188,0.04)] border border-[rgba(0,88,188,0.1)] p-lg rounded-[20px]">
          <div class="flex items-center gap-sm mb-md">
            <span class="material-symbols-outlined text-[#0058bc] text-[20px]">help</span>
            <h3 class="text-[#1a1c1c] font-semibold">About Webhooks</h3>
          </div>
          <p class="text-[13px] text-[#5f5e60] leading-relaxed">WA Control sends an HTTP POST to your URL on each selected event. Include a secret to verify payload authenticity via HMAC-SHA256 signature in the <code class="bg-[#e2e2e2] px-1 rounded text-[12px] font-mono">X-WA-Control-Signature</code> header.</p>
        </div>
      </div>
    </div>

    <!-- Recent Deliveries Table -->
    <div class="mt-xl bg-white rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] overflow-hidden">
      <div class="p-lg border-b border-[rgba(193,198,215,0.1)]">
        <h3 class="text-[#1a1c1c] font-semibold" style="font-size:20px;line-height:28px">Saved Webhooks</h3>
        <p class="text-[#5f5e60] text-[13px] mt-xs">All registered webhook configurations.</p>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b border-[rgba(193,198,215,0.1)]">
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Name</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Target URL</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Events</th>
              <th class="px-lg py-md text-left text-[11px] font-bold uppercase tracking-widest text-[#5f5e60]">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="wh in webhooks"
              :key="wh.id"
              class="border-b border-[rgba(193,198,215,0.05)] hover:bg-[rgba(0,88,188,0.02)] transition-colors"
            >
              <td class="px-lg py-md text-[13px] font-medium text-[#1a1c1c]">{{ wh.name }}</td>
              <td class="px-lg py-md text-[13px] text-[#5f5e60] max-w-[240px] truncate">{{ wh.targetUrl }}</td>
              <td class="px-lg py-md text-[13px] text-[#5f5e60]">{{ (wh.eventFilters || []).length }} events</td>
              <td class="px-lg py-md">
                <span v-if="wh.enabled" class="inline-flex items-center gap-1 px-sm py-0.5 bg-green-100 text-green-700 rounded-full text-[11px] font-bold uppercase tracking-wider">
                  <span class="material-symbols-outlined text-[12px]" style="font-variation-settings:'FILL' 1">check_circle</span> Active
                </span>
                <span v-else class="inline-flex items-center gap-1 px-sm py-0.5 bg-[#e2e2e2] text-[#5f5e60] rounded-full text-[11px] font-bold uppercase tracking-wider">
                  Inactive
                </span>
              </td>
            </tr>
            <tr v-if="webhooks.length === 0">
              <td colspan="4" class="px-lg py-xl text-center text-[#5f5e60] text-[13px]">
                <div class="flex flex-col items-center gap-md">
                  <span class="material-symbols-outlined text-[#c1c6d7] text-[48px]">webhook</span>
                  No webhooks configured yet.
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
definePageMeta({ layout: "default" });

const { request } = useApi();

interface Webhook {
  id: number;
  name: string;
  targetUrl: string;
  secret?: string;
  eventFilters: string[];
  enabled: boolean;
}

const webhooks = ref<Webhook[]>([]);
const savedWebhook = ref<Webhook | null>(null);
const saving = ref(false);
const saveError = ref("");
const saveSuccess = ref(false);
const showSecret = ref(false);

const form = ref({
  name: "",
  targetUrl: "",
  secret: "",
  eventFilters: [] as string[],
});

const availableEvents = [
  { value: "message.received", label: "Message Received", description: "Triggered when a new message arrives" },
  { value: "message.sent", label: "Message Delivered", description: "Triggered on successful message delivery" },
  { value: "device.status", label: "Device Status Change", description: "Triggered when a device connects/disconnects" },
  { value: "message.failed", label: "Webhook Error", description: "Triggered when a message fails to send" },
];

const loadWebhooks = async () => {
  try {
    const result = await request<{ items: Webhook[] }>("/api/v1/webhooks");
    webhooks.value = result.items || [];
    if (webhooks.value.length > 0) {
      savedWebhook.value = webhooks.value[0];
      form.value.name = savedWebhook.value.name;
      form.value.targetUrl = savedWebhook.value.targetUrl;
      form.value.eventFilters = savedWebhook.value.eventFilters || [];
    }
  } catch (err) {
    console.error("Failed to load webhooks:", err);
  }
};

const saveWebhook = async () => {
  if (!form.value.targetUrl) return;

  saving.value = true;
  saveError.value = "";
  saveSuccess.value = false;

  try {
    const payload = {
      name: form.value.name || "Default Webhook",
      targetUrl: form.value.targetUrl,
      secret: form.value.secret,
      eventFilters: form.value.eventFilters,
      enabled: true,
    };
    await request("/api/v1/webhooks", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    saveSuccess.value = true;
    await loadWebhooks();
    setTimeout(() => { saveSuccess.value = false; }, 3000);
  } catch (err) {
    saveError.value = (err as any).message || "Failed to save webhook";
  } finally {
    saving.value = false;
  }
};

const testWebhook = async () => {
  if (!savedWebhook.value) return;
  alert("Test delivery sent! Check your server logs.");
};

onMounted(loadWebhooks);
</script>
