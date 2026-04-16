# 侧边栏状态持久化与防闪烁 (FOUC) 优化方案

## 1. 背景与问题分析

在 VitePress (GoPress) 的主题实现中，侧边栏支持通过用户交互折叠/展开分组，并将状态持久化到 `localStorage` 中。

但是，在早期的实现中存在两个导致页面“一闪而过”（闪烁/卡顿）的体验问题：
1. **SPA 路由切换闪烁**：在使用 `fetch` 替换 `innerHTML` 进行无刷新页面跳转时，新生成的侧边栏 DOM 初始状态始终是展开的，随后才被 JavaScript 恢复为折叠状态。
2. **首次渲染闪烁 (FOUC)**：由于控制状态恢复的 `theme.js` 脚本是在页面底部作为 `<script type="module">` 引入的，浏览器会将其作为 `defer` 脚本延迟执行。这意味着浏览器在完成整个 HTML 页面 DOM 解析，甚至已经把默认展开的侧边栏绘制到屏幕上之后，才会执行 `theme.js` 去折叠它。

## 2. 优化方案设计

为了彻底解决这两种情况下的闪烁问题，我们实施了以下重构方案：

### 2.1 状态批量存储 (Batch Storage)
将原来分散存储的每一个侧边栏项（`gopress-sidebar-XXX`）合并存储为一个 JSON 对象 `gopress-sidebar-states`，降低在恢复状态时的 IO 读取频率，提高性能。

### 2.2 针对首屏渲染：`<head>` 注入 MutationObserver (防 FOUC)
要解决首次加载页面的 FOUC 问题，状态恢复逻辑必须在 DOM 渲染**之前**执行。

我们在 `layout.tmpl` 的 `<head>` 标签中注入了一段内联的同步阻塞脚本，并利用 `MutationObserver` 监听 DOM 树的构建：
- 内联脚本在页面加载的最开始执行。
- `MutationObserver` 会实时监控页面的解析，一旦解析器遇到了带有 `.VPSidebarItem.is-collapsible` 的节点被添加到 DOM 树中（此时浏览器还未进行重绘），立刻读取 `localStorage` 并给它们加上对应的 `collapsed` 类和 `display: none` 样式。
- 当页面解析完成（`DOMContentLoaded`）后，自动断开 Observer，避免影响后续的页面性能。

### 2.3 针对 SPA 跳转：DOM 替换后同步恢复
在 `theme.js` 中负责 SPA 跳转的 `navigateTo` 函数内，当通过 `innerHTML` 替换完页面布局后，立即同步调用一次 `restoreSidebarState()`。这样可以确保新页面内容在下一次浏览器重绘前，侧边栏状态已经被正确应用。

## 3. 核心代码实现

### 3.1 `layout.tmpl` 中的首屏防闪烁脚本

在 `<head>` 中添加如下逻辑：

```html
<!-- 提前恢复侧边栏折叠状态，在DOM解析期间立即应用，彻底避免页面渲染闪烁 -->
<script>
(function() {
    try {
        const saved = localStorage.getItem('gopress-sidebar-states');
        if (!saved) return;
        const states = JSON.parse(saved);
        
        const observer = new MutationObserver((mutations) => {
            for (const mutation of mutations) {
                for (const node of mutation.addedNodes) {
                    if (node.nodeType !== 1) continue;
                    
                    const items = node.classList && node.classList.contains('VPSidebarItem') 
                        ? [node] 
                        : node.querySelectorAll ? Array.from(node.querySelectorAll('.VPSidebarItem.is-collapsible')) : [];
                    
                    for (const item of items) {
                        if (!item.classList || !item.classList.contains('is-collapsible')) continue;
                        const textEl = item.querySelector('.item-row > .text');
                        const key = textEl ? textEl.textContent.trim() : '';
                        
                        if (key && states[key] !== undefined) {
                            let group = null;
                            for (let i = 0; i < item.children.length; i++) {
                                if (item.children[i].classList.contains('VPSidebarGroup')) {
                                    group = item.children[i];
                                    break;
                                }
                            }
                            if (states[key]) {
                                item.classList.add('collapsed');
                                if (group) group.style.display = 'none';
                            } else {
                                item.classList.remove('collapsed');
                                if (group) group.style.display = '';
                            }
                        }
                    }
                }
            }
        });
        observer.observe(document.documentElement, { childList: true, subtree: true });
        document.addEventListener('DOMContentLoaded', () => observer.disconnect());
    } catch (e) {}
})();
</script>
```

### 3.2 `theme.js` 中的状态管理与 SPA 跳转

**批量保存与恢复函数**：

```javascript
function getSidebarStateKey(item) {
	const textEl = item.querySelector('.item-row > .text');
	return textEl ? textEl.textContent.trim() : '';
}

function saveSidebarItemState(item, collapsed) {
	const key = getSidebarStateKey(item);
	if (!key) return;
	
	let states = {};
	try {
		const saved = localStorage.getItem('gopress-sidebar-states');
		if (saved) states = JSON.parse(saved);
	} catch (e) {}
	
	states[key] = collapsed;
	localStorage.setItem('gopress-sidebar-states', JSON.stringify(states));
}

function restoreSidebarState() {
	let states = null;
	try {
		const saved = localStorage.getItem('gopress-sidebar-states');
		if (saved) states = JSON.parse(saved);
	} catch (e) {}

	if (!states) return;

	document.querySelectorAll('.VPSidebarItem.is-collapsible').forEach(function(item) {
		const key = getSidebarStateKey(item);
		if (!key || states[key] === undefined) return;
        // ... 根据 states[key] 的布尔值恢复 class 和 style ...
	});
}
```

**SPA 跳转同步更新**：

```javascript
async function navigateTo(path, push = true) {
    // ...
    const newContent = doc.getElementById('Layout');
    if (newContent) {
        document.getElementById('Layout').innerHTML = newContent.innerHTML;
        // ...
        
        // 在替换完 DOM 后立即恢复状态，防止 SPA 跳转闪烁
        restoreSidebarState();
    }
    // ...
}
```

## 4. 总结与收益

1. **极致首屏体验**：借助 `MutationObserver` 将 DOM 修改前置到了浏览器重绘之前，实现了首屏侧边栏折叠状态“零闪烁”。
2. **SPA 无缝切换**：补齐了前端路由跳转时的状态恢复逻辑，保证了单页应用体验的一致性。
3. **更优的存储设计**：从分散存储改为批量 JSON 存储，减少了 IO 开销，使 LocalStorage 结构更加清晰整洁。