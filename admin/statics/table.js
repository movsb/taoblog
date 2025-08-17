class Table {
	constructor() {
		this.table = document.createElement('table');
		document.body.appendChild(this.table);

		this._reset = () => {
			this.init(8,8);
		};
		this.reset();

		/** @type {HTMLTableCellElement | null} */
		this.curCell = null;

		/** @type {HTMLTableCellElement[]} */
		// 始终为从左上到右下的顺序。
		this.selectedCells = [];
	
		this.table.addEventListener('click', (e) => {
			if (this.isCell(e.target)) {
				this.select(e.target);
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
				if(!this.selectRange(startCell, endCell)) {
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

	reset() {
		this._reset();
	}

	/**
	 * @param {HTMLTableCellElement} cell1 
	 * @param {HTMLTableCellElement} cell2 
	 */
	// 目前只支持从左上到右下的选择。
	// 如果设置成功，返回 true。
	selectRange(cell1, cell2) {
		console.log('selectRange:', cell1, cell2);

		const c1r1 = +cell1.dataset.r1;
		const c1c1 = +cell1.dataset.c1;
		const c2r2 = +cell2.dataset.r2;
		const c2c2 = +cell2.dataset.c2;

		this.clearSelection();

		let valid = true;

		Array.from(this.table.rows).forEach(row=> {
			Array.from(row.cells).forEach(cell=> {
				const r1 = +cell.dataset.r1;
				const c1 = +cell.dataset.c1;
				const r2 = +cell.dataset.r2;
				const c2 = +cell.dataset.c2;

				let some = false, all = true;

				// 被包含元素必须被完整包含。
				for(let i=r1; i<=r2; i++) {
					for(let j=c1; j<=c2; j++) {
						const within = c1r1 <= i && i <= c2r2 && c1c1 <= j && j <= c2c2;
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

	init(rows, cols) {
		this.table.innerHTML = '';

		for(let i=0; i<rows; i++) {
			const tr = this.table.insertRow();
			for(let j=0; j<cols; j++) {
				const td = tr.insertCell();
			}
		}

		this.calcCoords();
	}

	/** @param {HTMLTableCellElement} col */
	select(col) {
		if(this.curCell) {
			this.highlight(this.curCell, false);
		}

		this.clearSelection();

		this.curCell = col;
		this.highlight(col, true);
	}

	clearSelection() {
		this.selectedCells.forEach(cell => {
			this.highlight(cell, false);
		})
		this.selectedCells = [];
	}

	highlight(cell, on) {
		cell.style.backgroundColor = on ? 'green' : '';
	}

	addRow(position) {
		if (!this.curCell) {
			alert('Please select a cell first.');
			return;
		}

		/** @type {HTMLTableRowElement} */
		const row = this.curCell.parentElement;
		const rowIndex = row.rowIndex;
		const rowDiff = position == 'above' ? 0 : 1;
		const tr = this.table.insertRow(rowIndex + rowDiff);
		const maxCols = this.maxCols();
		for(let i=0; i<maxCols; i++) {
			tr.insertCell();
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
		this.selectedCells.forEach(cell => {
			const r2 = +cell.dataset.r2;
			const c2 = +cell.dataset.c2;
			if (r2 >= +lastCell.dataset.r2 && c2 >= +lastCell.dataset.c2) {
				lastCell = cell;
			}
		});

		const rowSpan = +lastCell.dataset.r2 - +firstCell.dataset.r1 + 1;
		const colSpan = +lastCell.dataset.c2 - +firstCell.dataset.c1 + 1;

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

		this.calcCoords();
	}

	maxCols() {
		let maxCol = 0;
		Array.from(this.table.rows).forEach(row=> {
			Array.from(row.cells).forEach(cell=> {
				maxCol = Math.max(maxCol, +cell.dataset.c2);
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
						const cc = this.table.rows[x].cells[y].dataset;
						const   ccr1 = +cc.r1,
								ccr2 = +cc.r2,
								ccc1 = +cc.c1,
								ccc2 = +cc.c2;
						if (ccr1 <= tr && tr <= ccr2 && ccc1 <= tc && tc <= ccc2) {
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
				const p = col.dataset;

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

				col.innerText = `${p.r1},${p.c1}`;
			});
		});
	}
}

let table = new Table();
