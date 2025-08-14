// 桌面设备上的模糊效果始终靠鼠标移除。
//
// 对于移动设备，模糊效果在点击图片时移除，
// 然后在指定时间后设置回来。
document.addEventListener(`DOMContentLoaded`, () => {
	const isMobileDevice = 'ontouchstart' in window || /iPhone|iPad|Android|XiaoMi/.test(navigator.userAgent);
	if (!isMobileDevice) { return; }

	const images = document.querySelectorAll('img.blur');
	images.forEach(img => {
		img.addEventListener('click', (e) => {
			if (!img.classList.contains('blur')) {
				// 如果已经移除模糊效果，则不再处理。
				return;
			}
			e.preventDefault();
			e.stopPropagation();
			img.classList.add('blur-removed');
			img.classList.remove('blur');
			setTimeout(()=>{
				img.classList.add('blur');
				img.classList.remove('blur-removed');
			}, 3000);
		}, { capture: true });
	});
});
