
const debug = (() => {
	const PACKET_SIZE_BYTES = 1500;
	const displayWrapper = document.querySelector('#display-wrapper');
	const chart = document.querySelector('#chart');

	const state = { rate: 200, prev: null, curr: null, stepSize: 5 }

	document.querySelector('form').addEventListener('submit', e => {
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
			.then(({ success, message, data }) => {
				state.rate = data.rate
				clearInterval(pollInterval);
				pollInterval = setInterval(poll, state.rate);
				// TODO - Handle discrepancy in X axis caused by rate change
			})
			.catch(console.error);
	});

	document.querySelector('#halt-button').addEventListener('click', () => clearInterval(pollInterval));

	document.querySelector('#step-size').addEventListener('change', (e) => {
		chart.querySelector('#lines').innerHTML = '';
		state.stepSize = parseInt(e.target.value);
	})

	function* calculateSpeed(prev, curr){
		for (const interfaceName in prev.interfaces){
			const pi = prev.interfaces[interfaceName];
			const ni = curr.interfaces[interfaceName];
			const receivedDiff = ni.received.count - pi.received.count;
			const transmittedDiff = ni.transmitted.count - pi.transmitted.count
			yield {
				interface: interfaceName,
				received: [receivedDiff, receivedDiff * PACKET_SIZE_BYTES / 1024],
				transmitted: [transmittedDiff, transmittedDiff * PACKET_SIZE_BYTES / 1024]
			}
		}
	}

	function updateDisplay(){
		const { prev, curr, stepSize } = state;

		const speeds = Array.from(calculateSpeed(prev, curr)).reduce((obj, speed) => ({ ...obj, [speed.interface]: speed }), {});
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
			trans.textContent = transmitted[0];
			dl.textContent = received[1].toFixed(0);
			up.textContent = transmitted[1].toFixed(0);

			let line = chart.querySelector(`[data-interface="${interface}"]`)
			if (!line) {
				line = document.createElementNS("http://www.w3.org/2000/svg", 'polyline');
				line.dataset.interface = interface;
				line.setAttribute('fill', 'none');
				line.setAttribute('stroke', 'blue');
				chart.querySelector('#lines').appendChild(line);
			}

			const x = (line.hasAttribute('points') ? line.getAttribute('points').split(' ').length : 0) * stepSize;
			const y = 200 - received[0];
			let points = `${line.hasAttribute('points') ? line.getAttribute('points') + ' ' : ''}${x},${y}`;
			if (x >= 400 + stepSize) {
				points = points.split(' ').slice(1).map(point => {
					let [x, y] = point.split(',').map(Number);
					x -= stepSize;
					return `${x},${y}`;
				}).join(' ');
			}
			line.setAttribute('points', points)
		}
	}

	function poll(){
		return fetch('/api').then(r => r.json())
			.then(({ success, message, data }) => {
				if (!state.prev) {
					state.prev = data;
					return;
				} else {
					state.prev = state.curr || state.prev;
				}

				state.curr = data;
				return updateDisplay();
			}).catch(error => {
				console.error(error);
				alert(error)
				clearInterval(pollInterval);
			});
	}
	poll();

	let pollInterval = setInterval(poll, state.rate);
	return { state, pollInterval };
})();