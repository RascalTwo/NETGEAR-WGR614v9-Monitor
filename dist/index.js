
const debug = (() => {
	const PACKET_SIZE_BYTES = 1500;
	const displayWrapper = document.querySelector('#display-wrapper');

	const state = { rate: 1000, prev: null, now: null }

	document.getElementsByTagName('form')[0].addEventListener('submit', e => {
		e.preventDefault();
		const data = Array.from(new FormData(document.forms[0]).entries())
			.reduce((obj, [key, value]) => ({ ...obj, [key]: value }), {});
		// Convert seconds to milliseconds, and also convert string to int.
		data.rate *= 1000;

		return fetch('/update', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(data)
		}).then(r => r.json())
			.then(console.log)
			.catch(console.error);
	})

	function* calculateSpeed(rate, prev, now){
		for (const interfaceName in prev.interfaces){
			const pi = prev.interfaces[interfaceName];
			const ni = now.interfaces[interfaceName];
			const receivedDiff = ni.received.count - pi.received.count;
			const transmittedDiff = ni.transmitted.count - pi.transmitted.count
			yield {
				interface: interfaceName,
				received: [receivedDiff, receivedDiff * PACKET_SIZE_BYTES / rate],
				transmitted: [transmittedDiff, transmittedDiff * PACKET_SIZE_BYTES / rate]
			}
		}
	}

	function updateDisplay(){
		const speeds = Array.from(calculateSpeed(state.rate, state.prev, state.now)).reduce((obj, speed) => ({ ...obj, [speed.interface]: speed }), {});
		for (const { interface, received, transmitted } of Object.values(speeds)){
			let row = displayWrapper.querySelector(`[data-interface="${interface}"]`);
			if (!row) {
				row = document.createElement('tr');
				row.dataset.interface = interface;
				row.appendChild([...Array(5)].reduce(frag => {
					frag.appendChild(document.createElement('td'));
					return frag
				}, document.createDocumentFragment()));
				displayWrapper.appendChild(row);
			}
			const [interfaceCell, recv, trans, dl, up ] = row.childNodes
			interfaceCell.textContent = interface;
			recv.textContent = received[0];
			trans.textContent = transmitted[1]
			dl.textContent = received[1];
			up.textContent = transmitted[1];
		}
	}

	function poll(){
		return fetch('/api').then(r => r.json())
			.then(({ success, message, data }) => {
				if (!state.prev) {
					state.prev = data;
					return;
				} else {
					state.prev = state.now || state.prev;
				}

				state.now = data;
				return updateDisplay();
			}).catch(error => {
				console.error(error);
				alert(error)
				clearInterval(pollInterval);
			});
	}
	poll();

	const pollInterval = setInterval(poll, 1000);
	return { state, pollInterval };
})();