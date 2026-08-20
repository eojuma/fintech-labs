# 🏦 African Vault — Digital SACCO Management System

A Go, GORM, and SQLite system evolving from a personal banking application into a digital SACCO management platform for Kenyan SACCOs, chamas, microfinance groups, and other member-owned savings organizations.

The existing account, transaction, role, M-Pesa, statement, and audit capabilities remain the foundation. SACCO-specific features are being added incrementally so that member ownership, loans, eligibility rules, and future distributions have explicit ledgers and audit trails.

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

### Phase 4 — Banking Foundation
- Multi-account support — current and savings accounts per user
- Transfer by phone number or account number
- Account statements — PDF and CSV download with date range selection
- User profile management — update contact details, change password, change PIN
- Balance visibility toggle
- Transaction receipts with unique reference numbers
- Transaction search, filtering and pagination
- Email notifications on every transaction
- SMS notifications via Africa's Talking
- Admin audit log tracking every admin action
- Transaction reports and analytics with 7-day chart
- Automated suspicious transaction flagging

### Phase 5 — Digital SACCO Management (Current)
- Member share capital tracked separately from spendable savings
- Administrator-recorded share contributions with an audit trail
- Loan application, approval, disbursement, repayment schedules, and balances
- Configurable loan eligibility rules tied to savings and/or share capital
- Savings interest and share dividend calculation and posting

SACCO features are delivered as separate tickets and commits. Features are not described as complete while migrations, posting operations, or required policy configuration remain outstanding.

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
| 14 | SMS notifications via Africa's Talking |
| 18 | Admin audit log |
| 19 | Transaction reports and analytics |
| 20 | Automated suspicious transaction flagging |
| 26 | Balance visibility toggle |
| 27 | Login with username or email |
| 31 | Account number on every transaction record |
| 32 | Responsive mobile layouts |
| 33 | Teller role assignment and teller operations |
| 34 | M-Pesa STK Push deposits |
| 35 | M-Pesa payment callback processing |
| 36 | SMS notification opt-out preference |
| 37 | Daily deposit, withdrawal, and transfer limits |
| 38 | Member account closure with balance and audit safeguards |
| 39 | Super-admin role for privileged administration |
| 40 | M-Pesa B2C withdrawal request client |
| 41 | Scheduled recurring transfer instructions and processing |
| 42 | Device registration and administrator approval gate |
| 43 | JWT token service for mobile API authentication |
| 44 | Production SMTP configuration validation |
| 45 | WebAuthn credential boundary and explicit provider gate |
| 46 | Minor-unit KES amount parsing utility |
| 47 | Member share-capital balance and contribution ledger |
| 48 | Loan application, approval, disbursement, schedules, and repayment tracking |
| 49 | Configurable savings and share-capital loan eligibility policy |
| 50 | Configurable annual savings interest and share dividend distributions |

---

## 🚧 Remaining Product Work

The remaining items require dedicated production integrations or migration planning:

- Complete WebAuthn ceremony integration with a supported authenticator service.
- Migrate persisted monetary columns from whole KES to cents after a controlled data migration.

The following features are implemented but require production credentials and operational setup:

- M-Pesa B2C payout callbacks and settlement reconciliation.
- Verified custom-domain SMTP sender configuration.
- JWT mobile API endpoints built on the token service.

---

## 🧭 SACCO Delivery Roadmap

### Ticket 1 — Member share capital

- Add a dedicated share-capital balance and contribution ledger rather than treating ownership capital as a spendable account.
- Allow administrators to record contributions atomically.
- Display share capital separately on the member dashboard.
- Integrate non-zero share capital explicitly into member exit checks.
- Schema migration: new GORM tables are required.

### Ticket 2 — Loan lifecycle

- Add loan applications with pending, approved, rejected, disbursed, and completed states.
- Record administrator decisions and disbursement.
- Generate scheduled repayments and track principal paid, interest paid, and outstanding balance.
- Schema migration: loan and repayment tables are required.

Automated collection from member accounts is not assumed; repayment posting is an explicit operation until a scheduler or payment mandate is implemented.

### Ticket 3 — Loan eligibility

- Store eligibility policy as SACCO configuration rather than hard-coding a common industry rule.
- Support a configurable multiple of savings, share capital, or both.
- Explain eligibility results before an application is accepted.
- Schema migration: SACCO configuration fields or a configuration table are required.

No default such as “three times savings” is treated as the SACCO's policy without administrator configuration.

### Ticket 4 — Interest and dividends

- Configure savings interest and share dividend rates and calculation periods.
- Calculate proposed allocations from ledger balances.
- Require an explicit administrator posting action before balances change.
- Record every posted allocation for audit and idempotency.
- Schema migration: distribution runs and member allocation records are required.

Rates, balance basis, pro-rating rules, and posting destinations vary by SACCO and must be configured. Earlier tickets do not assume interest or dividends already exist.

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
│   ├── handlers/       # HTTP handlers and auth/role middleware
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
│   │   ├── email.go
│   │   └── sms.go
│   ├── router/
│   │   └── router.go
│   ├── auth/
│   │   ├── jwt.go
│   │   └── webauthn.go
│   ├── mpesa/
│   │   └── mpesa.go
│   ├── services/       # business rules and transactional operations
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
- **Email:** SMTP via net/smtp with production sender validation
- **SMS:** Africa's Talking SMS API
- **Charts:** Chart.js
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
# Edit .env with your credentials

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
| `SMTP_PASS` | SMTP App Password |
| `SMTP_FROM` | Sender email address |
| `AT_USERNAME` | Africa's Talking username (use `sandbox` for testing) |
| `AT_API_KEY` | Africa's Talking API key |
| `MPESA_B2C_INITIATOR_NAME` | Safaricom B2C initiator name |
| `MPESA_B2C_SECURITY_CREDENTIAL` | Encrypted B2C security credential |
| `MPESA_B2C_TIMEOUT_URL` | Public B2C timeout callback URL |
| `MPESA_B2C_RESULT_URL` | Public B2C result callback URL |
| `JWT_SECRET` | At least 32 random characters for mobile API tokens |

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
| `/profile/close` | POST | Session | Close account after zero-balance and password checks |
| `/session/refresh` | POST | Session | Keepalive |
| `/admin` | GET | Admin | Admin dashboard |
| `/admin/deposit` | POST | Admin | Deposit to user account |
| `/admin/withdraw` | POST | Admin | Withdraw from user account |
| `/admin/share-contribution` | POST | Admin | Record a member share-capital contribution |
| `/loans/apply` | POST | Session | Submit a member loan application |
| `/admin/loans/decision` | POST | Admin | Approve or reject a pending loan |
| `/admin/loans/disburse` | POST | Admin | Disburse an approved loan to the current account |
| `/admin/loans/repayment` | POST | Admin | Record an explicit loan repayment |
| `/admin/loans/eligibility-policy` | POST | Admin | Configure loan eligibility rules |
| `/admin/distributions/policy` | POST | Admin | Configure interest and dividend rates |
| `/admin/distributions/preview` | POST | Admin | Preview a period distribution |
| `/admin/distributions/post` | POST | Admin | Explicitly post a previewed distribution |
| `/admin/toggle` | POST | Admin | Block or unblock account |
| `/admin/audit-log` | GET | Admin | View full audit log |
| `/admin/flagged` | GET | Admin | View flagged transactions |
| `/admin/assign-teller` | POST | Super admin | Assign teller role |
| `/admin/revoke-teller` | POST | Super admin | Revoke teller role |
| `/mpesa/deposit` | POST | Session | Initiate STK Push deposit |
| `/mpesa/callback` | POST | Safaricom | Process STK Push callback |

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
- Timing attack prevention on authentication
- Transaction PIN separate from login password
- Suspended accounts blocked at login with all sessions invalidated
- Admin cannot block their own account
- HTTPS enforced in production via redirect middleware
- Users can only view their own receipts
- Automated suspicious transaction flagging
- Daily transaction limits for deposits, withdrawals, and transfers
- SMS opt-out preference
- Device approval gate for recognized browsers
- Super-admin authorization for role management
- Idempotent and amount-validated M-Pesa callbacks
- Signed JWT token service with expiry and role claims

---

## 🚨 Fraud Detection Rules

Transactions are automatically flagged when:

| Rule | Threshold |
|------|-----------|
| Large single transaction | Amount ≥ KES 100,000 |
| Rapid successive transactions | 3 or more transactions within 5 minutes |
| Large withdrawal relative to balance | Withdrawal ≥ 80% of account balance |

Flagged transactions appear on the admin dashboard for review.

---

## 🗄️ Data Safety

- `.db` and `.env` files excluded from git via `.gitignore`
- Soft deletes via GORM DeletedAt — financial records never deleted
- Atomic DB transactions on every financial operation
- All session records cleaned up on logout and account suspension
- Every transaction has a unique reference number for tracing
- Admin audit log is permanent and never deletable

---

## 👥 Authors

**Evans Juma** — [@eojuma](https://github.com/eojuma)

Special thanks to Silas Lelei for peer-reviewing the GORM logic and testing the endpoints during the transition from maps to persistent storage.
