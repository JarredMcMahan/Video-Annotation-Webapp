import React, { Component } from "react";
import uuid from "uuid";
import AddItem from "./AddItem";
import Items from "./Items";
import "./List.css";

class List extends Component {
	constructor(props) {
		super(props);
		this.state = {
			Items: [],
		};
	}

	componentDidMount() {
		this.markComplete(this.state.Items.id);
		this.AddItem(this.state.Items.title);
		this.delitem(this.state.Items.id);
	}

	markComplete = (id) => {
		this.setState({
			Items: this.state.Items.map((item) => {
				if (item.id === id) {
					item.completed = !item.completed;
				}
				return item;
			}),
		});
	};

	delitem = (id) => {
		this.setState({
			Items: [...this.state.Items.filter((item) => item.id !== id)],
		});
	};

	AddItem = (title) => {
		this.setState({
			Items: [
				...this.state.Items,
				{
					id: uuid.v4(),
					title,
					completed: false,
				},
			],
		});
	};

	render() {
		return (
			<div className="container">
				<React.Fragment>
					<AddItem AddItem={this.AddItem} />
					<Items
						Items={this.state.Items}
						markComplete={this.markComplete}
						delitem={this.delitem}
					/>
					<label htmlFor={this.state.title}> {this.state.title} </label>{" "}
				</React.Fragment>
			</div>
		);
	}
}

export default List;
