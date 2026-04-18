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

// Automatically handle messages from Go redirects
window.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    // Clean URL so alert doesn't repeat on refresh
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }
});

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

function openModal(type, accountNumber) {
    const modal = document.getElementById('actionModal');
    if (!modal) return;

    const title = document.getElementById('modalTitle');
    const accountNumField = document.getElementById('accountNumber');
    const form = document.getElementById('quickActionForm');
    
    // UPDATED: Standardized routes to match main.go
    if (type === 'deposit') {
        title.innerText = '💰 Deposit to Account ' + accountNumber;
        form.action = '/admin/deposit';
    } else {
        title.innerText = '💸 Withdraw from Account ' + accountNumber;
        form.action = '/admin/withdraw';
    }
    
    accountNumField.value = accountNumber;
    document.getElementById('amount').value = '';
    modal.classList.add('active');
}

function closeModal() {
    const modal = document.getElementById('actionModal');
    if (modal) modal.classList.remove('active');
}

function validateTransfer() {
    const account = document.getElementById('to_account').value;
    const amount = document.getElementById('amount').value;
    const password = document.getElementById('password').value;

    if (!account || account.length !== 6) {
        showToast('Please enter a valid 6-digit account number', 'error');
        return false;
    }
    if (!amount || amount < 10) {
        showToast('Minimum transfer amount is KES 10', 'error');
        return false;
    }
    if (!password) {
        showToast('Please enter your password to confirm', 'error');
        return false;
    }
    return true;
}

// Automatically show messages from the URL (Success/Error redirects)
window.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    // Clean the URL so messages don't pop up again on refresh
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }
});