// 项目详情页面

import BaseComponent from '../components/BaseComponent.js';
import { projectAPI, chapterAPI } from '../api.js';
import { projectActions } from '../store.js';
import { showToast, formatRelativeTime } from '../utils.js';
import router from '../router.js';

export default class ProjectDetailPage extends BaseComponent {
    constructor(container, projectId) {
        super(container);
        this.projectId = projectId;
        this.project = null;
        this.chapters = [];
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

        if (!this.project) {
            return `
                <div class="container mt-4">
                    <div class="alert alert-danger">
                        <h4>项目未找到</h4>
                        <p>该项目不存在或已被删除</p>
                        <a href="#/" class="btn btn-primary">返回项目列表</a>
                    </div>
                </div>
            `;
        }

        return `
            <div class="container mt-4">
                <!-- 项目头部 -->
                <div class="d-flex justify-content-between align-items-center mb-4">
                    <div>
                        <nav aria-label="breadcrumb">
                            <ol class="breadcrumb mb-2">
                                <li class="breadcrumb-item"><a href="#/">项目</a></li>
                                <li class="breadcrumb-item active">${this.project.name || '未命名项目'}</li>
                            </ol>
                        </nav>
                        <h2 class="mb-0">${this.project.name || '未命名项目'}</h2>
                        <p class="text-muted mb-0">${this.project.description || '暂无描述'}</p>
                    </div>
                    <div>
                        <button class="btn btn-outline-secondary me-2" id="settingsBtn">
                            <i class="bi bi-gear me-1"></i>
                            设置
                        </button>
                        <button class="btn btn-primary" id="createChapterBtn">
                            <i class="bi bi-plus-lg me-1"></i>
                            新建章节
                        </button>
                    </div>
                </div>
                
                <!-- 项目内容 -->
                <div class="row">
                    <!-- 左侧：章节列表 -->
                    <div class="col-md-4">
                        <div class="card">
                            <div class="card-header bg-white">
                                <h5 class="mb-0">
                                    <i class="bi bi-list-ol me-2"></i>
                                    章节列表
                                </h5>
                            </div>
                            <div class="card-body p-0">
                                ${this.renderChapterList()}
                            </div>
                        </div>
                    </div>
                    
                    <!-- 右侧：项目信息 -->
                    <div class="col-md-8">
                        <div class="card">
                            <div class="card-header bg-white">
                                <h5 class="mb-0">
                                    <i class="bi bi-info-circle me-2"></i>
                                    项目信息
                                </h5>
                            </div>
                            <div class="card-body">
                                <div class="row mb-3">
                                    <div class="col-sm-3 text-muted">创建时间</div>
                                    <div class="col-sm-9">${formatRelativeTime(this.project.created_at)}</div>
                                </div>
                                <div class="row mb-3">
                                    <div class="col-sm-3 text-muted">更新时间</div>
                                    <div class="col-sm-9">${formatRelativeTime(this.project.updated_at)}</div>
                                </div>
                                <div class="row mb-3">
                                    <div class="col-sm-3 text-muted">章节数量</div>
                                    <div class="col-sm-9">${this.chapters.length} 章</div>
                                </div>
                                <div class="row">
                                    <div class="col-sm-3 text-muted">项目ID</div>
                                    <div class="col-sm-9"><code>${this.project.id}</code></div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 快速操作 -->
                        <div class="card mt-3">
                            <div class="card-header bg-white">
                                <h5 class="mb-0">
                                    <i class="bi bi-lightning me-2"></i>
                                    快速操作
                                </h5>
                            </div>
                            <div class="card-body">
                                <div class="d-grid gap-2">
                                    <button class="btn btn-outline-primary" id="worldSettingsBtn">
                                        <i class="bi bi-globe me-2"></i>
                                        世界设定
                                    </button>
                                    <button class="btn btn-outline-primary" id="charactersBtn">
                                        <i class="bi bi-people me-2"></i>
                                        角色管理
                                    </button>
                                    <button class="btn btn-outline-primary" id="narrativeBtn">
                                        <i class="bi bi-diagram-3 me-2"></i>
                                        叙事规划
                                    </button>
                                    <button class="btn btn-outline-primary" id="writerBtn">
                                        <i class="bi bi-pencil me-2"></i>
                                        AI写作
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            ${this.renderCreateChapterModal()}
        `;
    }

    renderChapterList() {
        if (this.chapters.length === 0) {
            return `
                <div class="empty-state py-4">
                    <i class="bi bi-file-earmark-text"></i>
                    <p class="mb-0">还没有章节</p>
                </div>
            `;
        }

        return `
            <ul class="chapter-list">
                ${this.chapters.map((chapter, index) => `
                    <li class="chapter-item" data-id="${chapter.id}">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <strong>第${index + 1}章</strong>
                                <span class="ms-2">${chapter.title || '未命名章节'}</span>
                            </div>
                            <button class="btn btn-sm btn-outline-danger delete-chapter-btn" data-id="${chapter.id}">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    </li>
                `).join('')}
            </ul>
        `;
    }

    renderCreateChapterModal() {
        return `
            <div class="modal fade" id="createChapterModal" tabindex="-1">
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">新建章节</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <form id="createChapterForm">
                            <div class="modal-body">
                                <div class="mb-3">
                                    <label for="chapterTitle" class="form-label">章节标题</label>
                                    <input type="text" class="form-control" id="chapterTitle" name="title" required>
                                </div>
                                <div class="mb-3">
                                    <label for="chapterContent" class="form-label">章节内容</label>
                                    <textarea class="form-control" id="chapterContent" name="content" rows="5"></textarea>
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
        await this.loadProject();
        await this.loadChapters();
    }

    bindEvents() {
        // 创建章节按钮
        const createBtn = this.$('#createChapterBtn');
        if (createBtn) {
            this.addEventListener(createBtn, 'click', () => this.showCreateChapterModal());
        }

        // 创建章节表单
        const form = this.$('#createChapterForm');
        if (form) {
            this.addEventListener(form, 'submit', (e) => this.handleCreateChapter(e));
        }

        // 快速操作按钮
        const buttons = {
            worldSettingsBtn: () => router.navigate(`/project/${this.projectId}/director`),
            charactersBtn: () => showToast('角色管理功能开发中', 'info'),
            narrativeBtn: () => router.navigate(`/project/${this.projectId}/director`),
            writerBtn: () => router.navigate(`/project/${this.projectId}/director`),
            settingsBtn: () => showToast('项目设置功能开发中', 'info')
        };

        Object.entries(buttons).forEach(([id, handler]) => {
            const btn = this.$(`#${id}`);
            if (btn) {
                this.addEventListener(btn, 'click', handler);
            }
        });
    }

    async loadProject() {
        try {
            this.loading = true;
            this.update();

            const response = await projectAPI.getProject(this.projectId);
            this.project = response?.data?.project || response?.data || null;
            projectActions.setCurrentProject(this.project);

            this.loading = false;
            this.update();
        } catch (error) {
            console.error('加载项目失败:', error);
            showToast('加载项目失败', 'error');
            this.project = null;
            this.loading = false;
            this.update();
        }
    }

    async loadChapters() {
        try {
            const response = await chapterAPI.getChapters(this.projectId);
            this.chapters = response?.data?.chapters || [];
            this.update();
        } catch (error) {
            console.error('加载章节失败:', error);
            this.chapters = [];
        }
    }

    showCreateChapterModal() {
        const modal = new bootstrap.Modal(this.$('#createChapterModal'));
        modal.show();
    }

    async handleCreateChapter(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const chapterData = {
            title: formData.get('title'),
            content: formData.get('content')
        };

        try {
            const submitBtn = this.$('#createChapterForm button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>创建中...';

            const response = await chapterAPI.createChapter(this.projectId, chapterData);
            const chapter = response?.data?.chapter || response?.data || response;

            this.chapters.push(chapter);
            showToast('章节创建成功！', 'success');

            // 关闭模态框
            const modal = bootstrap.Modal.getInstance(this.$('#createChapterModal'));
            modal.hide();

            // 重置表单
            e.target.reset();

            // 刷新列表
            this.update();
        } catch (error) {
            console.error('创建章节失败:', error);
            showToast(error.message || '创建章节失败', 'error');
        } finally {
            const submitBtn = this.$('#createChapterForm button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.innerHTML = '创建';
            }
        }
    }
}
