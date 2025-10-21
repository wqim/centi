"""
Microservice for testing purposes.
Also serves an example for creating other microservices.
"""
import json
import http.server
import socketserver
from base64 import b64encode, b64decode

PORT = 8001

public_keys = []
messages = []

class ApiHandler:
    def __init__(self):
        pass

    def _init_microservice(self, args):
        print("Initilalizing microservice with following configuration:", args)

    def _init_channels(self, channels):
        print("Creating channels:", channels)

    def _delete_channels(self, channels):
        print("Deleting channels:", channels)

    def _distribute_pk(self, params, pk):
        print("Distributing public key...")
        global public_keys
        public_keys.append( pk )

    def _collect_pks(self, params):
        print("Collecting public keys...")
        global public_keys
        # note: store current public keys for further usage
        # ( if they weren't updated somehow )
        pks = [{
            "platform": "test",
            "alias": "me",
            "content": i
        }for i in public_keys]
        return pks

    def _send(self, msg):
        print("Sending message:", msg["data"])
        global messages
        messages.append( msg )

    def _recv(self):
        print("Receiving messages...")
        global messages
        msgs = [ m for m in messages ]
        messages = []
        return msgs

    def _prepare_to_delete(self, data):
        print("Prepare-to-delete function called")
        return None

    def _delete(self, msg):
        print("Deleting message from the channel:", msg)

    def gen_response( self, msg ):
        msg_type = msg["message_type"]
        response = {"message_type": msg_type, "status": "success", "args": {}}

        if msg_type == "init_microservice":
            self._init_microservice( msg["args"] )

        elif msg_type == "init_channels":
            self._init_channels( msg["args"]["channels"] )

        elif msg_type == "delete_channels":
            self._delete_channels( msg["args"]["channels"] )
            
        elif msg_type == "distribute_pk":
            self._distribute_pk( msg["args"]["distribution_parameters"], msg["args"]["public_key"] )

        elif msg_type == "collect_pks":
            pks = self._collect_pks( msg["args"]["distribution_parameters"] )
            response["args"]["public_keys"] = pks

        elif msg_type == "send":
            self._send( msg["args"]["message"] )

        elif msg_type == "recv_messages":
            msgs = self._recv()
            response["args"]["messages"] = msgs

        elif msg_type == "prepare_to_delete":
            msg2 = self._prepare_to_delete( msg["args"]["data"] )
            response["args"]["message"] = msg2

        elif msg_type == "delete":
            self._delete( msg["args"]["message"] )

        elif msg_type == "message_from_bytes":
            args = {
                    "test_argument": "test_value"
            }
            response["args"] = args
        else:
            response["status"] = "failure"
            response["args"] = {
                    "error": f"Unknown message type: {msg_type}"
            }
        return response

    def response( self, api_message ):
        try:
            msg = json.loads( api_message )
            response = self.gen_response( msg )
            res = json.dumps( response )
            print( res )
            return res

        except Exception as e:
            print("[-] Failed to decode APIMessage:", e, "API Message:", api_message)
            return ""


class ReqHandler( http.server.SimpleHTTPRequestHandler ):

    def do_POST(self):
        content_length = int(self.headers["Content-Length"])
        post_data = self.rfile.read( content_length )
        post_data = post_data.decode()

        api_handler = ApiHandler()

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()

        response_message = api_handler.response( post_data )
        self.wfile.write( response_message.encode() )


with socketserver.TCPServer( ("127.0.0.1", PORT), ReqHandler ) as httpd:
    print(f"[*] Serving on port {PORT}")
    httpd.serve_forever()
