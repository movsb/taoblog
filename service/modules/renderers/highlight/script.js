function syncCodeScroll(img) {
	let container = img.parentElement;
	let tr = container.querySelector(':scope .lntable tr');
	let td =  container.querySelector(':scope .lntable .lntd:first-child');
	let rafId = null;
	tr.onscroll = () => {
		if (rafId !== null) return;
		rafId = requestAnimationFrame(() => {
			td.style.transform = `translateY(${-tr.scrollTop}px)`;
			rafId = null;
		});
	};
	img.remove();
}
