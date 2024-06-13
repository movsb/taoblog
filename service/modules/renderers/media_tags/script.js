function handleAudio(r) {
	let player = document.querySelector(`#${r}`);
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
	let capturing = false;
	progress.addEventListener('mousedown', e=>{
		capturing = true;
	}, true);
	progress.addEventListener('mouseup', e=>{
		capturing = false;
		audio.play();
	}, true);
	let progressDebouncing = undefined;
	progress.addEventListener('input', e=> {
		let tracking = +progress.value / +progress.max * audio.duration;
		time.innerText = displaySeconds(tracking);
		let set = () => audio.currentTime = tracking;
		if (audio.paused) { set(); return; }
		clearTimeout(progressDebouncing);
		progressDebouncing = setTimeout(set, 250);
	});
	audio.addEventListener('timeupdate', e => {
		if (!capturing) {
			progress.value = audio.currentTime / audio.duration * 100;
			time.innerText = displaySeconds(audio.currentTime);
		}
	});
	let play =player.querySelector('.play');
	let pause = player.querySelector('.pause');
	play.addEventListener('click', e=>{ audio.play(); })
	pause.addEventListener('click', e=>{ audio.pause(); })
}
