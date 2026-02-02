// 应用入口文件

import router from './router.js';
import Navbar from './components/Navbar.js';
import LoginPage from './pages/LoginPage.js';
import RegisterPage from './pages/RegisterPage.js';
import DashboardPage from './pages/DashboardPage.js';
import ProjectDetailPage from './pages/ProjectDetailPage.js';

// 初始化导航栏
const navbar = new Navbar('#navbar');

// 注册路由
router.registerRoutes([
    {
        path: '/login',
        meta: { requiresAuth: false },
        handler: async () => {
            navbar.unmount();
            const page = new LoginPage('#app');
            page.mount();
        }
    },
    {
        path: '/register',
        meta: { requiresAuth: false },
        handler: async () => {
            navbar.unmount();
            const page = new RegisterPage('#app');
            page.mount();
        }
    },
    {
        path: '/',
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
                <div class="container mt-4">
                    <div class="alert alert-info">
                        <h4>设置</h4>
                        <p>设置页面开发中...</p>
                        <a href="#/" class="btn btn-primary">返回首页</a>
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
                <div class="container mt-5 text-center">
                    <h1 class="display-1">404</h1>
                    <p class="lead">页面未找到</p>
                    <a href="#/" class="btn btn-primary">返回首页</a>
                </div>
            `;
        }
    }
]);

// 应用启动
console.log('NovelFlow 应用已启动');
