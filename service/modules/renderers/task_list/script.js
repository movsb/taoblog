// 文章编号目前来自于两个地方：
// <article id="xxx">
// TaoBlog.post_id = 1;
document.addEventListener('DOMContentLoaded', function() {
	let taskLists = document.querySelectorAll('ul.task-list');
	taskLists.forEach(list=>{
		list.querySelectorAll('.task-list-item > input[type=checkbox], .task-list-item > p > input[type=checkbox]').forEach(e => e.disabled = "");
	});
	taskLists.forEach(list =>list.addEventListener('click', async e => {
		let checkBox = e.target;
		if (checkBox.tagName != 'INPUT' || checkBox.type != 'checkbox') { return; }

		// 禁止父级任务列表响应事件。
		e.stopPropagation();
		// 没用，进入到 click 事件的时候已经状态改变了，只是没表现在界面上。
		// e.preventDefault();
		// 暂时取消 check，防止状态改变失败的时候界面上已经变化了。
		checkBox.checked = !checkBox.checked;

		let listItem = checkBox.parentElement;
		if (listItem.tagName == 'P') listItem = listItem.parentElement;
		if (!listItem.classList.contains('task-list-item')) { return; }
		let position = listItem.getAttribute('data-source-position');
		position = parseInt(position);
		if (!position) { return; }

		// 可能是文章、可能是评论。
		// 但是肯定先到达评论外层，然后才会到达文章。
		let id = undefined;
		let isPost = false;

		let node = listItem;
		while (node) {
			if (node.tagName == 'ARTICLE') {
				id = parseInt(node.getAttribute('id'));
				if (!id) { id = TaoBlog.post_id; }
				isPost = true;
				break
			} else if (node.classList.contains('comment-li')) {
				id = parseInt(node.getAttribute('id').split('-')[1]);
				isPost = false;
				break
			}
			node = node.parentElement;
		}
		if (!node) { return; } // 不应该。
		
		if (!id) {
			alert('没有找到对应的编号，不可操作任务。');
			return;
		}

		let willCheck = !checkBox.checked;

		if (!confirm(`确定${willCheck  ? "" : "取消"}完成任务？`)) {
			return;
		}

		let obj = isPost
			? TaoBlog.posts[id] 
			: TaoBlog.comments.find(c => c.id == id);
		if (!obj) {
			alert('意外错误。');
			return;
		}

		const api = new PostManagementAPI();
		let checks = [], unchecks = [];
		willCheck ? checks.push(position) : unchecks.push(position);
		try {
			let updated = await api.checkTaskListItems(id, isPost, obj.modified, checks, unchecks);
			obj.modified = updated.modification_time;
			if (willCheck) {
				listItem.classList.add('checked');
			} else {
				listItem.classList.remove('checked');
			}
			checkBox.checked = !checkBox.checked;
		} catch(e) {
			alert('任务更新失败：' + e.message ?? e);
		} finally {
		}
	}));
}, {once: true});
