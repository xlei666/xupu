// 基础组件类

export default class BaseComponent {
    constructor(container) {
        this.container = typeof container === 'string'
            ? document.querySelector(container)
            : container;
        this.state = {};
        this.eventListeners = [];
    }

    /**
     * 渲染组件 - 子类必须实现
     */
    render() {
        throw new Error('render() must be implemented by subclass');
    }

    /**
     * 挂载组件
     */
    mount() {
        if (!this.container) {
            console.error('Container not found');
            return;
        }

        const html = this.render();
        this.container.innerHTML = html;
        this.bindEvents();
        this.onMounted();
    }

    /**
     * 绑定事件 - 子类可选实现
     */
    bindEvents() {
        // 子类实现
    }

    /**
     * 组件挂载后回调
     */
    onMounted() {
        // 子类实现
    }

    /**
     * 更新组件
     */
    update() {
        this.mount();
    }

    /**
     * 卸载组件
     */
    unmount() {
        this.removeAllEventListeners();
        if (this.container) {
            this.container.innerHTML = '';
        }
        this.onUnmounted();
    }

    /**
     * 组件卸载后回调
     */
    onUnmounted() {
        // 子类实现
    }

    /**
     * 添加事件监听器（自动管理）
     */
    addEventListener(element, event, handler) {
        const el = typeof element === 'string'
            ? this.container.querySelector(element)
            : element;

        if (el) {
            el.addEventListener(event, handler);
            this.eventListeners.push({ element: el, event, handler });
        }
    }

    /**
     * 移除所有事件监听器
     */
    removeAllEventListeners() {
        this.eventListeners.forEach(({ element, event, handler }) => {
            element.removeEventListener(event, handler);
        });
        this.eventListeners = [];
    }

    /**
     * 查询元素
     */
    $(selector) {
        return this.container.querySelector(selector);
    }

    /**
     * 查询所有元素
     */
    $$(selector) {
        return this.container.querySelectorAll(selector);
    }
}
