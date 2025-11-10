import socket
import asyncio
from threading import Thread, Lock
from random import randint
from time import sleep

from globals import *
from ble import *

# general algorithm of ble part's work:
# 1. ensure bluetooth is enabled.
# 2. advertize ourselves.
# [in background]:
#       3. scan for devices near.
#       4. scan devices for specified service/characteristic.
#       5. connect to devices if they are related to centi network.
#       6. receive/send packets as usual.

default_channel = 4
messages_queue = []
mq_mtx = Lock()

def next_message():
    global mq_mtx
    global messages_queue
    mq_mtx.acquire()
    if messages_queue:
        msg = messages_queue[0]
        messages_queue = messages_queue[1:]
    else:
        msg = None
    mq_mtx.release()
    return msg

def push_message(msg):
    global mq_mtx
    global messages_queue
    mq_mtx.acquire()
    messages_queue.append( msg )
    mq_mtx.release()


def bool_from_str(s):
    s = s.strip().lower()
    if s == "true" or s == "1":
        return True
    return False

def get_mac_addr():
    pass

def get_known_devices():
    # return a list of [(mac_addr, channel)]
    # for all the nearby centi nodes
    pass

def recv_messages( conn, pksize = 1292, packet_size = 4096 ):
    peer_pk = conn.recv( pksize )
    add_public_key( peer_pk )
    while True:
        data = conn.recv( packet_size )
        add_message( data )

def handle_connection( conn ):
    # receive and send at the same time, mmm :P
    Thread( target = recv_messages, args=(conn,) ).start()
    while True:
        msg = next_message()
        if msg:
            conn.send( msg )

def run_client_in_background( characteristic_uuid ):
    conns = set()
    while True:
        addrs = get_known_devices() # [ (mac_addr, channel) ]
        for addr in addrs:
            if not(addr[0] in conns):
                # found a new node in the network, connecting....
                cli = socket.socket( socket.AF_BLUETOOTH, socket.SOCK_STREAM, socket.BTPROTO_RFCOMM )
                try:
                    cli.connect( addr )
                    conns.add( addr[0] )
                    Thread( target = handle_connection, args = (cli,) ).start()
                except:
                    cli.close()
        sleep( randint(1, 5) )


def run_server_in_background( server, characteristic_uuid ):
    while True:
        cli, addr = server.accept()
        print(f"[+] Accepted connection at {addr}")
        Thread( target = handle_connection, args=(cli,) ).start()

class ApiHandler:
    def __init__(self):
        # can be any adequate value
        self.characteristic_uuid = 'centi-bluetooth-characteristic-uuid'
        self.autodiscovery = False
        self.run_as_server = False
        self.channel = default_channel

    def _init_microservice(self, args):
        # setup basic things.
        self.autodiscovery = bool_from_str( args.get("autodiscovery", 'false') )
        self.run_as_server = bool_from_str( args.get("run_as_server", "false") )
        try:
            self.channel = int( args.get("channel_no", str(default_channel) ) )
        except:
            self.channl = default_channel

        print("Initilalizing microservice with following configuration:", args)

    def _init_channels(self, channels):
        print("Creating channels:", channels)
        # run bluetooth 'server' part
        if self.run_as_server:
            server = socket.socket( socket.AF_BLUETOOTH, socket.SOCK_STREAM, socket.BTPROTO_RFCOMM )
            try:
                server.bind( ( get_mac_addr(), self.channel ) )
                server.listen()
                Thread( target = run_server_in_background, args=(server, self.characteristic_uuid) ).start()
            except Exception as e:
                print(f"Failed to bind bluetooth server: {e}")

        # scan for nearby devices participating in the network
        # in the separate thread.
        Thread( target = run_client_in_background, args=(self.characteristic_uuid,) ).start()


    def _delete_channels(self, channels):
        print("Deleting channels:", channels)
        clear_peers()

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
            "platform": "bluetooth",
            "alias": "bluetooth:<peer-address-hash>",
            "content": i
        } for i in public_keys]
        return pks

    def _send(self, msg):
        print("Sending message:", msg["data"])
        push_message( msg )

    def _recv(self):
        print("Receiving messages...")
        return get_messages()

    def _delete(self, msg):
        pass

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
