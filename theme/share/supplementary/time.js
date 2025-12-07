class TimeWithZone {
	constructor(timestamp, zone) {
		const now = new Date();
		if (typeof timestamp != 'number') {
			timestamp = now.getTime() / 1000;
			zone = TimeWithZone.getTimezone();
		} else if (typeof zone != 'string' || zone == '') {
			zone = TimeWithZone.getTimezone();
		}
		this._timestamp = timestamp;
		this._zone = zone;
	}
	
	get time() { return this._timestamp; }
	get zone() { return this._zone;      }
	
	toJSON() {
		return new Date(this._timestamp*1000).toJSON();
	}

	static getTimezone() {
		try {
			return Intl.DateTimeFormat().resolvedOptions().timeZone;
		} catch {
			return '';
		}
	}
}
