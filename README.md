# go402 — x402 Micropayments Simulation



https://github.com/user-attachments/assets/36610aab-e297-4444-8be0-e6c1f52bdc35



Full-stack simulation of the [x402 protocol](https://x402.org) — the HTTP-native payment standard for machine-to-machine payments. Go backend + React frontend demonstrating how AI agents can autonomously pay for access to resources using HTTP `402 Payment Required`.

> x402 is now a Linux Foundation standard (April 2026), backed by Coinbase, Stripe, Cloudflare, and others.

## Quick Start

**Prerequisites:** [Go](https://go.dev/dl/) and [Node.js](https://nodejs.org) installed.

```bash
git clone https://github.com/mascarock/go402.git
cd go402
npm install
cd frontend && npm install && cd ..
npm run dev
```

That's it. `npm run dev` starts both the Go API (`localhost:8080`) and the React frontend (`localhost:5173`) in one terminal.

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
