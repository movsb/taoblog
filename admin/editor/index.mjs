import { PostFormUI } from './post-form.mjs';

document.addEventListener('DOMContentLoaded', () => {
	const post = TaoBlog.posts[TaoBlog.post_id];
	const form = new PostFormUI();
	form.initFrom(post);
	console.log('form', form);

	TaoBlog.vim.bind('[a]s', async ()=>{
		await form.triggerSubmit();
	});
}, {once: true});
