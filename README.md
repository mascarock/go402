# GoPayments — Micropayments Simulation Platform

Full-stack x402 micropayments demo: Go backend + React frontend simulating multi-tenant payment processing with HTTP 402 Payment Required flows.

## Quick Start

**Backend** (terminal 1):
```bash
cd backend
go run .
# → http://localhost:8080
```

**Frontend** (terminal 2):
```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173
```

## Architecture

```
backend/          Go API server (chi router, in-memory store)
├── handler/      HTTP handlers per domain
├── model/        Data types
├── store/        Thread-safe in-memory storage
├── middleware/   x402 payment middleware
└── service/      Business logic

frontend/         React + TypeScript + Vite + Tailwind
├── pages/        Dashboard, Tenant, PaymentDemo, Settlements
├── components/   Reusable UI components
└── api/          API client
```

## x402 Flow

1. `GET /api/protected/data` → returns `402 Payment Required` with payment headers
2. `POST /api/payments/process` with tenant credentials → returns receipt token
3. `GET /api/protected/data` with `X-Payment-Token` header → returns premium data

## Seeded Data

Three demo tenants pre-loaded:
- **Nexus Gaming** (NexusPlay) — $50.00 balance
- **Pulse Finance** (PulsePay) — $50.00 balance
- **Orbit Markets** (OrbitBet) — $50.00 balance

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/stats | Platform-wide statistics |
| GET | /api/tenants | List all tenants |
| POST | /api/tenants | Create tenant |
| GET | /api/tenants/:id | Get tenant |
| GET | /api/tenants/:id/wallet | Wallet balance |
| POST | /api/tenants/:id/deposit | Deposit funds |
| POST | /api/tenants/:id/withdraw | Withdraw funds |
| GET | /api/tenants/:id/transactions | Transaction history |
| GET | /api/settlements | List settlements |
| POST | /api/settlements | Create settlement |
| POST | /api/payments/process | Process micropayment |
| GET | /api/protected/data | x402-protected endpoint |
