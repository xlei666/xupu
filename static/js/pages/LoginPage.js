// 登录页面

import BaseComponent from '../components/BaseComponent.js';
import { authAPI } from '../api.js';
import { userActions } from '../store.js';
import { showToast, validateForm } from '../utils.js';
import router from '../router.js';

export default class LoginPage extends BaseComponent {
    render() {
        return `
            <div class="row justify-content-center align-items-center" style="min-height: 80vh;">
                <div class="col-md-5 col-lg-4">
                    <div class="card shadow">
                        <div class="card-body p-5">
                            <div class="text-center mb-4">
                                <h2 class="fw-bold">
                                    <i class="bi bi-book text-primary me-2"></i>
                                    NovelFlow 叙谱
                                </h2>
                                <p class="text-muted">AI驱动的小说创作平台</p>
                            </div>
                            
                            <form id="loginForm">
                                <div class="mb-3">
                                    <label for="username" class="form-label">用户名</label>
                                    <input type="text" class="form-control" id="username" name="username" required>
                                </div>
                                
                                <div class="mb-3">
                                    <label for="password" class="form-label">密码</label>
                                    <input type="password" class="form-control" id="password" name="password" required>
                                </div>
                                
                                <div class="mb-3 form-check">
                                    <input type="checkbox" class="form-check-input" id="remember">
                                    <label class="form-check-label" for="remember">记住我</label>
                                </div>
                                
                                <button type="submit" class="btn btn-primary w-100 mb-3">
                                    <i class="bi bi-box-arrow-in-right me-1"></i>
                                    登录
                                </button>
                                
                                <div class="text-center">
                                    <span class="text-muted">还没有账号？</span>
                                    <a href="#/register" class="text-decoration-none">立即注册</a>
                                </div>
                            </form>
                        </div>
                    </div>
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
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>登录中...';

            const response = await authAPI.login(username, password);

            if (response.token) {
                userActions.login(response.user, response.token);
                showToast('登录成功！', 'success');
                router.navigate('/');
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
                submitBtn.innerHTML = '<i class="bi bi-box-arrow-in-right me-1"></i>登录';
            }
        }
    }
}
