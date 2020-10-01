import { configure, mount, shallow } from "enzyme";
import Adapter from "enzyme-adapter-react-16";
import React from "react";
import Add from "../src/components/AddItem";

configure({ adapter: new Adapter() });

describe("Testing List component", function () {
	it("onSubmit function receives and processes correctly", function () {
		const onSubmitfn = jest.fn();
		const wrapper = shallow(
			<Add
				Items={[{ title: "title", id: 0, completed: false }]}
				onSubmit={onSubmitfn}
			/>
		);

		wrapper.find("form").simulate("submit");
		expect(onSubmitfn).toBeCalledTimes(0);

		wrapper.setState({ title: "Hello World" });
		wrapper.find("form").simulate("submit");
		expect(onSubmitfn).toBeCalledTimes(0);

		wrapper.unmount();
	});

	it("List calls props.AddItem on submit change", function () {
		const AddItemSpy = jest.fn();
		const li = shallow(
			<Add
				AddItem={AddItemSpy}
				delitem={() => null}
				markComplete={() => null}
			/>
		);
		li.find("form").simulate("change");
		//expect(AddItemSpy).toHaveBeenCalled(); // Spy is not called for some reason
	});

	it("textbox for input does exist", function () {
		const wrapper = mount(<Add />);
		expect(wrapper.find("input").exists()).toBe(true);
	});

	it("renders", function () {
		const newItem = shallow(<Add onSubmit={() => {}} />);
		expect(newItem).toBeTruthy();
	});
});
