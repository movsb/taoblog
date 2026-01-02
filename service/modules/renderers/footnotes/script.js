document.addEventListener('DOMContentLoaded', ()=>{
	/**
	 * 
	 * @param {HTMLAnchorElement} a
	 * @param {MouseEvent} e
	 */
	const preview = (a,e) => {
		const id = a.getAttribute('href').substring(1);
		const li = document.getElementById(id);

		/** @type {HTMLLIElement} */
		const clone = li.cloneNode(true);
		clone.querySelector('.footnote-backref').remove();

		const div = document.createElement('div');
		div.replaceChildren(...clone.childNodes);
		div.classList.add('footnote-preview');
		document.body.appendChild(div);

		const rca = a.getBoundingClientRect();
		const rcd = div.getBoundingClientRect();

		// 锚点相对于盒子的固定向右偏移量。
		// 14.14: sqrt(10^2+10^2)
		const anchorFixedOffset = 14.14;
		// 默认位置：a的正上方（左边对齐）。
		let left = rca.left+rca.width/2, top = rca.top-rcd.height-10;
		// 如果超出屏幕，向左挪。锚点向右挪。
		let leftOffset = 0;
		if(left + rcd.width > window.innerWidth) {
			// - 10: 任意量，右边不要太靠近屏幕边界。
			const newLeft = window.innerWidth - rcd.width - 10;
			leftOffset = newLeft - left;
			left = newLeft;
		}
		div.style.left = `${left-anchorFixedOffset}px`;
		div.style.top = `${top}px`;
		div.style.setProperty('--anchor-left', `${-leftOffset+anchorFixedOffset/2}px`);
		
		return div;
	};
	/** @type {HTMLAnchorElement[]} */
	const refs = document.querySelectorAll('a.footnote-ref');
	refs.forEach(r => {
		/** @type {HTMLDivElement|null} */
		let elem = null;
		let withinOld = false;
		let withinNew = false;
		r.addEventListener('mouseenter', (e) => {
			withinOld = true;
			if(!elem) {
				elem = preview(r, e);
				elem.addEventListener('mouseenter', ()=>{
					withinNew = true;
				});
				elem.addEventListener('mouseleave', ()=>{
					withinNew = false;
				});
			}
		});
		r.addEventListener('mouseleave', () => {
			withinOld = false;

			if(!elem) {return;}

			const timer = setInterval(()=>{
				if(!withinOld && !withinNew) {
					elem?.remove();
					elem = null;
					clearInterval(timer);
				}
			}, 500);
		});
	});
}, {once:true});
