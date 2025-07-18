document.addEventListener('DOMContentLoaded', () => {
	document.querySelectorAll('.live-photo').forEach(livePhoto => {
		const container = livePhoto.querySelector('.container');
		const icon = livePhoto.querySelector('.icon');
		const video = container.querySelector('video');
		const image = container.querySelector('img');
		const warning = livePhoto.querySelector('.warning');

		// TODO 优化：等图片可用的时候再把视频显示出来，测试出现
		// 过视频比图片先加载完成，导致 flicker。

		// 尽可能修复 Safari 无法播放的问题
		// if (/WebKit/.test(navigator.userAgent)) {
		// NOTE: 为了避免预加载，不应该提前 load。
		// video.load();
		// }

		// fix: 鼠标进入 → 开始加载 → 鼠标离开（加载成功前） → 加载失败。
		let within = false;

		const start = async (e) => {
			e.stopPropagation();
			e.preventDefault();

			within = true;

			try {
				video.currentTime = 0;
				await video.play();
				livePhoto.classList.add('zoom');
			}
			catch(e) {
				console.log(e);
				if (within && e instanceof DOMException) {
					if (['NotAllowedError','AbortError'].includes(e.name)) {
						warning.innerText = '浏览器未允许视频自动播放权限，无法播放实况照片。';
					} else if (['NotSupportedError'].includes(e.name)) {
						warning.innerText = '视频未加载完成或浏览器不支持播放此视频格式。';
					} else {
						warning.innerText = `其它错误：${e}`;
					}
					warning.classList.add('show');
				}
			}
		};

		const leave = (e) => {
			livePhoto.classList.remove('zoom');
			warning.classList.remove('show');

			// await play() 可能一直卡住不返回。
			// 在 pause 之前设置，如果  await play() 还没
			// 成功返回，就会进入异常处理。
			within = false;

			video.pause();
		};

		icon.addEventListener('mouseenter',   start);
		icon.addEventListener('mouseleave',   leave);

		image.addEventListener('touchstart',  start);
		image.addEventListener('touchend',    leave);
		image.addEventListener('touchcancel', leave);

		video.addEventListener('ended', () => {
			livePhoto.classList.remove('zoom');
		});
	});
});
