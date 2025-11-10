import os
import sys
import RNS
import json
import time
import argparse
import http.server
import socketserver
from threading import Thread, Lock
from random import randint
from base64 import b64encode, b64decode
from colorama import Fore

from api_handler import ApiHandler


api_handler = ApiHandler()
api_mtx = Lock()

# api server-related class.
class ReqHandler( http.server.SimpleHTTPRequestHandler ):

    def do_GET(self):
        # drop directory listing here
        self.send_response( 403 )

    def do_POST(self):

        global api_handler
        global api_mtx

        content_length = int(self.headers["Content-Length"])
        post_data = self.rfile.read( content_length )
        post_data = post_data.decode()

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()

        api_mtx.acquire()
        response_message = api_handler.response( post_data )
        api_mtx.release()
        self.wfile.write( response_message.encode() )


if __name__=='__main__':
    PORT = int( os.environ.get('RNS_PORT', '9000') )
    with socketserver.TCPServer( ("127.0.0.1", PORT), ReqHandler ) as httpd:
        print(f"[*] Serving on port {PORT}")
        httpd.serve_forever()
