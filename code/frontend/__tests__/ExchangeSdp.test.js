import sinon from "sinon";
import { ExchangeSdp } from "../src/ExchangeSdp";

let sandbox;
let pcStub;
let port;

beforeEach(() => {
	sandbox = sinon.createSandbox();
	pcStub = {
		localDescription: {
			Test: "Value",
		},
	};

	port = 3000;
	global.Headers = () => {};
});

it("ExchangeSdp -- constructor initializes state", () => {
	const exchanger = new ExchangeSdp(pcStub, port);

	expect(exchanger.portNum).toEqual(port);
	expect(exchanger.sdp).toEqual(btoa(JSON.stringify(pcStub.localDescription)));
});

it("ExchangeSdp -- postSdp posts", () => {
	class resolvableObject {
		async json() {
			console.log("inside json()");
			return { ServerSdp: "junk filled" };
		}
	}

	let mockFetch = (urlPath, opts) => {
		return new resolvableObject();
	};

	const exchanger = new ExchangeSdp(pcStub, port);
	return exchanger.postSdp(mockFetch).then((recievedSdp) => {
		expect(recievedSdp).toEqual(atob("junk filled"));
	});
});
