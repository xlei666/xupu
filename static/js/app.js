// 应用入口文件

import router from './router.js';
import { userActions } from './store.js';
import Navbar from './components/Navbar.js';
import LandingPage from './pages/LandingPage.js';
import LoginPage from './pages/LoginPage.js';
import RegisterPage from './pages/RegisterPage.js';
import DashboardPage from './pages/DashboardPage.js';
import ProjectDetailPage from './pages/ProjectDetailPage.js';

// 初始化导航栏
const navbar = new Navbar('#navbar');

// 注册路由
router.registerRoutes([
    {
        path: '/',
        meta: { requiresAuth: false, isLanding: true },
        handler: async () => {
            // 如果已登录，跳转到仪表盘
            if (userActions.isAuthenticated()) {
                router.navigate('/dashboard');
                return;
            }
            navbar.unmount();
            document.querySelector('#navbar').innerHTML = '';
            const page = new LandingPage('#app');
            page.mount();
        }
    },
    {
        path: '/login',
        meta: { requiresAuth: false },
        handler: async () => {
            navbar.unmount();
            document.querySelector('#navbar').innerHTML = '';
            const page = new LoginPage('#app');
            page.mount();
        }
    },
    {
        path: '/register',
        meta: { requiresAuth: false },
        handler: async () => {
            navbar.unmount();
            document.querySelector('#navbar').innerHTML = '';
            const page = new RegisterPage('#app');
            page.mount();
        }
    },
    {
        path: '/dashboard',
        meta: { requiresAuth: true },
        handler: async () => {
            navbar.mount();
            const page = new DashboardPage('#app');
            page.mount();
        }
    },
    {
        path: '/project/:id',
        meta: { requiresAuth: true },
        handler: async (params) => {
            navbar.mount();
            const projectId = params._pathParts[1];
            const page = new ProjectDetailPage('#app', projectId);
            page.mount();
        }
    },
    {
        path: '/settings',
        meta: { requiresAuth: true },
        handler: async () => {
            navbar.mount();
            // TODO: 实现设置页面
            document.querySelector('#app').innerHTML = `
                <div class="container py-4">
                    <div class="card">
                        <div class="card-body text-center py-5">
                            <div class="empty-state-icon" style="margin: 0 auto 1.5rem;">
                                <i class="bi bi-gear"></i>
                            </div>
                            <h3>设置</h3>
                            <p class="text-muted">设置页面开发中...</p>
                            <a href="#/dashboard" class="btn btn-primary">返回仪表盘</a>
                        </div>
                    </div>
                </div>
            `;
        }
    },
    {
        path: '/404',
        meta: { requiresAuth: false },
        handler: async () => {
            document.querySelector('#app').innerHTML = `
                <div class="auth-page">
                    <div class="auth-card text-center">
                        <h1 style="font-size: 4rem; color: var(--primary); margin-bottom: 1rem;">404</h1>
                        <h3>页面未找到</h3>
                        <p class="text-muted mb-4">您访问的页面不存在</p>
                        <a href="#/" class="btn btn-primary">返回首页</a>
                    </div>
                </div>
            `;
        }
    }
]);

// 应用启动
console.log('NovelFlow 应用已启动');
