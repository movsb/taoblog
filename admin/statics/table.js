class Table {
	constructor() {
		this.table = document.createElement('table');
		document.body.appendChild(this.table);

		/** @type {HTMLTableCellElement | null} */
		this.curCell = null;

		/** @type {HTMLTableCellElement[]} */
		// 始终为从左上到右下的顺序。
		this.selectedCells = [];
	
		this.table.addEventListener('click', (e) => {
			if (!this._isCell(e.target)) {
				return;
			}
			this._selectCell(e.target);
		});
		this.table.addEventListener('dblclick', (e) => {
			if (!this._isCell(e.target)) {
				return;
			}
			// console.log('双击');
			/** @type {HTMLTableCellElement} */
			const cell = e.target;
			if(cell == this.curCell && this._isEditing(cell)) {
				return;
			} 
			this._edit(cell, true);
		});
	
		this.table.addEventListener('mousedown', e => {
			// console.log('mousedown:', e.target);
			if(!this._isCell(e.target)) {
				// console.log('mousedown not on cell');
				return;
			}

			const startCell = e.target;

			const moveHandler = e => {
				// console.log('mousemove:', e.target);
				if (!this._isCell(e.target)) {
					// console.log('mousemove not on cell');
					return;
				}
				// 防止在同一个元素内移动时因频繁 clearSelection 导致失去编辑焦点。
				if (e.target == startCell) {
					return;
				}
				const table = e.target.closest('table');
				if (!table || table != this.table) {
					// console.log('mousemove on another table, ignoring');
					return;
				}
				const endCell = e.target;
				if(!this._selectRange(startCell, endCell)) {
					// console.log('选区无效。');
				}
			};

			document.addEventListener('mousemove', moveHandler);
			document.addEventListener('mouseup', ()=>{
				// console.log('mouseup:', e.target);
				document.removeEventListener('mousemove', moveHandler);
			}, { once: true });
		});
	}

	/**
	 * 
	 * @param {Number} rows 
	 * @param {Number} cols 
	 */
	reset(rows, cols) {
		if(!rows && !cols) {
			rows = this._rows;
			cols = this._cols;
		} else {
			this._rows = rows;
			this._cols = cols;
		}

		this.clearSelection();
		this.table.innerHTML = '';

		for(let i=0; i<rows; i++) {
			const tr = this.table.insertRow();
			for(let j=0; j<cols; j++) {
				const td = tr.insertCell();
			}
		}

		this._calcCoords();
	}

	remove() {
		this.table.remove();
	}

	/** @returns {string} */
	get content() {
		return this.table.outerHTML;
	}

	_forEachCell(callback) {
		Array.from(this.table.rows).forEach(row => {
			Array.from(row.cells).forEach(cell => {
				callback(cell);
			});
		});
	}

	/**
	 * 
	 * @param {HTMLTableCellElement} cell 
	 * @returns {{r1: Number,c1: Number,r2: Number,c2: Number}}
	 */
	_getCoords(cell) {
		return cell._coords;
	}
	_setCoords(cell, coords) {
		cell._coords = coords;
	}
	
	// 如果设置成功，返回 true。
	/**
	 * @param {Number} r1 
	 * @param {Number} c1 
	 * @param {Number} r2 
	 * @param {Number} c2 
	 */
	selectRange(r1,c1,r2,c2) {
		const cell1 = this.findCell(r1, c1);
		const cell2 = this.findCell(r2, c2);
		return this._selectRange(cell1, cell2);
	}

	/**
	 * @param {HTMLTableCellElement} cell1 
	 * @param {HTMLTableCellElement} cell2 
	 */
	_selectRange(cell1, cell2) {
		// console.log('selectRange:', cell1, cell2);

		let cc1 = this._getCoords(cell1);
		let cc2 = this._getCoords(cell2);

		if (!(cc1.r1 <= cc2.r1 && cc1.c1 <= cc2.c1)) { // ! ➡️↘️⬇️
			if(cc1.c1 == cc2.c1 || cc1.r1 == cc2.r1) { //   ⬆️⬅️
				const t = cell1;
				cell1 = cell2;
				cell2 = t;
				cc1 = this._getCoords(cell1);
				cc2 = this._getCoords(cell2);
			} else if (cc1.c1 > cc2.c1) { // ️↙️↖️
				return this.selectRange(cc1.r1, cc2.c1, cc2.r1, cc1.c1);
			} else { // ↗️
				return this.selectRange(cc2.r1, cc1.c1, cc1.r1, cc2.c1);
			}
		}

		this.clearSelection();

		let valid = true;

		Array.from(this.table.rows).forEach(row=> {
			Array.from(row.cells).forEach(cell=> {
				const cc = this._getCoords(cell);
				let some = false, all = true;

				// 被包含元素必须被完整包含。
				for(let i=cc.r1; i<=cc.r2; i++) {
					for(let j=cc.c1; j<=cc.c2; j++) {
						const within = cc1.r1 <= i && i <= cc2.r2 && cc1.c1 <= j && j <= cc2.c2;
						some |= within;
						all  &= within;
					}
				}

				if(!some) { return; }
				if(all) {
					this._highlight(cell, true);
					this.selectedCells.push(cell);
				} else {
					valid = false;
				}
			});
		});

		if(!valid) {
			this.clearSelection();
		}

		return valid;
	}

	_isCell(element) {
		return element.tagName == 'TD' || element.tagName == 'TH';
	}

	/**
	 * 
	 * @param {Number} row 
	 * @param {Number} col 
	 */
	selectCell(row, col) {
		const cell = this.findCell(row, col);
		this._selectCell(cell);
		return cell;
	}

	/** @param {HTMLTableCellElement} col */
	_selectCell(cell) {
		if(cell == this.curCell) {
			if(this._isEditing(cell)) return;
			else return this.clearSelection();
		}

		if(this.curCell) {
			this._highlight(this.curCell, false);
			this._edit(this.curCell, false);
		}

		this.clearSelection();

		this.curCell = cell;
		this._highlight(cell, true);
	}

	clearSelection() {
		if(this.curCell) {
			this._highlight(this.curCell, false);
			this._edit(this.curCell, false);
		}
		this.curCell = null;
		this.selectedCells.forEach(cell => {
			this._highlight(cell, false);
			this._edit(cell, false);
		})
		this.selectedCells = [];
	}

	/**
	 * 
	 * @param {HTMLTableCellElement} cell 
	 * @param {boolean} on 
	 */
	_highlight(cell, on) {
		if(on) cell.classList.add('selected');
		else cell.classList.remove('selected');
	}

	/**
	 * 
	 * @param {HTMLTableCellElement} cell 
	 * @param {boolean} on 
	 */
	_edit(cell, on) {
		if(on) {
			cell.contentEditable = 'plaintext-only';
			cell.classList.add('editing');
			cell.focus();

			const range = document.createRange();
			range.selectNodeContents(cell);
			const selection = window.getSelection();
			selection.removeAllRanges();
			selection.addRange(range);
		} else {
			// console.log('删除属性，移除类名', cell);
			cell.removeAttribute('contentEditable');
			cell.classList.remove('editing');
			// 会不会有误清除？
			window.getSelection().removeAllRanges();
		}
	}

	/**
	 * 
	 * @param {HTMLTableCellElement} cell 
	 * @returns {Boolean}
	 */
	_isEditing(cell) {
		return cell.classList.contains('editing');
	}

	/** @returns {HTMLTableCellElement | null} */
	findCell(r,c) {
		let ret = null;
		this._forEachCell(cell => {
			const cc = this._getCoords(cell);
			if(cc.r1 <= r && r <= cc.r2 && cc.c1 <= c && c <= cc.c2) {
				ret = cell;
				return;
			}
		});
		return ret;
	}

	addRowAbove() { return this._addRow('above'); }
	addRowBelow() { return this._addRow('below'); }

	_addRow(position) {
		if (!this.curCell) {
			alert('Please select a cell first.');
			return;
		}

		/** @type {HTMLTableRowElement} */
		const row = this.curCell.parentElement;
		const curCell = this.curCell;

		// 计算待插入行的逻辑行号。
		// 初始化为上方（above）插入。
		// 如果是下方，需要根据 rowspan 计算。
		let newRowIndex = position == 'above' 
			? row.rowIndex
			: row.rowIndex + curCell.rowSpan;

		// 如果上方或正文的行没有 colspan，则 maxCols 代表本来应该插入的列数。
		// 但是实际可能存在 colspan 和 rowspan，插入数量需要重新计算（更少）。
		const maxCols = this._maxCols();

		// 如果是第一行或最后一行，则不需要计算。
		if (newRowIndex == 0 || newRowIndex == this.table.rows.length) {
			const tr = this.table.insertRow(newRowIndex);
			for(let i=0; i<maxCols; i++) {
				tr.insertCell();
			}
		} else {
			// const refRow = this.table.rows[newRowIndex];
			// 计算待插入行的实际构成。

			let insertCount = 0;

			for(let i=0; i<maxCols; /*i++*/) {
				const rr = newRowIndex+1, rc = i+1;
				const cell = this.findCell(rr, rc);
				const cc = this._getCoords(cell);
				// 该单元格由自己组成。
				if(cc.r1 == cc.r2) {
					insertCount++;
					i += 1;
					continue;
				}
				// 由上面单元格的最下面的构成 || 由下面单元格的最上面的构成。
				if (position == 'above' && (cc.r1 == rr /*|| cc.r2 == rr*/) || position == 'below' && (cc.r1 == rr)) {
					insertCount++;
					i += 1;
					continue;
				}
				// 其它情况：扩展原有单元格。
				cell.rowSpan += 1;
				i += cell.colSpan;
			}

			const tr = this.table.insertRow(newRowIndex);
			for(let i=0; i<insertCount; i++) {
				tr.insertCell();
			}
		}

		this._calcCoords();
	}

	addColLeft()  { return this._addCol('left');  }
	addColRight() { return this._addCol('right'); }

	/** @param {string} position  */
	_addCol(position) {
		if (!this.curCell) {
			alert('Please select a cell first.');
			return;
		}

		/** @type {HTMLTableCellElement} */
		const cell = this.curCell;
		const cc = this._getCoords(cell);

		// 如果是第一列，不需要计算。
		// 如果是最后一列，直接追加。
		if (cc.c1 == 1 && position == 'left' || cc.c2 == this._maxCols() && position == 'right') {
			const rows = this.table.rows.length;
			for(let i=0; i<rows; i++) {
				const row = this.table.rows[i];
				row.insertCell(position=='left' ? 0 : -1);
			}
			this._calcCoords();
			return;
		}

		const left = position == 'left';
		const newColIndex = left ? cc.c1 - 1 : cc.c2;

		const rows = this.table.rows.length;
		for(let i=0; i<rows; i++) {
			const row = this.table.rows[i];
			const rr = i+1, rc = newColIndex+1;
			const cell = this.findCell(rr, rc);
			const cc = this._getCoords(cell);

			// 单元格由自己组成。
			if(cc.c1 == rc) {
				let pos = cell.cellIndex;
				// 上面插入过
				if(cell.rowSpan > 1 && i+1 != cc.r1) {
					pos--;
				}
				const td = row.insertCell(pos);
				this._setCoords(td, {
					r1: rr, c1: rc,
					r2: rr, c2: rc,
				});
				continue;
			}

			// 单元格由合并单元格组成，并且当前处在合并单元格的第一行。
			if(cc.r1 == i+1 && cc.c1 != rc) {
				cell.colSpan++;
				continue;
			}
		}

		this._calcCoords();
	}

	deleteRows() {
		const rows = [];
		if (this.curCell) {
			const cc = this._getCoords(this.curCell);
			for(let i=cc.r1; i<=cc.r2; i++) {
				rows.push(i);
			}
		} else if(this.selectedCells?.length > 0) {
			this.selectedCells.forEach(cell => {
				const cc = this._getCoords(cell);
				for(let i=cc.r1; i<=cc.r2; i++) {
					rows.push(i);
				}
			});
		}

		// descending
		const sorted = [...new Set(rows)].sort((a,b) => b-a);
		// console.log('deleteRows:', sorted);

		const maxCols = this._maxCols();
		sorted.forEach(r => {
			for(let c=1; c <= maxCols;) {
				const cell = this.findCell(r, c);
				const cc = this._getCoords(cell);
				// 单行元素。
				if (cell.rowSpan == 1) {
					c += cell.colSpan;
					continue;
				}
				// 向下展开。
				if (r == cc.r1) {
					this.selectCell(r, c);
					// 可以考虑再合并。
					this.split();
					c += cell.colSpan;
					continue;
				}
				// 来自上面。
				cell.rowSpan--;
				c += cell.colSpan;
			}
			this.table.deleteRow(r - 1);
		});

		this.clearSelection();
		this._calcCoords();
	}

	deleteCols() {
		const cols = [];
		if (this.curCell) {
			const cc = this._getCoords(this.curCell);
			for(let i=cc.c1; i<=cc.c2; i++) {
				cols.push(i);
			}
		} else if(this.selectedCells?.length > 0) {
			this.selectedCells.forEach(cell => {
				const cc = this._getCoords(cell);
				for(let i=cc.c1; i<=cc.c2; i++) {
					cols.push(i);
				}
			});
		}

		// descending
		const sorted = [...new Set(cols)].sort((a,b) => b-a);
		// console.log('deleteCols:', sorted);

		const rows = this.table.rows.length;
		sorted.forEach(c => {
			const toRemove = [];
			for(let r=1; r <= rows;) {
				const cell = this.findCell(r, c);
				const cc = this._getCoords(cell);
				// 单列元素。
				if (cell.colSpan == 1) {
					r += cell.rowSpan;
					toRemove.push(cell);
					continue;
				}
				// 向右展开。
				if(c == cc.c1) {
					this.selectCell(r, c);
					// 可以考虑再合并。
					this.split();
					// 修改过需要重新计算。
					this._calcCoords();
					// 拆分过了，只剩 1 行
					r += 1;
					toRemove.push(cell);
					continue;
				}
				// 来自左边。
				cell.colSpan--;
				r += cell.rowSpan;
			}
			toRemove.forEach(cell => cell.remove());
		});

		this.clearSelection();
		this._calcCoords();
	}

	merge() {
		if(this.selectedCells.length < 2) {
			alert('Please select at least two cells to merge.');
			return;
		}

		const firstCell = this.selectedCells[0];

		// 找最右最下的元素，并非一定是最后一个元素。
		// const lastCell = this.selectedCells[this.selectedCells.length - 1];

		let lastCell = firstCell;

		const firstCoords = this._getCoords(firstCell);
		let lastCoords = this._getCoords(lastCell);

		this.selectedCells.forEach(cell => {
			const cc = this._getCoords(cell);
			if (cc.r2 >= lastCoords.r2 && cc.c2 >= lastCoords.c2) {
				lastCell = cell;
				lastCoords = this._getCoords(lastCell);
			}
		});

		const rowSpan = lastCoords.r2 - firstCoords.r1 + 1;
		const colSpan = lastCoords.c2 - firstCoords.c1 + 1;

		// 移除所有其它元素。以第一个为合并标准。它总是位于最左上角位置，即第一个元素。
		for(let i=1; i<this.selectedCells.length; i++) {
			const cell = this.selectedCells[i];
			cell.remove();
		}

		if (rowSpan > 1) {
			firstCell.rowSpan = rowSpan;
		} else {
			firstCell.removeAttribute('rowspan');
		}
		if (colSpan > 1) {
			firstCell.colSpan = colSpan;
		} else {
			firstCell.removeAttribute('colspan');
		}

		// 合并后把当前单元格设置为第一个单元格。
		this._selectCell(firstCell);

		this._calcCoords();
	}

	split() {
		if (!this.curCell) {
			alert('Please select a cell first.');
			return;
		}

		const cell = this.curCell;
		if(cell.rowSpan == 1 && cell.colSpan == 1) {
			alert('not a merged cell');
			return;
		}

		const cc = this._getCoords(cell);
		const c1 = cc.c1;
		const rowCellIndices = [];
		// 每行都需要添加一个单元格。
		for(let i=cc.r1; i<=cc.r2; i++) {
			const row = this.table.rows[i-1];
			if(c1 == 1) {
				for(let k=0; k<cell.colSpan; k++) {
					// 最左上角的
					if(i==cc.r1 && k==0) {
						continue;
					}
					row.insertCell(k);
				}
				continue;
			}
			// 单元格可能是上面 rowspan 的的，需要向前找到第一个起始行为当前行的。
			for(let j=c1-1; j>=1; j--) {
				const left = this.findCell(i, j);
				const cl = this._getCoords(left);
				// 从上面挤下来的，或者左边还有元素。
				if(cl.r1 != i && cl.c1 > 1) {
					continue;
				}
				const cellIndices = [];
				for(let k=0; k<cell.colSpan; k++) {
					// 最左上角的
					if(i==cc.r1 && k==0) {
						continue;
					}
					const index = cl.r1 == i ? left.cellIndex+k+1 : k;
					cellIndices.push(index);
				}
				rowCellIndices.push(cellIndices);
				break;
			}
		}

		rowCellIndices.forEach((indices,relativeRowIndex) => {
			const realRowIndex = cc.r1 - 1 + relativeRowIndex;
			const row = this.table.rows[realRowIndex];
			indices.forEach(colIndex => {
				row.insertCell(colIndex);
			});
		});

		cell.removeAttribute('colspan');
		cell.removeAttribute('rowspan');

		this._calcCoords();
	}

	_maxCols() {
		let maxCol = 0;
		Array.from(this.table.rows).forEach(row=> {
			Array.from(row.cells).forEach(cell=> {
				const cc = this._getCoords(cell);
				maxCol = Math.max(maxCol, cc.c2);
			});
		});
		return maxCol;
	}

	_calcCoords() {
		// debugger;
		let calcC1 = (rowIndex, colIndex) => {
			let retry = (tr, tc) => {
				for(let x=0; x <= rowIndex; x++) {
					const cols = this.table.rows[x].cells.length;
					for(let y=0; y < cols; y++) {
						if(x == rowIndex && y == colIndex) {
							return tc;
						}
						const cc = this._getCoords(this.table.rows[x].cells[y]);
						if (cc.r1 <= tr && tr <= cc.r2 && cc.c1 <= tc && tc <= cc.c2) {
							tc++;
							return retry(tr, tc);
						}
					}
				}
			};

			const tr = rowIndex + 1;
			let tc = colIndex + 1;

			return retry(tr, tc);
		};

		Array.from(this.table.rows).forEach((row, rowIndex) => {
			Array.from(row.cells).forEach((col, colIndex) => {
				const p = {};

				p.r1 = rowIndex + 1;
				p.c1 = calcC1(rowIndex, colIndex);

				if(col.rowSpan == 0) {
					p.r2 = p.r1;
				} else {
					p.r2 = +p.r1 + col.rowSpan - 1;
				}

				if(col.colSpan == 0) {
					p.c2 = p.c1;
				} else {
					p.c2 = +p.c1 + col.colSpan - 1;
				}

				this._setCoords(col, p);

				col.innerText = `${p.r1},${p.c1}`;
			});
		});
	}
}

class TableTest {
	constructor() {
		/**
		 * @type {{ note: string, init: (t: Table) => void, html: string }[]}
		 */
		this.cases = [
			{
				init: t => { t.reset(2,2); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectCell(1,2); },
				html: '<table><tbody><tr><td>1,1</td><td class="selected">1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectCell(1,1); t.addRowAbove(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td class="selected">2,1</td><td>2,2</td></tr><tr><td>3,1</td><td>3,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectRange(1,2,2,2); t.merge(); },
				html: '<table><tbody><tr><td>1,1</td><td class="selected" rowspan="2">1,2</td></tr><tr><td>2,1</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectRange(1,2,2,2); t.merge(); t.selectCell(1,1); t.addRowAbove(); t.addRowBelow(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td class="selected">2,1</td><td class="" rowspan="3">2,2</td></tr><tr><td>3,1</td></tr><tr><td>4,1</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,1); t.selectRange(1,1,2,1); t.merge(); t.addColRight(); },
				html: '<table><tbody><tr><td class="selected" rowspan="2">1,1</td><td>1,2</td></tr><tr><td>2,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectRange(1,2,2,2); t.merge(); t.addColLeft(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td><td class="selected" rowspan="2">1,3</td></tr><tr><td>2,1</td><td>2,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(3,2); t.selectRange(1,1,3,1); t.merge(); t.selectCell(2,2); t.addColLeft(); },
				html: '<table><tbody><tr><td class="" rowspan="3">1,1</td><td>1,2</td><td>1,3</td></tr><tr><td>2,2</td><td class="selected">2,3</td></tr><tr><td>3,2</td><td>3,3</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(3,2); t.selectRange(1,1,3,1); t.merge(); t.selectCell(2,2); t.addColRight(); t.selectRange(2,2,2,3); t.merge(); t.addColLeft(); },
				html: '<table><tbody><tr><td class="" rowspan="3">1,1</td><td>1,2</td><td>1,3</td><td>1,4</td></tr><tr><td>2,2</td><td class="selected" colspan="2">2,3</td></tr><tr><td>3,2</td><td>3,3</td><td>3,4</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(3,3); t.selectRange(2,2,3,3); t.merge(); t.selectCell(1,2); t.addColLeft(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td><td class="selected">1,3</td><td>1,4</td></tr><tr><td>2,1</td><td>2,2</td><td class="" rowspan="2" colspan="2">2,3</td></tr><tr><td>3,1</td><td>3,2</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(3,3); t.selectRange(2,2,3,3); t.merge(); t.split(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td><td>1,3</td></tr><tr><td>2,1</td><td class="selected">2,2</td><td>2,3</td></tr><tr><td>3,1</td><td>3,2</td><td>3,3</td></tr></tbody></table>',
			},
			{
				init: t => { t.reset(2,2); t.selectRange(1,1,2,1); t.merge(); t.selectRange(1,2,2,2);  t.merge(); t.split(); },
				html: '<table><tbody><tr><td class="" rowspan="2">1,1</td><td class="selected">1,2</td></tr><tr><td>2,2</td></tr></tbody></table>',
			},
			{
				note: '删除行，单行元素',
				init: t => { t.reset(3,3); t.selectCell(1,2); t.deleteRows(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td><td>1,3</td></tr><tr><td>2,1</td><td>2,2</td><td>2,3</td></tr></tbody></table>',
			},
			{
				note: '删除行，多行元素，来自上面',
				init: t => { t.reset(3,3); t.selectRange(1,2,3,2); t.merge(); t.selectCell(3,1); t.deleteRows(); },
				html: '<table><tbody><tr><td>1,1</td><td class="" rowspan="2">1,2</td><td>1,3</td></tr><tr><td>2,1</td><td>2,3</td></tr></tbody></table>',
			},
			{
				note: '删除行，多行元素，向下展开',
				init: t => { t.reset(3,3); t.selectRange(1,2,3,2); t.merge(); t.selectCell(1,1); t.deleteRows(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td><td>1,3</td></tr><tr><td>2,1</td><td>2,2</td><td>2,3</td></tr></tbody></table>',
			},
			{
				note: '删除列，单列元素',
				init: t => { t.reset(3,3); t.selectCell(1,2); t.deleteCols(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr><tr><td>3,1</td><td>3,2</td></tr></tbody></table>',
			},
			{
				note: '删除列，多列元素，来自左边',
				init: t => { t.reset(3,3); t.selectRange(2,2,3,3); t.merge(); t.selectCell(1,3); t.deleteCols(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td>2,1</td><td class="" rowspan="2" colspan="1">2,2</td></tr><tr><td>3,1</td></tr></tbody></table>',
			},
			{
				note: '删除列，多列元素，向右展开',
				init: t => { t.reset(3,3); t.selectRange(2,2,3,3); t.merge(); t.selectCell(1,2); t.deleteCols(); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr><tr><td>3,1</td><td>3,2</td></tr></tbody></table>',
			},
		];
	}

	run() {
		this.cases.forEach((t, index) => {
			const table = new Table();
			try {
				t.init(table);
			} finally {
				table.remove();
			}
			const html = table.content;
			if(html != t.html) {
				console.table([`测试错误：${t.note ?? ''}`, t.html, html]);
				throw new Error(`测试错误: @${index}`);
			}
		});
	}
}

try {
	const tt = new TableTest();
	tt.run();
} catch(e) {
	console.error(e);
}

let table = new Table();
table.reset(3,3);
