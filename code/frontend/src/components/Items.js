import PropTypes from "prop-types";
import React from "react";
import ItemFunction from "./ItemFunction";

class Items extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			Items: [],
		};
	}

	componentDidMount() {
		this.render();
	}

	render() {
		return this.props.Items.map((item) => [
			<ItemFunction
				key={item.id}
				item={item}
				markComplete={this.props.markComplete}
				delitem={this.props.delitem}
			/>,
		]);
	}
}

Items.propTypes = {
	Items: PropTypes.array,
	markComplete: PropTypes.func,
	delitem: PropTypes.func,
};

export default Items;
