# 侧边栏折叠状态持久化方案（localStorage 版）

## 问题描述

当用户在侧边栏展开/折叠某个分组后，点击链接切换页面或刷新浏览器，`navigateTo()` 函数会用服务端返回的新 HTML 整体替换 `#Layout` 的 `innerHTML`，导致：

1. 侧边栏 DOM 被销毁并重建
2. 用户手动调整的折叠/展开状态全部丢失
3. 服务端渲染的初始状态（由 `Collapsed` 字段决定）重新生效

---

## 方案选择

### 为什么选 localStorage 而不是局部替换 DOM

| 方案 | 优点 | 缺点 |
|---|---|---|
| **localStorage 持久化** | 简单直接，无需修改导航核心逻辑；状态跨页面、跨刷新、跨重启浏览器都能保留 | 依赖文本 key 匹配，同名分组会有歧义 |
| 局部替换 DOM | 无歧义，状态天然保留 | 需处理复杂的页面跳转场景矩阵（4 种），改动范围大 |

**最终选择 localStorage 方案**，因为它改动最小，且能彻底解决跨页面刷新的状态丢失问题。

---

## 实现逻辑

### 1. 写入逻辑（用户操作触发）

在 `toggleSidebarGroup` 函数中，每次用户点击折叠/展开按钮时，将当前状态写入 `localStorage`：

```js
function saveSidebarItemState(item, collapsed) {
    const key = getSidebarStateKey(item);
    if (key) localStorage.setItem(key, collapsed ? '1' : '0');
}

function getSidebarStateKey(item) {
    const textEl = item.querySelector('.item-row > .text');
    return 'gopress-sidebar-' + (textEl ? textEl.textContent.trim() : '');
}
```

### 2. 读取逻辑（页面加载触发）

在 `DOMContentLoaded` 事件中，从 `localStorage` 读取之前保存的状态并应用到 DOM：

```js
function restoreSidebarState() {
    document.querySelectorAll('.VPSidebarItem.is-collapsible').forEach(function(item) {
        const key = getSidebarStateKey(item);
        const saved = localStorage.getItem(key);
        if (saved === null) return;

        let group = null;
        for (let i = 0; i < item.children.length; i++) {
            if (item.children[i].classList.contains('VPSidebarGroup')) {
                group = item.children[i];
                break;
            }
        }

        if (saved === '1') {
            item.classList.add('collapsed');
            if (group) group.style.display = 'none';
        } else {
            item.classList.remove('collapsed');
            if (group) group.style.display = '';
        }
    });
}

document.addEventListener('DOMContentLoaded', restoreSidebarState);
```

---

## 关键细节

### 状态优先级

1. **用户手动调整的状态**（localStorage）> 服务端配置的初始状态（Collapsed 字段）
2. 如果没有 localStorage 记录，完全按服务端配置渲染
3. 服务端配置仅作为首次访问的默认值

### 存储格式

| Key 格式 | Value | 含义 |
|---|---|---|
| `gopress-sidebar-{分组文本}` | `1` | 折叠状态 |
| `gopress-sidebar-{分组文本}` | `0` | 展开状态 |

### 边界情况处理

1. **同名分组**：如果侧边栏存在多个文本完全相同的分组，会共享同一个 localStorage 状态
2. **分组文本修改**：如果分组文本在配置中被修改，之前的状态会失效，回到服务端默认值
3. **清空 localStorage**：所有状态丢失，回到服务端默认值

---

## 需修改的文件

| 文件 | 修改内容 |
|---|---|
| `internal/theme/theme.js` | 新增状态写入和读取函数，修改 toggleSidebarGroup 同步状态，添加 DOMContentLoaded 事件监听 |

服务端（Go）、模板（layout.tmpl）、样式（theme.css）均**无需改动**。

---

## 测试步骤

1. 访问任意文档页，展开/折叠侧边栏分组
2. 点击链接跳转到其他文档页，检查折叠状态是否保留
3. 刷新浏览器，检查折叠状态是否保留
4. 重启浏览器，检查折叠状态是否保留
5. 打开新标签页访问同一站点，检查折叠状态是否同步（localStorage 同源共享）

---

## 优缺点总结

### 优点
- **改动最小**：仅修改 theme.js，不影响导航核心逻辑
- **状态持久化**：跨页面、跨刷新、跨重启浏览器都能保留
- **向后兼容**：没有 localStorage 记录时，完全按服务端配置渲染
- **简单直观**：代码逻辑清晰，易于维护

### 缺点
- **同名分组歧义**：如果侧边栏存在多个文本完全相同的分组，会共享同一个状态
- **依赖文本内容**：分组文本修改会导致之前的状态失效

### 优化建议

如果需要避免同名分组歧义，可以将 `getSidebarStateKey` 改为生成更稳定的 key，例如：

```js
function getSidebarStateKey(item) {
    // 生成唯一路径：父级文本 → 子级文本 → 当前文本
    let path = [];
    let current = item;
    while (current) {
        const textEl = current.querySelector('.item-row > .text');
        if (textEl) path.unshift(textEl.textContent.trim());
        current = current.closest('.VPSidebarGroup')?.parentElement?.closest('.VPSidebarItem');
    }
    return 'gopress-sidebar-' + path.join('>');
}
```

这样可以生成类似 `gopress-sidebar-一级菜单>二级菜单` 的唯一路径 key，彻底解决同名分组歧义问题。