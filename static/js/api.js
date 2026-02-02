// API服务层

import { userActions } from './store.js';
import { showToast, showLoading, hideLoading } from './utils.js';

const API_BASE_URL = window.location.origin;

/**
 * HTTP请求封装
 */
async function request(url, options = {}) {
    const token = userActions.getToken();

    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
        ...options,
        headers
    };

    try {
        const response = await fetch(`${API_BASE_URL}${url}`, config);

        // 处理401未授权
        if (response.status === 401) {
            userActions.logout();
            window.location.hash = '#/login';
            throw new Error('未授权，请重新登录');
        }

        // 处理其他错误状态
        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: '请求失败' }));
            throw new Error(error.message || `HTTP ${response.status}`);
        }

        // 处理204 No Content
        if (response.status === 204) {
            return null;
        }

        return await response.json();
    } catch (error) {
        console.error('API请求错误:', error);
        throw error;
    }
}

/**
 * GET请求
 */
async function get(url, options = {}) {
    return request(url, { ...options, method: 'GET' });
}

/**
 * POST请求
 */
async function post(url, data, options = {}) {
    return request(url, {
        ...options,
        method: 'POST',
        body: JSON.stringify(data)
    });
}

/**
 * PUT请求
 */
async function put(url, data, options = {}) {
    return request(url, {
        ...options,
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

/**
 * DELETE请求
 */
async function del(url, options = {}) {
    return request(url, { ...options, method: 'DELETE' });
}

/**
 * 认证API
 */
export const authAPI = {
    // 登录
    async login(username, password) {
        return post('/api/auth/login', { username, password });
    },

    // 注册
    async register(userData) {
        return post('/api/auth/register', userData);
    },

    // 获取当前用户信息
    async getCurrentUser() {
        return get('/api/auth/me');
    },

    // 登出
    async logout() {
        return post('/api/auth/logout');
    }
};

/**
 * 项目API
 */
export const projectAPI = {
    // 获取项目列表
    async getProjects() {
        return get('/api/projects');
    },

    // 获取项目详情
    async getProject(id) {
        return get(`/api/projects/${id}`);
    },

    // 创建项目
    async createProject(projectData) {
        return post('/api/projects', projectData);
    },

    // 更新项目
    async updateProject(id, projectData) {
        return put(`/api/projects/${id}`, projectData);
    },

    // 删除项目
    async deleteProject(id) {
        return del(`/api/projects/${id}`);
    }
};

/**
 * 章节API
 */
export const chapterAPI = {
    // 获取章节列表
    async getChapters(projectId) {
        return get(`/api/projects/${projectId}/chapters`);
    },

    // 获取章节详情
    async getChapter(projectId, chapterId) {
        return get(`/api/projects/${projectId}/chapters/${chapterId}`);
    },

    // 创建章节
    async createChapter(projectId, chapterData) {
        return post(`/api/projects/${projectId}/chapters`, chapterData);
    },

    // 更新章节
    async updateChapter(projectId, chapterId, chapterData) {
        return put(`/api/projects/${projectId}/chapters/${chapterId}`, chapterData);
    },

    // 删除章节
    async deleteChapter(projectId, chapterId) {
        return del(`/api/projects/${projectId}/chapters/${chapterId}`);
    }
};

/**
 * 叙事节点API
 */
export const narrativeAPI = {
    // 获取叙事节点
    async getNodes(projectId) {
        return get(`/api/projects/${projectId}/narrative-nodes`);
    },

    // 创建叙事节点
    async createNode(projectId, nodeData) {
        return post(`/api/projects/${projectId}/narrative-nodes`, nodeData);
    },

    // 更新叙事节点
    async updateNode(projectId, nodeId, nodeData) {
        return put(`/api/projects/${projectId}/narrative-nodes/${nodeId}`, nodeData);
    },

    // 删除叙事节点
    async deleteNode(projectId, nodeId) {
        return del(`/api/projects/${projectId}/narrative-nodes/${nodeId}`);
    }
};

/**
 * 世界设定API
 */
export const worldAPI = {
    // 获取世界设定
    async getWorldSettings(projectId) {
        return get(`/api/projects/${projectId}/world-settings`);
    },

    // 创建世界设定
    async createWorldSetting(projectId, settingData) {
        return post(`/api/projects/${projectId}/world-settings`, settingData);
    },

    // 更新世界设定
    async updateWorldSetting(projectId, settingId, settingData) {
        return put(`/api/projects/${projectId}/world-settings/${settingId}`, settingData);
    },

    // 生成世界设定（AI）
    async generateWorldSetting(projectId, params) {
        return post(`/api/projects/${projectId}/world-settings/generate`, params);
    }
};

/**
 * 角色API
 */
export const characterAPI = {
    // 获取角色列表
    async getCharacters(projectId) {
        return get(`/api/projects/${projectId}/characters`);
    },

    // 创建角色
    async createCharacter(projectId, characterData) {
        return post(`/api/projects/${projectId}/characters`, characterData);
    },

    // 更新角色
    async updateCharacter(projectId, characterId, characterData) {
        return put(`/api/projects/${projectId}/characters/${characterId}`, characterData);
    },

    // 删除角色
    async deleteCharacter(projectId, characterId) {
        return del(`/api/projects/${projectId}/characters/${characterId}`);
    }
};

/**
 * 写作器API
 */
export const writerAPI = {
    // 生成场景
    async generateScene(projectId, sceneData) {
        return post(`/api/projects/${projectId}/writer/generate-scene`, sceneData);
    },

    // 生成对话
    async generateDialogue(projectId, dialogueData) {
        return post(`/api/projects/${projectId}/writer/generate-dialogue`, dialogueData);
    }
};

export default {
    authAPI,
    projectAPI,
    chapterAPI,
    narrativeAPI,
    worldAPI,
    characterAPI,
    writerAPI
};
