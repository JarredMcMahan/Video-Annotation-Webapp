import { configure, mount } from "enzyme";
import Adapter from "enzyme-adapter-react-16";
import React from "react";
import sinon from "sinon";
import Atn, { Coord, UserAction } from "../src/components/Atn";

configure({ adapter: new Adapter() });

let sandbox;
let component;

beforeEach(() => {
	sandbox = sinon.createSandbox();
	component = new Atn();

	component.setState = function (state) {
		this.state = { ...this.state, ...state };
	};
});

afterEach(() => sandbox.restore());

it("Atn.js - renders without crashing", () => {
	const stub = sandbox.stub(Atn.prototype, "componentDidMount");
	const wrapper = mount(<Atn />);

	expect(wrapper.find("canvas").exists()).toEqual(true);
});

it("Atn.js - component creates object", () => {
	let atn = new Atn();

	expect(atn.state.pc).toBe(null);
	expect(atn.state.remoteSessionDescription).toEqual("");
	expect(atn.state.video).toBe(null);
	expect(atn.state.color).toEqual("#ff0000");
});

it("Atn.js - Testing Coord", () => {
	const testCoord = new Coord(1, 2);

	expect(testCoord.get_x()).toEqual(1);
	expect(testCoord.get_y()).toEqual(2);
});

it("Atn.js - Testing UserAction Draw", () => {
	const testAction = new UserAction("green", "draw", 1, 2);

	// Check if enabled
	expect(testAction.enabled).toBe(true);
	// Check color
	expect(testAction.get_color()).toEqual("green");
	// Check cords
	expect(testAction.drag[0].get_x()).toEqual(1);
	expect(testAction.drag[0].get_y()).toEqual(2);
	// Should not be text
	expect(testAction.is_text()).toBe(false);
	// Should be drawing
	expect(testAction.is_drag()).toBe(true);
	// Should be empty
	expect(testAction.is_empty()).toBe(false);
	// Setting line weight
	testAction.set_line_weight(2);
	expect(testAction.get_line_weight()).toEqual(2);
	// Drag length should be 1
	expect(testAction.get_drag_length()).toEqual(1);
	// Test event x
	expect(testAction.event_x()).toEqual(1);
	// Test event y
	expect(testAction.event_y()).toEqual(2);
	// Add cord
	testAction.add_coord(2, 3);
	expect(testAction.get_drag()[1].get_x()).toEqual(2);
	expect(testAction.get_drag()[1].get_y()).toEqual(3);
	// Disable
	testAction.disable();
	expect(testAction.disabled()).toBe(false);
});

it("Atn.js - Testing UserAction Text", () => {
	const testAction = new UserAction("blue", "text", 1, 2);

	// Check if enabled
	expect(testAction.enabled).toBe(true);
	// Check color
	expect(testAction.get_color()).toEqual("blue");
	// Check cords
	expect(testAction.drag[0].get_x()).toEqual(1);
	expect(testAction.drag[0].get_y()).toEqual(2);
	// Should  be text
	expect(testAction.is_text()).toBe(true);
	// Should not be drawing
	expect(testAction.is_drag()).toBe(false);
	// Should not be empty
	testAction.add_text("test");
	expect(testAction.is_empty()).toBe(false);
	// Text should say test
	expect(testAction.get_text()).toBe("test");
	// Test backspace
	testAction.backspace();
	expect(testAction.get_text()).toBe("tes");
	// Set font
	testAction.set_font("Times New Roman");
	testAction.set_font_size(12);
	expect(testAction.get_font()).toEqual("Times New Roman");
	expect(testAction.get_font_size()).toEqual(12);
	// Drag length should be 1
	expect(testAction.get_drag_length()).toEqual(1);
	// Test event x
	expect(testAction.event_x()).toEqual(1);
	// Test event y
	expect(testAction.event_y()).toEqual(2);
	// Disable
	testAction.disable();
	expect(testAction.disabled()).toBe(false);
});

it("Atn.js - componentDidMount()", () => {
	component.canvas = {
		current: {
			width: 0,
			height: 0,
		},
	};

	const keyPressStub = sandbox.stub(Atn.prototype, "keyPress");
	const createIntervalStub = sandbox.stub(Atn.prototype, "createInterval");

	component.createPeerConnection = sandbox.stub().returns("--pc--");

	component.componentDidMount();

	expect(keyPressStub.called).toBe(true);
	expect(createIntervalStub.called).toBe(true);
	expect(component.canvas.width).toEqual(component.state.videoWidth);
	expect(component.canvas.height).toEqual(component.state.videoHeight);

	expect(component.createPeerConnection.called).toBe(true);

	expect(component.state.videoStatus).toEqual("Playing");
	expect(component.state.pc).toEqual("--pc--");
});

describe("Atn.js - Timer related tests", () => {
	let sandbox;
	let component;
	let clock;

	beforeEach(() => {
		sandbox = sinon.createSandbox();
		clock = sandbox.useFakeTimers();

		component = new Atn();
		component.setState = function (state) {
			this.state = { ...this.state, ...state };
		};
	});

	afterEach(() => sandbox.restore());

	it("createInterval calls setInterval", () => {
		component.onInterval = sandbox.stub();

		component.createInterval();
		expect(component.onInterval.called).toBe(false);
		clock.tick(10);
		expect(component.onInterval.called).toBe(true);
		expect(component.state.intervalId.toString()).toEqual(
			Object.keys(clock.timers)[0]
		);
	});

	it("componentDidUnmount unregisters from the timer", () => {
		component.createInterval();
		component.componentWillUnmount();

		expect(clock.timers).toEqual({});
	});
});

it("Atn.js - canvasColor()", () => {
	component.canvasColor("green");
	expect(component.state.color).toEqual("green");
});

it("Atn.js - canvasTool()", () => {
	component.canvasTool("draw");
	expect(component.state.canvas_function).toEqual("draw");
});

it("Atn.js - canvasClear()", () => {
	component.state.user_actions.push(new UserAction("blue", "text", 1, 1));
	component.state.user_actions.push(new UserAction("green", "draw", 1, 2));
	component.canvasClear();
	expect(component.state.user_actions.length).toBe(0);
});

it("Atn.js - canvasUndo()", () => {
	component.state.user_actions.push(new UserAction("color", "draw", 1, 2));
	expect(component.state.user_actions.length).toBe(1);
	component.canvasUndo();
	expect(component.state.user_actions.length).toBe(0);
});

it("Atn.js - enterPressed()", () => {
	component.state.user_actions.push(new UserAction("blue", "text", 1, 1));
	component.enterPressed();
	expect(component.state.user_actions[0].disabled()).toBe(false);
});

it("Atn.js - mouseUp()", () => {
	component.mouseUp();
	expect(component.state.mouseDown).toBe(false);
});

it("Atn.js - handleClick()", () => {
	component.handleClick();
	expect(component.state.icon).toBe(true);
});

it("Atn.js - onIceCandidate", () => {
	let tmp = component.onIceCandidate({}, null);
	expect(tmp).toBe(undefined);
});

it("Atn.js - Key tests", () => {
	component.keyCode2ch({});
	expect(component.state.keys).toEqual(undefined);
	component.isKeyAlpha("key");
	expect(component.state.keys).toEqual(undefined);
	component.isKeyNumeric("key");
	expect(component.state.keys).toEqual(undefined);
	component.isKeySymbol("key");
	expect(component.state.keys).toEqual(undefined);
});
