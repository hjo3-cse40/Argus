// Argus Source Registration UI - Frontend Application

// DOM Elements
const form = document.getElementById('registration-form');
const submitButton = document.getElementById('submit-button');
const buttonText = submitButton.querySelector('.button-text');
const buttonSpinner = submitButton.querySelector('.loading-spinner');
const successMessage = document.getElementById('success-message');
const successText = document.getElementById('success-text');
const errorMessage = document.getElementById('error-message');
const errorText = document.getElementById('error-text');
const sourceList = document.getElementById('source-list');
const emptyState = document.getElementById('empty-state');
const listLoading = document.getElementById('list-loading');

// Form Fields
const nameInput = document.getElementById('source-name');
const typeSelect = document.getElementById('source-type');
const webhookInput = document.getElementById('discord-webhook');
const secretInput = document.getElementById('webhook-secret');
const repoUrlInput = document.getElementById('repository-url');

// Error Spans
const nameError = document.getElementById('name-error');
const typeError = document.getElementById('type-error');
const webhookError = document.getElementById('webhook-error');

// Validation State
let isFormValid = false;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    setupValidation();
    loadSources();
});

// Setup Form Validation
function setupValidation() {
    // Real-time validation on blur
    nameInput.addEventListener('blur', () => validateField('name'));
    typeSelect.addEventListener('blur', () => validateField('type'));
    webhookInput.addEventListener('blur', () => validateField('webhook'));
    
    // Re-validate on input to clear errors
    nameInput.addEventListener('input', () => {
        if (nameError.textContent) validateField('name');
        updateSubmitButton();
    });
    typeSelect.addEventListener('change', () => {
        if (typeError.textContent) validateField('type');
        updateSubmitButton();
    });
    webhookInput.addEventListener('input', () => {
        if (webhookError.textContent) validateField('webhook');
        updateSubmitButton();
    });
    
    // Form submission
    form.addEventListener('submit', handleSubmit);
    
    // Initial button state
    updateSubmitButton();
}

// Validate Individual Field
function validateField(fieldName) {
    let isValid = true;
    let errorMsg = '';
    
    switch (fieldName) {
        case 'name':
            const name = nameInput.value.trim();
            if (!name) {
                isValid = false;
                errorMsg = 'Source name is required';
            } else if (name.length > 100) {
                isValid = false;
                errorMsg = 'Source name must be 100 characters or less';
            }
            setFieldError(nameInput, nameError, errorMsg);
            break;
            
        case 'type':
            const type = typeSelect.value;
            if (!type) {
                isValid = false;
                errorMsg = 'Source type is required';
            }
            setFieldError(typeSelect, typeError, errorMsg);
            break;
            
        case 'webhook':
            const webhook = webhookInput.value.trim();
            if (!webhook) {
                isValid = false;
                errorMsg = 'Discord webhook URL is required';
            } else if (!webhook.startsWith('https://discord.com/api/webhooks/')) {
                isValid = false;
                errorMsg = 'Must be a valid Discord webhook URL (https://discord.com/api/webhooks/...)';
            }
            setFieldError(webhookInput, webhookError, errorMsg);
            break;
    }
    
    return isValid;
}

// Set Field Error State
function setFieldError(input, errorSpan, message) {
    if (message) {
        input.classList.add('error');
        input.setAttribute('aria-invalid', 'true');
        errorSpan.textContent = message;
    } else {
        input.classList.remove('error');
        input.removeAttribute('aria-invalid');
        errorSpan.textContent = '';
    }
}

// Validate All Fields
function validateForm() {
    const nameValid = validateField('name');
    const typeValid = validateField('type');
    const webhookValid = validateField('webhook');
    
    return nameValid && typeValid && webhookValid;
}

// Update Submit Button State
function updateSubmitButton() {
    const name = nameInput.value.trim();
    const type = typeSelect.value;
    const webhook = webhookInput.value.trim();
    
    // Enable button only if all required fields have values
    const hasAllFields = name && type && webhook;
    submitButton.disabled = !hasAllFields;
}

// Handle Form Submission
async function handleSubmit(event) {
    event.preventDefault();
    
    // Hide previous messages
    hideMessages();
    
    // Validate all fields
    if (!validateForm()) {
        showError('Please fix the validation errors before submitting');
        return;
    }
    
    // Disable form during submission
    setFormLoading(true);
    
    // Prepare data
    const formData = {
        name: nameInput.value.trim(),
        type: typeSelect.value,
        discord_webhook: webhookInput.value.trim(),
    };
    
    // Include webhook secret if provided
    const secret = secretInput.value.trim();
    if (secret) {
        formData.webhook_secret = secret;
    }
    
    // Include repository URL if provided
    const repoUrl = repoUrlInput.value.trim();
    if (repoUrl) {
        formData.repository_url = repoUrl;
    }
    
    try {
        const source = await createSource(formData);
        showSuccess(`Source "${source.name}" registered successfully!`);
        clearForm();
        loadSources(); // Refresh the list
    } catch (error) {
        showError(error.message);
    } finally {
        setFormLoading(false);
    }
}

// Set Form Loading State
function setFormLoading(loading) {
    submitButton.disabled = loading;
    
    if (loading) {
        buttonText.textContent = 'Registering...';
        buttonSpinner.hidden = false;
    } else {
        buttonText.textContent = 'Register Source';
        buttonSpinner.hidden = true;
    }
}

// Clear Form
function clearForm() {
    form.reset();
    nameInput.classList.remove('error');
    typeSelect.classList.remove('error');
    webhookInput.classList.remove('error');
    nameError.textContent = '';
    typeError.textContent = '';
    webhookError.textContent = '';
    updateSubmitButton();
}

// Show Success Message
function showSuccess(message) {
    successText.textContent = message;
    successMessage.hidden = false;
    errorMessage.hidden = true;
    
    // Auto-hide after 5 seconds
    setTimeout(() => {
        successMessage.hidden = true;
    }, 5000);
}

// Show Error Message
function showError(message) {
    errorText.textContent = message;
    errorMessage.hidden = false;
    successMessage.hidden = true;
}

// Hide All Messages
function hideMessages() {
    successMessage.hidden = true;
    errorMessage.hidden = true;
}

// API Functions (to be implemented in next tasks)
async function createSource(data) {
    const response = await fetchWithTimeout('/api/sources', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.details 
            ? errorData.details.join(', ')
            : errorData.error || `Server error: ${response.status}`;
        throw new Error(errorMsg);
    }
    
    return await response.json();
}

async function listSources(name = null) {
    const url = name ? `/api/sources?name=${encodeURIComponent(name)}` : '/api/sources';
    
    const response = await fetchWithTimeout(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    
    if (!response.ok) {
        throw new Error(`Failed to fetch sources: ${response.status}`);
    }
    
    return await response.json();
}

// Fetch with Timeout
async function fetchWithTimeout(url, options = {}, timeout = 10000) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    try {
        const response = await fetch(url, {
            ...options,
            signal: controller.signal,
        });
        clearTimeout(timeoutId);
        return response;
    } catch (error) {
        clearTimeout(timeoutId);
        
        if (error.name === 'AbortError') {
            throw new Error('Request timed out. Please try again.');
        }
        
        if (error instanceof TypeError) {
            throw new Error('Unable to connect to server. Please check your connection.');
        }
        
        throw error;
    }
}

// Load and Display Sources
async function loadSources() {
    // Show loading indicator
    listLoading.hidden = false;
    emptyState.hidden = true;
    sourceList.innerHTML = '';
    
    try {
        const sources = await listSources();
        renderSourceList(sources);
    } catch (error) {
        console.error('Failed to load sources:', error);
        // Show empty state on error
        listLoading.hidden = true;
        emptyState.hidden = false;
    }
}

// Render Source List
function renderSourceList(sources) {
    // Hide loading indicator
    listLoading.hidden = true;
    
    // Handle empty state
    if (!sources || sources.length === 0) {
        emptyState.hidden = false;
        sourceList.innerHTML = '';
        return;
    }
    
    // Hide empty state and render sources
    emptyState.hidden = true;
    sourceList.innerHTML = sources.map(source => createSourceCard(source)).join('');
}

// Create Source Card HTML
function createSourceCard(source) {
    const createdDate = new Date(source.created_at).toLocaleString();
    
    // Build repository URL section if present
    const repoUrlSection = source.repository_url 
        ? `<div class="source-detail">
                <strong>Repository:</strong>
                <a href="${escapeHtml(source.repository_url)}" target="_blank" rel="noopener noreferrer" class="source-repo-link">
                    ${escapeHtml(source.repository_url)}
                </a>
            </div>`
        : '';
    
    return `
        <div class="source-card">
            <div class="source-card-header">
                <h3 class="source-name">${escapeHtml(source.name)}</h3>
                <span class="source-type">${escapeHtml(source.type)}</span>
            </div>
            <div class="source-details">
                ${repoUrlSection}
                <div class="source-detail">
                    <strong>Discord Webhook:</strong>
                    <div class="source-webhook">${escapeHtml(source.discord_webhook)}</div>
                </div>
                <div class="source-detail">
                    <strong>Created:</strong> ${createdDate}
                </div>
                <div class="source-detail">
                    <strong>ID:</strong> <code>${escapeHtml(source.id)}</code>
                </div>
            </div>
        </div>
    `;
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

