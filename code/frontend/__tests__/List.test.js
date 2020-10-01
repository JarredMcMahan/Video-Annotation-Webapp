import { configure, shallow } from "enzyme";
import Adapter from "enzyme-adapter-react-16";
import React from "react";
import sinon from "sinon";
import ItemFunction from "../src/components/ItemFunction";
import Items from "../src/components/Items";
import List from "../src/components/List";

configure({ adapter: new Adapter() });

let wrapper;
let component;
let sandbox;

beforeEach(() => {
	sandbox = sinon.createSandbox();
	component = new List();
	wrapper = shallow(
		<List Items={[{ title: "title", id: 0, completed: false }]} />
	);
});

afterEach(() => sandbox.restore());

describe("Testing List component", function () {
	it("Should pass canary test", function () {
		expect(true).toBe(true);
	});

	it("Should throw exception for used as a function", function () {
		var toCall = function () {
			List(-1);
		};
		expect(toCall).toThrow("Cannot call a class as a function"); //
	});

	it("renders", function () {
		const list = shallow(<List />);
		expect(list).toBeTruthy();
	});

	it('expect to have state.title set to "Hello World" ', function () {
		wrapper.setState({ title: "Hello World" });
		expect(wrapper.state("title")).toEqual("Hello World");
	});

	it("expect to have state.complete set to false", function () {
		wrapper.setState({ completed: true });
		expect(wrapper.state("completed")).toEqual(true);
	});

	it("expect to have state.id set to 1", function () {
		wrapper.setState({ id: 1 });
		expect(wrapper.state("id")).toEqual(1);
	});

	it("Render ItemFunction wrapper", function () {
		wrapper = shallow(
			<ItemFunction
				item={{ title: "", id: "", completed: "" }}
				delitem={() => {}}
				markComplete={() => {}}
			/>
		);
		expect(wrapper).toBeTruthy();
	});

	it("design markComplete and delItem and AddItem", () => {
		const markCompleteStub = sandbox.stub(component, "markComplete");
		const delitemStub = sandbox.stub(component, "delitem");
		const AddItemStub = sandbox.stub(component, "AddItem");

		component.componentDidMount();

		expect(markCompleteStub.called).toBe(true);
		expect(delitemStub.called).toBe(true);
		expect(AddItemStub.called).toBe(true);
	});

	it("renders items", () => {
		component = new Items();
		const ItemStub = sandbox.stub(component, "render");

		component.componentDidMount();

		expect(ItemStub.called).toBe(true);
	});
});
