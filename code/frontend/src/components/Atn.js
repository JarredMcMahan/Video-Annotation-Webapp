import { Button, Grid } from "@material-ui/core";
import {
	Create,
	Delete,
	FontDownload,
	History,
	Pause,
	PlayArrow,
} from "@material-ui/icons";
import React, { Component } from "react";
import { ExchangeSdp } from "../ExchangeSdp.js";
import "./Atn.css";

export class Coord {
	constructor(x, y) {
		this.x = x;
		this.y = y;
    }
    
	get_x() {
		return this.x;
    }
    
	get_y() {
		return this.y;
	}
}

export class UserAction {
	constructor(color, type, x, y) {
		this.enabled = true;
		this.color = color;
		this.type = type;
		this.text = "";
		this.drag = [];
		this.add_coord(x, y);
    }
    
	backspace() {
		let temp = "";
		for (let i = 0; i < this.text.length - 1; i++) temp += this.text[i];
		this.text = temp;
    }
    
	disable() {
		this.enabled = false;
    }
    
	disabled() {
		return this.enabled;
    }
    
	is_text() {
		return this.type === "text";
    }
    
	is_drag() {
		return this.type === "draw";
    }
    
	is_empty() {
		return this.is_text() && this.text.length < 1;
    }
    
	set_font(font) {
		if (this.is_text()) this.font = font;
    }
    
	get_font() {
		return this.font;
    }
    
	set_font_size(size) {
		this.font_size = size;
    }
    
	get_font_size() {
		return this.font_size;
    }
    
	set_line_weight(weight) {
		this.line_weight = weight;
    }
    
	get_line_weight() {
		return this.line_weight;
    }
    
	add_text(text) {
		if (this.enabled && this.is_text()) this.text += text;
    }
    
	get_text() {
		return this.text;
    }
    
	get_color() {
		return this.color;
    }
    
	get_drag() {
		return this.drag;
    }
    
	get_drag_length() {
		return this.get_drag().length;
    }
    
	event_x() {
		return this.get_drag()[0].get_x();
    }
    
	event_y() {
		return this.get_drag()[0].get_y();
    }
    
	add_coord(x, y) {
		if (this.enabled) {
			this.drag.push(new Coord(x, y));
			this.drag_length += 1;
		}
	}
}

const PC_CONFIG = { iceServers: [{ urls: ["stun:stun.l.google.com:19302"] }] };

class Atn extends Component {
	constructor(props) {
		super(props);
		this.state = {
			pc: null,
			remoteSessionDescription: "",
			video: null,
			videoStatus: "",
			color: "#ff0000",
			lineThickness: 4,
			playPauseText: "Pause",
			mouseDown: false,
			icon: false,
			canvas_function: "draw",
			canvas_font_size: 30,
			canvas_font: "Arial",
			user_actions: [],
		};
		this.canvas = React.createRef();
	}

	createPeerConnection() {
		const peerConnection = new RTCPeerConnection(PC_CONFIG);
		peerConnection.onicecandidate = (event) =>
			this.onIceCandidate(peerConnection, event);
		peerConnection.oniceconnectionstatechange = () => {
			console.log("ICE state:", peerConnection.iceConnectionState);
			if (peerConnection.iceConnectionState !== "failed") {
				return;
			}
			// attempt restart
			peerConnection.createOffer({ iceRestart: true }).then(function (offer) {
				return peerConnection.setLocalDescription(offer);
			});
		};
		peerConnection.ontrack = (event) => this.getRemoteVideo(event);
		peerConnection.addTransceiver("video", { direction: "sendrecv" });
		peerConnection
			.createOffer()
			.then((desc) => {
				try {
					return peerConnection.setLocalDescription(desc);
				} catch (e) {
					console.log(e);
				}
			})
			.then(async () => {
				const exchanger = new ExchangeSdp(
					peerConnection,
					document.location.port
				);
				const returnedSdp = await exchanger.postSdp();

				this.setState({
					remoteSessionDescription: returnedSdp,
				});

				this.start();
			});

		return peerConnection;
	}

	componentDidMount() {
		let { pc } = this.state;

		this.keyPress();
		this.createInterval();

		pc = this.createPeerConnection();

		this.canvas.current.width = this.state.videoWidth;
		this.canvas.current.height = this.state.videoHeight;

		this.setState({
			video: React.createRef(),
			pc,
			videoStatus: "Playing",
		});
	}

	componentWillUnmount() {
		clearInterval(this.state.intervalId);
	}

	canvasUndo() {
		this.state.user_actions.pop();
    }
    
	canvasClear() {
		while (this.state.user_actions.length > 0) this.state.user_actions.pop();
    }
    
	canvasTool(tool) {
		this.setState({ canvas_function: tool });
    }
    
	canvasColor(color) {
		this.setState({ color: color });
    }
    
	onInterval() {
		const ctx = this.canvas.current.getContext("2d");
		ctx.clearRect(0, 0, this.canvas.current.width, this.canvas.current.height);
		let offset_x = this.canvas.current.getBoundingClientRect().left;
		let offset_y_canvas = this.canvas.current.getBoundingClientRect().top;
		let offset_y_scroll = window.pageYOffset;
		let offset_y = offset_y_canvas + offset_y_scroll;
		for (let i = 0; i < this.state.user_actions.length; i++) {
			if (this.state.user_actions[i].is_drag()) {
				if (this.state.user_actions[i].get_drag_length() > 1) {
					ctx.beginPath();
					ctx.strokeStyle = this.state.user_actions[i].get_color();
					ctx.lineWidth = this.state.user_actions[i].get_line_weight();
					ctx.moveTo(
						this.state.user_actions[i].event_x() - offset_x,
						this.state.user_actions[i].event_y() - offset_y
					);
					for (
						let j = 1;
						j < this.state.user_actions[i].get_drag_length() - 1;
						j++
					) {
						ctx.lineTo(
							this.state.user_actions[i].get_drag()[j].get_x() - offset_x,
							this.state.user_actions[i].get_drag()[j].get_y() - offset_y
						);
					}
					ctx.stroke();
				} else if (this.state.user_actions[i].get_drag_length() === 1) {
					ctx.fillStyle = this.state.user_actions[i].get_color();
					ctx.fillRect(
						this.state.user_actions[i].event_x() - offset_x,
						this.state.user_actions[i].event_y() - offset_y,
						(ctx.lineWidth = this.state.user_actions[i].get_line_weight()),
						(ctx.lineWidth = this.state.user_actions[i].get_line_weight())
					);
				}
			} else if (this.state.user_actions[i].is_text()) {
				ctx.fillStyle = this.state.user_actions[i].get_color();
				ctx.font =
					this.state.user_actions[i].get_font_size() +
					"px " +
					this.state.user_actions[i].get_font();
				ctx.fillText(
					this.state.user_actions[i].get_text(),
					this.state.user_actions[i].event_x() - offset_x,
					this.state.user_actions[i].event_y() - offset_y
				);
			}
		}
    }
    
	createInterval() {
		var intervalUpdateMS = 10;
		var intervalId = setInterval(
			function () {
				this.onInterval();
			}.bind(this),
			intervalUpdateMS
		);
		this.setState({ intervalId: intervalId });
    }
    
	enterPressed() {
		if (this.state.user_actions.length > 0)
			if (this.state.user_actions[this.state.user_actions.length - 1].is_text())
				this.state.user_actions[this.state.user_actions.length - 1].disable();
    }
    
	keyCode2ch(key) {
		return String.fromCharCode(key);
    }
    
	isKeyAlpha(key) {
		return key >= 65 && key <= 90;
    }
    
	isKeyNumeric(key) {
		return key >= 48 && key <= 57;
    }
    
	isKeySymbol(key) {
		return key === 32;
    }
    
	keyPress() {
		document.onkeydown = function (evt) {
			evt = evt || window.event;
			if (evt.keyCode === 13) this.enterPressed();
			if (evt.ctrlKey && evt.keyCode === 90) this.canvasUndo();
			else if (evt.keyCode === 8) {
				this.state.user_actions[this.state.user_actions.length - 1].backspace();
			} else if (
				this.isKeyAlpha(evt.keyCode) ||
				this.isKeyNumeric(evt.keyCode) ||
				this.isKeySymbol(evt.keyCode)
			) {
				var input_ch = this.keyCode2ch(evt.keyCode);
				if (this.isKeyAlpha(evt.keyCode))
					if (!evt.shiftKey) input_ch = input_ch.toLowerCase();
				if (this.state.user_actions.length > 0)
					if (
						this.state.user_actions[
							this.state.user_actions.length - 1
						].is_text()
					)
						this.state.user_actions[
							this.state.user_actions.length - 1
						].add_text(input_ch);
			}
		}.bind(this);
    }
    
	mouseDown() {
		this.setState({ mouseDown: true });
		let event = window.event;
		if (this.state.user_actions.length > 0) {
			if (
				this.state.user_actions[this.state.user_actions.length - 1].is_empty()
			)
				this.state.user_actions.pop();
			else
				this.state.user_actions[this.state.user_actions.length - 1].disable();
		}
		this.state.user_actions.push(
			new UserAction(
				this.state.color,
				this.state.canvas_function,
				event.clientX,
				event.clientY
			)
		);
		if (this.state.canvas_function === "text") {
			this.state.user_actions[this.state.user_actions.length - 1].set_font(
				this.state.canvas_font
			);
			this.state.user_actions[this.state.user_actions.length - 1].set_font_size(
				this.state.canvas_font_size
			);
		} else
			this.state.user_actions[
				this.state.user_actions.length - 1
			].set_line_weight(this.state.lineThickness);
	}

	mouseUp() {
		this.setState({ mouseDown: false });
	}
	mouseMove() {
		let event = window.event;
		if (this.state.canvas_function === "draw" && this.state.mouseDown)
			this.state.user_actions[this.state.user_actions.length - 1].add_coord(
				event.clientX,
				event.clientY
			);
	}

	showToolbar() {
		document.getElementById("canvas_control_container").style = "height:80px;";
		let tool_buttons = document.getElementsByClassName("tool_button");
		for (let i = 0; i < tool_buttons.length; i++)
			tool_buttons[i].style = "opacity:1;";
		document.getElementById("canvas_color_button_container").style =
			"opacity:1;";
	}

	hideToolbar() {
		document.getElementById("canvas_control_container").style = "height:0;";
		let tool_buttons = document.getElementsByClassName("tool_button");
		for (let i = 0; i < tool_buttons.length; i++)
			tool_buttons[i].style = "opacity:0;";
		document.getElementById("canvas_color_button_container").style =
			"opacity:0;";
	}

	playPauseVideo() {
		if (this.state.videoStatus === "Playing") {
			let { video } = this.state;
			video.current.pause();
			this.canvas.current.width = video.current.offsetWidth;
			this.canvas.current.height = video.current.offsetHeight;
			this.showToolbar();
			this.setState({
				video,
				videoStatus: "Paused",
				playPauseText: "Play",
			});
		} else {
			let { video } = this.state;
			this.canvasClear();
			video.current.play();
			this.hideToolbar();
			this.setState({
				video,
				videoStatus: "Playing",
				playPauseText: "Pause",
			});
		}
	}

	handleClick() {
		const { icon } = this.state;
		this.setState({ icon: !icon });
	}

	getRemoteVideo = (event) => {
		let { video } = this.state;
		let remoteVideo = video.current;
		if (remoteVideo.srcObject !== event.streams[0]) {
			remoteVideo.srcObject = event.streams[0];
		}
		this.setState({
			video,
			videoStatus: "Playing",
		});
	};

	onIceCandidate = (pc, event) => {
		if (event !== null) {
			return;
		}

		const l = btoa(JSON.stringify(pc.localDescription));
		this.setState({ localSessionDescription: l });
	};

	start() {
		let { remoteSessionDescription, pc } = this.state;
		if (remoteSessionDescription === "") {
			return alert("Remote Session Description must not be empty");
		}
		pc.createOffer()
			.then(() => {
				pc.setRemoteDescription(
					new RTCSessionDescription(JSON.parse(remoteSessionDescription))
				);
			})
			.catch((e) => console.log(e));
		this.setState({ pc });
	}

	render() {
		const { video, playPauseText, icon } = this.state;

		return (
			<Grid container spacing={3}>
				<div className="Atn">
					<Grid item>
						<div id="media_container">
							<div id="canvas_container">
								<canvas
									id="canvas"
									ref={this.canvas}
									onMouseMove={() => this.mouseMove()}
									onMouseDown={() => this.mouseDown()}
									onMouseUp={() => this.mouseUp()}
									onMouseLeave={() => this.mouseUp()}
								/>
							</div>
							<div id="video">
								<video
									id="video-element"
									className="video"
									ref={video}
									autoPlay
									muted
									src="https://media.giphy.com/media/Ph0oIVQeuvh0k/giphy.mp4"
								/>
							</div>
						</div>
					</Grid>
					<Grid item>
						<div id="canvas_control_container">
							<div id="canvas_function_button_container">
								<button
									id="canvas_undo_button"
									className="tool_button tool_function_button"
									onClick={() => this.canvasUndo()}
								>
									<History />
								</button>
								<button
									id="canvas_text_button"
									className="tool_button tool_function_button"
									onClick={() => this.canvasTool("text")}
								>
									<FontDownload />
								</button>
								<button
									id="canvas_draw_button"
									className="tool_button tool_function_button"
									onClick={() => this.canvasTool("draw")}
								>
									<Create />
								</button>
								<button
									id="canvas_clear_button"
									className="tool_button tool_function_button"
									onClick={() => this.canvasClear()}
								>
									<Delete />
								</button>
							</div>
							<div id="canvas_color_button_container">
								<button
									id="canvas_red_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#ff0000")}
								></button>
								<button
									id="canvas_orange_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#ff8800")}
								></button>
								<button
									id="canvas_yellow_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#ffff00")}
								></button>
								<button
									id="canvas_green_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#00ff00")}
								></button>
								<button
									id="canvas_cyan_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#00ffff")}
								></button>
								<button
									id="canvas_blue_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#0000ff")}
								></button>
								<button
									id="canvas_purple_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#cc00ff")}
								></button>
								<button
									id="canvas_white_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#ffffff")}
								></button>
								<button
									id="canvas_gray_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#888888")}
								></button>
								<button
									id="canvas_black_button"
									className="tool_button tool_color_button"
									onClick={() => this.canvasColor("#000000")}
								></button>
							</div>
						</div>
						<Button
							variant="contained"
							color="primary"
							id="pause_button"
							className="playPauseButton"
							onClick={() => {
								this.playPauseVideo();
								this.handleClick();
							}}
						>
							{playPauseText}
							{icon ? <PlayArrow /> : <Pause />}
						</Button>
					</Grid>
				</div>
			</Grid>
		);
	}
}

export default Atn;
