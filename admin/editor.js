class PostManagementAPI
{
	constructor() { }

	// 创建一条文章。
	async createPost(source, time) {
		let path = `/v3/posts`;
		let rsp = await fetch(path, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				date: time,
				type: 'tweet',
				status: 'public',
				source: source,
				source_type: 'markdown',
				status: 'public',
			}),
		});
		if (!rsp.ok) {
			throw new Error('发表失败：' + await rsp.text());
		}
		let c = await rsp.json();
		console.log(c);
		return c;
	}

	// 更新/“编辑”一条已有评论。
	// 返回更新后的评论项。
	// 参数：id        - 评论编号
	// 参数：source    - 评论 markdown 原文
	async updatePost(id, modified, source, created) {
		let path = `/v3/posts/${id}`;
		let rsp = await fetch(path, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				post: {
					source_type: 'markdown',
					source: source,
					date: created,
					modified: modified,
				},
				update_mask: 'source,sourceType,created'
			})
		});
		if (!rsp.ok) { throw new Error('更新失败：' + await rsp.text()); }
		let c = await rsp.json();
		console.log(c);
		return c;
	}
}

class PostFormUI {
	constructor() {
		this._form = document.querySelector('form');
	}

	get elemSource()    { return this._form['source'];  }
	get elemTime()      { return this._form['time'];    }

	get source()    { return this.elemSource.value;     }
	get time()      {
		let t = this.elemTime.value;
		let d = new Date(t).getTime() / 1000;
		if (!d) {
			d = new Date().getTime() / 1000;
		}
		d = Math.floor(d);
		return d;
	}
	set time(t)      {
		const convertToDateTimeLocalString = (date) => {
			const year = date.getFullYear();
			const month = (date.getMonth() + 1).toString().padStart(2, "0");
			const day = date.getDate().toString().padStart(2, "0");
			const hours = date.getHours().toString().padStart(2, "0");
			const minutes = date.getMinutes().toString().padStart(2, "0");
		  
			return `${year}-${month}-${day}T${hours}:${minutes}`;
		};
		this.elemTime.value = convertToDateTimeLocalString(new Date(t));
	}

	set source(v)   { this.elemSource.value = v;        }

	submit(callback) {
		let submit = document.querySelector('input[type=submit]');
		submit.addEventListener('click', (e) => {
			e.preventDefault();
			e.stopPropagation();
			callback();
		});
	}
}

let postAPI = new PostManagementAPI();
let formUI = new PostFormUI();
formUI.submit(async () => {
	try {
		let post = undefined;
		if (_post_id > 0) {
			post = await postAPI.updatePost(_post_id, _modified, formUI.source, formUI.time);
		} else {
			post = await postAPI.createPost(formUI.source, formUI.time);
		}
		alert('成功。');
		window.location = `/${post.id}/`;
	} catch(e) {
		alert(e);
	}
});
if (typeof _created == 'number') {
	formUI.time = _created*1000;
}
