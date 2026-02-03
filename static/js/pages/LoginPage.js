// 登录页面

import BaseComponent from '../components/BaseComponent.js';
import { authAPI } from '../api.js';
import { userActions } from '../store.js';
import { showToast, validateForm } from '../utils.js';
import router from '../router.js';

export default class LoginPage extends BaseComponent {
    render() {
        return `
            <div class="auth-page">
                <div class="auth-card">
                    <div class="auth-logo">
                        <svg width="36" height="36" viewBox="0 0 48 48" fill="none">
                            <rect x="8" y="16" width="24" height="28" rx="2" fill="#E8E8F4"/>
                            <rect x="12" y="12" width="24" height="28" rx="2" fill="#A5A7E8"/>
                            <rect x="16" y="8" width="24" height="28" rx="2" fill="#5B5FC7"/>
                        </svg>
                        <span>NovelFlow 叙谱</span>
                    </div>
                    <p class="auth-subtitle">AI驱动的小说创作平台</p>

                    <form id="loginForm">
                        <div class="mb-3">
                            <label for="username" class="form-label">用户名</label>
                            <input type="text" class="form-control" id="username" name="username"
                                   placeholder="请输入用户名" required>
                        </div>

                        <div class="mb-3">
                            <label for="password" class="form-label">密码</label>
                            <input type="password" class="form-control" id="password" name="password"
                                   placeholder="请输入密码" required>
                        </div>

                        <div class="mb-4 d-flex justify-content-between align-items-center">
                            <div class="form-check">
                                <input type="checkbox" class="form-check-input" id="remember">
                                <label class="form-check-label" for="remember">记住我</label>
                            </div>
                        </div>

                        <button type="submit" class="btn btn-primary w-100 btn-lg mb-3">
                            登录
                        </button>

                        <div class="auth-divider">或</div>

                        <div class="text-center">
                            <span class="text-muted">还没有账号？</span>
                            <a href="#/register" class="text-decoration-none" style="color: var(--primary); font-weight: 500;">立即注册</a>
                        </div>
                    </form>
                </div>
            </div>
        `;
    }

    bindEvents() {
        const form = this.$('#loginForm');
        if (form) {
            this.addEventListener(form, 'submit', (e) => this.handleSubmit(e));
        }
    }

    async handleSubmit(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const username = formData.get('username');
        const password = formData.get('password');

        // 表单验证
        const validation = validateForm({ username, password }, {
            username: {
                required: true,
                minLength: 3,
                message: '用户名至少3个字符'
            },
            password: {
                required: true,
                minLength: 6,
                message: '密码至少6个字符'
            }
        });

        if (!validation.valid) {
            const firstError = Object.values(validation.errors)[0];
            showToast(firstError, 'error');
            return;
        }

        try {
            const submitBtn = this.$('button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>登录中...';

            const response = await authAPI.login(username, password);

            if (response.token) {
                userActions.login(response.user, response.token);
                showToast('登录成功', 'success');
                router.navigate('/dashboard');
            } else {
                showToast('登录失败，请检查用户名和密码', 'error');
            }
        } catch (error) {
            console.error('登录错误:', error);
            showToast(error.message || '登录失败，请稍后重试', 'error');
        } finally {
            const submitBtn = this.$('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '登录';
            }
        }
    }
}
