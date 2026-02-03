// SPA路由系统

import { userActions } from './store.js';

class Router {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        this.beforeEachHooks = [];

        // 监听hash变化
        window.addEventListener('hashchange', () => this.handleRoute());
        window.addEventListener('load', () => this.handleRoute());
    }

    /**
     * 注册路由
     */
    register(path, handler, meta = {}) {
        this.routes.set(path, { handler, meta });
    }

    /**
     * 批量注册路由
     */
    registerRoutes(routes) {
        routes.forEach(route => {
            this.register(route.path, route.handler, route.meta);
        });
    }

    /**
     * 路由守卫
     */
    beforeEach(hook) {
        this.beforeEachHooks.push(hook);
    }

    /**
     * 处理路由
     */
    async handleRoute() {
        const hash = window.location.hash.slice(1) || '/';
        const { path, params } = this.parsePath(hash);

        const route = this.matchRoute(path);

        if (!route) {
            console.error('路由未找到:', path);
            this.navigate('/404');
            return;
        }

        // 执行路由守卫
        for (const hook of this.beforeEachHooks) {
            const result = await hook({ path, params, meta: route.meta });
            if (result === false) {
                return; // 阻止导航
            }
            if (typeof result === 'string') {
                this.navigate(result); // 重定向
                return;
            }
        }

        this.currentRoute = { path, params, meta: route.meta };

        try {
            await route.handler(params);
        } catch (error) {
            console.error('路由处理错误:', error);
        }
    }

    /**
     * 解析路径和参数
     */
    parsePath(hash) {
        const [path, queryString] = hash.split('?');
        const params = {};

        if (queryString) {
            queryString.split('&').forEach(param => {
                const [key, value] = param.split('=');
                params[decodeURIComponent(key)] = decodeURIComponent(value);
            });
        }

        // 解析路径参数 (例如: /project/:id)
        const pathParts = path.split('/').filter(Boolean);
        params._pathParts = pathParts;

        return { path, params };
    }

    /**
     * 匹配路由
     */
    matchRoute(path) {
        // 精确匹配
        if (this.routes.has(path)) {
            return this.routes.get(path);
        }

        // 动态路由匹配
        for (const [routePath, route] of this.routes) {
            if (this.isMatch(routePath, path)) {
                return route;
            }
        }

        return null;
    }

    /**
     * 判断路径是否匹配
     */
    isMatch(routePath, path) {
        const routeParts = routePath.split('/').filter(Boolean);
        const pathParts = path.split('/').filter(Boolean);

        if (routeParts.length !== pathParts.length) {
            return false;
        }

        return routeParts.every((part, i) => {
            return part.startsWith(':') || part === pathParts[i];
        });
    }

    /**
     * 导航到指定路径
     */
    navigate(path, params = {}) {
        let url = path;

        // 添加查询参数
        const queryString = Object.entries(params)
            .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
            .join('&');

        if (queryString) {
            url += '?' + queryString;
        }

        window.location.hash = url;
    }

    /**
     * 返回上一页
     */
    back() {
        window.history.back();
    }

    /**
     * 前进
     */
    forward() {
        window.history.forward();
    }

    /**
     * 获取当前路由
     */
    getCurrentRoute() {
        return this.currentRoute;
    }
}

// 创建全局路由实例
const router = new Router();

// 添加认证守卫
router.beforeEach(({ path, meta }) => {
    const requiresAuth = meta.requiresAuth === true;
    const isAuthenticated = userActions.isAuthenticated();

    // 如果需要认证但未登录，跳转到登录页
    if (requiresAuth && !isAuthenticated) {
        return '/login';
    }

    // 如果已登录访问登录/注册页，跳转到仪表盘
    if (isAuthenticated && (path === '/login' || path === '/register')) {
        return '/dashboard';
    }

    return true;
});

export default router;
