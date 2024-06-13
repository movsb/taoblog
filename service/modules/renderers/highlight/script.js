function syncCodeScroll(id) {
	let img = document.getElementById(id);
	let container = img.parentElement;
	let tr = container.querySelector(':scope .lntable tr');
	let td =  container.querySelector(':scope .lntable .lntd:first-child');
	tr.onscroll = e => td.style.top = `${-tr.scrollTop}px`;
	img.remove();
}
