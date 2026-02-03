// Landing Page - 首页着陆页

import BaseComponent from '../components/BaseComponent.js';

export default class LandingPage extends BaseComponent {
    render() {
        return `
            <div class="landing-page">
                <!-- Hero Section -->
                <section class="hero">
                    <nav class="landing-nav">
                        <div class="landing-nav-brand">
                            <svg width="28" height="28" viewBox="0 0 48 48" fill="none">
                                <rect x="8" y="16" width="24" height="28" rx="2" fill="#E8E8F4"/>
                                <rect x="12" y="12" width="24" height="28" rx="2" fill="#A5A7E8"/>
                                <rect x="16" y="8" width="24" height="28" rx="2" fill="#5B5FC7"/>
                            </svg>
                            <span>NovelFlow</span>
                        </div>
                        <div class="landing-nav-actions">
                            <a href="#/login" class="btn btn-ghost">登录</a>
                            <a href="#/register" class="btn btn-primary">免费开始</a>
                        </div>
                    </nav>

                    <div class="hero-content">
                        <h1 class="hero-title">
                            用 AI 释放你的<br>
                            <span class="gradient-text">创作潜能</span>
                        </h1>
                        <p class="hero-subtitle">
                            NovelFlow 叙谱是一个 AI 驱动的小说创作平台，<br>
                            帮助你构建世界观、管理人物、生成大纲，让创作更加流畅。
                        </p>
                        <div class="hero-actions">
                            <a href="#/register" class="btn btn-primary btn-xl">
                                <i class="bi bi-rocket-takeoff"></i>
                                免费开始创作
                            </a>
                            <a href="#features" class="btn btn-outline-secondary btn-xl">
                                了解更多
                            </a>
                        </div>
                    </div>

                    <div class="hero-visual">
                        <div class="hero-card">
                            <div class="hero-card-header">
                                <span class="dot red"></span>
                                <span class="dot yellow"></span>
                                <span class="dot green"></span>
                            </div>
                            <div class="hero-card-content">
                                <div class="typing-line">
                                    <span class="typing-label">角色</span>
                                    <span class="typing-text">林逸，一个来自江南小镇的少年...</span>
                                </div>
                                <div class="typing-line">
                                    <span class="typing-label">场景</span>
                                    <span class="typing-text">古老的藏书阁，月光透过窗棂...</span>
                                </div>
                                <div class="typing-line active">
                                    <span class="typing-label">AI</span>
                                    <span class="typing-text typing-cursor">正在生成精彩情节...</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                <!-- Features Section -->
                <section class="features" id="features">
                    <div class="section-header">
                        <h2 class="section-title">为创作者打造的全方位工具</h2>
                        <p class="section-subtitle">从灵感到成稿，NovelFlow 陪伴你完成每一步</p>
                    </div>

                    <div class="features-grid">
                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-globe-americas"></i>
                            </div>
                            <h3>世界观构建</h3>
                            <p>创建独特的世界设定，管理地理、历史、势力等元素，让你的故事世界更加真实丰满。</p>
                        </div>

                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-people"></i>
                            </div>
                            <h3>角色管理</h3>
                            <p>详细的角色档案，包括性格、背景、关系网络，AI 帮你保持角色一致性。</p>
                        </div>

                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-diagram-3"></i>
                            </div>
                            <h3>智能大纲</h3>
                            <p>AI 辅助生成故事大纲和章节结构，让你的情节发展更加紧凑有序。</p>
                        </div>

                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-magic"></i>
                            </div>
                            <h3>AI 续写</h3>
                            <p>遇到瓶颈？让 AI 为你提供多种续写方向，激发创作灵感。</p>
                        </div>

                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-journal-text"></i>
                            </div>
                            <h3>章节编辑</h3>
                            <p>专注的写作环境，实时保存，支持富文本编辑，让创作更加顺畅。</p>
                        </div>

                        <div class="feature-card">
                            <div class="feature-icon">
                                <i class="bi bi-cloud-check"></i>
                            </div>
                            <h3>云端同步</h3>
                            <p>作品安全存储在云端，随时随地继续创作，永不丢失。</p>
                        </div>
                    </div>
                </section>

                <!-- CTA Section -->
                <section class="cta">
                    <div class="cta-content">
                        <h2>准备好开始创作了吗？</h2>
                        <p>加入 NovelFlow，与 AI 一起开启你的创作之旅</p>
                        <a href="#/register" class="btn btn-primary btn-xl">
                            <i class="bi bi-pen"></i>
                            立即注册，免费使用
                        </a>
                    </div>
                </section>

                <!-- Footer -->
                <footer class="landing-footer">
                    <div class="footer-content">
                        <div class="footer-brand">
                            <svg width="24" height="24" viewBox="0 0 48 48" fill="none">
                                <rect x="8" y="16" width="24" height="28" rx="2" fill="#E8E8F4"/>
                                <rect x="12" y="12" width="24" height="28" rx="2" fill="#A5A7E8"/>
                                <rect x="16" y="8" width="24" height="28" rx="2" fill="#5B5FC7"/>
                            </svg>
                            <span>NovelFlow 叙谱</span>
                        </div>
                        <p class="footer-text">AI 驱动的小说创作平台</p>
                    </div>
                </footer>
            </div>
        `;
    }
}
