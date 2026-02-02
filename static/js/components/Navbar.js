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
            <div class="container-fluid">
                <a class="navbar-brand" href="#/">
                    <i class="bi bi-book me-2"></i>
                    NovelFlow 叙谱
                </a>
                <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
                    <span class="navbar-toggler-icon"></span>
                </button>
                <div class="collapse navbar-collapse" id="navbarNav">
                    <ul class="navbar-nav me-auto">
                        <li class="nav-item">
                            <a class="nav-link" href="#/">
                                <i class="bi bi-grid me-1"></i>
                                我的项目
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#/settings">
                                <i class="bi bi-gear me-1"></i>
                                设置
                            </a>
                        </li>
                    </ul>
                    <div class="d-flex align-items-center">
                        <span class="text-white me-3">
                            <i class="bi bi-person-circle me-1"></i>
                            ${this.user.username || '用户'}
                        </span>
                        <button class="btn btn-outline-light btn-sm" id="logoutBtn">
                            <i class="bi bi-box-arrow-right me-1"></i>
                            退出
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
