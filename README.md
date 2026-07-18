# 🏦 African Vault — Mobile Banking System

A production-grade mobile banking backend built from scratch in Go. African Vault is a learning project that teaches the internals of modern banking systems — ledger design, atomic transactions, session management, payment integrations, and security engineering.

> Built at Zone01 Kisumu — where we understand the *why* before writing the code.

---

## 📈 The Journey

### Phase 1 — Foundation
- Basic account creation and balance tracking
- In-memory storage with Go maps
- Simple deposit and withdrawal logic

### Phase 2 — Persistence & Integrity
- Migrated from maps to SQLite via GORM
- Atomic DB transactions — money never gets lost mid-operation
- Soft delete system preserving financial audit trails
- Clean architecture — models, services, handlers, router

### Phase 3 — Security Engineering
- Secure session management with random 32-byte token generation
- 10-minute inactivity timeout with browser warning popup
- Login rate limiting — 5 attempts before 15-minute lockout
- Timing attack prevention on authentication
- Secure cookie flags — HttpOnly, Secure, SameSite=Strict
- HTTPS enforcement in production
- 4-digit transaction PIN separate from login password
- Suspended accounts blocked at login with session invalidation
- Admin cannot block their own account

### Phase 4 — Banking Features (Current)
- Multi-account support — current and savings accounts per user
- Transfer by phone number or account number
- Account statements — PDF and CSV download with date range selection
- User profile management — update contact details, change password, change PIN
- Balance visibility toggle
- Transaction receipts with unique reference numbers
- Transaction search, filtering, and pagination
- Email notifications on every transaction
- Admin panel — deposit, withdraw, block/unblock accounts

---

## ✅ Completed Issues

| # | Feature |
|---|---------|
| 1 | Session expiry with 10-minute inactivity timeout |
| 2 | Login rate limiting with 15-minute lockout |
| 3 | Secure cookie flags and HTTPS enforcement |
| 4 | Transaction PIN on all financial operations |
| 5 | User profile page |
| 6 | Change password and change PIN |
| 7 | Multiple accounts per user (current + savings) |
| 8 | Transfer by phone number or account number |
| 9 | Transaction receipt page with print support |
| 10 | Transaction search, filtering and pagination |
| 12 | Account statement download (PDF + CSV) |
| 13 | Email notifications after every transaction |
| 26 | Balance visibility toggle |
| 31 | Account number on every transaction record |

---

## 🚧 Upcoming Features

- SMS notifications via Africa's Talking
- M-Pesa STK Push deposit and B2C withdrawal
- M-Pesa webhook callback handler
- Admin audit log and transaction analytics
- Suspicious transaction flagging
- Transaction limits management
- Scheduled recurring transfers
- Device verification and admin approval
- Role-based access control (teller, admin, super admin)
- Biometric authentication (WebAuthn)
- Currency precision migration to minor units
- Production email setup with custom domain
- Account closure flow
- Multiple account types

---

## 🏗️ Project Structure

```
fintech-labs/
├── cmd/
│   └── server/
│       └── main.go          
├── internal/
│   ├── db/
│   │   └── db.go
│   ├── handlers/
│   │   ├── accounts.go
│   │   ├── admin.go
│   │   ├── authentication.go
│   │   ├── profile.go
│   │   ├── receipts.go
│   │   ├── statements.go
│   │   ├── transactions.go
│   │   └── ui.go
│   ├── models/
│   │   └── models.go
│   ├── notifications/
│   │   └── email.go
│   ├── router/
│   │   └── router.go
│   ├── services/
│   │   └── services.go
│   └── utils/
│       └── utils.go
├── web/
│   ├── static/
│   │   ├── app.js
│   │   └── styles.css
│   └── templates/
│       ├── admin.html
│       ├── dashboard.html
│       ├── email.html
│       ├── login.html
│       ├── profile.html
│       ├── receipt.html
│       ├── register.html
│       └── register_admin.html
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **Database:** SQLite with GORM ORM
- **Frontend:** HTML, CSS, Vanilla JavaScript
- **Auth:** Custom session management with bcrypt
- **PDF Generation:** gofpdf
- **Email:** Gmail SMTP via net/smtp
- **Deployment:** Render (https://fintech-labs-uaph.onrender.com)

---

## 🚀 Getting Started

```bash
# Clone the repository
git clone https://github.com/eojuma/fintech-labs.git
cd fintech-labs

# Sync dependencies
go mod tidy

# Set up environment variables
cp .env.example .env
# Edit .env with your SMTP credentials

# Run the app
go run cmd/server/main.go
```

Visit `http://localhost:8080` to access the app.

---

## ⚙️ Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_PATH` | Path to SQLite database file (default: transaction.db) |
| `RENDER` | Set to `true` in production to enable secure cookies and HTTPS |
| `TZ` | Timezone (set to `Africa/Nairobi` on Render) |
| `SMTP_HOST` | SMTP server host (e.g. smtp.gmail.com) |
| `SMTP_PORT` | SMTP server port (e.g. 587) |
| `SMTP_USER` | SMTP username / email address |
| `SMTP_PASS` | SMTP password or App Password |
| `SMTP_FROM` | Sender email address |

---

## 📡 API Routes

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/login` | GET, POST | Public | User login |
| `/register-page` | GET | Public | Registration page |
| `/register` | POST | Public | Create account |
| `/logout` | POST | Session | Log out |
| `/dashboard` | GET | Session | User dashboard |
| `/deposit` | POST | Session + PIN | Deposit funds |
| `/withdraw` | POST | Session + PIN | Withdraw funds |
| `/transfer` | POST | Session + PIN | Send money |
| `/accounts/open` | POST | Session | Open savings account |
| `/statement/download` | GET | Session | Download statement |
| `/transactions/filter` | GET | Session | Filter transactions |
| `/receipt/{ref}` | GET | Session | View transaction receipt |
| `/profile` | GET | Session | View profile |
| `/profile/update` | POST | Session | Update contact details |
| `/profile/change-pin` | POST | Session | Change transaction PIN |
| `/profile/change-password` | POST | Session | Change password |
| `/session/refresh` | POST | Session | Keepalive |
| `/admin` | GET | Admin | Admin dashboard |
| `/admin/deposit` | POST | Admin | Deposit to user account |
| `/admin/withdraw` | POST | Admin | Withdraw from user account |
| `/admin/toggle` | POST | Admin | Block or unblock account |

---

## 🛡️ Security Features

- Passwords hashed with bcrypt
- Session tokens are cryptographically random 32-byte hex strings
- Sessions stored server-side and validated on every request
- Cookie reissued on every request to reset browser-side MaxAge
- Cookie flags: HttpOnly, Secure (production), SameSite=Strict
- Session expires after 10 minutes of inactivity
- Warning popup at 9 minutes with keepalive option
- Login locked after 5 failed attempts for 15 minutes
- Timing attack prevention — bcrypt runs even for non-existent users
- Transaction PIN separate from login password
- Suspended accounts blocked at login with all sessions invalidated
- Admin cannot block their own account
- HTTPS enforced in production via redirect middleware
- Users can only view their own receipts

---

## 🗄️ Data Safety

- `.db` and `.env` files excluded from git via `.gitignore`
- Soft deletes via GORM DeletedAt — financial records never deleted
- Atomic DB transactions on every financial operation
- All session records cleaned up on logout and account suspension
- Every transaction has a unique reference number for tracing

---

## 👥 Author

**Evans Juma** — [@eojuma](https://github.com/eojuma)

Special thanks to Silas Lelei for peer-reviewing the GORM logic and testing the endpoints during the transition from maps to persistent storage.
