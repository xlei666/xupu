// 注册页面

import BaseComponent from '../components/BaseComponent.js';
import { authAPI } from '../api.js';
import { showToast, validateForm } from '../utils.js';
import router from '../router.js';

export default class RegisterPage extends BaseComponent {
    render() {
        return `
            <div class="row justify-content-center align-items-center" style="min-height: 80vh;">
                <div class="col-md-5 col-lg-4">
                    <div class="card shadow">
                        <div class="card-body p-5">
                            <div class="text-center mb-4">
                                <h2 class="fw-bold">
                                    <i class="bi bi-person-plus text-primary me-2"></i>
                                    注册账号
                                </h2>
                                <p class="text-muted">加入NovelFlow开始创作</p>
                            </div>
                            
                            <form id="registerForm">
                                <div class="mb-3">
                                    <label for="username" class="form-label">用户名</label>
                                    <input type="text" class="form-control" id="username" name="username" required>
                                    <div class="form-text">3-20个字符，只能包含字母、数字和下划线</div>
                                </div>
                                
                                <div class="mb-3">
                                    <label for="email" class="form-label">邮箱</label>
                                    <input type="email" class="form-control" id="email" name="email" required>
                                </div>
                                
                                <div class="mb-3">
                                    <label for="password" class="form-label">密码</label>
                                    <input type="password" class="form-control" id="password" name="password" required>
                                    <div class="form-text">至少6个字符</div>
                                </div>
                                
                                <div class="mb-3">
                                    <label for="confirmPassword" class="form-label">确认密码</label>
                                    <input type="password" class="form-control" id="confirmPassword" name="confirmPassword" required>
                                </div>
                                
                                <button type="submit" class="btn btn-primary w-100 mb-3">
                                    <i class="bi bi-person-plus me-1"></i>
                                    注册
                                </button>
                                
                                <div class="text-center">
                                    <span class="text-muted">已有账号？</span>
                                    <a href="#/login" class="text-decoration-none">立即登录</a>
                                </div>
                            </form>
                        </div>
                    </div>
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
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>注册中...';

            await authAPI.register({ username, email, password });

            showToast('注册成功！请登录', 'success');
            router.navigate('/login');
        } catch (error) {
            console.error('注册错误:', error);
            showToast(error.message || '注册失败，请稍后重试', 'error');
        } finally {
            const submitBtn = this.$('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '<i class="bi bi-person-plus me-1"></i>注册';
            }
        }
    }
}
