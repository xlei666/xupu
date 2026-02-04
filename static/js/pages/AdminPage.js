
import BaseComponent from '../components/BaseComponent.js';
import { userActions } from '../store.js';
import { showToast } from '../utils.js';

export default class AdminPage extends BaseComponent {
    constructor(container) {
        super(container);
        this.currentTab = 'prompts';
        this.configs = [];
        this.prompts = [];
        this.structures = [];
    }

    render() {
        return `
        <div class="container py-4">
            <h2 class="mb-4">系统管理后台</h2>
            <div class="card shadow-sm">
                <div class="card-header">
                    <ul class="nav nav-tabs card-header-tabs">
                        <li class="nav-item">
                            <a class="nav-link ${this.currentTab === 'prompts' ? 'active' : ''}" href="#" data-tab="prompts">提示词管理</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link ${this.currentTab === 'configs' ? 'active' : ''}" href="#" data-tab="configs">系统配置</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link ${this.currentTab === 'structures' ? 'active' : ''}" href="#" data-tab="structures">叙事结构模板</a>
                        </li>
                    </ul>
                </div>
                <div class="card-body">
                    ${this.renderContent()}
                </div>
            </div>
        </div>

        <!-- Edit Modal -->
        <div class="modal fade" id="editModal" tabindex="-1">
            <div class="modal-dialog modal-lg">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">编辑</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <form id="editForm">
                            <input type="hidden" id="editKey">
                            <div class="mb-3">
                                <label class="form-label">Key/ID</label>
                                <input type="text" class="form-control" id="editKeyDisplay" disabled>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">内容</label>
                                <textarea class="form-control" id="editContent" rows="15" style="font-family: monospace;"></textarea>
                            </div>
                        </form>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                        <button type="button" class="btn btn-primary" id="saveBtn">保存</button>
                    </div>
                </div>
            </div>
        </div>
        `;
    }

    renderContent() {
        if (this.currentTab === 'prompts') return this.renderPrompts();
        if (this.currentTab === 'configs') return this.renderConfigs();
        if (this.currentTab === 'structures') return this.renderStructures();
        return '';
    }

    renderPrompts() {
        return `
            <div class="d-flex justify-content-between mb-3">
                <h5>提示词模板 (${this.prompts.length})</h5>
                <button class="btn btn-success btn-sm" id="syncPromptsBtn">
                    <i class="bi bi-arrow-repeat"></i> 从代码同步默认值
                </button>
            </div>
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead>
                        <tr>
                            <th>ID/Key</th>
                            <th>描述</th>
                            <th>版本</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this.prompts.map(p => `
                        <tr>
                            <td><code>${p.key}</code></td>
                            <td>${p.description || '-'}</td>
                            <td>v${p.version}</td>
                            <td>
                                <button class="btn btn-sm btn-outline-primary edit-prompt-btn" data-key="${p.key}">
                                    编辑
                                </button>
                            </td>
                        </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    renderConfigs() {
        return `
            <div class="alert alert-info">暂无配置项（开发中）</div>
        `;
    }

    renderStructures() {
        return `
            <div class="alert alert-info">叙事结构模板管理（开发中）</div>
        `;
    }

    mount() {
        super.mount();
        this.loadPrompts();

        // 绑定Tab切换
        this.container.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentTab = e.target.dataset.tab;
                this.refresh();
                if (this.currentTab === 'prompts') this.loadPrompts();
            });
        });

        // 绑定同步按钮
        const syncBtn = this.container.querySelector('#syncPromptsBtn');
        if (syncBtn) {
            syncBtn.addEventListener('click', () => this.syncPrompts());
        }

        // 绑定编辑按钮
        this.container.addEventListener('click', (e) => {
            if (e.target.closest('.edit-prompt-btn')) {
                const key = e.target.closest('.edit-prompt-btn').dataset.key;
                this.openEditModal(key);
            }
        });

        // 绑定保存
        const saveBtn = document.querySelector('#saveBtn');
        if (saveBtn) {
            // Remove old listener to prevent duplicates if mount called multiple times
            // A better way is to bind once or use proper component lifecycle cleanup
            const newBtn = saveBtn.cloneNode(true);
            saveBtn.parentNode.replaceChild(newBtn, saveBtn);
            newBtn.addEventListener('click', () => this.saveEdit());
        }
    }

    async loadPrompts() {
        try {
            const token = userActions.getToken();
            const res = await fetch('/api/v1/admin/prompts', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const data = await res.json();
            if (data.success) {
                this.prompts = data.data;
                this.refresh();
            }
        } catch (e) {
            console.error(e);
            showToast('加载提示词失败', 'error');
        }
    }

    async syncPrompts() {
        if (!confirm('确定要从代码中加载默认提示词吗？即使数据库已存在也会被保留，仅会添加缺失项。')) return;

        try {
            const token = userActions.getToken();
            const res = await fetch('/api/v1/admin/sync', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const data = await res.json();
            if (data.success) {
                showToast(`同步成功，新增 ${data.data.synced_count} 条`);
                this.loadPrompts();
            }
        } catch (e) {
            showToast('同步失败', 'error');
        }
    }

    openEditModal(key) {
        const item = this.prompts.find(p => p.key === key);
        if (!item) return;

        const modal = new bootstrap.Modal(document.getElementById('editModal'));
        document.getElementById('editKey').value = key;
        document.getElementById('editKeyDisplay').value = key;
        document.getElementById('editContent').value = item.content;

        modal.show();
    }

    async saveEdit() {
        const key = document.getElementById('editKey').value;
        const content = document.getElementById('editContent').value;

        try {
            const token = userActions.getToken();
            const res = await fetch(`/api/v1/admin/prompts/${key}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ content })
            });
            const data = await res.json();
            if (data.success) {
                showToast('保存成功');
                const modal = bootstrap.Modal.getInstance(document.getElementById('editModal'));
                modal.hide();
                this.loadPrompts();
            } else {
                showToast(data.message || '保存失败', 'error');
            }
        } catch (e) {
            showToast('保存失败', 'error');
        }
    }
}
