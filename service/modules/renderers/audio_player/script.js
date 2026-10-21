function handleAudio(img) {
	let player = img;
	while(player && !player.classList.contains('audio-player')) {
		player = player.parentElement;
	}
	let audio = player.querySelector(':scope audio');
	let time = player.querySelector('.time');
	audio.addEventListener('play', e => {
		player.classList.add('playing');
	});
	audio.addEventListener('pause', e => {
		player.classList.remove('playing');
	});
	let displaySeconds = t => {
		let seconds = Math.floor(t+.5);
		let minutes = Math.floor(seconds / 60);
		seconds = Math.floor(seconds % 60);
		return `${minutes<10?'0'+minutes:minutes}:${seconds<10?'0'+seconds:seconds}`;
	};
	audio.addEventListener('loadedmetadata', e => {
		console.log(audio.duration);
		const value = displaySeconds(audio.duration);
		time.innerText = value;
		time.setAttribute('title', `总时长：${value}`);
	});
	let progress = player.querySelector('.progress');
	progress.style.touchAction = 'none';
	let pointerID = null;
	let progressDebouncing = null;
	const trackingTime = () => {
		if (!Number.isFinite(audio.duration) || audio.duration <= 0 || +progress.max <= 0) { return NaN; }
		return +progress.value / +progress.max * audio.duration;
	};
	const seek = () => {
		clearTimeout(progressDebouncing);
		progressDebouncing = null;
		const tracking = trackingTime();
		if (!Number.isFinite(tracking)) { return; }
		audio.currentTime = tracking;
		time.innerText = displaySeconds(tracking);
	};
	progress.addEventListener('pointerdown', e => {
		if (pointerID !== null || e.button !== 0 || e.isPrimary === false) { return; }
		pointerID = e.pointerId;
		progress.setPointerCapture(pointerID);
	});
	const finish = e => {
		if (e.pointerId !== pointerID) { return; }
		pointerID = null;
		seek();
		if (progress.hasPointerCapture(e.pointerId)) { progress.releasePointerCapture(e.pointerId); }
	};
	for (const name of ['pointerup', 'pointercancel', 'lostpointercapture']) {
		progress.addEventListener(name, finish);
	}
	progress.addEventListener('change', seek);
	progress.addEventListener('input', e=> {
		const tracking = trackingTime();
		if (!Number.isFinite(tracking)) { return; }
		time.innerText = displaySeconds(tracking);
		clearTimeout(progressDebouncing);
		if (audio.paused) { seek(); return; }
		progressDebouncing = setTimeout(seek, 250);
	});
	audio.addEventListener('timeupdate', e => {
		if (pointerID === null && progressDebouncing === null && Number.isFinite(audio.duration) && audio.duration > 0) {
			progress.value = audio.currentTime / audio.duration * 100;
			time.innerText = displaySeconds(audio.currentTime);
		}
	});
	let play =player.querySelector('.play');
	let pause = player.querySelector('.pause');
	play.addEventListener('click', async () => {
		try { await audio.play(); }
		catch (error) { console.error('音频播放失败：', error); }
	});
	pause.addEventListener('click', e=>{ audio.pause(); })
	img.remove();
}
