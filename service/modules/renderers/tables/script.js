document.addEventListener('DOMContentLoaded', ()=>{
	/** @type {HTMLDivElement[]} */
	const tables = document.querySelectorAll('.table-wrapper');
	tables.forEach(wrap => {
		const table = wrap.querySelector('table');
		if(table.clientWidth > wrap.clientWidth || table.clientHeight > wrap.clientHeight) {
			const toolbar = document.createElement('div');
			toolbar.classList.add('toolbar');
			toolbar.insertAdjacentHTML('afterbegin', '<svg width="800px" height="800px" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M10 15H15V10H13.2V13.2H10V15ZM6 15V13.2H2.8V10H1V15H6ZM10 2.8H12.375H13.2V6H15V1H10V2.8ZM6 1V2.8H2.8V6H1V1H6Z"/></svg>');
			toolbar.title = '全屏';
			toolbar.addEventListener('click', ()=>{
				if(wrap.classList.toggle('fullscreen')) {
					/**
					 * 
					 * @param {KeyboardEvent} e 
					 */
					const escape = (e) => {
						console.log('awaiting to remove table escape handler');
						if(wrap.classList.contains('fullscreen')) {
							if(e.key == 'Escape') {
								e.preventDefault();
								e.stopImmediatePropagation();
								wrap.classList.remove('fullscreen');
								document.removeEventListener('keydown', escape);
							}
						} else {
							document.removeEventListener('keydown', escape);
						}
					};
					document.addEventListener('keydown', escape);
				}
			});
			wrap.insertAdjacentElement('afterbegin', toolbar);
		}
	});
}, {once: true});
