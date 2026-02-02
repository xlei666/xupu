// 状态管理 Store

class Store {
    constructor(initialState = {}) {
        this.state = initialState;
        this.listeners = [];
    }

    getState() {
        return this.state;
    }

    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.notify();
    }

    subscribe(listener) {
        this.listeners.push(listener);
        return () => {
            this.listeners = this.listeners.filter(l => l !== listener);
        };
    }

    notify() {
        this.listeners.forEach(listener => listener(this.state));
    }
}

// 全局状态
const store = new Store({
    user: null,
    token: localStorage.getItem('token') || null,
    projects: [],
    currentProject: null,
    theme: localStorage.getItem('theme') || 'light'
});

// 用户相关操作
export const userActions = {
    login(user, token) {
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
        store.setState({ user, token });
    },

    logout() {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        store.setState({ user: null, token: null });
    },

    getUser() {
        const user = store.getState().user;
        if (user) return user;

        const savedUser = localStorage.getItem('user');
        if (savedUser) {
            const parsedUser = JSON.parse(savedUser);
            store.setState({ user: parsedUser });
            return parsedUser;
        }

        return null;
    },

    getToken() {
        return store.getState().token || localStorage.getItem('token');
    },

    isAuthenticated() {
        return !!this.getToken();
    }
};

// 项目相关操作
export const projectActions = {
    setProjects(projects) {
        store.setState({ projects });
    },

    getProjects() {
        return store.getState().projects;
    },

    setCurrentProject(project) {
        store.setState({ currentProject: project });
    },

    getCurrentProject() {
        return store.getState().currentProject;
    },

    addProject(project) {
        const projects = [...store.getState().projects, project];
        store.setState({ projects });
    },

    updateProject(id, updates) {
        const projects = store.getState().projects.map(p =>
            p.id === id ? { ...p, ...updates } : p
        );
        store.setState({ projects });
    },

    deleteProject(id) {
        const projects = store.getState().projects.filter(p => p.id !== id);
        store.setState({ projects });
    }
};

// 主题相关操作
export const themeActions = {
    setTheme(theme) {
        localStorage.setItem('theme', theme);
        store.setState({ theme });
        document.documentElement.setAttribute('data-bs-theme', theme);
    },

    getTheme() {
        return store.getState().theme;
    },

    toggleTheme() {
        const currentTheme = this.getTheme();
        const newTheme = currentTheme === 'light' ? 'dark' : 'light';
        this.setTheme(newTheme);
    }
};

// 订阅状态变化
export function subscribe(listener) {
    return store.subscribe(listener);
}

// 获取完整状态
export function getState() {
    return store.getState();
}

// 初始化主题
themeActions.setTheme(themeActions.getTheme());

export default store;
