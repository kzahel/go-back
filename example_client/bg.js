
function test_game() {

    var STATE_IDLE = 0
    var STATE_CONNECTING = 1
    var STATE_AUTHED = 2
    var STATE_READY = 3

    var state = STATE_IDLE

    var ws = new WebSocket('ws://jstorrent.com:12345/echo');
    function onwsevent(evt, cb) {
	if (evt.data) {
	    var fr = new FileReader
	    fr.onload = function(evt) {
		var data = JSON.parse(evt.target.result)
		cb(data)

	    }
	    fr.readAsText( evt.data );
	}
    }
    ws.onopen = function(evt) {
	console.log('ws open!, authenticating')
	ws.send( JSON.stringify({command:"Authenticate", data:FB.getAccessToken()}) )
	state = STATE_AUTHED
    }

    ws.onclose = onwsevent
    ws.onerror = onwsevent
    ws.onmessage = function(evt) {
	console.log('onmessage',evt)
	if (state != STATE_READY) {
	    ws.send( JSON.stringify({command:"NewGame",invite:"12345"} ) )
	    state = STATE_READY
	}
	onwsevent(evt, function(data) {
	    console.log('msg from server',data);
	})
    }
}


/*
		console.log('server returns data',data)
		ws.send( JSON.stringify( {command: "JoinGame",
					  data: data.Roomid} ) );
*/