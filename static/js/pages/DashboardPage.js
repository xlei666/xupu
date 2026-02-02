// 项目列表页面（工作台首页）

import BaseComponent from '../components/BaseComponent.js';
import { projectAPI } from '../api.js';
import { projectActions } from '../store.js';
import { showToast, formatRelativeTime, confirm } from '../utils.js';
import router from '../router.js';

export default class DashboardPage extends BaseComponent {
    constructor(container) {
        super(container);
        this.projects = [];
        this.loading = true;
    }

    render() {
        if (this.loading) {
            return `
                <div class="container mt-4">
                    <div class="d-flex justify-content-center align-items-center" style="min-height: 60vh;">
                        <div class="spinner-border text-primary" role="status">
                            <span class="visually-hidden">加载中...</span>
                        </div>
                    </div>
                </div>
            `;
        }

        return `
            <div class="container mt-4">
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <h2 class="mb-0">
                        <i class="bi bi-grid me-2"></i>
                        我的项目
                    </h2>
                    <button class="btn btn-primary" id="createProjectBtn">
                        <i class="bi bi-plus-lg me-1"></i>
                        创建项目
                    </button>
                </div>
                
                ${this.renderProjects()}
            </div>
            
            ${this.renderCreateModal()}
        `;
    }

    renderProjects() {
        if (this.projects.length === 0) {
            return `
                <div class="empty-state">
                    <i class="bi bi-folder-x"></i>
                    <h3>还没有项目</h3>
                    <p>点击"创建项目"开始您的创作之旅</p>
                </div>
            `;
        }

        return `
            <div class="row g-4">
                ${this.projects.map(project => this.renderProjectCard(project)).join('')}
            </div>
        `;
    }

    renderProjectCard(project) {
        return `
            <div class="col-md-6 col-lg-4">
                <div class="card project-card" data-id="${project.id}">
                    <div class="card-body">
                        <h5 class="card-title">${project.name || '未命名项目'}</h5>
                        <p class="card-text">${project.description || '暂无描述'}</p>
                    </div>
                    <div class="card-footer d-flex justify-content-between align-items-center">
                        <small class="text-muted">
                            <i class="bi bi-clock me-1"></i>
                            ${formatRelativeTime(project.updated_at || project.created_at)}
                        </small>
                        <div>
                            <button class="btn btn-sm btn-outline-danger delete-btn" data-id="${project.id}">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    renderCreateModal() {
        return `
            <div class="modal fade" id="createProjectModal" tabindex="-1">
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">创建新项目</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <form id="createProjectForm">
                            <div class="modal-body">
                                <div class="mb-3">
                                    <label for="projectName" class="form-label">项目名称</label>
                                    <input type="text" class="form-control" id="projectName" name="name" required>
                                </div>
                                <div class="mb-3">
                                    <label for="projectDescription" class="form-label">项目描述</label>
                                    <textarea class="form-control" id="projectDescription" name="description" rows="3"></textarea>
                                </div>
                            </div>
                            <div class="modal-footer">
                                <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                                <button type="submit" class="btn btn-primary">创建</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
    }

    async onMounted() {
        await this.loadProjects();
    }

    bindEvents() {
        // 创建项目按钮
        const createBtn = this.$('#createProjectBtn');
        if (createBtn) {
            this.addEventListener(createBtn, 'click', () => this.showCreateModal());
        }

        // 创建项目表单
        const form = this.$('#createProjectForm');
        if (form) {
            this.addEventListener(form, 'submit', (e) => this.handleCreateProject(e));
        }

        // 项目卡片点击
        this.$$('.project-card').forEach(card => {
            this.addEventListener(card, 'click', (e) => {
                if (!e.target.closest('.delete-btn')) {
                    const id = card.dataset.id;
                    router.navigate(`/project/${id}`);
                }
            });
        });

        // 删除按钮
        this.$$('.delete-btn').forEach(btn => {
            this.addEventListener(btn, 'click', async (e) => {
                e.stopPropagation();
                const id = btn.dataset.id;
                await this.handleDeleteProject(id);
            });
        });
    }

    async loadProjects() {
        try {
            this.loading = true;
            this.update();

            const projects = await projectAPI.getProjects();
            this.projects = projects || [];
            projectActions.setProjects(this.projects);

            this.loading = false;
            this.update();
        } catch (error) {
            console.error('加载项目失败:', error);
            showToast('加载项目失败', 'error');
            this.loading = false;
            this.update();
        }
    }

    showCreateModal() {
        const modal = new bootstrap.Modal(this.$('#createProjectModal'));
        modal.show();
    }

    async handleCreateProject(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const projectData = {
            name: formData.get('name'),
            description: formData.get('description')
        };

        try {
            const submitBtn = this.$('#createProjectForm button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>创建中...';

            const project = await projectAPI.createProject(projectData);

            projectActions.addProject(project);
            this.projects.push(project);

            showToast('项目创建成功！', 'success');

            // 关闭模态框
            const modal = bootstrap.Modal.getInstance(this.$('#createProjectModal'));
            modal.hide();

            // 重置表单
            e.target.reset();

            // 刷新列表
            this.update();

            // 跳转到项目详情
            router.navigate(`/project/${project.id}`);
        } catch (error) {
            console.error('创建项目失败:', error);
            showToast(error.message || '创建项目失败', 'error');
        } finally {
            const submitBtn = this.$('#createProjectForm button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '创建';
            }
        }
    }

    async handleDeleteProject(id) {
        const confirmed = await confirm('确定要删除这个项目吗？此操作不可恢复。', '删除项目');

        if (!confirmed) return;

        try {
            await projectAPI.deleteProject(id);

            projectActions.deleteProject(id);
            this.projects = this.projects.filter(p => p.id !== id);

            showToast('项目已删除', 'success');
            this.update();
        } catch (error) {
            console.error('删除项目失败:', error);
            showToast(error.message || '删除项目失败', 'error');
        }
    }
}
