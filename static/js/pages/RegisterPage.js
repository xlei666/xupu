// 注册页面

import BaseComponent from '../components/BaseComponent.js';
import { authAPI } from '../api.js';
import { userActions } from '../store.js';
import { showToast, validateForm } from '../utils.js';
import router from '../router.js';

export default class RegisterPage extends BaseComponent {
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
                        <span>创建账号</span>
                    </div>
                    <p class="auth-subtitle">加入 NovelFlow 开始您的创作之旅</p>

                    <form id="registerForm">
                        <div class="mb-3">
                            <label for="username" class="form-label">用户名</label>
                            <input type="text" class="form-control" id="username" name="username"
                                   placeholder="请输入用户名" required>
                            <div class="form-text">3-20个字符，只能包含字母、数字和下划线</div>
                        </div>

                        <div class="mb-3">
                            <label for="email" class="form-label">邮箱</label>
                            <input type="email" class="form-control" id="email" name="email"
                                   placeholder="请输入邮箱地址" required>
                        </div>

                        <div class="mb-3">
                            <label for="password" class="form-label">密码</label>
                            <input type="password" class="form-control" id="password" name="password"
                                   placeholder="请设置密码" required>
                            <div class="form-text">至少6个字符</div>
                        </div>

                        <div class="mb-4">
                            <label for="confirmPassword" class="form-label">确认密码</label>
                            <input type="password" class="form-control" id="confirmPassword" name="confirmPassword"
                                   placeholder="请再次输入密码" required>
                        </div>

                        <button type="submit" class="btn btn-primary w-100 btn-lg mb-3">
                            创建账号
                        </button>

                        <div class="auth-divider">或</div>

                        <div class="text-center">
                            <span class="text-muted">已有账号？</span>
                            <a href="#/login" class="text-decoration-none" style="color: var(--primary); font-weight: 500;">立即登录</a>
                        </div>
                    </form>
                </div>
            </div>
        `;
    }

    bindEvents() {
        const form = this.$('#registerForm');
        if (form) {
            this.addEventListener(form, 'submit', (e) => this.handleSubmit(e));
        }
    }

    async handleSubmit(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const username = formData.get('username');
        const email = formData.get('email');
        const password = formData.get('password');
        const confirmPassword = formData.get('confirmPassword');

        // 表单验证
        const validation = validateForm({ username, email, password, confirmPassword }, {
            username: {
                required: true,
                minLength: 3,
                maxLength: 20,
                pattern: /^[a-zA-Z0-9_]+$/,
                message: '用户名格式不正确'
            },
            email: {
                required: true,
                email: true,
                message: '邮箱格式不正确'
            },
            password: {
                required: true,
                minLength: 6,
                message: '密码至少6个字符'
            },
            confirmPassword: {
                required: true,
                custom: (value, data) => value === data.password,
                message: '两次密码输入不一致'
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
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>注册中...';

            const response = await authAPI.register({ username, email, password });

            if (response.success && response.data?.tokens?.access_token) {
                // 注册成功后自动登录
                userActions.login(response.data.user, response.data.tokens.access_token);
                showToast('注册成功', 'success');
                router.navigate('/dashboard');
            } else {
                showToast('注册成功，请登录', 'success');
                router.navigate('/login');
            }
        } catch (error) {
            console.error('注册错误:', error);
            showToast(error.message || '注册失败，请稍后重试', 'error');
        } finally {
            const submitBtn = this.$('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '创建账号';
            }
        }
    }
}
