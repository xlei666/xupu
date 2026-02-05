
import BaseComponent from '../components/BaseComponent.js';
import { adminAPI } from '../api.js';
import { showToast } from '../utils.js';

export default class AdminPage extends BaseComponent {
    constructor(container) {
        super(container);
        this.currentTab = 'prompts';
        this.prompts = [];
        this.structures = [];
        this.configs = [];

        // Edit state
        this.editType = null;
        this.editKey = null;
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
                            <a class="nav-link ${this.currentTab === 'structures' ? 'active' : ''}" href="#" data-tab="structures">叙事结构模板</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link ${this.currentTab === 'configs' ? 'active' : ''}" href="#" data-tab="configs">系统配置</a>
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
                                <div class="form-text text-muted" id="editHelpText"></div>
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
        if (this.currentTab === 'structures') return this.renderStructures();
        if (this.currentTab === 'configs') return this.renderConfigs();
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
                <table class="table table-hover align-middle">
                    <thead class="table-light">
                        <tr>
                            <th>Key</th>
                            <th>描述</th>
                            <th>版本</th>
                            <th style="width: 100px;">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this.prompts.map(p => `
                        <tr>
                            <td><code>${p.key}</code></td>
                            <td>${p.description || '-'}</td>
                            <td><span class="badge bg-secondary">v${p.version}</span></td>
                            <td>
                                <button class="btn btn-sm btn-outline-primary edit-btn" data-type="prompt" data-key="${p.key}">
                                    编辑
                                </button>
                            </td>
                        </tr>
                        `).join('')}
                        ${this.prompts.length === 0 ? '<tr><td colspan="4" class="text-center text-muted py-4">暂无数据</td></tr>' : ''}
                    </tbody>
                </table>
            </div>
        `;
    }

    renderStructures() {
        return `
            <div class="d-flex justify-content-between mb-3">
                <h5>叙事结构模板 (${this.structures.length})</h5>
                <button class="btn btn-success btn-sm" id="syncStructuresBtn">
                    <i class="bi bi-arrow-repeat"></i> 同步默认结构
                </button>
            </div>
            <div class="table-responsive">
                <table class="table table-hover align-middle">
                    <thead class="table-light">
                        <tr>
                            <th>ID</th>
                            <th>名称</th>
                            <th>描述</th>
                            <th>状态</th>
                            <th style="width: 100px;">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this.structures.map(s => `
                        <tr>
                            <td><code>${s.id}</code></td>
                            <td><strong>${s.name}</strong></td>
                            <td class="text-muted small">${s.description || '-'}</td>
                            <td>
                                <span class="badge ${s.is_active ? 'bg-success' : 'bg-secondary'}">
                                    ${s.is_active ? '启用' : '禁用'}
                                </span>
                            </td>
                            <td>
                                <button class="btn btn-sm btn-outline-primary edit-btn" data-type="structure" data-key="${s.id}">
                                    编辑
                                </button>
                            </td>
                        </tr>
                        `).join('')}
                        ${this.structures.length === 0 ? '<tr><td colspan="5" class="text-center text-muted py-4">暂无数据</td></tr>' : ''}
                    </tbody>
                </table>
            </div>
        `;
    }

    renderConfigs() {
        return `
            <div class="d-flex justify-content-between mb-3">
                <h5>系统配置 (${this.configs.length})</h5>
                <button class="btn btn-success btn-sm" id="syncConfigsBtn">
                    <i class="bi bi-arrow-repeat"></i> 初始化/同步默认配置
                </button>
            </div>
            <div class="table-responsive">
                <table class="table table-hover align-middle">
                    <thead class="table-light">
                        <tr>
                            <th>Key</th>
                            <th>值</th>
                            <th>类型</th>
                            <th>描述</th>
                            <th style="width: 100px;">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${this.configs.map(c => `
                        <tr>
                            <td><code>${c.key}</code></td>
                            <td style="max-width: 300px;" class="text-truncate" title="${c.value}">${c.value}</td>
                            <td><span class="badge bg-light text-dark border">${c.type}</span></td>
                            <td class="text-muted small">${c.description || '-'}</td>
                            <td>
                                <button class="btn btn-sm btn-outline-primary edit-btn" data-type="config" data-key="${c.key}">
                                    编辑
                                </button>
                            </td>
                        </tr>
                        `).join('')}
                        ${this.configs.length === 0 ? '<tr><td colspan="5" class="text-center text-muted py-4">暂无数据</td></tr>' : ''}
                    </tbody>
                </table>
            </div>
        `;
    }

    mount() {
        super.mount();
        this.loadCurrentTab();
    }

    bindEvents() {
        // Bind Tab Switching
        this.container.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentTab = e.target.dataset.tab;
                this.update();
                this.loadCurrentTab();
            });
        });

        // Add Event Delegation for dynamic buttons
        // Use this.container directly as delegation root
        this.addEventListener(this.container, 'click', (e) => {
            const target = e.target;

            // Sync Buttons
            if (target.closest('#syncPromptsBtn')) this.handleSync('prompts');
            if (target.closest('#syncStructuresBtn')) this.handleSync('structures');
            if (target.closest('#syncConfigsBtn')) this.handleSync('configs');

            // Edit Buttons
            const editBtn = target.closest('.edit-btn');
            if (editBtn) {
                const { type, key } = editBtn.dataset;
                this.openEditModal(type, key);
            }
        });

        // Bind Save Button - Helper for modal
        const saveBtn = document.getElementById('saveBtn');
        if (saveBtn) {
            // Note: BaseComponent doesn't track listeners generated inside external modals easily if they are outside container
            // But here modal is part of render(), so it is inside container (likely).
            // However, bootstrap modal might move it to body. 
            // Let's check render(): modal is part of render string.
            // But if bootstrap moves it, we might lose it or duplicate binding.
            // Ideally we bind to document body or use the one inside if it stays.
            // Bootstrap 5 usually keeps it in place unless configured otherwise, or appends to body on show.
            // Safest is to bind click on the specific ID if found.
            // BaseComponent addEventListener wrapper is safer.
            this.addEventListener(saveBtn, 'click', () => this.saveEdit());
        }
    }

    loadCurrentTab() {
        if (this.currentTab === 'prompts') this.loadPrompts();
        if (this.currentTab === 'structures') this.loadStructures();
        if (this.currentTab === 'configs') this.loadConfigs();
    }

    async loadPrompts() {
        try {
            const res = await adminAPI.getPrompts();
            if (res.success) {
                this.prompts = res.data;
                this.update();
            }
        } catch (e) {
            showToast('加载提示词失败: ' + e.message, 'error');
        }
    }

    async loadStructures() {
        try {
            const res = await adminAPI.getStructures();
            if (res.success) {
                this.structures = res.data;
                this.update();
            }
        } catch (e) {
            showToast('加载结构失败: ' + e.message, 'error');
        }
    }

    async loadConfigs() {
        try {
            const res = await adminAPI.getConfigs();
            if (res.success) {
                this.configs = res.data;
                this.update();
            }
        } catch (e) {
            showToast('加载配置失败: ' + e.message, 'error');
        }
    }

    async handleSync(type) {
        if (!confirm('确定要执行同步操作吗？这将从系统代码中恢复缺少的数据项。')) return;

        try {
            let res;
            if (type === 'prompts') res = await adminAPI.syncPrompts();
            if (type === 'structures') res = await adminAPI.syncStructures();
            if (type === 'configs') res = await adminAPI.syncConfigs();

            if (res.success) {
                showToast(`同步成功，变更 ${res.data.synced_count} 项`);
                this.loadCurrentTab();
            }
        } catch (e) {
            showToast('同步失败: ' + e.message, 'error');
        }
    }

    openEditModal(type, key) {
        this.editType = type;
        this.editKey = key;

        const modal = new bootstrap.Modal(document.getElementById('editModal'));
        document.getElementById('editKey').value = key;
        document.getElementById('editKeyDisplay').value = key;

        const contentInput = document.getElementById('editContent');
        const helpText = document.getElementById('editHelpText');

        let content = '';
        if (type === 'prompt') {
            const item = this.prompts.find(p => p.key === key);
            content = item ? item.content : '';
            helpText.innerText = '支持使用 Go Template 语法，例如 {{.Variable}}';
        } else if (type === 'structure') {
            const item = this.structures.find(s => s.id === key);
            content = item ? JSON.stringify(item.structure, null, 2) : '';
            helpText.innerText = '必须是合法的 JSON 格式';
        } else if (type === 'config') {
            const item = this.configs.find(c => c.key === key);
            content = item ? item.value : '';
            helpText.innerText = `类型: ${item ? item.type : 'unknown'}`;
        }

        contentInput.value = content;
        modal.show();
    }

    async saveEdit() {
        const key = this.editKey;
        const content = document.getElementById('editContent').value;

        try {
            let res;
            if (this.editType === 'prompt') {
                res = await adminAPI.updatePrompt(key, content);
            } else if (this.editType === 'structure') {
                try {
                    const jsonContent = JSON.parse(content);
                    res = await adminAPI.updateStructure(key, jsonContent);
                } catch (e) {
                    showToast('无效的 JSON 格式', 'error');
                    return;
                }
            } else if (this.editType === 'config') {
                res = await adminAPI.updateConfig(key, content);
            }

            if (res && res.success) {
                showToast('保存成功');
                const modal = bootstrap.Modal.getInstance(document.getElementById('editModal'));
                modal.hide();
                this.loadCurrentTab();
            } else {
                showToast('保存失败', 'error');
            }
        } catch (e) {
            showToast('保存失败: ' + e.message, 'error');
        }
    }
}
