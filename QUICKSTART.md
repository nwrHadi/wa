# WhatsApp Bot Platform - Quick Start Guide

## Overview

Anda sekarang memiliki platform WhatsApp bot yang lengkap dengan:
- ✅ Dashboard untuk monitoring devices dan messages
- ✅ Device management (add, connect, disconnect)
- ✅ Message sending API
- ✅ Realtime status updates
- ✅ Webhook system untuk integrasi
- ✅ Audit logging

## 1. Memulai Sistem

### Terminal 1: Database Initialization
```bash
cd c:\Users\IT\Documents\HADI\WA\backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
go run ./cmd/dbinit
# Output: database ready: wa_platform
```

### Terminal 2: API Backend
```bash
cd c:\Users\IT\Documents\HADI\WA\backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
$env:REDIS_OPTIONAL='true'
go run ./cmd/api
# Output: [api] listening on :8080
```

### Terminal 3: Baileys Gateway (WhatsApp Connection)
```bash
cd c:\Users\IT\Documents\HADI\WA\baileys-gateway
npm install  # First time only, untuk install qrcode dependency
npm run dev
# Output: baileys gateway running on port 8090
```

### Terminal 4: Worker (Message Queue Processor)
```bash
cd c:\Users\IT\Documents\HADI\WA\backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
$env:GATEWAY_URL='http://localhost:8090'
go run ./cmd/worker
# Output: [worker] worker started...
```

### Terminal 5: Frontend
```bash
cd c:\Users\IT\Documents\HADI\WA\frontend
$env:NUXT_PUBLIC_API_BASE='http://localhost:8080'
npm run dev
# Output: ✔ Vite server built, listening on http://localhost:3000
```

## 2. Login ke Dashboard

1. Buka browser: `http://localhost:3000`
2. Login dengan:
   - Username: `admin`
   - Password: `admin123`
3. Anda akan masuk ke dashboard

## 3. Menambah Device WhatsApp

### Cara 1: Via Dashboard UI

1. Klik menu "Connected WhatsApp Devices"
2. Isi form:
   - **Device Key**: nama unik device (contoh: `device1`, `my_wa`)
   - **Device Label**: deskripsi friendly (contoh: `WhatsApp Marketing`)
3. Klik "Add Device"
4. Klik tombol "Connect" pada device yang baru dibuat
5. Sebuah QR code akan muncul
6. Buka WhatsApp di phone Anda
7. Buka Settings → Linked Devices → Link a Device
8. Scan QR code dari dashboard dengan camera
9. Tunggu sampai status berubah menjadi "connected" dan phone number muncul

### Cara 2: Via API

```bash
# Add Device
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceKey": "device1",
    "label": "My WhatsApp",
    "status": "disconnected"
  }'

# Connect Device (trigger QR generation)
curl -X POST http://localhost:8090/devices/device1/connect

# Get Device Status (with QR code)
curl http://localhost:8090/devices/device1/status
```

## 4. Mengirim Pesan WhatsApp

### Cara 1: Via Dashboard

1. Pastikan device sudah "connected"
2. Klik tombol "Send Message" pada device
3. Isi form:
   - **To (Phone Number)**: nomor tujuan (format: `+62812345678` atau `62812345678`)
   - **Message**: isi pesan
4. Klik "Send"
5. Message status akan berubah menjadi "processing" → "sent"

### Cara 2: Via API

```bash
curl -X POST http://localhost:8080/api/v1/messages/send \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceKey": "device1",
    "toNumber": "+62812345678",
    "messageBody": "Hello dari WhatsApp Bot!",
    "idempotencyKey": "unique-key-123"
  }'

# Response (202 Accepted):
{
  "messageId": "1",
  "externalRefId": "unique-key-123",
  "status": "processing",
  "createdAt": "2025-05-11T10:30:00Z"
}
```

## 5. Monitoring Messages

### Via Dashboard
1. Klik menu "Message Logs"
2. Lihat semua pesan yang dikirim/diterima
3. Filter by device, status, tanggal

### Via API
```bash
# Get all messages
curl -X GET http://localhost:8080/api/v1/logs/messages \
  -H "Authorization: Bearer <JWT_TOKEN>"

# Get message details
curl -X GET http://localhost:8080/api/v1/messages/{messageId} \
  -H "Authorization: Bearer <JWT_TOKEN>"

# Get messages by device
curl -X GET "http://localhost:8080/api/v1/devices/{deviceId}/messages?limit=50&offset=0" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## 6. Webhooks untuk Integrasi

Webhooks memungkinkan sistem Anda menerima notifikasi real-time tentang:
- Device connect/disconnect
- Message sent/failed
- Device error

### Setup Webhook

```bash
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Webhook",
    "targetUrl": "https://your-server.com/webhook",
    "secret": "webhook-secret-key",
    "eventFilters": "message.sent,message.failed",
    "enabled": true
  }'
```

### Event Payload Example

```json
{
  "type": "message.sent",
  "payload": {
    "deviceKey": "device1",
    "messageId": "123",
    "toNumber": "+62812345678",
    "body": "Hello!",
    "timestamp": "2025-05-11T10:30:00Z"
  },
  "timestamp": "2025-05-11T10:30:00Z"
}
```

Setiap webhook POST akan include header:
- `X-WA-Signature`: HMAC-SHA256 signature untuk verify webhook authenticity

Verify signature:
```python
import hmac
import hashlib

signature = request.headers.get('X-WA-Signature')
payload = request.get_data()
secret = 'webhook-secret-key'

expected_signature = hmac.new(
    secret.encode(),
    payload,
    hashlib.sha256
).hexdigest()

assert signature == expected_signature  # Valid webhook
```

## 7. API Endpoints Summary

### Authentication
- `POST /api/v1/auth/login` - Login (returns JWT)

### Devices
- `GET /api/v1/devices` - List all devices
- `POST /api/v1/devices` - Create device
- `POST /api/v1/devices/{deviceKey}/connect` - Connect device (trigger QR)
- `POST /api/v1/devices/{deviceKey}/disconnect` - Disconnect device
- `GET /api/v1/devices/{deviceKey}/status` - Get device status + QR

### Messages
- `POST /api/v1/messages/send` - Send message
- `GET /api/v1/messages/{messageId}` - Get message details
- `GET /api/v1/devices/{deviceId}/messages` - List messages for device
- `GET /api/v1/logs/messages` - List all messages (paginated)

### Dashboard
- `GET /api/v1/dashboard/summary` - KPI summary
- `GET /api/v1/dashboard/timeline` - Timeline view

### Webhooks
- `GET /api/v1/webhooks` - List webhooks
- `POST /api/v1/webhooks` - Create webhook
- `GET /api/v1/webhooks/{webhookId}` - Get webhook details

### Realtime
- `GET /api/v1/realtime` - SSE stream untuk real-time updates

## 8. Troubleshooting

### Device tidak connect
1. Pastikan gateway running di terminal 3
2. Check gateway logs untuk error
3. Pastikan phone memiliki WhatsApp installed
4. Coba scan QR ulang

### Message tidak terkirim
1. Check worker logs di terminal 4
2. Pastikan device dalam status "connected"
3. Cek format nomor (+62 atau 0)
4. Check message logs untuk error detail

### Frontend tidak bisa login
1. Pastikan backend running (terminal 2)
2. Check browser console untuk error
3. Cek JWT secret di backend config

### Redis error
1. Jika error "cannot connect to redis", itu normal
2. Sistem akan fallback ke in-process queue
3. Pastikan `REDIS_OPTIONAL='true'` di env variables

## 9. Deployment (Android Box - No Docker)

### Prerequisites
- Go 1.23+
- Node.js 20+
- MariaDB 10.6+
- PowerShell atau terminal bash

### Steps

1. **Build Backend**
```bash
cd backend
go build -o wa-api.exe ./cmd/api
go build -o wa-worker.exe ./cmd/worker
go build -o wa-dbinit.exe ./cmd/dbinit
```

2. **Build Gateway**
```bash
cd baileys-gateway
npm install --production
npm run build
```

3. **Build Frontend**
```bash
cd frontend
npm install --production
npm run build
# Hasilnya ada di frontend/.output/public
```

4. **Setup Environment**
```bash
# Create .env file
DB_HOST=127.0.0.1
DB_PORT=3307
DB_USER=waapp
DB_PASSWORD=waapp123
DB_NAME=wa_platform
API_PORT=8080
REDIS_OPTIONAL=true
GATEWAY_URL=http://localhost:8090
```

5. **Run Services**
```bash
# Terminal 1
./wa-dbinit.exe

# Terminal 2
./wa-api.exe

# Terminal 3
cd baileys-gateway && npm start

# Terminal 4
./wa-worker.exe

# Terminal 5
cd frontend && npm run preview
```

## 10. Next Steps

- Integrate dengan CRM/ticketing system
- Setup auto-responses
- Create dashboard untuk analytics
- Implement rate limiting
- Add multi-user support dengan role-based access

---

**Dokumentasi lengkap**: Baca file README.md di root project

**API Documentation**: Swagger/OpenAPI docs akan tersedia di `/api/v1/docs` (jika diimplementasi)

**Support**: Refer ke conversation history untuk technical details
