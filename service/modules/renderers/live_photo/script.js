document.addEventListener('DOMContentLoaded', () => {
	document.querySelectorAll('.live-photo').forEach(livePhoto => {
		const container = livePhoto.querySelector('.container');
		const icon = livePhoto.querySelector('.icon');
		const video = container.querySelector('video');
		const image = container.querySelector('img');

		const start = async (e) => {
			e.stopPropagation();
			e.preventDefault();
			container.classList.add('zoom');
			video.currentTime = 0;
			try { await video.play(); }
			catch(e) { console.log(e); }
		};

		const leave = (e) => {
			container.classList.remove('zoom');
			video.pause();
		};

		icon.addEventListener('mouseenter',   start);
		icon.addEventListener('mouseleave',   leave);

		image.addEventListener('touchstart',  start);
		image.addEventListener('touchend',    leave);
		image.addEventListener('touchcancel', leave);

		video.addEventListener('ended', () => {
			container.classList.remove('zoom');
		});
	});
});
