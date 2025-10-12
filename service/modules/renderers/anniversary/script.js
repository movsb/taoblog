function setupAnniversary(){

const date = new Date().toDateString();

if (!date.includes('Dec 24 2024') && !date.includes('Dec 25 2024')) {
	return;
}

if (location.pathname != '/') {
	return;
}

const canvas = document.createElement('canvas');
canvas.classList.add('anniversary');
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;
document.body.appendChild(canvas);

const ctx = canvas.getContext('2d');
const emojis = ['🎂', '🎄'];

const particles = [];

class Particle {
	constructor(x, y, emoji) {
		this.x = x;
		this.y = y;
		this.size = Math.random() * 40 + 20;
		this.speed = Math.random() + 1;
		this.emoji = emoji;
	}
	draw() {
		ctx.font = `${this.size}px Arial`;
		ctx.textAlign = 'center';
		ctx.fillText(this.emoji, this.x, this.y);
	}
	update() {
		this.y += this.speed;
		if (this.y > canvas.height) {
			this.y = -this.size;
			this.x = Math.random() * canvas.width;
		}
	}
}

function init() {
	for (let i = 0; i < 30; i++) {
		const randomEmoji = emojis[Math.floor(Math.random() * emojis.length)];
		const x = Math.random() * canvas.width;
		const y = Math.random() * canvas.height - canvas.height;
		particles.push(new Particle(x, y, randomEmoji));
	}
}

let timer = undefined;
timer = setTimeout(function(){
	clearTimeout(timer);
	timer = undefined;
}, 10000);

function animate() {
	ctx.clearRect(0, 0, canvas.width, canvas.height);
	particles.forEach(particle => {
		particle.update();
		particle.draw();
	});
	if (timer) {
		requestAnimationFrame(animate);
	} else {
		canvas.remove();
	}
}

init();
animate();

window.addEventListener('resize', () => {
	particles.length = 0;
	canvas.width = window.innerWidth;
	canvas.height = window.innerHeight;
	init();
});

}

document.addEventListener('DOMContentLoaded', setupAnniversary, {once: true});
