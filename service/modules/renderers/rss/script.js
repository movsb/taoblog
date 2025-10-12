document.addEventListener('DOMContentLoaded', function () {
	document.querySelectorAll('.rss-post-list').forEach(function (list) {
		list.addEventListener('click', async function (event) {
			if (event.target.classList.contains('mark-read')) {
				let parent = event.target.parentElement;
				while (parent && !parent.classList.contains('rss-post')) {
					parent = parent.parentElement;
				}
				if (parent) {
					let a = parent.querySelector('a');
					let url = a.getAttribute('href');
					let r = `${url}&r=1`;
					await fetch(r);
					parent.classList.add('read');
				}
			}
		});
	});
}, {once: true});
