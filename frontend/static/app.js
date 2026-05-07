/**
 * UTILITY: Toggle Password Visibility
 */
function togglePassword(inputId, iconElement) {
    const input = document.getElementById(inputId);
    if (!input) return;

    if (input.type === 'password') {
        input.type = 'text';
        iconElement.innerText = '👁️'; // Open eye
    } else {
        input.type = 'password';
        iconElement.innerText = '👁️‍🗨️'; // Closed eye
    }
}

/**
 * TOAST NOTIFICATIONS
 */
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast-notification ${type === 'success' ? 'toast-success' : 'toast-error'}`;
    toast.innerHTML = `<span>${type === 'success' ? '✅' : '❌'}</span> ${message}`;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 500);
    }, 4000);
}

/**
 * INITIALIZATION
 */
document.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }
});