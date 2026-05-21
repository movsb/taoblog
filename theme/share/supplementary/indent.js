// TODO: 没有选区的时候不能取消缩进。
class TextareaWithTab {
	/**
	 * 
	 * @param {HTMLTextAreaElement} editor 
	 */
	constructor(editor) {
		/** @type {HTMLTextAreaElement} */
		this.editor = editor;
		this.editor.addEventListener('keydown', (event)=>{
			this.handleTab(event);
			
			if(event.key == 'Enter' && this.handleEnter(event)) {
				event.preventDefault();
				return;
			}
		});
	}

	replace(start, end, text) {
		this.editor.focus();

		// 方案 A：优先用 execCommand，移动端（含 iOS Safari）更容易进撤销栈
		const canExec = typeof document.execCommand === 'function' &&
						(document.queryCommandSupported?.('insertText') ?? true);
		if (canExec) {
			this.editor.setSelectionRange(start, end);
			const ok = document.execCommand('insertText', false, text);
			if (ok) return;
		}

		// 方案 B：回退到 setRangeText（在桌面浏览器通常也能进入撤销栈）
		if (typeof this.editor.setRangeText === 'function') {
			this.editor.setRangeText(text, start, end, 'preserve');
			return;
		}

		// 最后兜底（不保 undo）
		this.editor.value = this.editor.value.slice(0, start) + text + this.editor.value.slice(end);
	}

	/**
	 * 
	 * @param {KeyboardEvent} e 
	 */
	handleEnter(e) {
		let start = this.editor.selectionStart;
		let end = this.editor.selectionEnd;
		// 必须在没有选区的时候尝试自动缩进。
		if(start != end) return false;

		const content = this.editor.value;

		// 按下回车的时候还没有插入，拷贝当前行前面的空白内容。
		let lineStart = start;
		let lastNonWhitespace = start;
		while(lineStart > 0) {
			const c = content[lineStart-1];
			if(c == '\n') { break; }
			if(c != ' ' && c != '\t') {
				lastNonWhitespace = lineStart-1;
			}
			lineStart--;
		}

		const prefix = content.slice(lineStart, lastNonWhitespace);
		this.replace(end, end, "\n"+prefix);
		return true;
	}

	handleTab(e) {
		if (e.key !== 'Tab') return;
		e.preventDefault();

		const content = this.editor.value;
		let start = this.editor.selectionStart;
		let end = this.editor.selectionEnd;
		let selection = content.slice(start, end);

		// 1. 如果是多行选区，则应该选区的每一行都缩进。
		//    所以这里把选区扩展到行首和行尾，并重新选区。
		const multi = selection.includes('\n');
		if (multi) {
			while (start > 0 && content[start - 1] !== '\n') start--;
			while (end < content.length && content[end] !== '\n') end++;
			selection = content.slice(start, end);
		}

		const isShift = e.shiftKey;
		const modified = selection
			.split('\n')
			.map(line => isShift ? line.replace(/^  |^\t/, '') : ('  ' + line))
			.join('\n');

		this.replace(start, end, modified);

		if (multi) {
			// 维持修改后的整段为选区，体验更接近编辑器
			const newEnd = start + modified.length;
			this.editor.setSelectionRange(start, newEnd);
		}
	}
}
