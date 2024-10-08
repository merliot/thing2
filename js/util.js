// Dynamically load CSS file into <head>
function loadCSS(href) {
	if (!document.querySelector(`link[href="${href}"]`)) {
		var link = document.createElement('link');
		link.rel = 'stylesheet';
		link.href = href;
		document.head.appendChild(link);
	}
}

// Dynamically load script file into <head>.  Callback is called once script is
// loaded.
function loadScript(src, callback) {
	if (!document.querySelector(`script[src="${src}"]`)) {
		var script = document.createElement('script');
		script.src = src;
		script.onload = callback;
		document.head.appendChild(script);
	} else {
		if (callback) callback();
	}
}

