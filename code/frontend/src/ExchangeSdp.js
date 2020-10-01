export class ExchangeSdp {
	constructor(pc, portNumber) {
		this.sdp = btoa(JSON.stringify(pc.localDescription));
		this.portNum = portNumber;
	}

	async postSdp(postFunc = fetch) {
		const requestOptions = {
			method: "POST",
			headers: new Headers({
				"Content-Type": "application/json",
			}),
			body: JSON.stringify({ BrowserSdp: this.sdp }),
		};

		let remote_sdp = await postFunc("/browsersdp", requestOptions);
		let json_val = await remote_sdp.json();
		return atob(json_val.ServerSdp);
	}
}
