import express from "express";
import pino from "pino";
import { startDevice, stopDevice, sendMessage, getDeviceStatus, listActiveSessions, } from "./device-manager.js";
const app = express();
const log = pino({ level: process.env.LOG_LEVEL || "info" });
const port = Number(process.env.PORT || 8090);
const allowedOrigin = process.env.CORS_ORIGIN || "http://localhost:3000";
app.use((req, res, next) => {
    res.header("Access-Control-Allow-Origin", allowedOrigin);
    res.header("Vary", "Origin");
    res.header("Access-Control-Allow-Methods", "GET,POST,OPTIONS");
    res.header("Access-Control-Allow-Headers", "Content-Type,Authorization");
    if (req.method === "OPTIONS") {
        return res.sendStatus(204);
    }
    return next();
});
app.use(express.json({ limit: "2mb" }));
app.get("/healthz", (_req, res) => {
    res.json({ status: "ok" });
});
// Get device status (QR code, connection state, phone number)
app.get("/devices/:deviceKey/status", (req, res) => {
    const { deviceKey } = req.params;
    const status = getDeviceStatus(deviceKey);
    res.json(status);
});
// Start device connection (triggers QR generation)
app.post("/devices/:deviceKey/connect", async (req, res) => {
    const { deviceKey } = req.params;
    try {
        await startDevice(deviceKey);
        res.status(202).json({ status: "connecting", deviceKey });
    }
    catch (err) {
        log.error({ err, deviceKey }, "failed to start device");
        res.status(500).json({ error: err.message });
    }
});
// Stop device connection
app.post("/devices/:deviceKey/disconnect", async (req, res) => {
    const { deviceKey } = req.params;
    try {
        await stopDevice(deviceKey);
        res.status(200).json({ status: "disconnected", deviceKey });
    }
    catch (err) {
        log.error({ err, deviceKey }, "failed to stop device");
        res.status(500).json({ error: err.message });
    }
});
// Send message from device
app.post("/devices/:deviceKey/messages/send", async (req, res) => {
    const { deviceKey } = req.params;
    const { toNumber, messageBody } = req.body || {};
    if (!toNumber || !messageBody) {
        return res
            .status(400)
            .json({ error: "toNumber and messageBody are required" });
    }
    try {
        const messageId = await sendMessage(deviceKey, toNumber, messageBody);
        res.status(200).json({ status: "sent", messageId });
    }
    catch (err) {
        log.error({ err, deviceKey, toNumber }, "failed to send message");
        res.status(500).json({ error: err.message });
    }
});
// List active sessions
app.get("/sessions", (_req, res) => {
    const sessions = listActiveSessions();
    res.json({ sessions });
});
app.listen(port, () => {
    log.info({ port }, "baileys gateway running");
});
