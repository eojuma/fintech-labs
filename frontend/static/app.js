/**
 * TOAST NOTIFICATIONS - Global UI Feedback
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
 * INITIALIZATION - Runs on page load
 */
document.addEventListener('DOMContentLoaded', () => {
    // 1. Alert Handling from Go Redirects
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }

    // 2. Page-Specific Initializers
    if (document.getElementById('admin-user-table')) initAdminPanel();
});

/**
 * ADMIN LOGIC - Using Event Delegation
 */
function initAdminPanel() {
    const table = document.getElementById('admin-user-table');
    const modal = document.getElementById('adminModal');
    const closeBtn = document.getElementById('closeModalBtn');

    if (!table || !modal) return;

    table.addEventListener('click', (e) => {
        const target = e.target;
        // Connects to the data-account attribute in our HTML template
        const accNum = target.getAttribute('data-account');
        
        if (target.classList.contains('btn-deposit-trigger')) {
            showAdminModal('deposit', accNum);
        } else if (target.classList.contains('btn-withdraw-trigger')) {
            showAdminModal('withdraw', accNum);
        }
    });

    closeBtn.onclick = () => modal.style.display = 'none';
    window.onclick = (e) => { if (e.target === modal) modal.style.display = 'none'; };
}

function showAdminModal(type, accNum) {
    const modal = document.getElementById('adminModal');
    const form = document.getElementById('adminActionForm');
    const title = document.getElementById('modalTitle');
    const hiddenAcc = document.getElementById('modalAccountNumber');

    hiddenAcc.value = accNum;
    document.getElementById('modalSubtitle').innerText = `Target: ${accNum}`;

    if (type === 'deposit') {
        title.innerText = "Admin Deposit";
        form.action = "/admin/deposit"; // Matches main.go route
    } else {
        title.innerText = "Admin Withdrawal";
        form.action = "/admin/withdraw"; // Matches main.go route
    }
    modal.style.display = 'flex';
}