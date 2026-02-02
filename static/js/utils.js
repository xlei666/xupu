// 工具函数库

/**
 * DOM操作辅助函数
 */
export const $ = (selector) => document.querySelector(selector);
export const $$ = (selector) => document.querySelectorAll(selector);

/**
 * 创建元素
 */
export function createElement(tag, attrs = {}, children = []) {
    const element = document.createElement(tag);
    
    Object.entries(attrs).forEach(([key, value]) => {
        if (key === 'className') {
            element.className = value;
        } else if (key === 'dataset') {
            Object.entries(value).forEach(([dataKey, dataValue]) => {
                element.dataset[dataKey] = dataValue;
            });
        } else if (key.startsWith('on')) {
            const eventName = key.substring(2).toLowerCase();
            element.addEventListener(eventName, value);
        } else {
            element.setAttribute(key, value);
        }
    });
    
    children.forEach(child => {
        if (typeof child === 'string') {
            element.appendChild(document.createTextNode(child));
        } else if (child instanceof Node) {
            element.appendChild(child);
        }
    });
    
    return element;
}

/**
 * 日期格式化
 */
export function formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
    const d = new Date(date);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    const seconds = String(d.getSeconds()).padStart(2, '0');
    
    return format
        .replace('YYYY', year)
        .replace('MM', month)
        .replace('DD', day)
        .replace('HH', hours)
        .replace('mm', minutes)
        .replace('ss', seconds);
}

/**
 * 相对时间格式化
 */
export function formatRelativeTime(date) {
    const now = new Date();
    const d = new Date(date);
    const diff = now - d;
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 7) {
        return formatDate(date, 'YYYY-MM-DD');
    } else if (days > 0) {
        return `${days}天前`;
    } else if (hours > 0) {
        return `${hours}小时前`;
    } else if (minutes > 0) {
        return `${minutes}分钟前`;
    } else {
        return '刚刚';
    }
}

/**
 * Toast提示
 */
export function showToast(message, type = 'info', duration = 3000) {
    const container = $('#toast-container');
    if (!container) return;
    
    const bgClass = {
        success: 'bg-success',
        error: 'bg-danger',
        warning: 'bg-warning',
        info: 'bg-info'
    }[type] || 'bg-info';
    
    const iconClass = {
        success: 'bi-check-circle',
        error: 'bi-exclamation-circle',
        warning: 'bi-exclamation-triangle',
        info: 'bi-info-circle'
    }[type] || 'bi-info-circle';
    
    const toast = createElement('div', {
        className: `toast align-items-center text-white ${bgClass} border-0`,
        role: 'alert',
        'aria-live': 'assertive',
        'aria-atomic': 'true'
    }, [
        createElement('div', { className: 'd-flex' }, [
            createElement('div', { className: 'toast-body' }, [
                createElement('i', { className: `bi ${iconClass} me-2` }),
                message
            ]),
            createElement('button', {
                type: 'button',
                className: 'btn-close btn-close-white me-2 m-auto',
                'data-bs-dismiss': 'toast',
                'aria-label': 'Close'
            })
        ])
    ]);
    
    container.appendChild(toast);
    
    const bsToast = new bootstrap.Toast(toast, { autohide: true, delay: duration });
    bsToast.show();
    
    toast.addEventListener('hidden.bs.toast', () => {
        toast.remove();
    });
}

/**
 * 表单验证
 */
export function validateForm(formData, rules) {
    const errors = {};
    
    Object.entries(rules).forEach(([field, rule]) => {
        const value = formData[field];
        
        if (rule.required && !value) {
            errors[field] = rule.message || `${field}不能为空`;
            return;
        }
        
        if (rule.minLength && value.length < rule.minLength) {
            errors[field] = rule.message || `${field}长度不能少于${rule.minLength}个字符`;
            return;
        }
        
        if (rule.maxLength && value.length > rule.maxLength) {
            errors[field] = rule.message || `${field}长度不能超过${rule.maxLength}个字符`;
            return;
        }
        
        if (rule.pattern && !rule.pattern.test(value)) {
            errors[field] = rule.message || `${field}格式不正确`;
            return;
        }
        
        if (rule.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
            errors[field] = rule.message || '邮箱格式不正确';
            return;
        }
        
        if (rule.custom && !rule.custom(value, formData)) {
            errors[field] = rule.message || `${field}验证失败`;
            return;
        }
    });
    
    return {
        valid: Object.keys(errors).length === 0,
        errors
    };
}

/**
 * 防抖函数
 */
export function debounce(func, wait = 300) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * 节流函数
 */
export function throttle(func, limit = 300) {
    let inThrottle;
    return function executedFunction(...args) {
        if (!inThrottle) {
            func(...args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

/**
 * 显示加载遮罩
 */
export function showLoading() {
    const existing = $('.loading-overlay');
    if (existing) return;
    
    const overlay = createElement('div', { className: 'loading-overlay' }, [
        createElement('div', { className: 'spinner-border text-primary', role: 'status' }, [
            createElement('span', { className: 'visually-hidden' }, ['加载中...'])
        ])
    ]);
    
    document.body.appendChild(overlay);
}

/**
 * 隐藏加载遮罩
 */
export function hideLoading() {
    const overlay = $('.loading-overlay');
    if (overlay) {
        overlay.remove();
    }
}

/**
 * 确认对话框
 */
export function confirm(message, title = '确认') {
    return new Promise((resolve) => {
        const modalId = 'confirmModal' + Date.now();
        const modal = createElement('div', {
            className: 'modal fade',
            id: modalId,
            tabindex: '-1'
        }, [
            createElement('div', { className: 'modal-dialog' }, [
                createElement('div', { className: 'modal-content' }, [
                    createElement('div', { className: 'modal-header' }, [
                        createElement('h5', { className: 'modal-title' }, [title]),
                        createElement('button', {
                            type: 'button',
                            className: 'btn-close',
                            'data-bs-dismiss': 'modal'
                        })
                    ]),
                    createElement('div', { className: 'modal-body' }, [message]),
                    createElement('div', { className: 'modal-footer' }, [
                        createElement('button', {
                            type: 'button',
                            className: 'btn btn-secondary',
                            'data-bs-dismiss': 'modal',
                            onClick: () => resolve(false)
                        }, ['取消']),
                        createElement('button', {
                            type: 'button',
                            className: 'btn btn-primary',
                            'data-bs-dismiss': 'modal',
                            onClick: () => resolve(true)
                        }, ['确定'])
                    ])
                ])
            ])
        ]);
        
        document.body.appendChild(modal);
        const bsModal = new bootstrap.Modal(modal);
        bsModal.show();
        
        modal.addEventListener('hidden.bs.modal', () => {
            modal.remove();
        });
    });
}

/**
 * 深拷贝
 */
export function deepClone(obj) {
    if (obj === null || typeof obj !== 'object') return obj;
    if (obj instanceof Date) return new Date(obj);
    if (obj instanceof Array) return obj.map(item => deepClone(item));
    
    const clonedObj = {};
    for (const key in obj) {
        if (obj.hasOwnProperty(key)) {
            clonedObj[key] = deepClone(obj[key]);
        }
    }
    return clonedObj;
}
