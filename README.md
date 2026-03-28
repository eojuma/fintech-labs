💳 Fintech Labs:A persistent ledger and transaction audit logs

This repository documents my professional journey from building basic Go experiments to developing a persistent, production-ready fintech API.

📈 The Journey So Far

What started as a collection of practice scripts has evolved into a structured system centered on Data Integrity and User Persistence.

Phase 1: The Foundation (Original Goals)

    Learn fintech concepts by building from scratch.

    Practice backend development using the Go standard library.

    Explore payment systems and transaction simulations.

Phase 2: Professional Grade (Current State)

    GORM & SQLite Integration: Transitioned from temporary memory (maps) to permanent disk storage to prevent data loss.

    Atomic Transactions: Implemented db.Transaction logic to ensure that money is never "lost in the air" during a crash.

    Soft Delete System: Created an "Inactive/Reactivate" flow to preserve financial audit trails—essential for fintech compliance.

    Clean Architecture: Separated the project into models, services, and handlers for better scalability and team collaboration.

🛠️ Technical Stack

    Language: Go (Golang)

    Database: SQLite (Local persistence)

    ORM: GORM (Object Relational Mapping)

    Environment: Developed on Linux(Ubuntu)

🚀 Getting Started

To run this project locally and see the evolution in action:

    Sync Dependencies:
    Bash

    go mod tidy

    Initialize & Run:
    Bash

    go run .

    GORM will automatically perform an AutoMigration and create your transaction.db file locally.

📡 API Reference
| Endpoint | Method | Purpose |
| :--- | :--- | :--- |
| `/balance` | `GET` | Fetch real-time account balance from SQLite. |
| `/deposit` | `POST` | Securely add funds via GORM transactions. |
| `/delete` | `DELETE` | Mark an account as **Inactive** (Soft Delete). |
| `/reactivate` | `POST` | Restore an account to **Active** status. |


🛡️ Data Safety & Git Hygiene

We utilize a .gitignore to ensure that local *.db files stay on the developer's machine. This prevents sensitive test data from being pushed to GitHub and avoids merge conflicts between team members.


👥 Authors & Contributors

    Evans Juma - [@eojuma](https://github.com/eojuma) - Lead Backend Developer

    Special thanks to Silas Lelei for peer-reviewing,the GORM logic and testing the endpoints during transition from maps to a persistent storage.

📝 Note

This is a living project. It started simple and will continue to improve as I explore more complex DevOps and Cloud-Native technologies for the Kenyan and remote markets.
