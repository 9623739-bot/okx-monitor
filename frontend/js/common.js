/**
 * OKX Monitor - 公共JS
 */

// API 基础路径
const API_BASE = '/api';

// 存储 token
const TOKEN_KEY = 'okx_monitor_token';

// ===== Token 管理 =====
function getToken() {
    return localStorage.getItem(TOKEN_KEY);
}

function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
}

function clearToken() {
    localStorage.removeItem(TOKEN_KEY);
}

function isLoggedIn() {
    return !!getToken();
}

// ===== HTTP 请求 =====
async function apiRequest(method, path, body) {
    const url = API_BASE + path;
    const headers = {
        'Content-Type': 'application/json'
    };

    const token = getToken();
    if (token) {
        headers['Authorization'] = 'Bearer ' + token;
    }

    const opts = {
        method: method,
        headers: headers
    };

    if (body && method !== 'GET') {
        opts.body = JSON.stringify(body);
    }

    try {
        const resp = await fetch(url, opts);
        const data = await resp.json();

        if (data.code === 401) {
            clearToken();
            window.location.href = '/index.html';
            return null;
        }

        return data;
    } catch (err) {
        console.error('API Error:', err);
        showToast('网络请求失败', 'error');
        return null;
    }
}

function apiGet(path) {
    return apiRequest('GET', path);
}

function apiPost(path, body) {
    return apiRequest('POST', path, body);
}

// ===== Toast 提示 =====
function showToast(message, type) {
    type = type || 'info';
    var toast = document.createElement('div');
    toast.className = 'toast toast-' + type;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(function () {
        toast.remove();
    }, 3000);
}

// ===== 格式化 =====
function formatPrice(price) {
    if (price >= 1000) {
        return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    } else if (price >= 1) {
        return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 4 });
    } else {
        return price.toLocaleString('en-US', { minimumFractionDigits: 4, maximumFractionDigits: 8 });
    }
}

function formatVolume(vol) {
    if (vol >= 1000000) {
        return (vol / 1000000).toFixed(1) + 'M';
    } else if (vol >= 1000) {
        return (vol / 1000).toFixed(1) + 'K';
    }
    return vol.toFixed(1);
}

function formatTime(dateStr) {
    var d = new Date(dateStr);
    var pad = function (n) { return n < 10 ? '0' + n : n; };
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
        ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
}

function formatPct(pct, showSign) {
    var s = pct.toFixed(2) + '%';
    if (showSign && pct > 0) s = '+' + s;
    return s;
}

function formatPnl(pnl) {
    var prefix = pnl >= 0 ? '+' : '';
    return prefix + pnl.toFixed(2) + ' USDT';
}

// 去后缀：RLS-USDT-SWAP → RLS
function formatSymbol(s) {
    return s ? s.replace('-USDT-SWAP', '') : s;
}

// ===== 数字动画 =====
function animateValue(el, start, end, duration) {
    var startTime = null;
    var isFloat = (end % 1) !== 0;
    var decimals = isFloat ? 2 : 0;

    function step(timestamp) {
        if (!startTime) startTime = timestamp;
        var progress = Math.min((timestamp - startTime) / duration, 1);
        var current = start + (end - start) * progress;
        el.textContent = isFloat ? current.toFixed(decimals) : Math.floor(current);
        if (progress < 1) {
            requestAnimationFrame(step);
        }
    }
    requestAnimationFrame(step);
}

// ===== 侧边栏激活状态 =====
function setActiveNav() {
    var path = window.location.pathname;
    var links = document.querySelectorAll('.sidebar-nav a');
    links.forEach(function (link) {
        var href = link.getAttribute('href');
        if (href && path.endsWith(href.replace('./', ''))) {
            link.classList.add('active');
        }
    });
}

// ===== 退出登录 =====
function logout() {
    clearToken();
    window.location.href = '/index.html';
}

// ===== 页面初始化 =====
document.addEventListener('DOMContentLoaded', function () {
    setActiveNav();

    // 非登录页检查登录状态
    if (!window.location.pathname.includes('index.html') && 
        window.location.pathname !== '/' &&
        !isLoggedIn()) {
        window.location.href = '/index.html';
    }
});
