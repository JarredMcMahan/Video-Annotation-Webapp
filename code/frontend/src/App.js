import { Grid } from "@material-ui/core";
import React from "react";
import "./App.css";
import Atn from "./components/Atn";
import List from "./components/List";

function App() {
	return (
		<Grid container>
			<Grid item>
				<div
					style={{ display: "flex", alignItems: "flex-start" }}
					className="App"
				>
					<Atn />
					<List />
				</div>
			</Grid>
		</Grid>
	);
}

export default App;
