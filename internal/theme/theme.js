// Minimal SPA Router to emulate GoPress fast navigation
document.addEventListener('click', e => {
	const link = e.target.closest('a');
	if (!link || !link.href) return;
	
	const url = new URL(link.href);
	if (url.origin !== window.location.origin) return; // External link
	if (url.pathname === window.location.pathname) return; // Same page

	e.preventDefault();
	navigateTo(url.pathname);
});

window.addEventListener('popstate', () => {
	navigateTo(window.location.pathname, false);
});

async function navigateTo(path, push = true) {
	try {
		const res = await fetch(path);
		const html = await res.text();
		
		const parser = new DOMParser();
		const doc = parser.parseFromString(html, 'text/html');
		
		const newContent = doc.getElementById('Layout');
		if (newContent) {
			document.getElementById('Layout').innerHTML = newContent.innerHTML;
			document.title = doc.title;
			if (push) {
				window.history.pushState({}, '', path);
			}
			window.scrollTo(0, 0);
			
			if (window.closeSidebar) {
				window.closeSidebar();
			}
			
			restoreSidebarState();
		} else {
			window.location.href = path; // Fallback
		}
	} catch (e) {
		console.error('Navigation error:', e);
		window.location.href = path; // Fallback
	}
}

window.toggleDarkMode = function() {
	const isDark = document.documentElement.classList.toggle('dark');
	localStorage.setItem('gopress-theme-appearance', isDark ? 'dark' : 'light');
};

window.toggleSidebar = function() {
	const sidebar = document.getElementById('VPSidebar');
	if (sidebar) {
		sidebar.classList.toggle('open');
	}
};
window.closeSidebar = function() {
	const sidebar = document.getElementById('VPSidebar');
	if (sidebar) {
		sidebar.classList.remove('open');
	}
};

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
	});
}

window.toggleSidebarGroup = function(btn) {
	const item = btn.closest('.VPSidebarItem');
	if (!item) return;

	const isCollapsed = item.classList.contains('collapsed');

	let group = null;
	for (let i = 0; i < item.children.length; i++) {
		if (item.children[i].classList.contains('VPSidebarGroup')) {
			group = item.children[i];
			break;
		}
	}

	if (isCollapsed) {
		item.classList.remove('collapsed');
		if (group) group.style.display = '';
		saveSidebarItemState(item, false);
	} else {
		item.classList.add('collapsed');
		if (group) group.style.display = 'none';
		saveSidebarItemState(item, true);
	}
};

// 在模块加载完成时立即执行一次，作为补充（主要防止内联脚本未生效等情况）
restoreSidebarState();