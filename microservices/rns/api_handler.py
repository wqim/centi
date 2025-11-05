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

from globals import *
from callbacks import *
from announce_handler import AnnounceHandler

def announce_loop( dst ):
    while True:
        dst.announce()
        time.sleep( randint(1, 10) / 10 )

class ApiHandler:
    # dummy constructor
    def __init__(self):
        #print('__init__()')
        self.app_name = 'centi-rns-microservice'
        self.aspect_filter = 'centi-rns-microservice.announce'
        self.tags = []
        self.running = False
        self.autoconf = False
        self.servers = set()
        self.peers = []
        self.mtx = Lock() # mutex is related to peer's list (self.peers)

    def _init_microservice(self, args):
        print("Initilalizing microservice with following configuration:", args)

        config_path = args.get("config_path", "")
        self.max_attempts = int( args.get("max_attempts", "10") )
        self.rns = RNS.Reticulum( config_path )
        self.identity = RNS.Identity()

        run_as_server = args.get("run_as_server", False)
        if run_as_server in ["true", "True"]:
            print("[+] running as both server and client....")
            # server-related part: just announce ourselves and work in background.
            self.server_dst = RNS.Destination(
                    self.identity,
                    RNS.Destination.IN,
                    RNS.Destination.SINGLE,
                    self.app_name,
                    'announce'
            )

            self.server_dst.set_link_established_callback( client_connected )
            # announce ourselves in the background
            Thread( target = announce_loop, args=(self.server_dst,) ).start()

        # autoconf = autodiscovery, just look for announcing nodes
        self.autoconf = args.get("autodiscovery", False)
        if self.autoconf == "true" or self.autoconf == "True":
            self.autoconf = True
            announce_handler = AnnounceHandler(
                    callback = self,
                    aspect_filter = self.aspect_filter,
            )
            RNS.Transport.register_announce_handler( announce_handler )
        #print("_init_microservice():", len(self.peers) )

    def _add_known_server(self, destination_hash):
        if have_public_key():
            if destination_hash not in self.servers:
                self.servers.add( destination_hash )
                # print("new amount of known servers(nodes):", len(self.servers))
                self._connect_to_addr( destination_hash )

    def _init_channels(self, channels):
        servers_list = channels
        #print("_init_channels():", len(self.peers) )
        self._connect_to_nodes( servers_list )
        #print("_init_channels() (2):", len(self.peers) )

    def _connect_to_nodes(self, servers_list):
        # client-related
        if have_public_key():
            for _, addr in servers_list:
                #print( addr )
                self._connect_to_addr( addr )

    def _connect_to_addr(self, addr):
        # check if name of node is valid
        if have_public_key() == False:
            return

        try:
            dst_len = (RNS.Reticulum.TRUNCATED_HASHLENGTH // 8) * 2
            if len(addr) != dst_len:
                print("Destination length is invalid. Must be {hex} hexadecimal characters ({byte} bytes)".format(
                    hex = dst_len, byte = dst_len // 2
                ))
                sys.exit(-1)

            destination_hash = bytes.fromhex( addr )
        except Exception as e:
            print(e)
            return

        attempt = 0
        # check if we are able to connect to server
        if not RNS.Transport.has_path( destination_hash ):
            # okay, we don't know thee path to a target node
            # just request it
            RNS.Transport.request_path( destination_hash )
            while not RNS.Transport.has_path( destination_hash ):
                time.sleep( 0.1 )
                attempt += 1
                if attempt == self.max_attempts:
                    break

        # we succeed in connecting to the server
        if attempt < self.max_attempts:
            peer_identity = RNS.Identity.recall( destination_hash )
            peer_destination = RNS.Destination(
                    peer_identity,
                    RNS.Destination.OUT,
                    RNS.Destination.SINGLE,
                    self.app_name,
                    'announce'
            )
            link = RNS.Link( peer_destination )
            #send_public_key( link )
            set_callbacks( link, packet_received )
            link.set_link_established_callback( client_connected )

            # we can use this link later during data (re)sending
            self.mtx.acquire()
            self.peers.append( link ) # peer, received_pk
            print("[*] self._connect_to_addr(): len(self.peers) =", len(self.peers))
            self.mtx.release()
            RNS.log(Fore.LIGHTGREEN_EX + "[+] successfully connected to " + Fore.RESET + RNS.prettyhexrep( destination_hash ) )
        else:
            RNS.log(Fore.LIGHTRED_EX +"[-] failed to connect to " + Fore.RESET + RNS.prettyhexrep( destination_hash ) )
        
        #print("[*] self._connect_to_addr() (2): len(self.peers) =", len(self.peers))

    def _delete_channels(self, channels):
        print("Deleting channels:", channels)
        #print("_delete_channels():", len(self.peers) )
        for link in self.peers:
            link.teardown()
        time.sleep( 1.5 )
        self.peers = []
        self.servers = set()
        # should we delete channels we have connected to...?

    def _distribute_pk(self, params, pk):
        print("Distributing public key... (already done by callback functions.)")
        set_public_key( pk )
        #print("_distribute_pk():", len(self.peers) )
        # send public key as a message to everyone we didn't sent it yet.
        #self._send( {"data": pk}, True )

    def _collect_pks(self, params):
        print("Collecting public keys...")
        #print("_collect_pks():", len(self.peers) )
        pks = get_public_keys()
        print("[*] Amount of public keys collected:", len(pks))
        # note: store current public keys for further usage
        # ( if they weren't updated somehow )
        return pks

    def _send_to_peers(self, peers_, msg):
        #print("_send_to_peers():", len(self.peers), len(peers_) )
        for i in range(len(peers_)):
            peer = peers_[i]
            # sending public key, obviously
            decoded = b64decode( msg["data"] )
            RNS.Packet( peer, decoded ).send()


    def _send(self, msg):
        # print("Sending message:", msg["data"])
        #print("Senging message...")
        peers_ = get_peers()
        ##print("[*] Amount of peers:", len(peers_))
        self._send_to_peers( peers_, msg )

        self.mtx.acquire()
        #print("[*] Amount of self.peers:", len(self.peers))
        self._send_to_peers( self.peers, msg )
        self.mtx.release()

    def _recv(self):
        #print("Receiving messages...")
        msgs = get_messages()
        #print(Fore.LIGHTCYAN_EX + "[*] Amount of received messages:" + Fore.RESET, len(msgs))
        return msgs

    # for reticulum it's just a couple of dummies here...
    def _prepare_to_delete(self, data):
        return None
    def _delete(self, msg):
        pass # print("Deleting message from the channel:", msg)

    # function which really handles the api call and all the format
    # transformations.
    def gen_response( self, msg ):
        msg_type = msg["message_type"]
        response = {"message_type": msg_type, "status": "success", "args": {}}
        # print(msg_type)
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
            response["args"] = {}
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
            #print( res )
            return res

        except Exception as e:
            print("[-] Failed to decode APIMessage:", e)
            return ""
