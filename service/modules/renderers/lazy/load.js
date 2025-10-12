// 视频元素没有 lazy 加载，自动 video.load() 在包含多篇文章的页面上体验不好（大量加载）。
// 这里用观察者模式监测视频元素的位置手动触发加载。
document.addEventListener('DOMContentLoaded', ()=>{
	const videos = document.querySelectorAll('video');
	if (videos.length <= 0) return;

	const observer = new IntersectionObserver(events=> {
		events.forEach(event=>{
			const video = event.target;
			if (event.isIntersecting) {
				observer.unobserve(video);

				if (video.preload == 'none') {
					video.preload = 'metadata';
				}

				video.load();

				console.log('进入视野：', video);
			}
		});
	});

	videos.forEach(v => {
		observer.observe(v);
	});
}, {once: true});
