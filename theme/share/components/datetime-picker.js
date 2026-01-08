class DateTimePicker {
	constructor(anchor, ts, tz, confirm) {
		const outer = document.createElement('div');
		outer.innerHTML = DateTimePicker._template;
		this._dialog = outer.querySelector('dialog');
		this._dialog.remove();
		/** @type {HTMLFormElement} */
		this._form = this._dialog.querySelector('form');
		/** @type {HTMLSelectElement} */
		this._tzSelect = this._form.querySelector('.tz');
		/** @type {HTMLTableElement} */
		this._table = this._form.querySelector('table');
		/** @type {HTMLSelectElement} */
		this._year = this._form.querySelector('select.year');
		/** @type {HTMLSelectElement} */
		this._month = this._form.querySelector('select.month');
		/** @type {HTMLInputElement} */
		this._time = this._form.querySelector('.time');

		this._confirm = confirm;

		// 4 digits
		this._curYear = 0;
		// [1,12]
		this._curMonth = 0;
		this._curDay = 0;
		// [0,23]
		this._curHour = 0;
		this._curMinute = 0;
		this._curSecond = 0;
		this._curTZ = '';

		/** @type {HTMLTableCellElement|null} */
		this._prevActive = null;

		this._initTimezones();
		this._handleEvents();

		if(!ts) { ts = new Date().getTime()/1000; }
		if(!tz) { tz = Intl.DateTimeFormat().resolvedOptions().timeZone; }

		const t = new Date(ts*1000);

		this._select(t.getFullYear(), t.getMonth()+1, t.getDate());
		this._setTime(ts, tz);

		document.body.appendChild(this._dialog);

		this._dialog.inert = true;
		this._dialog.showModal();
		requestAnimationFrame(()=>{
			this._dialog.inert = false;
		});

		// 显示之前拿不到位置。
		this._placeAnchor(anchor);
	}

	get value() {
		const d = new Date(this._curYear, this._curMonth-1, this._curDay, this._curHour, this._curMinute, this._curSecond);
		return {unix: d.getTime()/1000, timezone: this._curTZ};
	}

	/**
	 * 
	 * @param {HTMLElement} anchor 
	 */
	_placeAnchor(anchor) {
		const r = anchor.getBoundingClientRect();
		const d = this._dialog.getBoundingClientRect();
		const left = r.left;
		let top = r.bottom + 5;
		if(top + d.height > window.innerHeight) {
			top = r.top - 5 - d.height;
		}
		this._dialog.style.left = `${left}px`;
		this._dialog.style.top = `${top}px`;
	}

	_handleEvents() {
		this._table.addEventListener('click', (e)=>{
			/** @type {HTMLElement} */
			const target = e.target;
			if(target.tagName == 'TD') {
				this._selectDay(+target.dataset.day);
			}
		});

		this._year.addEventListener('change', ()=>{
			this._select(+this._year.value, this._curMonth, this._curDay);
		});

		this._month.addEventListener('change', ()=>{
			this._select(this._curYear, +this._month.value, this._curDay);
		});

		this._form.querySelector('.now').addEventListener('click', (e)=>{
			e.preventDefault();
			e.stopPropagation();
			const t = new Date();
			const z = Intl.DateTimeFormat().resolvedOptions().timeZone;
			this._select(t.getFullYear(), t.getMonth()+1, t.getDate());
			this._setTime(t.getTime()/1000, z);
		});

		this._dialog.addEventListener('click', e => {
			/** @type {HTMLDialogElement} */
			const target = e.target;
			if(target != this._dialog) { return; }

			const r = this._dialog.getBoundingClientRect();
			if(e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) {
				this.close();
			}
		});

		this._time.addEventListener('change', ()=>{
			/** @type {Date} */
			const totalSeconds = Math.floor(this._time.valueAsNumber/1000);
			this._curHour = Math.floor(totalSeconds / 3600);
			this._curMinute = Math.floor((totalSeconds % 3600) / 60);
			this._curSecond = totalSeconds % 60;
		});

		this._tzSelect.addEventListener('change', ()=>{
			this._curTZ = this._tzSelect.value;
		});
	}

	close() {
		if(this._confirm) {
			this._confirm(this.value);
		}
		this._dialog.close();
		this._dialog.remove();
	}

	_select(year, month, day) {
		const maxDays = new Date(year, month-1+1, 0).getDate();
		const t = new Date(year, month-1, Math.min(maxDays, day));
		this._buildTable(t.getTime()/1000);
		this._selectDay(t.getDate());
	}

	_setTime(ts, tz) {
		/** @type {HTMLOptionElement|null} */
		const opt = this._tzSelect.querySelector(`option[data-name="${tz}"`);
		if(opt) {
			opt.selected = true;
		}
		this._curTZ = tz;

		const d = new Date(ts*1000);
		const h = d.getHours();
		const m = d.getMinutes();
		const hs = h<10 ? h.toString().padStart(2,'0') : h.toString();
		const ms = m<10 ? m.toString().padStart(2,'0') : m.toString();
		this._time.value = `${hs}:${ms}`;
		this._curHour = h;
		this._curMinute = m;
		this._curSecond = d.getSeconds();
	}

	_selectDay(n) {
		const active = this._table.querySelector(`td[data-day="${n}"]`);
		if(!active) {return;}

		if(this._prevActive) {
			this._prevActive.classList.remove('active');
		}

		active.classList.add('active');
		this._prevActive = active;

		this._curDay = n;
	}

	_initTimezones() {
		Intl.supportedValuesOf('timeZone').forEach(z => {
			const opt = document.createElement('option');
			opt.textContent = z;
			opt.dataset.name = z;
			this._tzSelect.add(opt);
		});
	}

	_buildTable(ts) {
		const now = new Date(ts*1000);
		if(now.getFullYear() == this._curYear && now.getMonth()+1==this._curMonth) {
			return;
		}

		while(this._table.rows.length>1) {
			this._table.rows[this._table.rows.length-1].remove();
		}

		let firstDayOfWeek = new Date(now.getFullYear(), now.getMonth(), 1).getDay();
		if(firstDayOfWeek==0) {firstDayOfWeek = 7;}

		let tr = this._table.insertRow();
		for(let i=1; i<firstDayOfWeek; i++) {
			tr.insertCell();
		}

		const daysOfMonth = new Date(now.getFullYear(), now.getMonth()+1,0).getDate();
		for(let i=1; i<=daysOfMonth; i++) {
			const td = tr.insertCell();
			td.textContent = `${i}`;
			td.dataset.day = i;
			if((i-1+firstDayOfWeek)%7 == 0) {
				tr = this._table.insertRow();
			}
		}
		for(let i=daysOfMonth+1;;i++) {
			tr.insertCell();
			if((i-1+firstDayOfWeek)%7==0) {
				break;
			}
		}

		this._curYear = now.getFullYear();
		const y = this._year.querySelector(`option[data-year="${this._curYear}"]`)
		if(!y) {
			this._year.textContent = '';
			const now = new Date();
			for(let n=now.getFullYear(); n >= 1990; n--) {
				const opt = document.createElement('option');
				opt.textContent = `${n}`;
				opt.dataset.year = n;
				this._year.add(opt);
			}
		}
		this._year.querySelector(`option[data-year="${this._curYear}"]`).selected = true;
		this._curMonth = now.getMonth()+1;
		const m = this._month.querySelector(`option[data-month="${this._curMonth}"]`)
		if(!m) {
			this._month.textContent = '';
			for(let i=1; i<=12;i++) {
				const opt = document.createElement('option');
				let s = i.toString();
				if(i<10) s = s.padStart(2,'0');
				opt.textContent = `${s}`;
				opt.dataset.month = i;
				this._month.add(opt);
			}
		}
		this._month.querySelector(`option[data-month="${this._curMonth}"]`).selected = true;
	}

	static _template = `
<dialog class="datetime-picker">
	<form autocomplete="off" method="dialog">
		<div style="display: flex; justify-content: space-between;">
			<div>
				年：<select class="year"></select>
				月：<select class="month"></select>
			</div>
			<button class="now">现在</button>
		</div>
		<div>
			<table style="width: 100%; border: 1px solid lightgray;">
				<tr><th>一</th><th>二</th><th>三</th><th>四</th><th>五</th><th>六</th><th>日</th></tr>
			</table>
		</div>
		<div style="display: flex; justify-content: space-between; gap: 5px; align-items: center;">
			时间：<input type="time" class="time"><select class="tz"></select>
		</div>
	</form>
</dialog>`;
}
