// 导航栏组件

import BaseComponent from './BaseComponent.js';
import { userActions, subscribe } from '../store.js';
import router from '../router.js';

export default class Navbar extends BaseComponent {
    constructor(container) {
        super(container);
        this.user = userActions.getUser();

        // 订阅状态变化
        this.unsubscribe = subscribe((state) => {
            if (state.user !== this.user) {
                this.user = state.user;
                this.update();
            }
        });
    }

    render() {
        if (!this.user) {
            return ''; // 未登录不显示导航栏
        }

        return `
            <div class="container-fluid d-flex align-items-center" style="height: 100%;">
                <a class="navbar-brand" href="#/">
                    <svg width="24" height="24" viewBox="0 0 48 48" fill="none">
                        <rect x="8" y="16" width="24" height="28" rx="2" fill="#E8E8F4"/>
                        <rect x="12" y="12" width="24" height="28" rx="2" fill="#A5A7E8"/>
                        <rect x="16" y="8" width="24" height="28" rx="2" fill="#5B5FC7"/>
                    </svg>
                    NovelFlow
                </a>
                <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
                    <span class="navbar-toggler-icon"></span>
                </button>
                <div class="collapse navbar-collapse" id="navbarNav">
                    <ul class="navbar-nav me-auto">
                        <li class="nav-item">
                            <a class="nav-link" href="#/">
                                <i class="bi bi-grid-3x3-gap me-1"></i>
                                项目
                            </a>
                        </li>
                    </ul>
                    <div class="d-flex align-items-center gap-2">
                        <div class="d-flex align-items-center gap-2 px-2 py-1" style="background: var(--bg-hover); border-radius: var(--radius-md);">
                            <i class="bi bi-person-circle" style="color: var(--text-secondary);"></i>
                            <span style="color: var(--text-primary); font-weight: 500; font-size: 0.875rem;">
                                ${this.user.username || '用户'}
                            </span>
                        </div>
                        <button class="btn btn-ghost btn-sm" id="logoutBtn" title="退出登录">
                            <i class="bi bi-box-arrow-right"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    bindEvents() {
        const logoutBtn = this.$('#logoutBtn');
        if (logoutBtn) {
            this.addEventListener(logoutBtn, 'click', () => this.handleLogout());
        }
    }

    handleLogout() {
        userActions.logout();
        router.navigate('/login');
    }

    unmount() {
        super.unmount();
        if (this.unsubscribe) {
            this.unsubscribe();
        }
    }
}
