// 项目详情页面

import BaseComponent from '../components/BaseComponent.js';
import { projectAPI, chapterAPI, worldAPI, characterAPI, blueprintAPI } from '../api.js';
import { projectActions, userActions } from '../store.js';
import { showToast, formatRelativeTime } from '../utils.js';
import router from '../router.js';

export default class ProjectDetailPage extends BaseComponent {
    constructor(container, projectId) {
        super(container);
        this.projectId = projectId;
        this.project = null;
        this.chapters = [];
        this.worldSettings = null;
        this.characters = [];
        this.blueprint = null;
        this.loading = true;

        // State for setting generation
        this.isGenerating = false;
        this.generatedSettingContent = '';
        this.API_BASE = '/api/v1';
    }

    async onMounted() {
        await this.loadProject();
    }

    async loadProject() {
        try {
            this.loading = true;
            this.update();

            const response = await projectAPI.getProject(this.projectId);
            this.project = response?.data?.project || response?.data || null;

            if (this.project) {
                projectActions.setCurrentProject(this.project);

                // Load related data in parallel
                const promises = [
                    chapterAPI.getChapters(this.projectId).then(res => this.chapters = res?.data?.chapters || []),
                    this.project.world_id ? worldAPI.getWorldSettings(this.projectId).then(res => this.worldSettings = res?.data || null) : Promise.resolve(null),
                    characterAPI.getCharacters(this.projectId).then(res => this.characters = res?.data?.characters || []),
                    // If we have narrative_id, try to fetch blueprint. 
                    // Note: API might not expose getBlueprint by project_id easily, but let's try assuming narrative_id is blueprint_id
                    this.project.narrative_id ? blueprintAPI.getBlueprint(this.project.narrative_id).then(res => this.blueprint = res?.data || null).catch(() => null) : Promise.resolve(null)
                ];

                await Promise.allSettled(promises);
            }

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

        const isNewWork = !this.worldSettings && (!this.chapters || this.chapters.length === 0);

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
                        <!-- Removed New Chapter button -->
                    </div>
                </div>
                
                ${isNewWork ? this.renderNewWorkState() : this.renderProjectContent()}
            </div>
            
            ${this.renderSettingModal()}
        `;
    }

    renderNewWorkState() {
        return `
            <div class="empty-state py-5 text-center bg-light rounded-3">
                <i class="bi bi-journal-plus display-1 text-muted"></i>
                <h3 class="mt-3">开始你的创作之旅</h3>
                <p class="text-muted mb-4">该项目尚未初始化，请先生成核心设定</p>
                <button class="btn btn-primary btn-lg px-5" id="startSettingBtn">
                    <i class="bi bi-magic me-2"></i>
                    开始设定
                </button>
            </div>
        `;
    }

    renderProjectContent() {
        return `
            <div class="row">
                <!-- 左侧：章节列表 (保留作为导航) -->
                <div class="col-md-3">
                    <div class="card mb-3">
                        <div class="card-header bg-white">
                            <h5 class="mb-0">章节列表</h5>
                        </div>
                        <div class="card-body p-0" style="max-height: 600px; overflow-y: auto;">
                            ${this.renderChapterList()}
                        </div>
                    </div>
                </div>
                
                <!-- 右侧：主要内容 -->
                <div class="col-md-9">
                    <!-- 核心设定 -->
                    <div class="card mb-4">
                        <div class="card-header bg-white d-flex justify-content-between align-items-center">
                            <h5 class="mb-0">核心设定</h5>
                            <button class="btn btn-sm btn-outline-primary" id="viewFullSettingBtn">查看详情</button>
                        </div>
                        <div class="card-body">
                            ${this.renderCoreSettingPreview()}
                        </div>
                    </div>

                    <!-- 角色列表 -->
                    <div class="card mb-4">
                        <div class="card-header bg-white">
                            <h5 class="mb-0">角色列表</h5>
                        </div>
                        <div class="card-body p-0">
                            ${this.renderCharacterList()}
                        </div>
                    </div>

                    <!-- 叙事蓝图详情 -->
                    ${this.renderBlueprintSection()}

                    <!-- 作品信息 -->
                    <div class="card mb-4">
                        <div class="card-header bg-white">
                            <h5 class="mb-0">作品信息</h5>
                        </div>
                        <div class="card-body">
                            <div class="row">
                                <div class="col-md-6">
                                    <p><strong>章节数量：</strong> ${this.chapters.length} 章</p>
                                    <p><strong>创建时间：</strong> ${formatRelativeTime(this.project.created_at)}</p>
                                </div>
                                <div class="col-md-6">
                                    <p><strong>总字数：</strong> ${this.calculateTotalWords()} 字</p>
                                    <p><strong>更新时间：</strong> ${formatRelativeTime(this.project.updated_at)}</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- 下一步操作 -->
                    <div class="d-grid gap-2 mb-5">
                        <button class="btn btn-primary btn-lg" id="startNarrativeBtn">
                            <i class="bi bi-diagram-3 me-2"></i>
                            开始叙事规划
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    renderChapterList() {
        if (!this.chapters || this.chapters.length === 0) {
            return '<div class="p-3 text-muted text-center">暂无章节</div>';
        }
        return `
            <ul class="list-group list-group-flush">
                ${this.chapters.map(ch => `
                    <li class="list-group-item d-flex justify-content-between align-items-center">
                        <span class="text-truncate">第${ch.chapter_num}章 ${ch.title}</span>
                        <span class="badge bg-secondary rounded-pill">${ch.status === 'completed' ? '完' : '草'}</span>
                    </li>
                `).join('')}
            </ul>
        `;
    }

    renderCoreSettingPreview() {
        if (!this.worldSettings) return '<div class="text-muted">暂无设定</div>';

        // Try to extract some core info
        const philosophy = this.worldSettings.philosophy?.core_question || '未定义';
        const worldview = this.worldSettings.worldview?.cosmology?.structure || '未定义';

        return `
            <div class="row">
                <div class="col-md-12">
                    <p><strong>核心问题：</strong> ${philosophy}</p>
                    <p><strong>世界结构：</strong> ${worldview}</p>
                    <p><strong>世界类型：</strong> ${this.worldSettings.type} (${this.worldSettings.scale})</p>
                </div>
            </div>
        `;
    }

    renderCharacterList() {
        if (!this.characters || this.characters.length === 0) {
            return '<div class="p-3 text-muted text-center">暂无角色</div>';
        }

        // Show top 5 characters
        const displayChars = (Array.isArray(this.characters) ? this.characters : []).slice(0, 5);

        return `
            <table class="table table-hover mb-0">
                <thead class="table-light">
                    <tr>
                        <th>姓名</th>
                        <th>性别</th>
                        <th>种族</th>
                        <th>职业</th>
                    </tr>
                </thead>
                <tbody>
                    ${displayChars.map(c => `
                        <tr>
                            <td>${c.name}</td>
                            <td>${c.static_profile?.gender || '-'}</td>
                            <td>${c.static_profile?.race || '-'}</td>
                            <td>${c.static_profile?.occupation || '-'}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
            ${this.characters.length > 5 ? `<div class="p-2 text-center border-top"><small class="text-muted">查看全部 ${this.characters.length} 个角色</small></div>` : ''}
        `;
    }

    renderBlueprintSection() {
        if (!this.blueprint) {
            return `
                <div class="card mb-4">
                    <div class="card-header bg-white">
                        <h5 class="mb-0">叙事蓝图详情</h5>
                    </div>
                    <div class="card-body text-center text-muted py-4">
                        <p>尚未生成叙事蓝图</p>
                    </div>
                </div>
            `;
        }

        const outline = this.blueprint.story_outline || {};
        const act1 = outline.act1?.setup || '未定义';
        const type = outline.structure_type || '未定义';

        return `
            <div class="card mb-4">
                <div class="card-header bg-white">
                    <h5 class="mb-0">叙事蓝图详情</h5>
                </div>
                <div class="card-body">
                    <p><strong>结构类型：</strong> ${type}</p>
                    <div class="alert alert-light border">
                        <strong>开篇设定：</strong> ${act1}
                    </div>
                    <p class="text-muted"><small>包含 ${this.blueprint.chapter_plans ? this.blueprint.chapter_plans.length : 0} 个章节规划</small></p>
                </div>
            </div>
        `;
    }

    calculateTotalWords() {
        return this.chapters.reduce((sum, ch) => sum + (ch.word_count || 0), 0);
    }

    renderSettingModal() {
        return `
            <div class="modal fade" id="settingModal" tabindex="-1" data-bs-backdrop="static">
                <div class="modal-dialog modal-lg">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">创建核心设定</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal" id="closeSettingModalBtn"></button>
                        </div>
                        <div class="modal-body">
                            <div id="settingFormArea">
                                <div class="mb-3">
                                    <label class="form-label">故事核心（一句话亮点）</label>
                                    <input type="text" class="form-control" id="storyCore" placeholder="例如：穿越者利用现代知识在修仙世界建立科技帝国">
                                </div>
                                <div class="mb-3">
                                    <label class="form-label">小说类型</label>
                                    <input type="text" class="form-control" id="genre" placeholder="例如：玄幻、修仙、都市、科幻">
                                </div>
                                <div class="mb-3">
                                    <label class="form-label">补充说明（可选）</label>
                                    <textarea class="form-control" id="description" rows="3" placeholder="其他你想要的设定细节..."></textarea>
                                </div>
                            </div>
                            
                            <div id="settingResultArea" style="display: none;">
                                <div class="alert alert-info border-0 bg-light">
                                    <div id="settingOutput" style="white-space: pre-wrap; font-family: monospace; max-height: 400px; overflow-y: auto;"></div>
                                </div>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <div id="settingFormActions">
                                <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                                <button type="button" class="btn btn-primary" id="genSettingBtn">
                                    <i class="bi bi-stars me-1"></i>生成设定
                                </button>
                            </div>
                            <div id="settingResultActions" style="display: none;">
                                <button type="button" class="btn btn-outline-secondary" id="regenSettingBtn">重新生成</button>
                                <button type="button" class="btn btn-primary" id="confirmSettingBtn">确认使用</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    bindEvents() {
        // Start Setting Button
        const startBtn = this.$('#startSettingBtn');
        if (startBtn) {
            this.addEventListener(startBtn, 'click', () => this.showSettingModal());
        }

        // Generate Setting Button
        const genBtn = this.$('#genSettingBtn');
        if (genBtn) {
            this.addEventListener(genBtn, 'click', () => this.generateSettings());
        }

        // Regenerate Setting Button
        const regenBtn = this.$('#regenSettingBtn');
        if (regenBtn) {
            this.addEventListener(regenBtn, 'click', () => this.resetSettingForm());
        }

        // Confirm Setting Button
        const confirmBtn = this.$('#confirmSettingBtn');
        if (confirmBtn) {
            this.addEventListener(confirmBtn, 'click', () => this.confirmSettings());
        }

        // Start Narrative Planning Button
        const narrativeBtn = this.$('#startNarrativeBtn');
        if (narrativeBtn) {
            this.addEventListener(narrativeBtn, 'click', () => {
                router.navigate(`/project/${this.projectId}/director`);
            });
        }

        // View Full Setting
        const viewSettingBtn = this.$('#viewFullSettingBtn');
        if (viewSettingBtn) {
            this.addEventListener(viewSettingBtn, 'click', () => {
                router.navigate(`/project/${this.projectId}/director`); // Director page shows full settings
            });
        }
    }

    showSettingModal() {
        const modal = new bootstrap.Modal(this.$('#settingModal'));
        modal.show();
    }

    resetSettingForm() {
        this.$('#settingFormArea').style.display = 'block';
        this.$('#settingResultArea').style.display = 'none';
        this.$('#settingFormActions').style.display = 'block';
        this.$('#settingResultActions').style.display = 'none';
        this.$('#settingOutput').innerHTML = '';
    }

    async generateSettings() {
        const storyCore = this.$('#storyCore').value.trim();
        const genre = this.$('#genre').value.trim();
        const description = this.$('#description').value.trim();

        if (!storyCore || !genre) {
            showToast('请填写故事核心和小说类型', 'warning');
            return;
        }

        // UI Update
        this.$('#settingFormArea').style.display = 'none';
        this.$('#settingResultArea').style.display = 'block';
        this.$('#settingFormActions').style.display = 'none';
        this.$('#settingResultActions').style.display = 'none';

        const outputEl = this.$('#settingOutput');
        // Initialize structure ONCE
        outputEl.innerHTML = `
            <div class="p-4 text-center">
                <div id="genStatus" class="mb-2 fw-bold">正在初始化AI...</div>
                <div class="progress" style="height: 10px;">
                    <div id="genProgress" class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style="width: 0%"></div>
                </div>
            </div>
            <div id="streamingResult" class="mt-3 text-start border rounded bg-white" style="max-height: 500px; overflow-y: auto; font-family: monospace; display: none;"></div>
        `;

        try {
            const token = userActions.getToken();
            const response = await fetch(`${this.API_BASE}/projects/${this.projectId}/world-gacha`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({
                    context: `${storyCore}\n${description}`,
                    settings: {
                        world_type: genre,
                        style: "通用"
                    }
                })
            });

            if (!response.ok) throw new Error(`API Error: ${response.status}`);

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let currentEvent = '';
            let finalData = null;

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('event:')) {
                        currentEvent = line.substring(6).trim();
                    } else if (line.startsWith('data:')) {
                        const dataStr = line.substring(5).trim();
                        if (!dataStr) continue;

                        try {
                            const data = JSON.parse(dataStr);

                            if (currentEvent === 'progress') {
                                const statusEl = document.getElementById('genStatus');
                                const progressEl = document.getElementById('genProgress');
                                if (statusEl) statusEl.textContent = data.message || '生成中...';
                                if (progressEl) progressEl.style.width = (data.percent || 0) + '%';

                            } else if (currentEvent === 'stage_data') {
                                const streamDiv = document.getElementById('streamingResult');
                                if (streamDiv) {
                                    streamDiv.style.display = 'block';
                                    const stageName = data.name || data.stage;
                                    const content = JSON.stringify(data.content, null, 2);

                                    const newEntry = document.createElement('div');
                                    newEntry.className = 'p-3 border-bottom';
                                    newEntry.innerHTML = `
                                        <h6 class="text-primary fw-bold mb-2">✓ ${stageName} 生成完成</h6>
                                        <pre class="bg-light p-2 rounded mb-0" style="font-size: 0.85em; white-space: pre-wrap;">${content}</pre>
                                    `;
                                    streamDiv.appendChild(newEntry);
                                    streamDiv.scrollTop = streamDiv.scrollHeight;
                                }
                            } else if (currentEvent === 'debug_prompt') {
                                const streamDiv = document.getElementById('streamingResult');
                                if (streamDiv) {
                                    streamDiv.style.display = 'block';
                                    const stageName = data.stage;
                                    const prompt = data.prompt;

                                    const newEntry = document.createElement('div');
                                    newEntry.className = 'p-3 border-bottom bg-light';
                                    newEntry.innerHTML = `
                                        <div class="d-flex justify-content-between align-items-center mb-1">
                                            <span class="badge bg-secondary">Prompt</span>
                                            <small class="text-muted">${stageName}</small>
                                        </div>
                                        <div class="text-muted" style="font-size: 0.8em; white-space: pre-wrap; border-left: 3px solid #6c757d; padding-left: 10px;">${prompt}</div>
                                    `;
                                    streamDiv.appendChild(newEntry);
                                    streamDiv.scrollTop = streamDiv.scrollHeight;
                                }
                            } else if (currentEvent === 'result') {
                                finalData = typeof data === 'string' ? JSON.parse(data) : data;
                            } else if (currentEvent === 'error') {
                                throw new Error(data.message);
                            }
                        } catch (e) {
                            console.error("Parse error", e);
                        }
                    }
                }
            }

            if (finalData && finalData.data) {
                const s = finalData.data.stages;
                let text = '';
                if (s.philosophy) text += `【哲学与价值观】\n${JSON.stringify(s.philosophy, null, 2)}\n\n`;
                if (s.worldview) text += `【世界观】\n${JSON.stringify(s.worldview, null, 2)}\n\n`;
                if (s.laws) text += `【法则】\n${JSON.stringify(s.laws, null, 2)}\n\n`;
                if (s.geography) text += `【地理】\n共${s.geography.regions ? s.geography.regions.length : 0}个区域\n\n`;

                outputEl.innerHTML = `<pre>${text}</pre>`;
                this.$('#settingResultActions').style.display = 'block';
            } else {
                outputEl.innerHTML += '<div class="text-warning mt-3">生成结束，但未收到完整结果，请尝试刷新页面。</div>';
                this.$('#settingResultActions').style.display = 'block';
            }

        } catch (err) {
            outputEl.innerHTML = `<div class="text-danger p-3">生成失败: ${err.message}</div>`;
            this.$('#settingResultActions').style.display = 'block';
        }
    }

    async confirmSettings() {
        const modal = bootstrap.Modal.getInstance(this.$('#settingModal'));
        modal.hide();

        showToast('核心设定已保存', 'success');

        // Reload project to get the new world_id and render content
        await this.loadProject();
    }
}
