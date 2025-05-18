function syncCodeScroll(img) {
	let container = img.parentElement;
	let tr = container.querySelector(':scope .lntable tr');
	let td =  container.querySelector(':scope .lntable .lntd:first-child');
	tr.onscroll = e => td.style.top = `${-tr.scrollTop}px`;
	img.remove();
}
