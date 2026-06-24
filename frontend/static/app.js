/**
 * UTILITY: Toggle Password Visibility
 */
function togglePassword(inputId, iconElement) {
  const input = document.getElementById(inputId);
  if (!input) return;

  if (input.type === "password") {
    input.type = "text";
    iconElement.innerText = "👁️";
  } else {
    input.type = "password";
    iconElement.innerText = "👁️‍🗨️";
  }
}

/**
 * TOAST NOTIFICATIONS
 */
function showToast(message, type = "success") {
  const toast = document.createElement("div");
  toast.className = `toast-notification ${type === "success" ? "toast-success" : "toast-error"}`;
  toast.innerHTML = `<span>${type === "success" ? "✅" : "❌"}</span> ${message}`;
  document.body.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = "0";
    setTimeout(() => toast.remove(), 500);
  }, 4000);
}

/**
 * INITIALIZATION
 */
document.addEventListener("DOMContentLoaded", () => {
  const urlParams = new URLSearchParams(window.location.search);
  const successMsg = urlParams.get("success");
  const errorMsg = urlParams.get("error");

  if (successMsg) showToast(decodeURIComponent(successMsg.replace(/\+/g, " ")), "success");
  if (errorMsg) showToast(decodeURIComponent(errorMsg.replace(/\+/g, " ")), "error");
  if (successMsg || errorMsg) {
    window.history.replaceState({}, document.title, window.location.pathname);
  }

  // Client-side registration password confirmation
  const regForm = document.getElementById("registerForm");
  if (regForm) {
    regForm.addEventListener("submit", (e) => {
      const pwd = regForm.querySelector('input[name="password"]').value;
      const cpwd = regForm.querySelector('input[name="confirm_password"]').value;
      if (pwd !== cpwd) {
        e.preventDefault();
        showToast("Passwords do not match", "error");
        return false;
      }
      return true;
    });

    const params = new URLSearchParams(window.location.search);
    const fields = ["fullname", "username", "email", "phone", "id_number"];
    fields.forEach((f) => {
      const val = params.get(f);
      if (val) {
        const input = regForm.querySelector(`[name="${f}"]`);
        if (input) input.value = decodeURIComponent(val);
      }
    });
  }

  // Start session timer on protected pages
  if (!window.location.pathname.includes("/login") &&
    !window.location.pathname.includes("/register")) {
    startSessionTimer();
  }
});

/**
 * ADMIN MODAL TRIGGERS — event delegation
 */
document.addEventListener("click", (e) => {
  const adminModal = document.getElementById("adminModal")
  if (!adminModal) return

  const modalTitle = document.getElementById("modalTitle")
  const modalSubtitle = document.getElementById("modalSubtitle")
  const modalAccountNumber = document.getElementById("modalAccountNumber")
  const adminActionForm = document.getElementById("adminActionForm")
  const modalSubmitBtn = document.getElementById("modalSubmitBtn")

  if (e.target.classList.contains("btn-deposit-trigger")) {
    const account = e.target.getAttribute("data-account")
    modalTitle.textContent = "Deposit Funds"
    modalSubtitle.textContent = `Account: ${account}`
    modalAccountNumber.value = account
    adminActionForm.action = "/admin/deposit"
    modalSubmitBtn.textContent = "Deposit"
    modalSubmitBtn.style.background = "var(--success)"
    adminModal.style.display = "flex"
  }

  if (e.target.classList.contains("btn-withdraw-trigger")) {
    const account = e.target.getAttribute("data-account")
    modalTitle.textContent = "Withdraw Funds"
    modalSubtitle.textContent = `Account: ${account}`
    modalAccountNumber.value = account
    adminActionForm.action = "/admin/withdraw"
    modalSubmitBtn.textContent = "Withdraw"
    modalSubmitBtn.style.background = "var(--warning)"
    adminModal.style.display = "flex"
  }

  if (e.target.id === "closeModalBtn" || e.target === adminModal) {
    adminModal.style.display = "none"
  }
})

/**
 * SESSION EXPIRY WARNING
 * Shows a warning popup at 9 minutes, auto-logs out at 10 minutes
 */
function startSessionTimer() {
  const TIMEOUT = 10 * 60 * 1000;
  const WARNING = 9 * 60 * 1000;

  let warningTimer = setTimeout(showSessionWarning, WARNING);
  let logoutTimer = setTimeout(forceLogout, TIMEOUT);

  function showSessionWarning() {
    const overlay = document.createElement("div");
    overlay.id = "session-warning-overlay";

    Object.assign(overlay.style, {
      position: "fixed",
      top: "0",
      left: "0",
      width: "100vw",
      height: "100vh",
      background: "rgba(0, 0, 0, 0.5)",
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      zIndex: "999999",
    });

    overlay.innerHTML = `
      <div style="
        background: #ffffff;
        border-radius: 12px;
        padding: 2.5rem;
        max-width: 420px;
        width: 90%;
        text-align: center;
        box-shadow: 0 25px 60px rgba(0,0,0,0.4);
        border: 2px solid #f59e0b;
        position: relative;
        z-index: 1000000;
      ">
        <div style="font-size: 3rem; margin-bottom: 1rem;">⏳</div>
        <h3 style="font-size: 1.3rem; font-weight: 700; color: #1e293b; margin-bottom: 0.75rem;">Session Expiring Soon</h3>
        <p style="color: #64748b; font-size: 0.95rem; margin-bottom: 1.5rem; line-height: 1.6;">
          Your session will expire in <strong>1 minute</strong> due to inactivity. Do you want to stay logged in?
        </p>
        <div style="display: flex; gap: 12px; justify-content: center;">
          <button id="btn-stay-connected" style="
            background: #004a99; color: white; padding: 12px 24px;
            border-radius: 8px; border: none; font-weight: 600;
            cursor: pointer; font-size: 0.95rem;">Stay Connected</button>
          <button id="btn-logout-now" style="
            background: #ef4444; color: white; padding: 12px 24px;
            border-radius: 8px; border: none; font-weight: 600;
            cursor: pointer; font-size: 0.95rem;">Logout</button>
        </div>
      </div>
    `;

    document.body.appendChild(overlay);

    document.getElementById("btn-stay-connected").addEventListener("click", () => {
      fetch("/session/refresh", { method: "POST" })
        .then(() => {
          overlay.remove();
          clearTimeout(warningTimer);
          clearTimeout(logoutTimer);
          warningTimer = setTimeout(showSessionWarning, WARNING);
          logoutTimer = setTimeout(forceLogout, TIMEOUT);
        })
        .catch(() => forceLogout());
    });

    document.getElementById("btn-logout-now").addEventListener("click", () => {
      forceLogout();
    });
  }

  function forceLogout() {
    const form = document.createElement("form");
    form.method = "POST";
    form.action = "/logout";
    document.body.appendChild(form);
    form.submit();
  }
}