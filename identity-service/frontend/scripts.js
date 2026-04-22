const API_BASE = 'http://127.0.0.1:5000/api';

let currentUser = null;
let accessToken = localStorage.getItem('access_token');
let refreshToken = localStorage.getItem('refresh_token');

const navbar = document.getElementById('navbar');
const contentDiv = document.getElementById('content');
const userGreeting = document.getElementById('user-greeting');
const notification = document.getElementById('notification');

function showNotification(message, type = 'success') {
    notification.textContent = message;
    notification.className = `notification ${type}`;
    setTimeout(() => notification.classList.add('hidden'), 3000);
}

async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
    }

    let response = await fetch(url, { ...options, headers });
    if (response.status === 401 && refreshToken) {
        const refreshResp = await fetch(`${API_BASE}/auth/refresh`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${refreshToken}` }
        });
        if (refreshResp.ok) {
            const data = await refreshResp.json();
            accessToken = data.access_token;
            localStorage.setItem('access_token', accessToken);
            headers['Authorization'] = `Bearer ${accessToken}`;
            response = await fetch(url, { ...options, headers });
        } else {
            logout();
            throw new Error('Сессия истекла');
        }
    }
    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Ошибка запроса');
    }
    return response;
}

function logout() {
    localStorage.clear();
    accessToken = null;
    refreshToken = null;
    currentUser = null;
    navbar.style.display = 'none';
    renderLoginPage();
    showNotification('Вы вышли из системы');
}

function renderLoginPage() {
    contentDiv.innerHTML = `
        <h2>Вход</h2>
        <form id="login-form">
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" required>
            </div>
            <div class="form-group">
                <label>Пароль</label>
                <input type="password" name="password" required>
            </div>
            <button type="submit">Войти</button>
            <p>Нет аккаунта? <a href="#" id="show-register">Зарегистрироваться</a></p>
        </form>
    `;
    document.getElementById('login-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        try {
            const resp = await fetch(`${API_BASE}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(Object.fromEntries(formData))
            });
            if (!resp.ok) throw new Error('Неверный email или пароль');
            const data = await resp.json();
            accessToken = data.access_token;
            refreshToken = data.refresh_token;
            localStorage.setItem('access_token', accessToken);
            localStorage.setItem('refresh_token', refreshToken);
            currentUser = data.user;
            showNotification('Успешный вход!');
            renderApp();
        } catch (err) {
            showNotification(err.message, 'error');
        }
    });
    document.getElementById('show-register').addEventListener('click', (e) => {
        e.preventDefault();
        renderRegisterPage();
    });
}

function renderRegisterPage() {
    contentDiv.innerHTML = `
        <h2>Регистрация</h2>
        <form id="register-form">
            <div class="form-group">
                <label>Имя</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" required>
            </div>
            <div class="form-group">
                <label>Пароль</label>
                <input type="password" name="password" required minlength="6">
            </div>
            <button type="submit">Зарегистрироваться</button>
            <p>Уже есть аккаунт? <a href="#" id="show-login">Войти</a></p>
        </form>
    `;
    document.getElementById('register-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        try {
            const resp = await fetch(`${API_BASE}/auth/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(Object.fromEntries(formData))
            });
            if (!resp.ok) {
                const err = await resp.json();
                throw new Error(err.error);
            }
            showNotification('Регистрация успешна! Теперь войдите.');
            renderLoginPage();
        } catch (err) {
            showNotification(err.message, 'error');
        }
    });
    document.getElementById('show-login').addEventListener('click', (e) => {
        e.preventDefault();
        renderLoginPage();
    });
}

async function renderProfilePage() {
    const resp = await apiRequest('/profile');
    const user = await resp.json();
    contentDiv.innerHTML = `
        <h2>Профиль</h2>
        <form id="profile-form">
            <div class="form-group">
                <label>Имя</label>
                <input type="text" name="name" value="${user.name}" required>
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" value="${user.email}" required>
            </div>
            <div class="form-group">
                <label>Новый пароль (оставьте пустым, чтобы не менять)</label>
                <input type="password" name="password" minlength="6">
            </div>
            <button type="submit">Сохранить</button>
        </form>
    `;
    document.getElementById('profile-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData);
        if (!data.password) delete data.password;
        try {
            await apiRequest('/profile', {
                method: 'PUT',
                body: JSON.stringify(data)
            });
            showNotification('Профиль обновлён');
            // Обновить приветствие
            currentUser.name = data.name;
            userGreeting.textContent = `Привет, ${currentUser.name}!`;
        } catch (err) {
            showNotification(err.message, 'error');
        }
    });
}

async function renderProjectsPage() {
    const resp = await apiRequest('/projects');
    const projects = await resp.json();
    let html = `
        <h2>Мои проекты</h2>
        <form id="add-project-form" style="margin-bottom: 20px;">
            <div class="form-group" style="display: flex; gap: 10px;">
                <input type="text" name="name" placeholder="Название проекта" required style="flex:1;">
                <button type="submit">Создать</button>
            </div>
        </form>
    `;
    if (projects.length) {
        html += `<table>
            <tr><th>ID</th><th>Название</th><th>Действия</th></tr>`;
        projects.forEach(p => {
            html += `<tr>
                <td>${p.id}</td>
                <td>${p.name}</td>
                <td>
                    <button class="action-btn delete-btn" data-id="${p.id}">Удалить</button>
                </td>
            </tr>`;
        });
        html += `</table>`;
    } else {
        html += `<p>У вас пока нет проектов.</p>`;
    }
    contentDiv.innerHTML = html;

    document.getElementById('add-project-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = e.target.name.value;
        try {
            await apiRequest('/projects', {
                method: 'POST',
                body: JSON.stringify({ name })
            });
            e.target.reset();
            showNotification('Проект создан');
            renderProjectsPage();
        } catch (err) {
            showNotification(err.message, 'error');
        }
    });

    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = btn.dataset.id;
            if (confirm('Удалить проект?')) {
                try {
                    await apiRequest(`/projects/${id}`, { method: 'DELETE' });
                    showNotification('Проект удалён');
                    renderProjectsPage();
                } catch (err) {
                    showNotification(err.message, 'error');
                }
            }
        });
    });
}

async function renderApiKeysPage() {
    const resp = await apiRequest('/api-keys');
    const keys = await resp.json();
    let html = `
        <h2>API-ключи</h2>
        <p>Используйте эти ключи для доступа агента к API.</p>
        <form id="add-key-form" style="margin-bottom: 20px;">
            <div class="form-group" style="display: flex; gap: 10px;">
                <input type="text" name="name" placeholder="Название ключа" required style="flex:1;">
                <button type="submit">Создать ключ</button>
            </div>
        </form>
    `;
    if (keys.length) {
        html += `<table>
            <tr><th>Название</th><th>Ключ</th><th>Статус</th><th>Действия</th></tr>`;
        keys.forEach(k => {
            html += `<tr>
                <td>${k.name}</td>
                <td><code>${k.key}</code></td>
                <td>${k.is_active ? 'Активен' : 'Отозван'}</td>
                <td>
                    ${k.is_active ? `<button class="action-btn delete-btn" data-id="${k.id}">Отозвать</button>` : ''}
                </td>
            </tr>`;
        });
        html += `</table>`;
    } else {
        html += `<p>У вас нет API-ключей.</p>`;
    }
    contentDiv.innerHTML = html;

    document.getElementById('add-key-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = e.target.name.value;
        try {
            await apiRequest('/api-keys', {
                method: 'POST',
                body: JSON.stringify({ name })
            });
            e.target.reset();
            showNotification('Ключ создан');
            renderApiKeysPage();
        } catch (err) {
            showNotification(err.message, 'error');
        }
    });

    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = btn.dataset.id;
            if (confirm('Отозвать ключ?')) {
                try {
                    await apiRequest(`/api-keys/${id}`, { method: 'DELETE' });
                    showNotification('Ключ отозван');
                    renderApiKeysPage();
                } catch (err) {
                    showNotification(err.message, 'error');
                }
            }
        });
    });
}

function renderApp() {
    navbar.style.display = 'flex';
    userGreeting.textContent = `Привет, ${currentUser.name}!`;
    renderProfilePage();
}

document.getElementById('nav-profile').addEventListener('click', renderProfilePage);
document.getElementById('nav-projects').addEventListener('click', renderProjectsPage);
document.getElementById('nav-apikeys').addEventListener('click', renderApiKeysPage);
document.getElementById('logout-btn').addEventListener('click', logout);

if (accessToken) {
    fetch(`${API_BASE}/profile`, {
        headers: { 'Authorization': `Bearer ${accessToken}` }
    }).then(async resp => {
        if (resp.ok) {
            currentUser = await resp.json();
            renderApp();
        } else {
            localStorage.clear();
            renderLoginPage();
        }
    }).catch(() => {
        renderLoginPage();
    });
} else {
    renderLoginPage();
}