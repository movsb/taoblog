class IOSToggle extends HTMLElement {
	static observedAttributes = [
		'checked',
		'disabled',
		'label',
		'name',
		'required',
		'value',
		'aria-label',
	];

	connectedCallback() {
		if (this._input) {
			return;
		}

		const label = document.createElement('label');
		label.className = 'ios-toggle';

		this._input = document.createElement('input');
		this._input.type = 'checkbox';
		this._input.className = 'ios-toggle__input';

		const track = document.createElement('span');
		track.className = 'ios-toggle__track';
		track.setAttribute('aria-hidden', 'true');

		this._label = document.createElement('span');
		this._label.className = 'ios-toggle__label';

		label.append(this._input, track, this._label);
		this.replaceChildren(label);

		for (const name of IOSToggle.observedAttributes) {
			this._syncAttribute(name);
		}
	}

	attributeChangedCallback(name) {
		if (this._input) {
			this._syncAttribute(name);
		}
	}

	get checked() {
		return this._input ? this._input.checked : this.hasAttribute('checked');
	}

	set checked(checked) {
		this.toggleAttribute('checked', Boolean(checked));
	}

	get disabled() {
		return this.hasAttribute('disabled');
	}

	set disabled(disabled) {
		this.toggleAttribute('disabled', Boolean(disabled));
	}

	get name() {
		return this.getAttribute('name') || '';
	}

	set name(name) {
		this.setAttribute('name', name);
	}

	get value() {
		return this._input ? this._input.value : (this.getAttribute('value') || 'on');
	}

	set value(value) {
		this.setAttribute('value', value);
	}

	_syncAttribute(name) {
		switch (name) {
		case 'checked':
			this._input.defaultChecked = this.hasAttribute('checked');
			this._input.checked = this._input.defaultChecked;
			break;
		case 'disabled':
		case 'required':
			this._input[name] = this.hasAttribute(name);
			break;
		case 'label':
			this._label.textContent = this.getAttribute('label') || '';
			this._label.hidden = !this._label.textContent;
			break;
		case 'name':
			this._input.name = this.getAttribute('name') || '';
			break;
		case 'value':
			this._input.value = this.getAttribute('value') || 'on';
			break;
		case 'aria-label': {
			const ariaLabel = this.getAttribute('aria-label');
			if (ariaLabel) {
				this._input.setAttribute('aria-label', ariaLabel);
			} else {
				this._input.removeAttribute('aria-label');
			}
			break;
		}
		}
	}
}

customElements.define('ios-toggle', IOSToggle);
