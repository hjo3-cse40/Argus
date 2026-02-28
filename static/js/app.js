// Argus Hierarchical Source Management UI - Frontend Application

// DOM Elements - Messages
const successMessage = document.getElementById('success-message');
const successText = document.getElementById('success-text');
const successClose = document.getElementById('success-close');
const errorMessage = document.getElementById('error-message');
const errorText = document.getElementById('error-text');
const errorClose = document.getElementById('error-close');

// DOM Elements - Platforms
const addPlatformBtn = document.getElementById('add-platform-btn');
const platformsList = document.getElementById('platforms-list');
const platformsLoading = document.getElementById('platforms-loading');
const platformsEmpty = document.getElementById('platforms-empty');

// DOM Elements - Platform Modal
const platformModal = document.getElementById('platform-modal');
const platformModalClose = document.getElementById('platform-modal-close');
const platformForm = document.getElementById('platform-form');
const platformNameSelect = document.getElementById('platform-name');
const platformWebhookInput = document.getElementById('platform-webhook');
const platformSecretInput = document.getElementById('platform-secret');
const platformSubmitBtn = document.getElementById('platform-submit-btn');
const platformCancelBtn = document.getElementById('platform-cancel-btn');
const platformNameError = document.getElementById('platform-name-error');
const platformWebhookError = document.getElementById('platform-webhook-error');

// DOM Elements - Subsources
const subsourcesList = document.getElementById('subsources-list');
const subsourcesLoading = document.getElementById('subsources-loading');
const subsourcesEmpty = document.getElementById('subsources-empty');

// DOM Elements - Subsource Modal
const subsourceModal = document.getElementById('subsource-modal');
const subsourceModalClose = document.getElementById('subsource-modal-close');
const subsourceForm = document.getElementById('subsource-form');
const subsourcePlatformId = document.getElementById('subsource-platform-id');
const subsourcePlatformName = document.getElementById('subsource-platform-name');
const subsourceNameInput = document.getElementById('subsource-name');
const subsourceIdentifierInput = document.getElementById('subsource-identifier');
const subsourceUrlInput = document.getElementById('subsource-url');
const subsourceSubmitBtn = document.getElementById('subsource-submit-btn');
const subsourceCancelBtn = document.getElementById('subsource-cancel-btn');
const subsourceNameError = document.getElementById('subsource-name-error');
const subsourceIdentifierError = document.getElementById('subsource-identifier-error');

// DOM Elements - Confirmation Modal
const confirmModal = document.getElementById('confirm-modal');
const confirmMessage = document.getElementById('confirm-message');
const confirmOkBtn = document.getElementById('confirm-ok-btn');
const confirmCancelBtn = document.getElementById('confirm-cancel-btn');

// State
let currentPlatforms = [];
let currentSubsources = [];
let confirmCallback = null;
let editingPlatformId = null;
let editingSubsourceId = null;

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    console.log('Page loaded, message states:', {
        successHidden: successMessage.hidden,
        errorHidden: errorMessage.hidden
    }); // Debug log
    setupEventListeners();
    loadData();
});

// Setup Event Listeners
function setupEventListeners() {
    // Message close buttons
    successClose.addEventListener('click', () => {
        console.log('Success close button clicked'); // Debug log
        successMessage.hidden = true;
    });
    errorClose.addEventListener('click', () => {
        console.log('Error close button clicked'); // Debug log
        errorMessage.hidden = true;
    });
    
    // Platform modal
    addPlatformBtn.addEventListener('click', openPlatformModal);
    platformModalClose.addEventListener('click', closePlatformModal);
    platformCancelBtn.addEventListener('click', closePlatformModal);
    platformForm.addEventListener('submit', handlePlatformSubmit);
    
    // Platform form validation
    platformNameSelect.addEventListener('change', () => {
        if (platformNameError.textContent) validatePlatformField('name');
    });
    platformWebhookInput.addEventListener('input', () => {
        if (platformWebhookError.textContent) validatePlatformField('webhook');
    });
    
    // Subsource modal
    subsourceModalClose.addEventListener('click', closeSubsourceModal);
    subsourceCancelBtn.addEventListener('click', closeSubsourceModal);
    subsourceForm.addEventListener('submit', handleSubsourceSubmit);
    
    // Subsource form validation
    subsourceNameInput.addEventListener('input', () => {
        if (subsourceNameError.textContent) validateSubsourceField('name');
    });
    subsourceIdentifierInput.addEventListener('input', () => {
        if (subsourceIdentifierError.textContent) validateSubsourceField('identifier');
    });
    
    // Confirmation modal
    confirmCancelBtn.addEventListener('click', () => {
        console.log('Confirm cancel clicked'); // Debug log
        closeConfirmModal();
    });
    confirmOkBtn.addEventListener('click', () => {
        console.log('Confirm OK clicked'); // Debug log
        handleConfirmOk();
    });
    
    // Close modals on background click
    platformModal.addEventListener('click', (e) => {
        if (e.target === platformModal) closePlatformModal();
    });
    subsourceModal.addEventListener('click', (e) => {
        if (e.target === subsourceModal) closeSubsourceModal();
    });
    confirmModal.addEventListener('click', (e) => {
        if (e.target === confirmModal) closeConfirmModal();
    });
}

// Load All Data
async function loadData() {
    await loadPlatforms();
    await loadSubsources();
}

// ===== PLATFORM FUNCTIONS =====

// Load Platforms
async function loadPlatforms() {
    platformsLoading.hidden = false;
    platformsEmpty.hidden = true;
    platformsList.innerHTML = '';
    
    try {
        const platforms = await fetchPlatforms();
        currentPlatforms = platforms;
        renderPlatforms(platforms);
    } catch (error) {
        console.error('Failed to load platforms:', error);
        platformsLoading.hidden = true;
        platformsEmpty.hidden = false;
    }
}

// Fetch Platforms from API
async function fetchPlatforms() {
    const response = await fetchWithTimeout('/api/platforms', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
    });
    
    if (!response.ok) {
        throw new Error(`Failed to fetch platforms: ${response.status}`);
    }
    
    return await response.json();
}

// Render Platforms List
function renderPlatforms(platforms) {
    platformsLoading.hidden = true;
    
    if (!platforms || platforms.length === 0) {
        platformsEmpty.hidden = false;
        platformsList.innerHTML = '';
        return;
    }
    
    platformsEmpty.hidden = true;
    platformsList.innerHTML = platforms.map(platform => createPlatformCard(platform)).join('');
    
    console.log('Attaching event listeners to', platforms.length, 'platforms'); // Debug log
    
    // Attach event listeners to buttons
    platforms.forEach(platform => {
        // Edit button
        const editBtn = document.getElementById(`edit-platform-${platform.id}`);
        if (editBtn) {
            editBtn.addEventListener('click', () => openEditPlatformModal(platform));
        }
        
        // Delete button
        const deleteBtn = document.getElementById(`delete-platform-${platform.id}`);
        console.log('Delete button for', platform.name, ':', deleteBtn); // Debug log
        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => {
                console.log('Delete button clicked for platform:', platform.name); // Debug log
                confirmDeletePlatform(platform);
            });
        }
        
        // Add Subsource button
        const addSubsourceBtn = document.getElementById(`add-subsource-platform-${platform.id}`);
        if (addSubsourceBtn) {
            addSubsourceBtn.addEventListener('click', () => openSubsourceModal(platform));
        }
    });
}

// Create Platform Card HTML
function createPlatformCard(platform) {
    const createdDate = new Date(platform.created_at).toLocaleString();
    const platformLabel = platform.name.charAt(0).toUpperCase() + platform.name.slice(1);
    
    return `
        <div class="platform-card">
            <div class="platform-card-header">
                <div class="platform-info">
                    <span class="platform-badge platform-${platform.name}">${platformLabel}</span>
                    <span class="platform-date">${createdDate}</span>
                </div>
                <div class="platform-actions">
                    <button 
                        id="add-subsource-platform-${platform.id}" 
                        class="btn btn-primary btn-small"
                        aria-label="Add subsource to ${platformLabel}"
                    >
                        Add Subsource
                    </button>
                    <button 
                        id="edit-platform-${platform.id}" 
                        class="btn btn-secondary btn-small"
                        aria-label="Edit ${platformLabel} platform"
                    >
                        Edit
                    </button>
                    <button 
                        id="delete-platform-${platform.id}" 
                        class="btn btn-danger btn-small"
                        aria-label="Delete ${platformLabel} platform"
                    >
                        Delete
                    </button>
                </div>
            </div>
            <div class="platform-details">
                <div class="platform-detail">
                    <strong>Webhook:</strong>
                    <div class="platform-webhook">${escapeHtml(platform.discord_webhook)}</div>
                </div>
                <div class="platform-detail">
                    <strong>ID:</strong> <code>${escapeHtml(platform.id)}</code>
                </div>
            </div>
        </div>
    `;
}

// Open Platform Modal
function openPlatformModal() {
    editingPlatformId = null;
    platformForm.reset();
    clearPlatformErrors();
    document.getElementById('platform-modal-title').textContent = 'Add Platform';
    platformSubmitBtn.querySelector('.button-text').textContent = 'Create Platform';
    platformNameSelect.disabled = false;
    platformModal.hidden = false;
    platformNameSelect.focus();
}

// Open Edit Platform Modal
function openEditPlatformModal(platform) {
    editingPlatformId = platform.id;
    platformForm.reset();
    clearPlatformErrors();
    
    document.getElementById('platform-modal-title').textContent = 'Edit Platform';
    platformSubmitBtn.querySelector('.button-text').textContent = 'Update Platform';
    
    // Pre-fill form with existing data
    platformNameSelect.value = platform.name;
    platformNameSelect.disabled = true; // Can't change platform name
    platformWebhookInput.value = platform.discord_webhook;
    
    platformModal.hidden = false;
    platformWebhookInput.focus();
}

// Close Platform Modal
function closePlatformModal() {
    platformModal.hidden = true;
    platformForm.reset();
    clearPlatformErrors();
}

// Clear Platform Form Errors
function clearPlatformErrors() {
    setFieldError(platformNameSelect, platformNameError, '');
    setFieldError(platformWebhookInput, platformWebhookError, '');
}

// Validate Platform Field
function validatePlatformField(fieldName) {
    let isValid = true;
    let errorMsg = '';
    
    switch (fieldName) {
        case 'name':
            const name = platformNameSelect.value;
            if (!name) {
                isValid = false;
                errorMsg = 'Platform is required';
            }
            setFieldError(platformNameSelect, platformNameError, errorMsg);
            break;
            
        case 'webhook':
            const webhook = platformWebhookInput.value.trim();
            console.log('Validating webhook:', webhook); // Debug log
            if (!webhook) {
                isValid = false;
                errorMsg = 'Discord webhook URL is required';
            } else {
                // Check if it's a valid Discord webhook URL
                const isDiscordWebhook = webhook.startsWith('https://discord.com/api/webhooks/') || 
                                       webhook.startsWith('https://discordapp.com/api/webhooks/');
                
                if (!isDiscordWebhook) {
                    isValid = false;
                    errorMsg = 'Must be a valid Discord webhook URL (https://discord.com/api/webhooks/... or https://discordapp.com/api/webhooks/...)';
                    console.log('Webhook validation failed for:', webhook); // Debug log
                } else {
                    console.log('Webhook validation passed for:', webhook); // Debug log
                }
            }
            setFieldError(platformWebhookInput, platformWebhookError, errorMsg);
            break;
    }
    
    return isValid;
}

// Validate Platform Form
function validatePlatformForm() {
    const nameValid = validatePlatformField('name');
    const webhookValid = validatePlatformField('webhook');
    return nameValid && webhookValid;
}

// Handle Platform Form Submit
async function handlePlatformSubmit(event) {
    event.preventDefault();
    
    hideMessages();
    
    if (!validatePlatformForm()) {
        showError('Please fix the validation errors before submitting');
        return;
    }
    
    const isEditing = editingPlatformId !== null;
    const buttonText = isEditing ? 'Updating...' : 'Creating...';
    setFormLoading(platformSubmitBtn, true, buttonText);
    
    const formData = {
        discord_webhook: platformWebhookInput.value.trim(),
    };
    
    if (!isEditing) {
        formData.name = platformNameSelect.value;
    }
    
    const secret = platformSecretInput.value.trim();
    if (secret) {
        formData.webhook_secret = secret;
    }
    
    try {
        if (isEditing) {
            await updatePlatform(editingPlatformId, formData);
            showSuccess('Platform updated successfully!');
        } else {
            await createPlatform(formData);
            showSuccess(`Platform "${formData.name}" created successfully!`);
        }
        closePlatformModal();
        await loadData();
    } catch (error) {
        showError(error.message);
    } finally {
        const finalButtonText = isEditing ? 'Update Platform' : 'Create Platform';
        setFormLoading(platformSubmitBtn, false, finalButtonText);
    }
}

// Create Platform via API
async function createPlatform(data) {
    const response = await fetchWithTimeout('/api/platforms', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.error || `Server error: ${response.status}`;
        throw new Error(errorMsg);
    }
    
    return await response.json();
}

// Update Platform via API
async function updatePlatform(platformId, data) {
    const response = await fetchWithTimeout(`/api/platforms/${platformId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.error || `Server error: ${response.status}`;
        throw new Error(errorMsg);
    }
    
    return await response.json();
}

// Confirm Delete Platform
function confirmDeletePlatform(platform) {
    console.log('confirmDeletePlatform called for:', platform); // Debug log
    const platformLabel = platform.name.charAt(0).toUpperCase() + platform.name.slice(1);
    confirmMessage.textContent = `Are you sure you want to delete the ${platformLabel} platform? This will also delete all associated subsources.`;
    confirmCallback = async () => await deletePlatform(platform.id);
    confirmModal.hidden = false;
    console.log('Confirmation modal should be visible now'); // Debug log
}

// Delete Platform via API
async function deletePlatform(platformId) {
    console.log('deletePlatform called for ID:', platformId); // Debug log
    try {
        console.log('Sending DELETE request to:', `/api/platforms/${platformId}`); // Debug log
        const response = await fetchWithTimeout(`/api/platforms/${platformId}`, {
            method: 'DELETE',
        });
        
        console.log('Delete response status:', response.status); // Debug log
        console.log('Delete response ok:', response.ok); // Debug log
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            const errorMsg = errorData.error || `Failed to delete platform: ${response.status}`;
            throw new Error(errorMsg);
        }
        
        console.log('Platform deleted successfully, reloading data...'); // Debug log
        showSuccess('Platform deleted successfully');
        await loadData();
        console.log('Data reloaded after delete'); // Debug log
    } catch (error) {
        console.error('Delete platform error:', error); // Debug log
        showError(error.message);
    }
}

// ===== SUBSOURCE FUNCTIONS =====

// Load Subsources
async function loadSubsources() {
    subsourcesLoading.hidden = false;
    subsourcesEmpty.hidden = true;
    subsourcesList.innerHTML = '';
    
    try {
        const subsources = await fetchAllSubsources();
        currentSubsources = subsources;
        renderSubsources(subsources);
    } catch (error) {
        console.error('Failed to load subsources:', error);
        subsourcesLoading.hidden = true;
        subsourcesEmpty.hidden = false;
    }
}

// Fetch All Subsources from API
async function fetchAllSubsources() {
    // Fetch subsources for each platform
    const allSubsources = [];
    
    for (const platform of currentPlatforms) {
        try {
            const subsources = await fetchSubsources(platform.id);
            allSubsources.push(...subsources);
        } catch (error) {
            console.error(`Failed to fetch subsources for platform ${platform.name}:`, error);
        }
    }
    
    return allSubsources;
}

// Fetch Subsources for a Platform
async function fetchSubsources(platformId) {
    const response = await fetchWithTimeout(`/api/platforms/${platformId}/subsources`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
    });
    
    if (!response.ok) {
        throw new Error(`Failed to fetch subsources: ${response.status}`);
    }
    
    return await response.json();
}

// Render Subsources Grouped by Platform
function renderSubsources(subsources) {
    subsourcesLoading.hidden = true;
    
    if (!subsources || subsources.length === 0) {
        subsourcesEmpty.hidden = false;
        subsourcesList.innerHTML = '';
        return;
    }
    
    subsourcesEmpty.hidden = true;
    
    // Group subsources by platform
    const grouped = {};
    subsources.forEach(subsource => {
        const platformName = subsource.platform_name || 'Unknown';
        if (!grouped[platformName]) {
            grouped[platformName] = [];
        }
        grouped[platformName].push(subsource);
    });
    
    // Render each platform group
    const html = Object.entries(grouped).map(([platformName, platformSubsources]) => {
        const platform = currentPlatforms.find(p => p.name === platformName);
        return createPlatformGroup(platformName, platform, platformSubsources);
    }).join('');
    
    subsourcesList.innerHTML = html;
    
    // Attach event listeners
    currentPlatforms.forEach(platform => {
        const addBtn = document.getElementById(`add-subsource-${platform.id}`);
        if (addBtn) {
            addBtn.addEventListener('click', () => openSubsourceModal(platform));
        }
    });
    
    subsources.forEach(subsource => {
        const editBtn = document.getElementById(`edit-subsource-${subsource.id}`);
        if (editBtn) {
            editBtn.addEventListener('click', () => openEditSubsourceModal(subsource));
        }
        
        const deleteBtn = document.getElementById(`delete-subsource-${subsource.id}`);
        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => confirmDeleteSubsource(subsource));
        }
    });
}

// Create Platform Group HTML
function createPlatformGroup(platformName, platform, subsources) {
    const platformLabel = platformName.charAt(0).toUpperCase() + platformName.slice(1);
    const platformId = platform ? platform.id : '';
    
    const subsourcesHtml = subsources.map(subsource => createSubsourceCard(subsource)).join('');
    
    return `
        <div class="platform-group">
            <div class="platform-group-header">
                <div class="platform-group-title">
                    <span class="platform-badge platform-${platformName}">${platformLabel}</span>
                    <span class="subsource-count">${subsources.length} subsource${subsources.length !== 1 ? 's' : ''}</span>
                </div>
                ${platform ? `
                    <button 
                        id="add-subsource-${platformId}" 
                        class="btn btn-primary btn-small"
                        aria-label="Add subsource to ${platformLabel}"
                    >
                        Add Subsource
                    </button>
                ` : ''}
            </div>
            <div class="subsources-grid">
                ${subsourcesHtml}
            </div>
        </div>
    `;
}

// Create Subsource Card HTML
function createSubsourceCard(subsource) {
    const createdDate = new Date(subsource.created_at).toLocaleString();
    
    const urlSection = subsource.url 
        ? `<div class="subsource-detail">
                <strong>URL:</strong>
                <a href="${escapeHtml(subsource.url)}" target="_blank" rel="noopener noreferrer" class="subsource-link">
                    ${escapeHtml(subsource.url)}
                </a>
            </div>`
        : '';
    
    return `
        <div class="subsource-card">
            <div class="subsource-card-header">
                <h4 class="subsource-name">${escapeHtml(subsource.name)}</h4>
                <div class="subsource-actions">
                    <button 
                        id="edit-subsource-${subsource.id}" 
                        class="btn btn-secondary btn-small"
                        aria-label="Edit ${escapeHtml(subsource.name)}"
                    >
                        Edit
                    </button>
                    <button 
                        id="delete-subsource-${subsource.id}" 
                        class="btn btn-danger btn-small"
                        aria-label="Delete ${escapeHtml(subsource.name)}"
                    >
                        Delete
                    </button>
                </div>
            </div>
            <div class="subsource-details">
                <div class="subsource-detail">
                    <strong>Identifier:</strong> <code>${escapeHtml(subsource.identifier)}</code>
                </div>
                ${urlSection}
                <div class="subsource-detail">
                    <strong>Created:</strong> ${createdDate}
                </div>
            </div>
        </div>
    `;
}

// Open Subsource Modal
function openSubsourceModal(platform) {
    editingSubsourceId = null;
    subsourceForm.reset();
    clearSubsourceErrors();
    
    document.getElementById('subsource-modal-title').textContent = 'Add Subsource';
    subsourceSubmitBtn.querySelector('.button-text').textContent = 'Create Subsource';
    
    subsourcePlatformId.value = platform.id;
    const platformLabel = platform.name.charAt(0).toUpperCase() + platform.name.slice(1);
    subsourcePlatformName.textContent = platformLabel;
    subsourcePlatformName.className = `platform-badge platform-${platform.name}`;
    
    subsourceModal.hidden = false;
    subsourceNameInput.focus();
}

// Open Edit Subsource Modal
function openEditSubsourceModal(subsource) {
    editingSubsourceId = subsource.id;
    subsourceForm.reset();
    clearSubsourceErrors();
    
    document.getElementById('subsource-modal-title').textContent = 'Edit Subsource';
    subsourceSubmitBtn.querySelector('.button-text').textContent = 'Update Subsource';
    
    // Find the platform for this subsource
    const platform = currentPlatforms.find(p => p.id === subsource.platform_id);
    if (platform) {
        subsourcePlatformId.value = platform.id;
        const platformLabel = platform.name.charAt(0).toUpperCase() + platform.name.slice(1);
        subsourcePlatformName.textContent = platformLabel;
        subsourcePlatformName.className = `platform-badge platform-${platform.name}`;
    }
    
    // Pre-fill form with existing data
    subsourceNameInput.value = subsource.name;
    subsourceIdentifierInput.value = subsource.identifier;
    subsourceUrlInput.value = subsource.url || '';
    
    subsourceModal.hidden = false;
    subsourceNameInput.focus();
}

// Close Subsource Modal
function closeSubsourceModal() {
    subsourceModal.hidden = true;
    subsourceForm.reset();
    clearSubsourceErrors();
}

// Clear Subsource Form Errors
function clearSubsourceErrors() {
    setFieldError(subsourceNameInput, subsourceNameError, '');
    setFieldError(subsourceIdentifierInput, subsourceIdentifierError, '');
}

// Validate Subsource Field
function validateSubsourceField(fieldName) {
    let isValid = true;
    let errorMsg = '';
    
    switch (fieldName) {
        case 'name':
            const name = subsourceNameInput.value.trim();
            if (!name) {
                isValid = false;
                errorMsg = 'Subsource name is required';
            }
            setFieldError(subsourceNameInput, subsourceNameError, errorMsg);
            break;
            
        case 'identifier':
            const identifier = subsourceIdentifierInput.value.trim();
            if (!identifier) {
                isValid = false;
                errorMsg = 'Identifier is required';
            }
            setFieldError(subsourceIdentifierInput, subsourceIdentifierError, errorMsg);
            break;
    }
    
    return isValid;
}

// Validate Subsource Form
function validateSubsourceForm() {
    const nameValid = validateSubsourceField('name');
    const identifierValid = validateSubsourceField('identifier');
    return nameValid && identifierValid;
}

// Handle Subsource Form Submit
async function handleSubsourceSubmit(event) {
    event.preventDefault();
    
    hideMessages();
    
    if (!validateSubsourceForm()) {
        showError('Please fix the validation errors before submitting');
        return;
    }
    
    const isEditing = editingSubsourceId !== null;
    const buttonText = isEditing ? 'Updating...' : 'Creating...';
    setFormLoading(subsourceSubmitBtn, true, buttonText);
    
    const platformId = subsourcePlatformId.value;
    const formData = {
        name: subsourceNameInput.value.trim(),
        identifier: subsourceIdentifierInput.value.trim(),
    };
    
    const url = subsourceUrlInput.value.trim();
    if (url) {
        formData.url = url;
    }
    
    try {
        if (isEditing) {
            await updateSubsource(editingSubsourceId, formData);
            showSuccess('Subsource updated successfully!');
        } else {
            await createSubsource(platformId, formData);
            showSuccess(`Subsource "${formData.name}" created successfully!`);
        }
        closeSubsourceModal();
        await loadSubsources();
    } catch (error) {
        showError(error.message);
    } finally {
        const finalButtonText = isEditing ? 'Update Subsource' : 'Create Subsource';
        setFormLoading(subsourceSubmitBtn, false, finalButtonText);
    }
}

// Create Subsource via API
async function createSubsource(platformId, data) {
    const response = await fetchWithTimeout(`/api/platforms/${platformId}/subsources`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.error || `Server error: ${response.status}`;
        throw new Error(errorMsg);
    }
    
    return await response.json();
}

// Update Subsource via API
async function updateSubsource(subsourceId, data) {
    const response = await fetchWithTimeout(`/api/subsources/${subsourceId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.error || `Server error: ${response.status}`;
        throw new Error(errorMsg);
    }
    
    return await response.json();
}

// Confirm Delete Subsource
function confirmDeleteSubsource(subsource) {
    confirmMessage.textContent = `Are you sure you want to delete the subsource "${subsource.name}"?`;
    confirmCallback = async () => await deleteSubsource(subsource.id);
    confirmModal.hidden = false;
}

// Delete Subsource via API
async function deleteSubsource(subsourceId) {
    console.log('deleteSubsource called for ID:', subsourceId); // Debug log
    try {
        console.log('Sending DELETE request to:', `/api/subsources/${subsourceId}`); // Debug log
        const response = await fetchWithTimeout(`/api/subsources/${subsourceId}`, {
            method: 'DELETE',
        });
        
        console.log('Delete subsource response status:', response.status); // Debug log
        console.log('Delete subsource response ok:', response.ok); // Debug log
        
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            const errorMsg = errorData.error || `Failed to delete subsource: ${response.status}`;
            throw new Error(errorMsg);
        }
        
        console.log('Subsource deleted successfully, reloading subsources...'); // Debug log
        showSuccess('Subsource deleted successfully');
        await loadSubsources();
        console.log('Subsources reloaded after delete'); // Debug log
    } catch (error) {
        console.error('Delete subsource error:', error); // Debug log
        showError(error.message);
    }
}

// ===== CONFIRMATION MODAL =====

// Handle Confirm OK
async function handleConfirmOk() {
    console.log('handleConfirmOk called, callback:', confirmCallback); // Debug log
    
    // Store callback before closing modal (which sets it to null)
    const callback = confirmCallback;
    closeConfirmModal();
    
    if (callback) {
        console.log('Executing callback...'); // Debug log
        await callback();
    } else {
        console.log('No callback to execute'); // Debug log
    }
}

// Close Confirm Modal
function closeConfirmModal() {
    confirmModal.hidden = true;
    confirmCallback = null;
}

// ===== UTILITY FUNCTIONS =====

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

// Set Form Loading State
function setFormLoading(button, loading, text) {
    const buttonText = button.querySelector('.button-text');
    const buttonSpinner = button.querySelector('.loading-spinner');
    
    button.disabled = loading;
    
    if (loading) {
        buttonText.textContent = text;
        buttonSpinner.hidden = false;
    } else {
        buttonText.textContent = text;
        buttonSpinner.hidden = true;
    }
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

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
