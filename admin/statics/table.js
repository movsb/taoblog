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
			if (this.isCell(e.target)) {
				this._selectCell(e.target);
			}
		});
	
		this.table.addEventListener('mousedown', e => {
			console.log('mousedown:', e.target);
			if(!this.isCell(e.target)) {
				console.log('mousedown not on cell');
				return;
			}

			const startCell = e.target;
			if(this.curCell) {
				this.highlight(this.curCell, false);
			}
			this.clearSelection();

			const moveHandler = e => {
				console.log('mousemove:', e.target);
				if (!this.isCell(e.target)) {
					console.log('mousemove not on cell');
					return;
				}
				const table = e.target.closest('table');
				if (!table || table != this.table) {
					console.log('mousemove on another table, ignoring');
					return;
				}
				const endCell = e.target;
				if(!this._selectRange(startCell, endCell)) {
					console.log('选区无效。');
				}
			};

			document.addEventListener('mousemove', moveHandler);
			document.addEventListener('mouseup', ()=>{
				console.log('mouseup:', e.target);
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
		this.clearSelection();
		this.table.innerHTML = '';

		for(let i=0; i<rows; i++) {
			const tr = this.table.insertRow();
			for(let j=0; j<cols; j++) {
				const td = tr.insertCell();
			}
		}

		this.calcCoords();
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

	_getCoords(cell) {
		return cell._coords;
	}
	_setCoords(cell, coords) {
		cell._coords = coords;
	}
	
	// 目前只支持从左上到右下的选择。
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
		console.log('selectRange:', cell1, cell2);

		const cc1 = this._getCoords(cell1);
		const cc2 = this._getCoords(cell2);

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
					this.highlight(cell, true);
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

	isCell(element) {
		return element.tagName == 'TD' || element.tagName == 'TH';
	}

	/**
	 * 
	 * @param {Number} row 
	 * @param {Number} col 
	 */
	selectCell(row, col) {
		const cell = this.findCell(row, col);
		return this._selectCell(cell);
	}

	/** @param {HTMLTableCellElement} col */
	_selectCell(col) {
		if(this.curCell) {
			this.highlight(this.curCell, false);
		}

		this.clearSelection();

		this.curCell = col;
		this.highlight(col, true);
	}

	clearSelection() {
		this.curCell = null;
		this.selectedCells.forEach(cell => {
			this.highlight(cell, false);
		})
		this.selectedCells = [];
	}

	/**
	 * 
	 * @param {HTMLTableCellElement} cell 
	 * @param {boolean} on 
	 */
	highlight(cell, on) {
		if(on) cell.classList.add('selected');
		else cell.classList.remove('selected');
	}

	/** @returns {HTMLTableCellElement} */
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

	addRow(position) {
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
		const maxCols = this.maxCols();

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

		this.calcCoords();
	}

	/** @param {string} position  */
	addCol(position) {
		if (!this.curCell) {
			alert('Please select a cell first.');
			return;
		}

		/** @type {HTMLTableCellElement} */
		const col = this.curCell;
		const colIndex = col.cellIndex;
		const colDiff = position == 'left' ? 0 : 1;
		const rowCount = this.table.rows.length;
		for(let i=0; i<rowCount; i++) {
			const td = this.table.rows[i].insertCell(colIndex+colDiff);
		}

		this.calcCoords();
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
		const lastCoords = this._getCoords(lastCell);

		this.selectedCells.forEach(cell => {
			const cc = this._getCoords(cell);
			if (cc.r2 >= lastCoords.r2 && cc.c2 >= lastCoords.c2) {
				lastCell = cell;
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

		this.calcCoords();
	}

	maxCols() {
		let maxCol = 0;
		Array.from(this.table.rows).forEach(row=> {
			Array.from(row.cells).forEach(cell=> {
				const cc = this._getCoords(cell);
				maxCol = Math.max(maxCol, cc.c2);
			});
		});
		return maxCol;
	}

	calcCoords() {
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
		 * @type {{ run: (t: Table) => void, html: string }[]}
		 */
		this.cases = [
			{
				run: t => { t.reset(2,2); },
				html: '<table><tbody><tr><td>1,1</td><td>1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr></tbody></table>',
			},
			{
				run: t => { t.reset(2,2); t.selectCell(1,2); },
				html: '<table><tbody><tr><td>1,1</td><td class="selected">1,2</td></tr><tr><td>2,1</td><td>2,2</td></tr></tbody></table>',
			},
		];
	}

	run() {
		this.cases.forEach((t, index) => {
			const table = new Table();
			t.run(table);
			const html = table.content;
			table.remove();
			if(html != t.html) {
				console.table(['测试错误：', html, t.html]);
				throw new Error(`测试错误: @${index}`);
			}
		});
	}
}

new TableTest().run();

let table = new Table();
table.reset(2,2);
