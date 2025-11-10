import os
import http.server
import socketserver

class ReqHandler( http.server.SimpleHTTPRequestHandler ):

    def do_GET(self):
        self.send_response(403)

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


if __name__ == '__main__':
    port = int(os.environ.get('BLE_PORT', '3333'))
    with socketserver.TCPServer( ("127.0.0.1", port), ReqHandler ) as httpd:
        print(f"[*] Serving on port {port}")
        httpd.serve_forever()
