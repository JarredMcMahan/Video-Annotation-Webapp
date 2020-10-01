import {
	Checkbox,
	IconButton,
	ListItem,
	ListItemText,
	Paper,
} from "@material-ui/core";
import DeleteOutlined from "@material-ui/icons/DeleteOutlined";
import PropTypes from "prop-types";
import React from "react";

export class ItemFunction extends React.Component {
	render() {
		const { title, id, completed } = this.props.item;
		return (
			<Paper
				justify="space-between" // Add it here :)
				Container
				spacing={24}
				style={{
					margin: ".2px",
					borderBottom: "1px #000",
					textDecoration: this.props.item.completed ? "line-through" : "none",
					padding: "2px",
					border: ".5px solid #008cba",
					wordWrap: "break-word",
				}}
			>
				<ListItem>
					<Checkbox
						type="checkbox"
						checked={completed}
						onClick={this.props.markComplete.bind(this, id)}
						style={{
							color: "#008cba",
						}}
					/>{" "}
					<ListItemText primary={title} />
					<IconButton
						aria-label="Delete item"
						onClick={this.props.delitem.bind(this, id)}
						style={btnStyle}
					>
						<DeleteOutlined />
					</IconButton>
				</ListItem>
			</Paper>
		);
	}
}

ItemFunction.propTypes = {
	item: PropTypes.object,
	markComplete: PropTypes.func,
	delitem: PropTypes.func,
};

const btnStyle = {
	background: "#fff",
	color: "#ff0000",
	cursor: "pointer",
};
export default ItemFunction;
