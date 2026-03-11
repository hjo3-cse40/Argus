/**
 * auth.js — Argus login handler
 * Handles POST /api/auth/login, session cookie set by server.
 */

// If already logged in, go straight to dashboard
if (window.location.pathname === '/login.html' || window.location.pathname === '/register.html' || window.location.pathname === '/' || window.location.pathname === '/index.html') {
  fetch('/api/auth/me', { credentials: 'same-origin' })
    .then(res => {
      if (res.ok) {
        window.location.replace('/dashboard.html');
      } else {
        document.body.style.visibility = 'visible';
      }
    })
    .catch(() => {
      document.body.style.visibility = 'visible';
    });
} else {
  document.body.style.visibility = 'visible';
}


(function () {
  'use strict';

  const LOGIN_ENDPOINT = '/api/auth/login';
  const DASHBOARD_URL  = '/dashboard.html';
  const LOGIN_URL      = '/login.html';

  const form       = document.getElementById('login-form');
  const emailInput = document.getElementById('email');
  const passInput  = document.getElementById('password');
  const submitBtn  = document.getElementById('submit-btn');
  const errorMsg   = document.getElementById('error-message');
  const errorText  = document.getElementById('error-text');
  const emailError = document.getElementById('email-error');
  const passError  = document.getElementById('password-error');

  if (!form) return;

  function setLoading(on) {
    if (submitBtn) submitBtn.disabled = on;
  }

  function showError(msg) {
    errorText.textContent = msg;
    errorMsg.hidden = false;
  }

  function clearErrors() {
    errorMsg.hidden = true;
    emailError.textContent = '';
    passError.textContent = '';
    emailInput.removeAttribute('aria-invalid');
    passInput.removeAttribute('aria-invalid');
  }

  function validate() {
    let ok = true;
    const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (!emailInput.value.trim()) {
      emailError.textContent = 'Email is required.';
      emailInput.setAttribute('aria-invalid', 'true');
      ok = false;
    } else if (!emailRe.test(emailInput.value.trim())) {
      emailError.textContent = 'Enter a valid email address.';
      emailInput.setAttribute('aria-invalid', 'true');
      ok = false;
    }

    if (!passInput.value) {
      passError.textContent = 'Password is required.';
      passInput.setAttribute('aria-invalid', 'true');
      ok = false;
    }

    return ok;
  }

  form.addEventListener('submit', async function (e) {
    e.preventDefault();
    clearErrors();
    if (!validate()) return;

    try {
      const res = await fetch(LOGIN_ENDPOINT, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ email: emailInput.value.trim(), password: passInput.value }),
      });

      if (res.ok) {
        window.location.replace(DASHBOARD_URL);
        return;
      }

      if (res.status === 401) {
        passInput.value = '';
        passInput.focus();
        showError('Invalid email or password.');
        return;
      }

      showError('Server error. Please try again.');

    } catch (err) {
      showError('Cannot reach the server. Check your connection.');
    }
  });

  // Session guard — call AuthGuard.checkSession() on protected pages
  window.AuthGuard = {
    checkSession: async function () {
      try {
        const res = await fetch('/api/auth/me', { credentials: 'same-origin' });
        if (res.status === 401) window.location.replace(LOGIN_URL);
      } catch (_) {}
    },
    logout: async function () {
      try {
        await fetch('/api/auth/logout', { method: 'POST', credentials: 'same-origin' });
      } catch (_) {}
      window.location.replace(LOGIN_URL);
    },
  };

})();