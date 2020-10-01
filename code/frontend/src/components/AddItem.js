import { Button, Grid, TextField } from "@material-ui/core";
import PropTypes from "prop-types";
import React, { Component } from "react";

export class AddItem extends Component {
	constructor(props) {
		super(props);
		this.state = {
			title: "",
			completed: false,
		};
	}

	onSubmit = (e) => {
		if (!this.state.title) {
			return;
		}
		try {
			this.props.AddItem(this.state.title);
		} finally {
			this.setState({ title: "" });
			try {
				e.preventDefault();
			} finally {
				return;
			}
		}
	};

	onChange = (e) => this.setState({ [e.target.name]: e.target.value });

	render() {
		return (
			<form onSubmit={this.onSubmit} style={{ display: "flex" }}>
				<Grid container>
					<Grid xs={9} md={9} item>
						<TextField
							fullWidth
							type="text"
							name="title"
							style={{
								flex: "10",
								padding: "10px",
								border: ".8px solid",
								background: "#ffffff",
								borderColor: "#008cba",
								borderBottomColor: "#008cba",
							}}
							placeholder="What's next?"
							value={this.state.title}
							onChange={this.onChange}
						/>
					</Grid>
					<Grid xs={2} md={2} item>
						<Button
							fullWidth
							color="secondary"
							variant="outlined"
							type="submit"
							value="Submit"
							borderRadius="5%"
							style={{
								color: "#008cba",
								border: "2px solid",
								background: "#ffffff",
								padding: "13px 48px 13px 48px",
							}}
						>
							Add
						</Button>
					</Grid>
				</Grid>
			</form>
		);
	}
}

AddItem.propTypes = {
	AddItem: PropTypes.func,
};

export default AddItem;
