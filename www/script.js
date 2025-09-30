const reqPksBtn = document.getElementById("request-pks");
const keysHolder = document.getElementById("known-keys");
const keyAliasInput = document.getElementById("key-alias");
const connBtn = document.getElementById("connect-btn");
const peersHolder = document.getElementById("peers-holder");
const peerName = document.getElementById("peer");
const dataInput = document.getElementById("data");
const sendBtn = document.getElementById("send-request");
const messagesView = document.getElementById("received-messages");
const errorsView = document.getElementById("errors");
const debugView = document.getElementById("debug");

// identifiers and timestamp of messages we are waiting reply for
const Ping = 2000; // ms to ping local server.
var messages = [];
var messagesTable = document.createElement("table");
var connections = [];
var firstTime = true;

const ConnectionType = 1;
const DataType = 2;

function clearView( holder ) {
	holder.innerHTML = "";
	holder.innerText = "";
}

function renderParagraphs( pars, holder ) {
	clearView( holder );
	for (let par of pars) {
		p = document.createElement( "p" );
		p.innerText = par;
		holder.appendChild( p );
	}
}

function renderMessages() {

	//console.log("Rendering messages:", messages);
	messagesTable.innerHtml = "";
	messagesTable.innerText = "";

	tr = document.createElement("tr");
	td1 = document.createElement("td");
	td2 = document.createElement("td");
	td1.innerText = "Message ID";
	td2.innerText = "Time stamp";
	tr.appendChild( td1 );
	tr.appendChild( td2 );

	messagesTable.appendChild( tr );

	for (let m of messages) {
		tr = document.createElement("tr");
		newElem = document.createElement("td");
		newElem.innerText = m["message_id"];
		tr.appendChild( newElem );
		newElem2 = document.createElement("td");
		newElem2.innerText = m["timestamp"];
		tr.appendChild( newElem2 );
		messagesTable.appendChild( tr );
	}
	if (firstTime === true) {
		firstTime = false;
		debugView.appendChild( messagesTable );
	}
}

async function listPublicKeys() {
	try {
		const response = await fetch("/api/public-keys");
		if (!response.ok) {
			console.log(`Response status: ${response.status}`);
		}
		const json = await response.json();
		console.log( json );
		//keysHolder.innerText = json;
		const resp = json; //const resp = JSON.parse( json );
		keysHolder.innerHtml = "";
		keysHolder.innerText = "";
		for (let k of resp) {
			//console.log("Adding an element:", k["alias"]);
			newElement = document.createElement("p");
			newElement.innerText = k["alias"];
			keysHolder.appendChild( newElement );
		}
	} catch (error) {
		console.error( error.message );
	}
}

reqPksBtn.onclick = async () => {
	try {
		// check from who we want to receive public keys
		if (keyAliasInput.value.trim().length != 0) {
			const response = await fetch("/api/request-public-keys", {
				method: "POST",
				body: JSON.stringify( {
					peer_alias: keyAliasInput.value.trim(),
				}),
			});
			
			console.log( "awaited response from server:", response );

			if (!response.ok) {
				console.log(`Response status: ${response.status}`);
			}
			
			const resp = await response.json();
			console.log( resp );
			if ( resp["errors"].length != 0 ) {
				renderParagraphs( resp["errors"], errorsView );
			} else {
				clearView( errorsView );
				console.log("[+] Requested public keys from server");
				keyAliasInput.value = "";
			}
		}
	} catch(error) {
		console.error( error.message );
	}
}

connBtn.onclick = async () => {
	try {
		if (keyAliasInput.value.trim().length != 0) {
			const request = {
				key_alias: keyAliasInput.value,
			};


			const response = await fetch("/api/connect", {
				method: "POST",
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				console.log(`Response status: ${response.status}`);
			}

			const resp = await response.json();
			//console.log( resp );
			if ( resp["errors"].length == 0 ) {
				console.log("pushing a message (1)");
				messages.push({
					message_id: resp["message_id"],
					timestamp: resp["timestamp"],
					type: ConnectionType,
				});
				keyAliasInput.value = "";
				clearView( errorsView );
			} else {
				renderParagraphs( resp["errors"], errorsView );
			}
		}

	} catch (error) {
		console.error( error.message );
	}
}


sendBtn.onclick = async () => {
	try {
		if ( peerName.value.trim().length != 0 && dataInput.value.trim().length != 0 ) {
			const request = {
				dst: peerName.value,
				data: btoa(dataInput.value),
			};

			console.log("Trying to send request:", JSON.stringify(request));
			
			const response = await fetch("/api/request", {
				method: "POST",
				body: JSON.stringify(request)
			});
			console.log("[fetched]");
			if (!response.ok) {
				console.log(`Response status: ${response.status}`);
			}
			const result = await response.json();
			//console.log( result );
			if ( result["errors"].length == 0 ) {

				clearView( errorsView );
				messageDiv = document.createElement("div");
				messageDiv.className = "message-div";

				senderElem = document.createElement("p");
				senderElem.innerText = "You";
				senderElem.className = "sender-name";

				dataElem = document.createElement("p");
				dataElem.innerText = dataInput.value;

				messageDiv.appendChild( senderElem );
				messageDiv.appendChild( dataElem );
				messagesView.appendChild( messageDiv );

				//peerName.value = "";
				dataInput.value = "";

				console.log("pushing a message (2)");
				messages.push( {
					message_id: result["message_id"],
					timestamp: result["timestamp"],
					type: DataType,
				});
			} else {
				renderParagraphs( result["errors"], errorsView );
			}
		} else {
			console.log("Invalid input length");
		}

	} catch (error) {
		console.error( error.message );
	}
}

async function collectMessages() {
	try {
		const response = await( fetch("/api/messages") );
		if (!response.ok) {
			console.log(`Response status: ${response.status}`);
		} else {
			const result = await response.json();
			for (let m of result) {
				// render each message in the messages view...
				messageDiv = document.createElement("div");
				messageDiv.className = "message-div";

				senderElem = document.createElement("p");
				senderElem.innerText = m[0];
				senderElem.className = "sender-name";

				dataElem = document.createElement("p");
				dataElem.innerText = atob( m[1] );

				messageDiv.appendChild( senderElem );
				messageDiv.appendChild( dataElem );

				messagesView.appendChild( messageDiv );
			}
		}
	} catch (error) {
		console.error( error.message );
	}
}

window.setInterval( async () => {

	// we are awaiting for receiving a message
	await collectMessages();

	// and waiting some peers to be connected.
	try {
		const response = await fetch("/api/peers");
		if (!response.ok) {
			console.log(`Response status: ${response.status}`);
		}
		const result = await response.json();
		// result is an array of peers, just render it.
		renderParagraphs( result, peersHolder );
	} catch(error) {
		console.error( error.message );
	}

	await listPublicKeys();

}, Ping );
