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
