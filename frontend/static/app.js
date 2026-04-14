
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast-notification ${type === 'success' ? 'toast-success' : 'toast-error'}`;
    toast.innerHTML = `<span>${type === 'success' ? '✅' : '❌'}</span> ${message}`;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 500);
    }, 3000);
}

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
    
    if (type === 'deposit') {
        title.innerText = '💰 Add Funds to Account ' + accountNumber;
        form.action = '/admin/api/deposit';
    } else {
        title.innerText = '💸 Withdraw Funds from Account ' + accountNumber;
        form.action = '/admin/api/withdraw';
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

window.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const successMsg = urlParams.get('success');
    const errorMsg = urlParams.get('error');

    if (successMsg) showToast(decodeURIComponent(successMsg), 'success');
    if (errorMsg) showToast(decodeURIComponent(errorMsg), 'error');
    
    if (successMsg || errorMsg) {
        window.history.replaceState({}, document.title, window.location.pathname);
    }
});