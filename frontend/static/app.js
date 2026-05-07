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
        iconElement.innerText = '👁️‍🗨️'; // Eye with stroke
    }
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

    if (document.getElementById('admin-user-table')) initAdminPanel();
});

/**
 * ADMIN PANEL LOGIC
 */
function initAdminPanel() {
    const table = document.getElementById('admin-user-table');
    const modal = document.getElementById('adminModal');
    const closeBtn = document.getElementById('closeModalBtn');

    if (!table || !modal) return;

    table.addEventListener('click', (e) => {
        const target = e.target;
        const accNum = target.getAttribute('data-account');
        if (!accNum) return;

        if (target.classList.contains('btn-deposit-trigger')) {
            showAdminModal('deposit', accNum);
        } else if (target.classList.contains('btn-withdraw-trigger')) {
            showAdminModal('withdraw', accNum);
        }
    });

    closeBtn.onclick = () => modal.style.display = 'none';
}

function showAdminModal(type, accNum) {
    const modal = document.getElementById('adminModal');
    const form = document.getElementById('adminActionForm');
    const hiddenAcc = document.getElementById('modalAccountNumber');
    
    hiddenAcc.value = accNum;
    document.getElementById('modalSubtitle').innerText = `Account: ${accNum}`;
    
    if (type === 'deposit') {
        document.getElementById('modalTitle').innerText = "Admin Deposit";
        form.action = "/admin/deposit";
    } else {
        document.getElementById('modalTitle').innerText = "Admin Withdrawal";
        form.action = "/admin/withdraw";
    }
    modal.style.display = 'flex';
}