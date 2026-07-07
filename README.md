# 🏦 African Vault — Mobile Banking System

A production-grade mobile banking  built from scratch in Go. African Vault is a learning project that teaches the internals of modern banking systems — ledger design, atomic transactions, session management, payment integrations, and security engineering.

---

## 📈 The Journey

What started as a collection of Go experiments has evolved into a structured, secure, and scalable banking system built with real-world standards in mind.

### Phase 1 — Foundation
- Basic account creation and balance tracking
- In-memory storage with Go maps and slices
- Simple deposit and withdrawal logic

### Phase 2 — Persistence & Integrity
- Migrated from maps to SQLite via GORM
- Atomic DB transactions — money never gets lost mid-operation
- Soft delete system preserving financial audit trails
- Clean architecture — models, services, handlers, router

### Phase 3 — Security Engineering (Current)
- Secure session management with random token generation and server-side expiry
- 10-minute inactivity timeout with browser warning popup
- Login rate limiting — 5 attempts before 15-minute lockout
- Timing attack prevention on authentication
- Secure cookie flags — HttpOnly, Secure, SameSite=Strict
- HTTPS enforcement in production
- 4-digit transaction PIN separate from login password
- Suspended accounts blocked at login with session invalidation

### Phase 4 — Banking Features (In Progress)
- Multi-account support — current and savings accounts per user
- Transfer by phone number or account number
- Account statements — PDF and CSV download with date range
- User profile management — update contact details, change password, change PIN
- Balance visibility toggle
- Admin panel — deposit, withdraw, block/unblock accounts

---

## Completed Issues

| # | Feature |
|---|---------|
| 1 | Session expiry with 10-minute inactivity timeout |
| 2 | Login rate limiting with lockout |
| 3 | Secure cookie flags and HTTPS enforcement |
| 4 | Transaction PIN on all financial operations |
| 5 | User profile page |
| 6 | Change password and change PIN |
| 7 | Multiple accounts per user (current + savings) |
| 8 | Transfer by phone number or account number |
| 12 | Account statement download (PDF + CSV) |
| 26 | Balance visibility toggle |

---

## Upcoming Features

- Transaction search and filtering
- Transaction receipts
- Email and SMS notifications
- M-Pesa STK Push deposit and B2C withdrawal
- Admin audit log and transaction analytics
- Suspicious transaction flagging
- Scheduled recurring transfers
- Device verification and admin approval
- Role-based access control (teller, admin, super admin)
- Biometric authentication (WebAuthn)
- Currency precision migration to minor units

---

## Tech Stack

- Language: Go (Golang)
- Database: SQLite with GORM ORM
- Frontend: HTML, CSS, Vanilla JavaScript
- Auth: Custom session management with bcrypt
- PDF Generation: gofpdf
- Deployment: Render (https://fintech-labs-uaph.onrender.com)

---

## Getting Started

Clone the repository:

    git clone https://github.com/eojuma/fintech-labs.git
    cd fintech-labs

Sync dependencies:

    go mod tidy

Run the app:

    go run .

GORM will automatically run AutoMigrate and create transaction.db locally. Visit http://localhost:8080 to access the app.

---

## API Routes

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| /login | GET, POST | Public | User login |
| /register-page | GET | Public | Registration page |
| /register | POST | Public | Create account |
| /logout | POST | Session | Log out |
| /dashboard | GET | Session | User dashboard |
| /deposit | POST | Session + PIN | Deposit funds |
| /withdraw | POST | Session + PIN | Withdraw funds |
| /transfer | POST | Session + PIN | Send money |
| /accounts/open | POST | Session | Open savings account |
| /statement/download | GET | Session | Download statement |
| /profile | GET | Session | View profile |
| /profile/update | POST | Session | Update contact details |
| /profile/change-pin | POST | Session | Change transaction PIN |
| /profile/change-password | POST | Session | Change password |
| /session/refresh | POST | Session | Keepalive |
| /admin | GET | Admin | Admin dashboard |
| /admin/deposit | POST | Admin | Deposit to user account |
| /admin/withdraw | POST | Admin | Withdraw from user account |
| /admin/toggle | POST | Admin | Block or unblock account |

---

## Security Features

- Passwords hashed with bcrypt
- Session tokens are cryptographically random 32-byte hex strings
- Sessions stored server-side and validated on every request
- Cookie flags: HttpOnly, Secure (production), SameSite=Strict
- Session expires after 10 minutes of inactivity
- Warning popup at 9 minutes with keepalive option
- Login locked after 5 failed attempts for 15 minutes
- Timing attack prevention on authentication
- Transaction PIN separate from login password
- Suspended accounts blocked at login, all sessions invalidated
- Admin cannot block their own account
- HTTPS enforced in production

---

## Data Safety

- .db files excluded from git via .gitignore
- Soft deletes via GORM DeletedAt — financial records never deleted
- Atomic DB transactions on every financial operation
- All session records cleaned up on logout and account suspension

---

## Authors

Evans Juma — @eojuma (https://github.com/eojuma)

Special thanks to Silas Lelei for peer-reviewing the GORM logic and testing the endpoints during the transition from maps to persistent storage.

---

## Note

African Vault is a living project built to understand banking internals deeply. It started simple and grows with every issue closed. Built at Zone01 Kisumu — where we understand the why before writing the code.