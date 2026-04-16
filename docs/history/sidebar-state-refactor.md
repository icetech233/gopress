# 侧边栏折叠状态重构技术方案

## 1. 背景与问题分析

在原有的 VitePress (GoPress) 主题实现中，侧边栏的折叠状态存在页面切换时“一闪而过”的卡顿和闪烁问题。经过分析，主要原因有以下两点：

1. **频繁且分散的 `localStorage` 读写**：
   原逻辑中，针对每一个可折叠的侧边栏项（`VPSidebarItem`），都会单独在 `localStorage` 中生成一条记录（如 `gopress-sidebar-XXX`）。在页面初始化或恢复状态时，遍历所有侧边栏项并在循环体内频繁调用同步阻塞的 `localStorage.getItem()`，导致 IO 性能较差，当侧边栏项较多时容易引发渲染卡顿。
   
2. **SPA 页面切换时的状态丢失**：
   项目使用了 `navigateTo` 方法实现无刷新的 SPA 路由跳转。在通过 `fetch` 获取新页面 HTML 并替换 `Layout` 容器的 `innerHTML` 后，新生成的 DOM 元素（包括侧边栏）并未重新应用之前保存的折叠状态。这导致新页面的侧边栏会短暂显示为默认的“全部展开”状态，从而产生视觉上的闪烁。

## 2. 优化方案与重构目标

针对上述问题，我们对侧边栏状态的管理逻辑进行了以下重构：

- **状态批量存储 (Batch Storage)**：将所有侧边栏项的折叠状态统一合并为一个 JSON 对象，并保存在单一的 `localStorage` 键（`gopress-sidebar-states`）下。这样在状态恢复时，只需读取和解析一次 `localStorage`，极大地降低了 IO 开销。
- **状态即时恢复 (Immediate Restoration)**：在 SPA 路由跳转（`navigateTo`）替换页面内容之后，立即主动调用 `restoreSidebarState()`，确保新渲染的侧边栏 DOM 能够在浏览器下一次重绘前应用正确的折叠状态，消除闪烁。

## 3. 具体实现细节

### 3.1 状态存储与读取机制的改造

**修改前**：
```javascript
function saveSidebarItemState(item, collapsed) {
	const key = getSidebarStateKey(item);
	if (key) localStorage.setItem(key, collapsed ? '1' : '0');
}
```

**修改后**：
改为集中式的 JSON 存储机制：
```javascript
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
```

在 `restoreSidebarState` 方法中，改为一次性读取解析：
```javascript
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

### 3.2 SPA 路由跳转后的状态同步

在 `navigateTo` 函数中，完成 DOM 替换后，显式调用状态恢复逻辑：

```javascript
async function navigateTo(path, push = true) {
    // ... fetch 新页面内容 ...
    const newContent = doc.getElementById('Layout');
    if (newContent) {
        document.getElementById('Layout').innerHTML = newContent.innerHTML;
        // ... 更新 title 和 history ...
        
        // 【新增】页面切换后立即恢复侧边栏状态，避免闪烁
        restoreSidebarState();
    } else {
        window.location.href = path; // Fallback
    }
}
```

## 4. 总结与收益

通过此次重构：
1. **性能提升**：侧边栏状态的 `localStorage` 读取次数由 `O(N)` 降为 `O(1)`，显著减少了同步阻塞时间。
2. **体验优化**：彻底解决了 SPA 无刷新页面切换时的侧边栏折叠状态闪烁问题，保证了导航切换时的视觉连贯性和顺滑度。
3. **存储整洁**：避免了在用户浏览器的 `localStorage` 中散落大量的、碎片化的 `gopress-sidebar-*` 键值对，使缓存管理更加规范。