import {
  default as makeWASocket,
  useMultiFileAuthState,
  WASocket,
  Browsers,
  DisconnectReason,
  fetchLatestBaileysVersion,
  isJidBroadcast,
  proto,
} from "@whiskeysockets/baileys";
import pino from "pino";
import axios from "axios";
import fs from "fs";
import path from "path";
import { Boom } from "@hapi/boom";

const log = pino({ level: process.env.LOG_LEVEL || "info" });

interface DeviceSession {
  socket: WASocket | null;
  qrCode: string | null;
  phoneNumber: string | null;
  isConnected: boolean;
}

const SESSIONS_DIR = process.env.SESSIONS_DIR || "./sessions";
const BACKEND_EVENT_URL =
  process.env.BACKEND_EVENT_URL || "http://localhost:8080/internal/events/baileys";
const BACKEND_SHARED_TOKEN = process.env.BACKEND_SHARED_TOKEN || "change-me";
const MAX_RECONNECT_ATTEMPTS = Number(process.env.MAX_RECONNECT_ATTEMPTS || 6);
const ENABLE_PLACEHOLDER_QR = process.env.ENABLE_PLACEHOLDER_QR === "true";
const AUTO_RESET_AUTH_ON_LOGGED_OUT =
  process.env.AUTO_RESET_AUTH_ON_LOGGED_OUT !== "false";

// Store active device sessions in memory
const activeSessions = new Map<string, DeviceSession>();
const reconnectAttempts = new Map<string, number>();
const reconnectTimers = new Map<string, NodeJS.Timeout>();

function clearReconnectTimer(deviceKey: string) {
  const timer = reconnectTimers.get(deviceKey);
  if (timer) {
    clearTimeout(timer);
    reconnectTimers.delete(deviceKey);
  }
}

function resetReconnectState(deviceKey: string) {
  clearReconnectTimer(deviceKey);
  reconnectAttempts.delete(deviceKey);
}

function removeSessionFiles(deviceKey: string) {
  const sessionPath = path.join(SESSIONS_DIR, deviceKey);
  try {
    if (fs.existsSync(sessionPath)) {
      fs.rmSync(sessionPath, { recursive: true, force: true });
      log.warn({ deviceKey, sessionPath }, "removed stale auth session files");
    }
  } catch (err) {
    log.error({ err, deviceKey, sessionPath }, "failed to remove session files");
  }
}

export async function startDevice(deviceKey: string): Promise<void> {
  log.info({ deviceKey }, "starting device connection");
  clearReconnectTimer(deviceKey);

  // Gracefully end existing socket (if any) but preserve current session data.
  const existingSession = activeSessions.get(deviceKey);
  if (existingSession?.socket) {
    log.info({ deviceKey }, "closing existing device socket before restart");
    try {
      existingSession.socket.end(undefined);
    } catch (err) {
      log.error({ err, deviceKey }, "error closing existing socket before restart");
    }
  }

  const sessionPath = path.join(SESSIONS_DIR, deviceKey);
  if (!fs.existsSync(sessionPath)) {
    fs.mkdirSync(sessionPath, { recursive: true });
  }

  const { state, saveCreds } = await useMultiFileAuthState(sessionPath);

  if (ENABLE_PLACEHOLDER_QR) {
    // Optional fallback for UI testing only.
    try {
      const QRCode = await import("qrcode");
      const placeholderQr = `DEMO-${deviceKey}-${Date.now()}`;
      const qrDataUrl = await QRCode.toDataURL(placeholderQr, {
        width: 300,
        margin: 2,
        color: { dark: "#000000", light: "#FFFFFF" },
      });

      const current = activeSessions.get(deviceKey);
      activeSessions.set(deviceKey, {
        socket: null,
        qrCode: qrDataUrl,
        phoneNumber: current?.phoneNumber ?? null,
        isConnected: false,
      });

      await publishEvent("device.qr_updated", {
        deviceKey,
        qrText: placeholderQr,
        qrDataUrl,
        timestamp: new Date().toISOString(),
      });
      log.info({ deviceKey }, "generated placeholder QR code");
    } catch (err) {
      log.error({ err, deviceKey }, "failed to generate placeholder QR");
    }
  }

  try {
    const { version, isLatest } = await fetchLatestBaileysVersion();
    log.info({ deviceKey, version, isLatest }, "resolved baileys whatsapp version");

    const socket = makeWASocket({
      auth: state,
      logger: pino({ level: process.env.BAILEYS_LOG_LEVEL || "silent" }),
      printQRInTerminal: false,
      version,
      browser: Browsers.macOS("WA Control Center"),
      defaultQueryTimeoutMs: 60000,
      connectTimeoutMs: 60000,
      keepAliveIntervalMs: 30000,
      emitOwnEvents: true,
      generateHighQualityLinkPreview: true,
      markOnlineOnConnect: false,
      syncFullHistory: false,
    });

    // Handle QR code display
    socket.ev.on("connection.update", async (update) => {
      const { connection, lastDisconnect, qr } = update;
      log.info({ deviceKey, update: { connection, qr: !!qr, hasDisconnect: !!lastDisconnect } }, "connection update received");

      if (qr) {
        log.info({ deviceKey }, "QR code detected, generating data URL");
        // Generate QR code data URL
        try {
          const QRCode = await import("qrcode");
          const qrDataUrl = await QRCode.toDataURL(qr, {
            width: 300,
            margin: 2,
            color: { dark: "#000000", light: "#FFFFFF" },
          });

          // Store QR in active session
          const session = activeSessions.get(deviceKey);
          if (session) {
            session.qrCode = qrDataUrl;
          }

          // Publish QR update event to backend
          await publishEvent("device.qr_updated", {
            deviceKey,
            qrText: qr,
            qrDataUrl,
            timestamp: new Date().toISOString(),
          });

          log.info({ deviceKey }, "QR code generated");
        } catch (err) {
          log.error({ err, deviceKey }, "failed to generate QR code");
        }
      }

      if (connection === "connecting") {
        await publishEvent("device.connecting", {
          deviceKey,
          status: "connecting",
          timestamp: new Date().toISOString(),
        });
      }

      if (connection === "open") {
        const phoneNumber = socket.user?.id?.split(":")[0] || "";
        const session = activeSessions.get(deviceKey);
        if (session) {
          session.isConnected = true;
          session.phoneNumber = phoneNumber;
          session.qrCode = null; // Clear QR after connection
        }
        resetReconnectState(deviceKey);

        await publishEvent("device.connected", {
          deviceKey,
          status: "connected",
          phoneNumber,
          timestamp: new Date().toISOString(),
        });

        log.info({ deviceKey, phoneNumber }, "device connected");
      }

      if (connection === "close") {
        const boomError = lastDisconnect?.error as Boom | undefined;
        const statusCode = boomError?.output?.statusCode;
        const reason = statusCode || "unknown";
        const shouldReconnect =
          statusCode !== DisconnectReason.loggedOut &&
          statusCode !== DisconnectReason.forbidden;
        const currentSession = activeSessions.get(deviceKey);

        log.warn(
          {
            deviceKey,
            statusCode,
            disconnectReason: DisconnectReason[statusCode as DisconnectReason] || "unknown",
            errorMessage: boomError?.message,
          },
          "device connection closed",
        );

        if (shouldReconnect) {
          const attempts = (reconnectAttempts.get(deviceKey) || 0) + 1;
          reconnectAttempts.set(deviceKey, attempts);

          // Keep the session visible as disconnected while retrying.
          if (currentSession) {
            currentSession.socket = null;
            currentSession.isConnected = false;
          }

          if (attempts > MAX_RECONNECT_ATTEMPTS) {
            resetReconnectState(deviceKey);
            await publishEvent("device.disconnected", {
              deviceKey,
              status: "disconnected",
              reason: reason || "max_retries_exceeded",
              timestamp: new Date().toISOString(),
            });
            log.warn({ deviceKey, attempts }, "max reconnect attempts reached, stopping retries");
            return;
          }

          const delayMs = Math.min(2000 * attempts, 15000);
          const timer = setTimeout(() => {
            startDevice(deviceKey).catch((err) => {
              log.error({ err, deviceKey }, "failed to reconnect device");
            });
          }, delayMs);
          reconnectTimers.set(deviceKey, timer);

          await publishEvent("device.reconnecting", {
            deviceKey,
            status: "connecting",
            reason,
            attempts,
            delayMs,
            timestamp: new Date().toISOString(),
          });
        } else {
          // Permanent disconnect
          if (currentSession) {
            currentSession.socket = null;
            currentSession.isConnected = false;
            currentSession.qrCode = null;
          }
          resetReconnectState(deviceKey);

          const isLoggedOut = statusCode === DisconnectReason.loggedOut;
          if (isLoggedOut && AUTO_RESET_AUTH_ON_LOGGED_OUT) {
            removeSessionFiles(deviceKey);
            setTimeout(() => {
              startDevice(deviceKey).catch((err) => {
                log.error({ err, deviceKey }, "failed to restart after auth reset");
              });
            }, 1000);
          }

          await publishEvent("device.disconnected", {
            deviceKey,
            status: "disconnected",
            reason,
            timestamp: new Date().toISOString(),
          });

          log.info({ deviceKey, reason }, "device disconnected");
        }
      }
    });

    // Handle incoming messages
    socket.ev.on("messages.upsert", async (m) => {
      for (const msg of m.messages) {
        if (!msg.key.fromMe) {
          const sender = msg.key.remoteJid || "";
          const body =
            msg.message?.conversation ||
            msg.message?.extendedTextMessage?.text ||
            "[media/unsupported message type]";

          await publishEvent("message.received", {
            deviceKey,
            messageId: msg.key.id,
            sender,
            body,
            timestamp: msg.messageTimestamp
              ? new Date(parseInt(msg.messageTimestamp.toString()) * 1000).toISOString()
              : new Date().toISOString(),
          });
        }
      }
    });

    // Save credentials whenever they're updated
    socket.ev.on("creds.update", saveCreds);

    // Store active session
    const current = activeSessions.get(deviceKey);
    activeSessions.set(deviceKey, {
      socket,
      qrCode: current?.qrCode ?? null,
      phoneNumber: current?.phoneNumber ?? null,
      isConnected: false,
    });

    log.info({ deviceKey }, "device socket initialized");
  } catch (err) {
    log.error({ err, deviceKey }, "failed to start device");
    activeSessions.delete(deviceKey);
    throw err;
  }
}

export async function stopDevice(deviceKey: string): Promise<void> {
  log.info({ deviceKey }, "stopping device");
  resetReconnectState(deviceKey);

  const session = activeSessions.get(deviceKey);
  if (session && session.socket) {
    try {
      session.socket.end(undefined);
    } catch (err) {
      log.error({ err, deviceKey }, "error stopping socket");
    }
  }

  activeSessions.delete(deviceKey);

  await publishEvent("device.disconnected", {
    deviceKey,
    status: "disconnected",
    reason: "user_initiated",
    timestamp: new Date().toISOString(),
  });
}

export async function sendMessage(
  deviceKey: string,
  toNumber: string,
  messageBody: string,
): Promise<string> {
  log.info({ deviceKey, toNumber }, "sending message");

  const session = activeSessions.get(deviceKey);
  if (!session || !session.socket || !session.isConnected) {
    throw new Error("device not connected");
  }

  try {
    const jid = normalizeRecipientJid(toNumber);

    const result = await session.socket.sendMessage(jid, {
      text: messageBody,
    });

    const messageId = result?.key?.id || "";

    await publishEvent("message.sent", {
      deviceKey,
      messageId,
      toNumber,
      body: messageBody,
      timestamp: new Date().toISOString(),
    });

    return messageId;
  } catch (err) {
    log.error({ err, deviceKey, toNumber }, "failed to send message");

    await publishEvent("message.failed", {
      deviceKey,
      toNumber,
      body: messageBody,
      error: (err as Error).message,
      timestamp: new Date().toISOString(),
    });

    throw err;
  }
}

export function getDeviceStatus(deviceKey: string): {
  isConnected: boolean;
  qrCode: string | null;
  phoneNumber: string | null;
} {
  const session = activeSessions.get(deviceKey);
  if (!session) {
    return { isConnected: false, qrCode: null, phoneNumber: null };
  }

  return {
    isConnected: session.isConnected,
    qrCode: session.qrCode,
    phoneNumber: session.phoneNumber,
  };
}

export function listActiveSessions(): string[] {
  return Array.from(activeSessions.keys());
}

function normalizeRecipientJid(toNumber: string): string {
  const trimmedNumber = toNumber.trim();
  if (trimmedNumber.includes("@")) {
    return trimmedNumber;
  }

  let digits = trimmedNumber.replace(/\D/g, "");
  if (digits.startsWith("00")) {
    digits = digits.slice(2);
  } else if (digits.startsWith("0")) {
    digits = `62${digits.slice(1)}`;
  } else if (digits.startsWith("8")) {
    digits = `62${digits}`;
  }

  return `${digits}@s.whatsapp.net`;
}

async function publishEvent(type: string, payload: unknown) {
  try {
    await axios.post(
      BACKEND_EVENT_URL,
      { type, payload },
      {
        timeout: 5000,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${BACKEND_SHARED_TOKEN}`,
        },
      },
    );
  } catch (error) {
    log.error({ error, type }, "failed to publish event to backend");
  }
}
