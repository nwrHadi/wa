# Implementasi: WhatsApp Bot dengan Device Management dan Message Sending

## Ringkasan Implementasi

Saya telah mengimplementasikan fitur lengkap untuk menambah device WhatsApp dan mengirim pesan. Berikut adalah komponen yang telah ditambahkan:

### ✅ Baileys Gateway Integration

**File baru**: `baileys-gateway/src/device-manager.ts`

Fitur:
- Real Baileys library integration untuk WhatsApp connection
- QR code generation dan display
- Device session persistence (multi-device support)
- Incoming message handler
- Auto-reconnect untuk transient errors
- Event publishing ke backend

Metode utama:
- `startDevice(deviceKey)` - Mulai koneksi WhatsApp dan generate QR code
- `stopDevice(deviceKey)` - Disconnect device
- `sendMessage(deviceKey, toNumber, messageBody)` - Kirim pesan
- `getDeviceStatus(deviceKey)` - Get QR code, connection status, phone number

### ✅ Backend Message Send API

**File baru**: `backend/internal/http/handlers/messages.go`

Endpoints:
- `POST /api/v1/messages/send` - Send message dari device
  - Request: `{ deviceKey, toNumber, messageBody, idempotencyKey }`
  - Response: `{ messageId, externalRefId, status, createdAt }`
- `GET /api/v1/messages/{messageId}` - Get message details
- `GET /api/v1/devices/{deviceId}/messages` - List messages with pagination

Fitur:
- Idempotency key untuk prevent duplicate sends
- Device connection validation
- Realtime event publishing (SSE)
- Queue job enqueue untuk async processing
- Status tracking (processing → sent/failed)

### ✅ Worker for Queue Processing

**Update**: `backend/internal/worker/processor.go`

Fitur:
- Handle `message.send` jobs dari queue
- Call Baileys gateway `/devices/{deviceKey}/messages/send`
- Update message status di database
- Error logging dan retry handling

### ✅ Store Layer Methods

**Update**: `backend/internal/store/store.go`

Metode baru:
- `CreateMessage()` - Create message record dengan LastInsertId
- `GetMessage()` - Get message by ID
- `GetMessageByExternalRef()` - Get message by idempotency key
- `ListMessagesByDevice()` - List dengan pagination
- `UpdateMessageStatusByID()` - Update status by message ID

### ✅ Frontend Devices Page

**Update**: `frontend/pages/devices.vue`

Fitur UI:
- Device list dengan status display (connected/connecting/disconnected)
- Add device form
- Connect device button (trigger QR generation)
- QR code modal dengan display untuk scan
- Message send form dengan modal
- Status polling setiap 5 detik
- Device management (disconnect)

Flow:
1. User klik "Add Device" → simpan ke database
2. User klik "Connect" → trigger gateway, polling untuk QR
3. QR muncul di modal → user scan dengan WhatsApp
4. Status otomatis update ke "connected"
5. "Send Message" button aktif
6. User isi form dan klik send → message queued dan diproses

### ✅ Router Updates

**Update**: `backend/internal/http/server/router.go`

Routes baru:
```go
api.POST("/messages/send", h.Messages.Send)
api.GET("/messages/:messageId", h.Messages.Get)
api.GET("/devices/:deviceId/messages", h.Messages.ListByDevice)
```

## Flow Lengkap: Add Device dan Kirim Pesan

### 1. Add Device

```
User UI (Devices page)
  ↓ (fill form: deviceKey, label)
  ↓ POST /api/v1/devices
Backend API
  ↓ (UpsertDevice → database)
  ↓ (publish device.created event)
Hub (realtime)
  ↓ (broadcast to connected clients)
Frontend Dashboard
  ↓ (device list updated)
```

### 2. Connect Device (Generate QR)

```
User UI
  ↓ (click Connect button)
  ↓ POST /devices/{deviceKey}/connect  [to gateway port 8090]
Baileys Gateway
  ↓ (startDevice → makeWASocket)
  ↓ (QR generated from auth code)
  ↓ (POST /internal/events/baileys with device.qr_updated)
Backend API
  ↓ (store QR in device_sessions)
  ↓ (publish event to realtime hub)
Frontend
  ↓ (polling /devices/{deviceKey}/status)
  ↓ (show QR modal dengan image)
  ↓ (user scan with phone)
Baileys
  ↓ (connection.update event)
  ↓ (POST device.connected event)
Backend
  ↓ (update device status → connected)
  ↓ (store phone number)
  ↓ (publish event)
Frontend
  ↓ (device status → connected, show phone number)
```

### 3. Send Message

```
User UI
  ↓ (fill: toNumber, messageBody)
  ↓ (click Send)
  ↓ POST /api/v1/messages/send
Backend API (messages.go)
  ↓ (check device exists & connected)
  ↓ (check idempotency)
  ↓ (CreateMessage in database)
  ↓ (status = processing)
  ↓ (publish message.created event)
  ↓ (enqueue message.send job)
  ↓ (return 202 Accepted)
Frontend
  ↓ (show success message)
  ↓ (redirect to logs or show status)

Background Processing:
Queue/Redis
  ← (message.send job from API)
Worker (go run ./cmd/worker)
  ↓ (poll queue every 1 second)
  ↓ (handleMessageSend)
  ↓ POST /devices/{deviceKey}/messages/send
Baileys Gateway
  ↓ (socket.sendMessage)
  ↓ (get messageId from response)
  ↓ (return { status: "sent", messageId })
Worker
  ↓ (UpdateMessageStatusByID → status: sent)
  ↓ (waMessageId stored)
  ↓ (logging)

Realtime Updates (if SSE connected):
  → Backend publishes message.sent event
  → Hub broadcasts to connected clients
  → Frontend updates logs in real-time
```

## API Usage Examples

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Response: { "token": "eyJ0eXAi..." }
```

### Add Device
```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceKey": "device1",
    "label": "My WhatsApp",
    "status": "disconnected"
  }'

# Response:
{
  "id": 1,
  "deviceKey": "device1",
  "label": "My WhatsApp",
  "status": "disconnected",
  "createdAt": "2025-05-11T10:30:00Z"
}
```

### Connect Device
```bash
curl -X POST http://localhost:8090/devices/device1/connect

# Response: { "status": "connecting", "deviceKey": "device1" }

# Poll status
curl http://localhost:8090/devices/device1/status

# Response (before scan):
{
  "isConnected": false,
  "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "phoneNumber": null
}

# Response (after scan):
{
  "isConnected": true,
  "qrCode": null,
  "phoneNumber": "+62812345678"
}
```

### Send Message
```bash
curl -X POST http://localhost:8080/api/v1/messages/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceKey": "device1",
    "toNumber": "+62812345678",
    "messageBody": "Hello from WhatsApp Bot!",
    "idempotencyKey": "msg-001"
  }'

# Response (202 Accepted):
{
  "messageId": 1,
  "externalRefId": "msg-001",
  "status": "processing",
  "createdAt": "2025-05-11T10:31:00Z"
}

# Check status
curl http://localhost:8080/api/v1/messages/1 \
  -H "Authorization: Bearer YOUR_TOKEN"

# Response (after worker processes):
{
  "id": 1,
  "deviceId": 1,
  "externalRefId": "msg-001",
  "waMessageId": "3EB0DA...",
  "toNumber": "+62812345678",
  "direction": "outbound",
  "status": "sent",
  "body": "Hello from WhatsApp Bot!",
  "createdAt": "2025-05-11T10:31:00Z",
  "updatedAt": "2025-05-11T10:31:05Z"
}
```

## Startup Instructions

### Terminal 1: Database Init
```bash
cd backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
go run ./cmd/dbinit
```

### Terminal 2: API
```bash
cd backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
$env:REDIS_OPTIONAL='true'
go run ./cmd/api
```

### Terminal 3: Gateway
```bash
cd baileys-gateway
npm install  # first time only - untuk install qrcode
npm run dev
```

### Terminal 4: Worker
```bash
cd backend
$env:DB_HOST='127.0.0.1'
$env:DB_PORT='3307'
$env:DB_USER='waapp'
$env:DB_PASSWORD='waapp123'
$env:DB_NAME='wa_platform'
$env:GATEWAY_URL='http://localhost:8090'
go run ./cmd/worker
```

### Terminal 5: Frontend
```bash
cd frontend
$env:NUXT_PUBLIC_API_BASE='http://localhost:8080'
npm run dev
```

### Browser
```
http://localhost:3000
Login: admin / admin123
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Nuxt 3)                    │
│          ├── Dashboard (KPI, device status)                 │
│          ├── Devices page (add, connect, send msg)          │
│          ├── Logs page (message history)                    │
│          └── SSE realtime stream                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
    ┌──────────────────┼──────────────────┐
    ▼                  │                  ▼
┌────────────┐    ┌─────────┐     ┌──────────────┐
│  API       │    │  Events │     │  Realtime    │
│  Handler   │    │  Handler│     │  Hub (SSE)   │
│  (Go)      │    │         │     │              │
└────────────┘    └─────────┘     └──────────────┘
    ▼                  ▼                  ▲
┌────────────────────────────────────────────────┐
│         Backend API (Echo v4)                  │
│         Port 8080                              │
└────────────────────────────────────────────────┘
    ▼                  ▼                  ▼
┌────────────┐    ┌──────────┐    ┌──────────────┐
│ Store/Repo │    │  Queue   │    │  Services    │
│ (DB access)│    │  (Redis) │    │  (Events)    │
└────────────┘    └──────────┘    └──────────────┘
         │              ▲               ▼
         ▼              │          ┌──────────────┐
    ┌──────────┐        │          │  Webhook    │
    │MariaDB   │        │          │  Service    │
    │Database  │        │          └──────────────┘
    └──────────┘        │
                        │
                   ┌────┴──────────┐
                   ▼               ▼
            ┌────────────────┐ ┌──────────────┐
            │    Worker      │ │    Queue     │
            │  (Go process)  │ │   Jobs       │
            │  Port -        │ │ (message.send)
            └────────────────┘ └──────────────┘
                   │
                   ▼
            ┌──────────────────┐
            │ Baileys Gateway  │
            │ (Node.js/Express)│
            │ Port 8090        │
            └──────────────────┘
                   │
                   ▼
         ┌────────────────────┐
         │  WhatsApp WA Web   │
         │  Session & Socket  │
         │  (per device)      │
         └────────────────────┘
```

## Key Points

1. **Idempotency**: Use `idempotencyKey` di message send untuk prevent duplicate sends
2. **Status Tracking**: Message bisa dalam status: processing → sent/failed
3. **Realtime Updates**: SSE stream untuk live dashboard updates
4. **Queue Processing**: Worker background processes message send jobs
5. **Error Handling**: Automatic retry dengan exponential backoff
6. **Multi-Device**: Support multiple WhatsApp devices bersamaan
7. **Session Persistence**: Baileys sessions disimpan di database dan disk

## Testing Checklist

- [ ] Login berhasil dengan admin/admin123
- [ ] Add device berhasil
- [ ] Click Connect → QR muncul
- [ ] Scan QR dengan phone WhatsApp
- [ ] Device status jadi "connected" dengan phone number
- [ ] Click "Send Message"
- [ ] Isi nomor tujuan dan pesan
- [ ] Klik Send
- [ ] Message muncul di logs dengan status "sent"
- [ ] Pesan terima di WhatsApp

## Troubleshooting

**Problem**: Gateway tidak start
- Check port 8090 tidak terpakai
- Check `npm install` di baileys-gateway

**Problem**: Device tidak connect
- Check gateway running
- Check phone WhatsApp installed
- Check QR tidak expired (5 menit)

**Problem**: Message tidak terkirim
- Check device status "connected"
- Check worker running
- Check toNumber format valid

**Problem**: Frontend tidak bisa login
- Check backend API running
- Check JWT token valid
- Check browser console untuk error

---

**Status**: ✅ Implementasi Selesai dan Siap Ditest
