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
