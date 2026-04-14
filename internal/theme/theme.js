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