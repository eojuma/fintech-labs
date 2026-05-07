/**
 * TOAST NOTIFICATIONS
 * Handles success/error messages at the top of the screen
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
 * Runs as soon as the page loads
 */
document.addEventListener('DOMContentLoaded', () => {
    // 1. Check URL for Success/Error messages from Go Redirects
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    // Clean URL so refresh doesn't show message again
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }

    // 2. Initialize Admin Logic
    initAdminPanel();
});

/**
 * ADMIN PANEL LOGIC
 * Uses Event Delegation for maximum performance and learning
 */
function initAdminPanel() {
    const table = document.getElementById('admin-user-table');
    const modal = document.getElementById('adminModal');
    const closeBtn = document.getElementById('closeModalBtn');

    if (!table || !modal) return; // Exit if we aren't on the admin page

    // Listen for clicks on the table instead of every single button
    table.addEventListener('click', (e) => {
        const target = e.target;
        
        // Find if a Deposit or Withdraw trigger was clicked
        if (target.classList.contains('btn-deposit-trigger') || target.classList.contains('btn-withdraw-trigger')) {
            const accountNumber = target.getAttribute('data-account');
            const type = target.classList.contains('btn-deposit-trigger') ? 'deposit' : 'withdraw';
            
            showAdminModal(type, accountNumber);
        }
    });

    // Modal Close Logic
    closeBtn.onclick = () => modal.style.display = 'none';
    window.onclick = (e) => { if (e.target === modal) modal.style.display = 'none'; };
}

/**
 * MODAL CONFIGURATION
 * Dynamically switches the modal between "Deposit" and "Withdraw" mode
 */
function showAdminModal(type, accNum) {
    const modal = document.getElementById('adminModal');
    const form = document.getElementById('adminActionForm');
    const title = document.getElementById('modalTitle');
    const subtitle = document.getElementById('modalSubtitle');
    const submitBtn = document.getElementById('modalSubmitBtn');
    const hiddenAccInput = document.getElementById('modalAccountNumber');

    // Set the account number in the hidden input
    hiddenAccInput.value = accNum;
    subtitle.innerText = `Target Account: ${accNum}`;

    if (type === 'deposit') {
        title.innerText = "Admin Deposit";
        form.action = "/admin/deposit";
        submitBtn.className = "btn-deposit";
        submitBtn.innerText = "Confirm Deposit";
    } else {
        title.innerText = "Admin Withdrawal";
        form.action = "/admin/withdraw";
        submitBtn.className = "btn-withdraw";
        submitBtn.innerText = "Confirm Withdrawal";
    }

    modal.style.display = 'flex';
}

/**
 * UTILITY: Toggle Password Visibility
 */
function togglePassword(inputId, iconElement) {
    const input = document.getElementById(inputId);
    if (input.type === 'password') {
        input.type = 'text';
        iconElement.innerText = '👁️';
    } else {
        input.type = 'password';
        iconElement.innerText = '👁️‍🗨️';
    }
}