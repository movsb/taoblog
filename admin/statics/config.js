/**
 * @import * from '../dynamic/utils.js';
 */

/** @type {HTMLImageElement} */
const iconImg = document.getElementById('icon');
iconImg.addEventListener('click', () => {
	/** @type {HTMLInputElement} */
	const file = document.getElementById('icon-file');
	file.onchange = async () => {
		console.log(file.files);
		if(file.files.length != 1) {
			return;
		}

		const url = await readFileAsDataURL(file.files[0]);
		console.log(url);

		iconImg.src = url;
		iconImg.dataset.changed = '1';
	};
	file.click();
});

/** @type {HTMLFormElement} */
const formBasic = document.getElementById('form-basic');
formBasic.onsubmit = async (e) => {
	e.preventDefault();

	const data = {
		config: {
			name:           formBasic.elements['name'].value,
			description:    formBasic.elements['description'].value,
			home:           formBasic.elements['home'].value,
		},
	};

	const iconChanged = iconImg.dataset.changed == '1';
	if(iconChanged) {
		data.config.icon = iconImg.src;
		data.update_icon = true;
	}

	const saveBasicIcon = document.getElementById('save-basic-icon');
	saveBasicIcon.classList.add('icon-loading');

	try {
		const rsp = await fetch(`/v3/configs/site`, {
			method: `POST`,
			body: JSON.stringify(data),
		});
		
		const output = await decodeResponse(rsp);

		if(iconChanged) {
			iconImg.src = output.config.icon;
			delete iconImg.dataset.changed;
		}
	} catch (e) {
		alert('更新失败：' + e);
	} finally {
		saveBasicIcon.classList.remove('icon-loading');
	}
};

const fontFamilyPattern = '|\\s*(?:[a-zA-Z0-9]+|"[^"]+")\\s*(,\\s*(?:[a-zA-Z0-9]+|"[^"]+")\\s*)*';
/** @type {HTMLInputElement[]} */
document.querySelectorAll('.font-pattern').forEach(input => input.setAttribute('pattern', fontFamilyPattern));

/** @type {HTMLFormElement} */
const themeForm = document.getElementById('form-theme');
/** @type {HTMLInputElement} */
const accentColorText = themeForm.elements['accent_color_text'];
/** @type {HTMLInputElement} */
const accentColorButton = themeForm.elements['accent_color_button'];
/** @type {HTMLInputElement} */
const accentColorColor = themeForm.elements['accent_color_color'];
/** @type {HTMLParagraphElement[]} */
const accentColorBoxes = themeForm.querySelectorAll('.color-boxes p');
const syncAccentColor = () => {
	accentColorBoxes[0].style.color = accentColorText.value;
	accentColorBoxes[1].style.color = accentColorText.value;
	accentColorColor.value = accentColorText.value;
};
syncAccentColor();
accentColorText.addEventListener('input', ()=>{
	syncAccentColor();
});
accentColorButton.addEventListener('click', () =>{
	accentColorColor.click();
});
accentColorColor.addEventListener('input', ()=>{
	accentColorText.value = accentColorColor.value;
	syncAccentColor();
});

const updateConfig = async (path, value) => {
	try {
		const pod = typeof value == 'boolean' || typeof value == 'number' || typeof value == 'string';
		const rsp = await fetch('/v3/configs', {
			method: 'PATCH',
			body: JSON.stringify({
				path: path,
				yaml: pod ? value : JSON.stringify(value),
			}),
		});
		return await decodeResponse(rsp);
	} catch (e) {
		alert(`path: ${path}\nerror: ${e}`);
		throw e;
	}
};

/** @type {HTMLFormElement} */
themeForm.onsubmit = async(e) => {
	e.preventDefault();

	/** @type {string} */
	const accentColorValue = themeForm.elements['accent_color_text'].value;

	const saveIcon = document.getElementById('save-theme-icon');
	saveIcon.classList.add('icon-loading');

	const requests = [
		updateConfig('theme.variables.font.family', themeForm.elements['font_custom'].value),
		updateConfig('theme.variables.font.mono', themeForm.elements['mono_custom'].value),
		updateConfig('theme.variables.font.size', themeForm.elements['font_size'].value),
		updateConfig('theme.variables.colors.accent', accentColorValue),
	];

	if(accentColorValue != '') {
		requests.push(updateConfig('theme.variables.colors.highlight', `rgb(from ${accentColorValue} r g b / 60%)`));
		requests.push(updateConfig('theme.variables.colors.selection', `rgb(from ${accentColorValue} r g b / 60%)`));
	}

	const responses = await Promise.allSettled(requests);

	saveIcon.classList.remove('icon-loading');
};

/** @type {HTMLFormElement} */
const formOthers = document.getElementById('others-config');
formOthers.onsubmit = async(e) => {
	e.preventDefault();

	const saveIcon = document.getElementById('save-others-icon');
	saveIcon.classList.add('icon-loading');

	try {
		const requests = [
			updateConfig('others.whois.api_layer.key', formOthers.elements['api_layer_key'].value),
			updateConfig('others.geo.gaode.key', formOthers.elements['gaode_api_key'].value),
		];

		const responses = await Promise.allSettled(requests);
	} finally {
		saveIcon.classList.remove('icon-loading');
	}
};

/** @type {HTMLFormElement} */
const formNotify = document.getElementById('notify-config');
formNotify.onsubmit = async(e) => {
	e.preventDefault();

	const saveIcon = document.getElementById('save-notify-icon');
	saveIcon.classList.add('icon-loading');

	try {
		const requests = [
			updateConfig('notify.bark.token', formNotify.elements['bark_token'].value),
			updateConfig('notify.mailer.server', formNotify.elements['mail_server'].value),
			updateConfig('notify.mailer.account', formNotify.elements['mail_account'].value),
			updateConfig('notify.mailer.password', formNotify.elements['mail_password'].value),
		];
		const responses = await Promise.allSettled(requests);
	} finally {
		saveIcon.classList.remove('icon-loading');
	}
};
