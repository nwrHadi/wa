<template>
  <div>
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:justify-between sm:items-end gap-4 mb-xl">
      <div>
        <div class="flex items-center gap-sm text-[#5f5e60] mb-sm">
          <span class="text-[13px] font-medium">Management</span>
          <span class="material-symbols-outlined text-[16px]">chevron_right</span>
          <span class="text-[13px] font-semibold text-[#1a1c1c]">Devices</span>
        </div>
        <h2 class="text-[#1a1c1c] font-semibold text-[26px] leading-[34px] sm:text-[32px] sm:leading-[40px]" style="letter-spacing:-0.01em">Connected Devices</h2>
        <p class="text-[#5f5e60] text-[14px] sm:text-[15px] mt-xs">Manage your active WhatsApp instances and monitor their health.</p>
      </div>
      <div class="flex gap-2 sm:gap-md">
        <div class="flex items-center gap-sm px-3 sm:px-md py-sm bg-white rounded-lg border border-[rgba(193,198,215,0.2)] shadow-sm">
          <span class="w-2 h-2 bg-green-500 rounded-full shrink-0"></span>
          <span class="text-[12px] sm:text-[13px] font-medium text-[#414755] whitespace-nowrap">{{ connectedCount }} Active</span>
        </div>
        <div class="flex items-center gap-sm px-3 sm:px-md py-sm bg-white rounded-lg border border-[rgba(193,198,215,0.2)] shadow-sm">
          <span class="w-2 h-2 bg-red-500 rounded-full shrink-0"></span>
          <span class="text-[12px] sm:text-[13px] font-medium text-[#414755] whitespace-nowrap">{{ offlineCount }} Offline</span>
        </div>
      </div>
    </div>

    <!-- Devices Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-gutter mb-xl">
      <!-- Device Cards -->
      <div
        v-for="device in devices"
        :key="device.id"
        :class="[
          'bg-white p-lg rounded-[20px] shadow-[0px_4px_20px_rgba(0,0,0,0.04)] transition-all border',
          device.status === 'disconnected'
            ? 'border-[rgba(193,198,215,0.2)] opacity-80 hover:opacity-100'
            : 'border-transparent hover:border-[rgba(0,88,188,0.1)] hover:shadow-[0px_10px_30px_rgba(0,0,0,0.08)]'
        ]"
      >
        <div class="flex justify-between items-start mb-lg">
          <div class="flex items-center gap-md">
            <div class="w-12 h-12 bg-[rgba(0,88,188,0.05)] rounded-2xl flex items-center justify-center text-[#0058bc]">
              <span class="material-symbols-outlined text-[28px]">smartphone</span>
            </div>
            <div>
              <h3 class="text-[#1a1c1c] font-semibold text-[18px] leading-6 sm:text-[20px] sm:leading-[28px]">{{ device.label }}</h3>
              <p class="text-[#5f5e60] text-[13px] font-medium">{{ device.deviceKey }}</p>
              <p v-if="device.phoneNumber" class="text-[#0058bc] text-[13px] font-medium">{{ device.phoneNumber }}</p>
            </div>
          </div>
          <!-- Status Badge -->
          <div
            v-if="device.status === 'connected'"
            class="px-sm py-1 bg-green-100 text-green-700 rounded-full flex items-center gap-1"
          >
            <span class="material-symbols-outlined text-[14px]" style="font-variation-settings:'FILL' 1">check_circle</span>
            <span class="text-[11px] font-bold uppercase tracking-wider">Connected</span>
          </div>
          <div
            v-else-if="device.status === 'connecting'"
            class="px-sm py-1 bg-yellow-100 text-yellow-700 rounded-full flex items-center gap-1"
          >
            <span class="material-symbols-outlined text-[14px]">sync</span>
            <span class="text-[11px] font-bold uppercase tracking-wider">Connecting</span>
          </div>
          <div
            v-else
            class="px-sm py-1 bg-[#e2e2e2] text-[#5f5e60] rounded-full flex items-center gap-1"
          >
            <span class="material-symbols-outlined text-[14px]">error</span>
            <span class="text-[11px] font-bold uppercase tracking-wider">Offline</span>
          </div>
        </div>

        <!-- QR Code Display -->
        <div v-if="device.qrDataUrl && qrModals[device.id]" class="mb-lg">
          <div class="bg-[#f3f3f4] rounded-xl p-md flex flex-col items-center gap-md">
            <img :src="device.qrDataUrl" :alt="`QR for ${device.label}`" class="w-40 h-40 rounded-lg" />
            <p class="text-[13px] text-[#5f5e60] text-center">Scan with WhatsApp to connect</p>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex flex-wrap gap-1.5 sm:gap-sm border-t border-[rgba(193,198,215,0.1)] pt-md">
          <button
            v-if="device.status === 'connected'"
            @click="showMessageForm(device)"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-medium bg-[#eeeeee] hover:bg-[#e2e2e2] text-[#414755] rounded-lg transition-colors px-3"
          >
            Send Message
          </button>
          <button
            v-if="device.status === 'disconnected'"
            @click="connectDevice(device.deviceKey)"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-semibold bg-[#0058bc] text-white rounded-lg shadow-sm hover:opacity-90 transition-opacity px-3"
          >
            Connect
          </button>
          <button
            v-if="device.status === 'connecting' && qrModals[device.id]"
            @click="qrModals[device.id] = false"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-medium bg-[#eeeeee] hover:bg-[#e2e2e2] text-[#414755] rounded-lg transition-colors px-3"
          >
            Hide QR
          </button>
          <button
            v-if="device.status === 'connecting' && !qrModals[device.id] && device.qrDataUrl"
            @click="qrModals[device.id] = true"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-medium bg-[#eeeeee] hover:bg-[#e2e2e2] text-[#414755] rounded-lg transition-colors px-3"
          >
            Show QR
          </button>
          <button
            v-if="device.status !== 'disconnected'"
            @click="disconnectDevice(device.deviceKey)"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-medium bg-[rgba(186,26,26,0.05)] hover:bg-[rgba(186,26,26,0.1)] text-[#ba1a1a] rounded-lg transition-colors px-3"
          >
            Disconnect
          </button>
          <button
            @click="openDeleteDialog(device)"
            :disabled="deletingDeviceKey === device.deviceKey"
            class="flex-1 sm:flex-initial py-2 text-[12px] sm:text-[13px] font-medium bg-[rgba(186,26,26,0.08)] hover:bg-[rgba(186,26,26,0.16)] text-[#ba1a1a] rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed px-3"
          >
            {{ deletingDeviceKey === device.deviceKey ? "Deleting..." : "Delete" }}
          </button>
        </div>
      </div>

      <!-- Add Device Card -->
      <div
        class="border-2 border-dashed border-[rgba(193,198,215,0.4)] rounded-[20px] flex flex-col items-center justify-center p-lg min-h-[220px] group hover:border-[rgba(0,88,188,0.4)] hover:bg-[rgba(0,88,188,0.02)] transition-all cursor-pointer"
        @click="showAddDeviceModal = true"
      >
        <div class="w-14 h-14 rounded-full bg-[#eeeeee] flex items-center justify-center text-[#5f5e60] mb-md group-hover:scale-110 group-hover:bg-[#0058bc] group-hover:text-white transition-all duration-300">
          <span class="material-symbols-outlined text-[32px]">add_to_home_screen</span>
        </div>
        <span class="text-[18px] sm:text-[20px] font-semibold text-[#414755]">Add New Instance</span>
        <p class="text-[13px] text-[#5f5e60] text-center max-w-[200px] mt-xs">Connect a new WhatsApp account to WA Control.</p>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="devices.length === 0" class="text-center py-xl">
      <div class="w-16 h-16 bg-[rgba(0,88,188,0.1)] rounded-full flex items-center justify-center mx-auto mb-md">
        <span class="material-symbols-outlined text-[#0058bc] text-[32px]">smartphone</span>
      </div>
      <p class="text-[#5f5e60] text-[15px]">No devices yet. Add one to get started!</p>
    </div>

    <!-- Add Device Modal -->
    <div
      v-if="showAddDeviceModal"
      class="fixed inset-0 bg-[rgba(47,49,49,0.4)] backdrop-blur-sm z-[100] flex items-center justify-center p-md"
      @click="showAddDeviceModal = false"
    >
      <div class="bg-white w-full max-w-md rounded-[24px] shadow-[0px_20px_40px_rgba(0,0,0,0.1)] overflow-hidden" @click.stop>
        <div class="p-xl">
          <div class="flex justify-between items-center mb-lg">
            <div>
              <h2 class="text-[#1a1c1c] font-semibold" style="font-size:22px;line-height:30px">Add New Device</h2>
              <p class="text-[#5f5e60] text-[13px] mt-xs">Connect a new WhatsApp instance.</p>
            </div>
            <button @click="showAddDeviceModal = false" class="p-2 hover:bg-[#eeeeee] rounded-full transition-colors text-[#5f5e60]">
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <div class="flex flex-col gap-md">
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Device Key</label>
              <input
                v-model="newDevice.deviceKey"
                placeholder="e.g., sales-bot-01"
                class="w-full bg-[#f3f3f4] border-none rounded-xl px-md py-3 text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
              />
            </div>
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Device Label</label>
              <input
                v-model="newDevice.label"
                placeholder="e.g., Sales Support Bot"
                class="w-full bg-[#f3f3f4] border-none rounded-xl px-md py-3 text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
              />
            </div>
          </div>
        </div>
        <div class="px-xl pb-xl flex gap-md">
          <button @click="showAddDeviceModal = false" class="flex-1 py-3 bg-[#eeeeee] text-[#1a1c1c] font-semibold rounded-xl hover:bg-[#e2e2e2] transition-colors">
            Cancel
          </button>
          <button
            @click="addDeviceAndClose"
            :disabled="!newDevice.deviceKey || loading"
            class="flex-1 py-3 bg-[#0058bc] text-white font-semibold rounded-xl shadow-md hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loading ? "Adding..." : "Add Device" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Message Send Modal -->
    <div
      v-if="selectedDevice && showMessageModal"
      class="fixed inset-0 bg-[rgba(47,49,49,0.4)] backdrop-blur-sm z-[100] flex items-center justify-center p-md"
      @click="showMessageModal = false"
    >
      <div class="bg-white w-full max-w-md rounded-[24px] shadow-[0px_20px_40px_rgba(0,0,0,0.1)] overflow-hidden" @click.stop>
        <div class="p-xl">
          <div class="flex justify-between items-center mb-lg">
            <div>
              <h2 class="text-[#1a1c1c] font-semibold" style="font-size:22px;line-height:30px">Send Message</h2>
              <p class="text-[#5f5e60] text-[13px] mt-xs">via <strong>{{ selectedDevice.label }}</strong></p>
            </div>
            <button @click="showMessageModal = false" class="p-2 hover:bg-[#eeeeee] rounded-full transition-colors text-[#5f5e60]">
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <div class="flex flex-col gap-md">
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Recipient Number</label>
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-md flex items-center pointer-events-none">
                  <span class="material-symbols-outlined text-[#717786] text-[20px]">phone</span>
                </div>
                <input
                  v-model="messageForm.toNumber"
                  placeholder="+62 or phone number"
                  class="w-full bg-[#f3f3f4] border-none rounded-xl py-3 pl-11 pr-md text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all"
                />
              </div>
            </div>
            <div class="flex flex-col gap-xs">
              <label class="text-[13px] font-medium text-[#414755] px-1">Message</label>
              <textarea
                v-model="messageForm.body"
                placeholder="Type your message..."
                rows="4"
                class="w-full bg-[#f3f3f4] border-none rounded-xl px-md py-3 text-[15px] focus:outline-none focus:ring-2 focus:ring-[#0058bc] focus:bg-white transition-all resize-none"
              ></textarea>
            </div>
          </div>
        </div>
        <div class="px-xl pb-xl flex gap-md">
          <button @click="showMessageModal = false" class="flex-1 py-3 bg-[#eeeeee] text-[#1a1c1c] font-semibold rounded-xl hover:bg-[#e2e2e2] transition-colors">
            Cancel
          </button>
          <button
            @click="sendMessage"
            :disabled="!messageForm.toNumber || !messageForm.body || messageSending"
            class="flex-1 py-3 bg-[#0058bc] text-white font-semibold rounded-xl shadow-md hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-sm"
          >
            <span class="material-symbols-outlined text-[20px]">send</span>
            {{ messageSending ? "Sending..." : "Send" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Device Modal -->
    <div
      v-if="showDeleteModal && deviceToDelete"
      class="fixed inset-0 bg-[rgba(47,49,49,0.4)] backdrop-blur-sm z-[110] flex items-center justify-center p-md"
      @click="closeDeleteDialog"
    >
      <div class="bg-white w-full max-w-md rounded-[24px] shadow-[0px_20px_40px_rgba(0,0,0,0.1)] overflow-hidden" @click.stop>
        <div class="p-xl">
          <div class="w-12 h-12 rounded-2xl bg-[rgba(186,26,26,0.1)] text-[#ba1a1a] flex items-center justify-center mb-md">
            <span class="material-symbols-outlined">delete</span>
          </div>
          <h2 class="text-[#1a1c1c] font-semibold" style="font-size:22px;line-height:30px">Delete Device?</h2>
          <p class="text-[#5f5e60] text-[13px] mt-sm">
            Device <strong>{{ deviceToDelete.label }}</strong> ({{ deviceToDelete.deviceKey }}) akan dihapus permanen dari dashboard.
          </p>
          <p v-if="deleteError" class="text-[#ba1a1a] text-[13px] mt-md flex items-center gap-xs">
            <span class="material-symbols-outlined text-[16px]">error</span>
            {{ deleteError }}
          </p>
        </div>
        <div class="px-xl pb-xl flex gap-md">
          <button
            @click="closeDeleteDialog"
            :disabled="deletingDeviceKey === deviceToDelete.deviceKey"
            class="flex-1 py-3 bg-[#eeeeee] text-[#1a1c1c] font-semibold rounded-xl hover:bg-[#e2e2e2] transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            @click="confirmDeleteDevice"
            :disabled="deletingDeviceKey === deviceToDelete.deviceKey"
            class="flex-1 py-3 bg-[#ba1a1a] text-white font-semibold rounded-xl shadow-md hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ deletingDeviceKey === deviceToDelete.deviceKey ? "Deleting..." : "Delete" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";

definePageMeta({ layout: "default" });

const { request } = useApi();
const config = useRuntimeConfig();
const gatewayBase = `${config.public.gatewayBase || "/gateway"}`.replace(/\/$/, "");

interface Device {
  id: string;
  deviceKey: string;
  label: string;
  status: string;
  phoneNumber?: string;
  qrText?: string;
  qrDataUrl?: string;
}

const devices = ref<Device[]>([]);
const newDevice = ref({ deviceKey: "", label: "" });
const selectedDevice = ref<Device | null>(null);
const showMessageModal = ref(false);
const showAddDeviceModal = ref(false);
const messageForm = ref({ toNumber: "", body: "" });
const loading = ref(false);
const messageSending = ref(false);
const deletingDeviceKey = ref("");
const showDeleteModal = ref(false);
const deviceToDelete = ref<Device | null>(null);
const deleteError = ref("");
const qrModals = ref<Record<string, boolean>>({});

const connectedCount = computed(() => devices.value.filter(d => d.status === "connected").length);
const offlineCount = computed(() => devices.value.filter(d => d.status === "disconnected").length);

const gatewayRequest = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await $fetch<T>(`${gatewayBase}${path}`, {
    ...init,
    retry: 0,
    timeout: 10000,
  });

  return response;
};

const loadDevices = async () => {
  try {
    const result = await request<{ items: Device[] }>("/api/v1/devices");
    const previousByKey = new Map(devices.value.map((d) => [d.deviceKey, d]));

    devices.value = (result.items || []).map((d) => {
      const previous = previousByKey.get(d.deviceKey);
      const id = String(d.id || previous?.id || d.deviceKey);

      const keepConnecting = previous?.status === "connecting" && qrModals.value[id];

      return {
        ...d,
        id,
        status: d.status || (keepConnecting ? "connecting" : "disconnected"),
        phoneNumber: d.phoneNumber || previous?.phoneNumber,
        qrText: d.qrText || previous?.qrText,
        qrDataUrl: previous?.qrDataUrl,
      };
    });
  } catch (err) {
    console.error("Failed to load devices:", err);
  }
};

const addDevice = async () => {
  if (!newDevice.value.deviceKey) {
    alert("Device key is required");
    return;
  }

  loading.value = true;
  try {
    await request("/api/v1/devices", {
      method: "POST",
      body: JSON.stringify({
        deviceKey: newDevice.value.deviceKey,
        label: newDevice.value.label || newDevice.value.deviceKey,
        status: "disconnected",
      }),
    });
    newDevice.value = { deviceKey: "", label: "" };
    await loadDevices();
  } catch (err) {
    alert("Failed to add device: " + (err as any).message);
  } finally {
    loading.value = false;
  }
};

const addDeviceAndClose = async () => {
  await addDevice();
  if (!loading.value) {
    showAddDeviceModal.value = false;
  }
};

const connectDevice = async (deviceKey: string) => {
  try {
    const encodedDeviceKey = encodeURIComponent(deviceKey);
    await gatewayRequest(`/devices/${encodedDeviceKey}/connect`, {
      method: "POST",
    });

    let attempts = 0;
    const interval = setInterval(async () => {
      try {
        const status = await gatewayRequest<any>(`/devices/${encodedDeviceKey}/status`);

        const device = devices.value.find((d) => d.deviceKey === deviceKey);
        if (device) {
          if (status.qrCode) {
            device.qrDataUrl = status.qrCode;
            device.status = "connecting";
            qrModals.value[device.id] = true;
          }

          if (status.isConnected) {
            device.status = "connected";
            device.phoneNumber = status.phoneNumber || device.phoneNumber;
            device.qrText = undefined;
            device.qrDataUrl = undefined;
            qrModals.value[device.id] = false;
            clearInterval(interval);
            await loadDevices();
          }
        }
      } catch (err) {
        console.error("Failed to check status:", err);
      }

      attempts++;
      if (attempts > 120) {
        clearInterval(interval);
      }
    }, 1000);

    await loadDevices();
  } catch (err) {
    alert("Failed to connect device: " + (err as any).message);
  }
};

const disconnectDevice = async (deviceKey: string) => {
  try {
    const encodedDeviceKey = encodeURIComponent(deviceKey);
    await gatewayRequest(`/devices/${encodedDeviceKey}/disconnect`, {
      method: "POST",
    });

    await loadDevices();
  } catch (err) {
    alert("Failed to disconnect device: " + (err as any).message);
  }
};

const openDeleteDialog = (device: Device) => {
  deviceToDelete.value = device;
  deleteError.value = "";
  showDeleteModal.value = true;
};

const closeDeleteDialog = () => {
  if (deviceToDelete.value && deletingDeviceKey.value === deviceToDelete.value.deviceKey) {
    return;
  }
  showDeleteModal.value = false;
  deviceToDelete.value = null;
  deleteError.value = "";
};

const confirmDeleteDevice = async () => {
  const device = deviceToDelete.value;
  if (!device) {
    return;
  }

  deleteError.value = "";
  deletingDeviceKey.value = device.deviceKey;
  try {
    if (device.status !== "disconnected") {
      try {
        await disconnectDevice(device.deviceKey);
      } catch {
        // Tetap lanjut coba hapus walau disconnect gagal.
      }
    }

    try {
      await request(`/api/v1/devices/${encodeURIComponent(device.deviceKey)}`, {
        method: "DELETE",
      });
    } catch (err: any) {
      const statusCode = err?.statusCode ?? err?.response?.status;
      if (statusCode !== 404) {
        throw err;
      }

      // Fallback untuk backend lama/proxy yang tidak meneruskan path param.
      await request(`/api/v1/devices?deviceKey=${encodeURIComponent(device.deviceKey)}`, {
        method: "DELETE",
      });
    }

    delete qrModals.value[device.id];
    await loadDevices();
    deletingDeviceKey.value = "";
    closeDeleteDialog();
  } catch (err) {
    deleteError.value = "Failed to delete device. Pastikan backend sudah restart dengan route terbaru.";
  } finally {
    deletingDeviceKey.value = "";
  }
};

const showMessageForm = (device: Device) => {
  selectedDevice.value = device;
  messageForm.value = { toNumber: "", body: "" };
  showMessageModal.value = true;
};

const sendMessage = async () => {
  if (!selectedDevice.value || !messageForm.value.toNumber || !messageForm.value.body) {
    alert("Please fill in all fields");
    return;
  }

  messageSending.value = true;
  try {
    await request("/api/v1/messages/send", {
      method: "POST",
      body: JSON.stringify({
        deviceKey: selectedDevice.value.deviceKey,
        toNumber: messageForm.value.toNumber,
        messageBody: messageForm.value.body,
      }),
    });
    
    showMessageModal.value = false;
    messageForm.value = { toNumber: "", body: "" };
    alert("Message sent! Check the logs for delivery status.");
  } catch (err) {
    alert("Failed to send message: " + (err as any).message);
  } finally {
    messageSending.value = false;
  }
};

const startPolling = () => {
  setInterval(loadDevices, 5000);
};

onMounted(() => {
  loadDevices();
  startPolling();
});
</script>
