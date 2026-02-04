
import BaseComponent from '../components/BaseComponent.js';
import { userActions } from '../store.js';
import router from '../router.js';
import { showToast } from '../utils.js';

export default class DirectorPage extends BaseComponent {
    constructor(container, projectId) {
        super(container);
        this.projectId = projectId;
        // API Proxy in frontend usually handles /api -> backend:8080/api/v1 or similar
        // But here we construct full path or relative path
        this.API_BASE = '/api/v1';

        // State
        this.currentSettingId = null;
        this.currentDaGangId = null;
        this.currentXiGangId = null;
        this.chapters = [];
        this.generatedSettingContent = '';
        this.generatedDagangContent = '';
        this.isGenerating = false;
        this.currentTab = 'setting';
    }

    render() {
        return `
        <div class="director-mode">
            <div class="top-bar">
                <span class="logo">Novelflow</span>
                <span class="project-name" id="projectName">Loading...</span>
                <div class="top-spacer"></div>
                <span class="mode-badge">导演模式</span>
                <button class="icon-btn close-btn ms-3" id="exitBtn" title="退出导演模式">
                    <i class="bi bi-x-lg"></i>
                </button>
            </div>

            <div class="main-layout">
                <!-- 左侧：章节列表 + 全书大纲 -->
                <div class="left-sidebar">
                    <div class="sidebar-header">
                        <span>章节</span>
                    </div>
                    <div class="sidebar-content" style="flex: 1;" id="chapterList">
                        <div class="empty-state" style="height: 100px;">
                            <div>暂无章节</div>
                        </div>
                    </div>
                    <div style="height: 280px; border-top: 1px solid var(--border); display: flex; flex-direction: column;">
                        <div class="sidebar-header">
                            <span>全书大纲</span>
                        </div>
                        <div class="sidebar-content" style="flex: 1; font-size: 12px;" id="dagangPreview">
                            <div style="color: var(--text-muted);">暂无大纲</div>
                        </div>
                    </div>
                </div>

                <!-- 中央：编辑区 -->
                <div class="center-area">
                    <!-- Tab栏 -->
                    <div class="editor-tabs">
                        <div class="editor-tab active" data-tab="setting" id="tab-btn-setting">核心设定</div>
                        <div class="editor-tab disabled" data-tab="dagang" id="tab-btn-dagang">全书大纲</div>
                        <div class="editor-tab disabled" data-tab="xigang" id="tab-btn-xigang">章节细纲</div>
                        <div class="editor-tab disabled" data-tab="content" id="tab-btn-content">正文</div>
                    </div>

                    <!-- 编辑内容区 -->
                    <div class="editor-area-container">
                        <!-- 核心设定Tab -->
                        <div class="editor-content active" id="tab-setting">
                            <div class="form-container">
                                <div class="form-title">创建核心设定</div>
                                <div class="form-subtitle">定义你的故事核心，AI将基于此生成完整设定</div>

                                <div id="settingFormArea">
                                    <div class="form-group">
                                        <label class="form-label">故事核心（一句话亮点）</label>
                                        <input type="text" class="form-input" id="storyCore" placeholder="例如：穿越者利用现代知识在修仙世界建立科技帝国">
                                    </div>
                                    <div class="form-group">
                                        <label class="form-label">小说类型</label>
                                        <input type="text" class="form-input" id="genre" placeholder="例如：玄幻、修仙、都市、科幻，可组合">
                                    </div>
                                    <div class="form-group">
                                        <label class="form-label">补充说明（可选）</label>
                                        <textarea class="form-input form-textarea" id="description" placeholder="其他你想要的设定细节..."></textarea>
                                    </div>
                                    <div class="btn-group">
                                        <button class="btn btn-primary" id="genSettingBtn">生成设定</button>
                                    </div>
                                </div>

                                <div id="settingResultArea" style="display: none;">
                                    <div class="form-label">生成结果</div>
                                    <div class="stream-output" id="settingOutput"></div>

                                    <div class="feedback-section">
                                        <div class="form-label">不满意？输入反馈让AI修改</div>
                                        <textarea class="form-input" id="feedbackInput" placeholder="例如：主角名字换一个，金手指要更强..." style="min-height: 60px;"></textarea>
                                        <div class="btn-group">
                                            <button class="btn" id="regenFeedbackBtn" disabled>根据反馈修改</button>
                                            <button class="btn" id="regenSettingBtn">重新抽卡</button>
                                            <div style="flex:1"></div>
                                            <button class="btn btn-primary" id="confirmSettingBtn">确认使用</button>
                                        </div>
                                    </div>
                                </div>

                                <!-- 已有设定时显示 -->
                                <div id="settingViewArea" style="display: none;">
                                    <div class="form-label">当前设定</div>
                                    <div class="stream-output" id="settingView" style="max-height: 400px;"></div>
                                    <div class="btn-group">
                                        <button class="btn" id="editSettingBtn">修改设定</button>
                                        <button class="btn" id="recreateSettingBtn">重新生成</button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 全书大纲Tab -->
                        <div class="editor-content" id="tab-dagang">
                            <div class="form-container">
                                <div class="form-title">全书大纲</div>
                                <div class="form-subtitle">基于核心设定生成整本书的大纲框架</div>

                                <div id="dagangFormArea">
                                    <div class="btn-group">
                                        <button class="btn btn-primary" id="genDaGangBtn">生成大纲</button>
                                    </div>
                                </div>

                                <div id="dagangResultArea" style="display: none;">
                                    <div class="stream-output" id="dagangOutput" style="max-height: 400px;"></div>
                                    <div class="btn-group">
                                        <button class="btn" id="regenDaGangBtn">重新生成</button>
                                        <div style="flex:1"></div>
                                        <button class="btn btn-primary" id="confirmDaGangBtn">确认使用</button>
                                    </div>
                                </div>

                                <div id="dagangViewArea" style="display: none;">
                                    <textarea class="form-input form-textarea" id="dagangEdit" style="min-height: 400px;"></textarea>
                                    <div class="btn-group">
                                        <button class="btn" id="recreateDaGangBtn">重新生成</button>
                                        <button class="btn btn-primary" id="saveDaGangEditBtn">保存修改</button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- 章节细纲Tab -->
                        <div class="editor-content" id="tab-xigang">
                            <div class="form-container">
                                <div class="form-title">章节细纲</div>
                                <div class="form-subtitle">为每个章节生成详细的写作指引</div>

                                <div id="xigangFormArea">
                                    <div class="btn-group">
                                        <button class="btn btn-primary" id="genXiGangBtn">生成细纲</button>
                                    </div>
                                </div>

                                <div id="xigangResultArea" style="display: none;">
                                    <div id="xigangList"></div>
                                </div>
                            </div>
                        </div>

                        <!-- 正文Tab -->
                        <div class="editor-content" id="tab-content">
                            <div id="contentFormArea">
                                <div class="content-header">
                                    <h1 class="content-title" id="contentTitle">选择章节</h1>
                                    <div class="content-meta" id="contentMeta">从左侧选择章节开始写作</div>
                                </div>
                                <div id="contentBody" style="font-family: 'Georgia', 'Noto Serif SC', serif; font-size: 17px; line-height: 2; white-space: pre-wrap;">
                                    <div class="empty-state">
                                        <div>请从左侧选择章节</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- 右侧：上下文面板 -->
                <div class="right-sidebar">
                    <div class="context-section">
                        <div class="context-header" data-toggle="collapse">
                            <span>状态追踪</span>
                            <span class="context-toggle">v</span>
                        </div>
                        <div class="context-content" id="statusContext">
                            <div class="context-item">
                                <div class="context-item-label">核心设定</div>
                                <div class="context-item-value" id="statusSetting">未完成</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">全书大纲</div>
                                <div class="context-item-value" id="statusDagang">未完成</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">章节细纲</div>
                                <div class="context-item-value" id="statusXigang">0 章</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">正文进度</div>
                                <div class="context-item-value" id="statusContent">0 / 0 章</div>
                            </div>
                        </div>
                    </div>

                    <div class="context-section">
                        <div class="context-header" data-toggle="collapse">
                            <span>设定追踪</span>
                            <span class="context-toggle">v</span>
                        </div>
                        <div class="context-content" id="settingContext">
                            <div style="color: var(--text-muted);">暂无设定</div>
                        </div>
                    </div>

                    <div class="context-section">
                        <div class="context-header" data-toggle="collapse">
                            <span>当前章节</span>
                            <span class="context-toggle">v</span>
                        </div>
                        <div class="context-content" id="chapterContext">
                            <div style="color: var(--text-muted);">未选择章节</div>
                        </div>
                    </div>

                    <div class="context-section" style="flex: 1;">
                        <div class="context-header" data-toggle="collapse">
                            <span>操作日志</span>
                            <span class="context-toggle">v</span>
                        </div>
                        <div class="context-content" id="logContent" style="max-height: none;">
                            <div style="color: var(--text-muted);">等待操作...</div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 底部操作面板 -->
            <div class="bottom-panel" id="bottomPanel" style="height: 220px;">
                <div class="panel-header" id="bottomPanelHeader">
                    <div class="panel-title">
                        <span>操作面板</span>
                        <span style="font-size: 11px; color: var(--text-muted);">| 细纲 / 四条线 / 快捷操作</span>
                    </div>
                    <span class="panel-toggle">v</span>
                </div>
                <div class="panel-content">
                    <div class="panel-section">
                        <div class="panel-section-header">
                            <span>本章细纲</span>
                            <button class="icon-btn" id="regenCurrentXiGangBtn" title="重新生成">↻</button>
                        </div>
                        <div class="panel-section-content" id="currentXigang">
                            <div style="color: var(--text-muted);">选择章节查看细纲</div>
                        </div>
                    </div>

                    <div class="panel-section">
                        <div class="panel-section-header">四条叙事线</div>
                        <div class="panel-section-content" id="fourLines">
                            <div class="context-item">
                                <div class="context-item-label">情感线</div>
                                <div class="context-item-value" id="emotionLine">-</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">快感线</div>
                                <div class="context-item-value" id="pleasureLine">-</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">伏笔线</div>
                                <div class="context-item-value" id="foreshadowLine">-</div>
                            </div>
                            <div class="context-item">
                                <div class="context-item-label">价值线</div>
                                <div class="context-item-value" id="valueLine">-</div>
                            </div>
                        </div>
                    </div>

                    <div class="panel-section">
                        <div class="panel-section-header">快捷操作</div>
                        <div class="panel-section-content">
                            <button class="btn" style="width: 100%; margin-bottom: 8px;" id="writeBtn" disabled>开始写作本章</button>
                            <button class="btn" style="width: 100%; margin-bottom: 8px;" id="continueBtn" disabled>继续写作</button>
                            <!-- <button class="btn btn-sm" style="width: 100%;" id="showDetailBtn">查看完整细纲</button> -->
                        </div>
                    </div>
                </div>
            </div>
        </div>
        `;
    }

    async onMounted() {
        this.$('#projectName').textContent = `Project: ${this.projectId}`;
        this.log('系统就绪');
    }

    bindEvents() {
        // Tab switching
        ['setting', 'dagang', 'xigang', 'content'].forEach(tab => {
            const btn = this.$(`#tab-btn-${tab}`);
            if (btn) {
                this.addEventListener(btn, 'click', () => this.switchTab(tab));
            }
        });

        // Exit
        const exitBtn = this.$('#exitBtn');
        if (exitBtn) {
            this.addEventListener(exitBtn, 'click', () => {
                router.navigate(`/project/${this.projectId}`);
            });
        }

        // Toggles
        this.container.querySelectorAll('.context-header[data-toggle="collapse"]').forEach(header => {
            this.addEventListener(header, 'click', () => {
                header.parentElement.classList.toggle('collapsed');
            });
        });

        const bottomHeader = this.$('#bottomPanelHeader');
        if (bottomHeader) {
            this.addEventListener(bottomHeader, 'click', () => {
                this.$('#bottomPanel').classList.toggle('collapsed');
            });
        }

        // Setting Actions
        this.addEventListener(this.$('#genSettingBtn'), 'click', () => this.generateSettingPreview());
        this.addEventListener(this.$('#regenFeedbackBtn'), 'click', () => this.regenerateWithFeedback());
        this.addEventListener(this.$('#regenSettingBtn'), 'click', () => this.regenerateSetting());
        this.addEventListener(this.$('#confirmSettingBtn'), 'click', () => this.confirmSetting());
        this.addEventListener(this.$('#editSettingBtn'), 'click', () => this.editSetting());
        this.addEventListener(this.$('#recreateSettingBtn'), 'click', () => this.regenerateSetting());

        // DaGang Actions
        this.addEventListener(this.$('#genDaGangBtn'), 'click', () => this.generateDaGang());
        this.addEventListener(this.$('#regenDaGangBtn'), 'click', () => this.regenerateDaGang());
        this.addEventListener(this.$('#confirmDaGangBtn'), 'click', () => this.confirmDaGang());
        this.addEventListener(this.$('#recreateDaGangBtn'), 'click', () => this.regenerateDaGang());
        this.addEventListener(this.$('#saveDaGangEditBtn'), 'click', () => this.saveDaGangEdit());

        // XiGang Actions
        this.addEventListener(this.$('#genXiGangBtn'), 'click', () => this.generateXiGang());

        // Content Actions
        this.addEventListener(this.$('#writeBtn'), 'click', () => this.generateContent());
        this.addEventListener(this.$('#continueBtn'), 'click', () => this.continueContent());
    }

    // --- Helpers ---

    log(msg) {
        const logEl = this.$('#logContent');
        if (!logEl) return;
        const time = new Date().toLocaleTimeString().slice(0, 5);
        logEl.innerHTML = `<div>[${time}] ${msg}</div>` + logEl.innerHTML;
    }

    switchTab(tab) {
        const tabEl = this.$(`#tab-btn-${tab}`);
        if (!tabEl || tabEl.classList.contains('disabled')) return;

        this.container.querySelectorAll('.editor-tab').forEach(t => t.classList.remove('active'));
        tabEl.classList.add('active');

        this.container.querySelectorAll('.editor-content').forEach(c => c.classList.remove('active'));
        this.$(`#tab-${tab}`).classList.add('active');

        this.currentTab = tab;
    }

    enableTab(tab) {
        this.$(`#tab-btn-${tab}`).classList.remove('disabled');
    }

    updateStatus() {
        this.$('#statusSetting').textContent = this.currentSettingId ? '已完成' : '未完成';
        this.$('#statusDagang').textContent = this.currentDaGangId ? '已完成' : '未完成';
        this.$('#statusXigang').textContent = `${this.chapters.length} 章`;

        const completedChapters = this.chapters.filter(c => c.content).length;
        this.$('#statusContent').textContent = `${completedChapters} / ${this.chapters.length} 章`;
    }

    async fetchWithAuth(url, options = {}) {
        const token = userActions.getToken();
        const headers = {
            'Content-Type': 'application/json',
            ...(options.headers || {}),
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        if (options.body && typeof options.body === 'object') {
            options.body = JSON.stringify(options.body);
        }

        const finalUrl = this.API_BASE + url;

        const response = await fetch(finalUrl, {
            ...options,
            headers
        });

        if (response.status === 401) {
            router.navigate('/login');
            throw new Error('未授权，请登录');
        }

        return response;
    }

    // --- Logic Implementation ---

    async generateSettingPreview() {
        const storyCore = this.$('#storyCore').value.trim();
        const genre = this.$('#genre').value.trim();
        const description = this.$('#description').value.trim();

        if (!storyCore || !genre) {
            alert('请填写故事核心和小说类型');
            return;
        }

        this.isGenerating = true;
        this.$('#settingFormArea').style.display = 'none';
        this.$('#settingResultArea').style.display = 'block';
        this.$('#settingViewArea').style.display = 'none';

        const outputEl = this.$('#settingOutput');
        outputEl.innerHTML = '<span class="stream-cursor"></span>';
        outputEl.classList.add('generating');
        this.generatedSettingContent = '';

        this.log('正在调用AI生成世界设定...');

        try {
            // Using POST /api/v1/projects/:id/world-stages/gacha
            const response = await this.fetchWithAuth(`/projects/${this.projectId}/world-stages/gacha`, {
                method: 'POST',
                body: {
                    context: `${storyCore}\n${description}`,
                    settings: {
                        world_type: genre,
                        style: "通用"
                    }
                }
            });

            if (!response.ok) throw new Error(`API Error: ${response.status}`);

            const data = await response.json();

            if (data.data && data.data.stages) {
                const s = data.data.stages;
                let text = '';
                if (s.philosophy) text += `【哲学与价值观】\n${JSON.stringify(s.philosophy, null, 2)}\n\n`;
                if (s.worldview) text += `【世界观】\n${JSON.stringify(s.worldview, null, 2)}\n\n`;
                if (s.laws) text += `【法则】\n${JSON.stringify(s.laws, null, 2)}`;

                this.generatedSettingContent = text;
                this.currentSettingId = data.data.world_id; // Capture WorldID
                outputEl.innerHTML = `<pre>${text}</pre>`;
            } else {
                this.generatedSettingContent = "生成完成，但返回格式未知。请查看控制台。";
                outputEl.innerHTML = this.generatedSettingContent;
                console.log(data);
            }

            outputEl.classList.remove('generating');
            this.log('世界设定生成完成');

        } catch (err) {
            this.log('错误: ' + err.message);
            outputEl.innerHTML = '<span style="color: var(--danger);">生成失败: ' + err.message + '</span>';
        } finally {
            this.isGenerating = false;
        }
    }

    regenerateSetting() {
        this.$('#settingResultArea').style.display = 'none';
        this.$('#settingFormArea').style.display = 'block';
        this.$('#settingViewArea').style.display = 'none';
    }

    regenerateWithFeedback() {
        alert("反馈修改功能暂未连接到后端");
    }

    async confirmSetting() {
        this.log('设定已确认');

        const storyCore = this.$('#storyCore').value.trim();
        const genre = this.$('#genre').value.trim();

        this.$('#settingResultArea').style.display = 'none';
        this.$('#settingViewArea').style.display = 'block';
        this.$('#settingView').innerHTML = `<pre>${this.generatedSettingContent}</pre>`;

        this.$('#statusSetting').textContent = '已完成';

        this.$('#settingContext').innerHTML = `
            <div class="context-item">
                <div class="context-item-label">故事核</div>
                <div class="context-item-value">${storyCore}</div>
            </div>
            <div class="context-item">
                <div class="context-item-label">类型</div>
                <div class="context-item-value">${genre}</div>
            </div>
        `;

        this.enableTab('dagang');
        this.updateStatus();
    }

    editSetting() {
        this.$('#settingViewArea').style.display = 'none';
        this.$('#settingFormArea').style.display = 'block';
    }

    // --- DaGang ---

    async generateDaGang() {
        this.log('正在生成全书大纲...');
        // Mock implementation since /api/v1/narrative/dagang is complex or missing
        // Could try calling CreateBlueprint if we had params

        this.$('#dagangFormArea').style.display = 'none';
        this.$('#dagangResultArea').style.display = 'block';
        this.$('#dagangViewArea').style.display = 'none';

        const outputEl = this.$('#dagangOutput');
        outputEl.innerHTML = 'AI大纲生成功能接入中...<br>目前仅支持查看自动生成的世界设定。';
        this.log('大纲生成暂时跳过（Mock）');

        this.generatedDagangContent = "大纲内容占位符...";

        setTimeout(() => {
            this.enableTab('xigang');
        }, 1000);
    }

    // Stub other methods for safety
    regenerateDaGang() { this.generateDaGang(); }
    confirmDaGang() {
        this.$('#dagangResultArea').style.display = 'none';
        this.$('#dagangViewArea').style.display = 'block';
        this.$('#dagangEdit').value = this.generatedDagangContent;
        this.currentDaGangId = "mock_dagang";
        this.updateStatus();
    }
    saveDaGangEdit() { this.log("大纲已本地保存"); }
    generateXiGang() {
        this.log("细纲生成功能接入中...");
        // Mock Chapters
        this.chapters = [
            { chapter_num: 1, chapter_name: "开端", instruction: "主角登场..." },
            { chapter_num: 2, chapter_name: "冲突", instruction: "遇到敌人..." }
        ];
        this.renderXiGangList();
        this.renderChaptersNav();
        this.enableTab('content');
    }

    renderChaptersNav() {
        const html = this.chapters.map((ch, i) => `
            <div class="dir-chapter-item" data-index="${i}">
                <span class="chapter-status"></span>
                第${ch.chapter_num}章 ${ch.chapter_name}
            </div>
        `).join('');
        this.$('#chapterList').innerHTML = html;
        this.container.querySelectorAll('.dir-chapter-item').forEach(el => {
            this.addEventListener(el, 'click', () => this.selectChapter(parseInt(el.dataset.index)));
        });
    }

    renderXiGangList() {
        const html = this.chapters.map((ch, i) => `
            <div class="xigang-card" data-index="${i}">
                <div class="xigang-card-title">第${ch.chapter_num}章 ${ch.chapter_name}</div>
                <div class="xigang-card-info">${ch.instruction || ''}</div>
            </div>
        `).join('');
        this.$('#xigangList').innerHTML = html;
        this.container.querySelectorAll('.xigang-card').forEach(el => {
            this.addEventListener(el, 'click', () => this.selectChapter(parseInt(el.dataset.index)));
        });
    }

    selectChapter(index) {
        this.currentXiGangId = index;
        const ch = this.chapters[index];

        this.container.querySelectorAll('.dir-chapter-item').forEach((el, i) => el.classList.toggle('active', i === index));
        this.container.querySelectorAll('.xigang-card').forEach((el, i) => el.classList.toggle('active', i === index));

        this.$('#currentXigang').innerHTML = `
            <div style="font-weight: 500; margin-bottom: 8px;">第${ch.chapter_num}章 ${ch.chapter_name}</div>
            <div style="white-space: pre-wrap;">${ch.instruction || '无详细内容'}</div>
        `;

        this.$('#contentTitle').textContent = `第${ch.chapter_num}章 ${ch.chapter_name}`;
        this.$('#contentBody').innerHTML = `<div class="empty-state"><div>AI写作功能接入中...</div></div>`;
        this.$('#writeBtn').disabled = false;
    }

    generateContent() { alert("AI写作功能开发中"); }
    continueContent() { }
}
