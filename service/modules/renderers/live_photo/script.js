document.addEventListener('DOMContentLoaded', () => {
	document.querySelectorAll('.live-photo').forEach(livePhoto => {
		const container = livePhoto.querySelector('.container');
		const icon = livePhoto.querySelector('.icon');
		const video = container.querySelector('video');
		const image = container.querySelector('img');
		const warning = livePhoto.querySelector('.warning');

		// TODO 优化：等图片可用的时候再把视频显示出来，测试出现
		// 过视频比图片先加载完成，导致 flicker。

		const start = async (e) => {
			e.stopPropagation();
			e.preventDefault();
			try {
				video.currentTime = 0;
				await video.play();
				container.classList.add('zoom');
			}
			catch(e) {
				console.log(e);
				if (e instanceof DOMException) {
					if (['NotAllowedError','AbortError'].includes(e.name)) {
						warning.innerText = '浏览器未允许视频自动播放权限，无法播放实况照片。';
					} else if (['NotSupportedError'].includes(e.name)) {
						warning.innerText = '视频错误，无法播放。';
					}
					warning.classList.add('show');
				}
			}
		};

		const leave = (e) => {
			container.classList.remove('zoom');
			warning.classList.remove('show');
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
