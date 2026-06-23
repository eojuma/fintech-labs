/**
 * UTILITY: Toggle Password Visibility
 */
function togglePassword(inputId, iconElement) {
  const input = document.getElementById(inputId);
  if (!input) return;

  if (input.type === "password") {
    input.type = "text";
    iconElement.innerText = "👁️"; // Open eye
  } else {
    input.type = "password";
    iconElement.innerText = "👁️‍🗨️"; // Closed eye
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
      const cpwd = regForm.querySelector(
        'input[name="confirm_password"]',
      ).value;
      if (pwd !== cpwd) {
        e.preventDefault();
        showToast("Passwords do not match", "error");
        return false;
      }
      return true;
    });

    // Prefill form inputs from URL query params (preserve fields after redirect)
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
});


/**
 * SESSION EXPIRY WARNING
 * Shows a warning popup at 9 minutes, auto-logs out at 10 minutes
 */
function startSessionTimer() {
  const TIMEOUT = 10 * 60 * 1000;      // 10 minutes in milliseconds
  const WARNING = 9 * 60 * 1000;       // 9 minutes in milliseconds

  let warningTimer = setTimeout(showSessionWarning, WARNING);
  let logoutTimer = setTimeout(forceLogout, TIMEOUT);

  function showSessionWarning() {
    // Create the popup
    const overlay = document.createElement("div");
    overlay.id = "session-warning-overlay";
    overlay.innerHTML = `
      <div class="session-warning-box">
        <div class="session-warning-icon">⏳</div>
        <h3>Session Expiring Soon</h3>
        <p>Your session will expire in <strong>1 minute</strong> due to inactivity. Do you want to stay logged in?</p>
        <div class="session-warning-actions">
          <button id="btn-stay-connected">Stay Connected</button>
          <button id="btn-logout-now">Logout</button>
        </div>
      </div>
    `;
    document.body.appendChild(overlay);

    // Stay Connected — refresh the session and reset timers
    document.getElementById("btn-stay-connected").addEventListener("click", () => {
      fetch("/session/refresh", { method: "POST" })
        .then(() => {
          overlay.remove();
          clearTimeout(warningTimer);
          clearTimeout(logoutTimer);
          // Restart the timers fresh
          warningTimer = setTimeout(showSessionWarning, WARNING);
          logoutTimer = setTimeout(forceLogout, TIMEOUT);
        })
        .catch(() => {
          // If refresh fails, force logout
          forceLogout();
        });
    });

    // Logout immediately
    document.getElementById("btn-logout-now").addEventListener("click", () => {
      forceLogout();
    });
  }

  function forceLogout() {
    // Submit the logout form programmatically
    const form = document.createElement("form");
    form.method = "POST";
    form.action = "/logout";
    document.body.appendChild(form);
    form.submit();
  }
}

// Only start the session timer on protected pages (dashboard, admin)
// Not on login or register pages
if (!window.location.pathname.includes("/login") && 
    !window.location.pathname.includes("/register")) {
  startSessionTimer();
}