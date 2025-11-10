import os
import sys
import RNS
from colorama import Fore
from globals import *

def client_connected( link ):
    
    while have_public_key() == False:
        pass

    #send_public_key( link )
    set_callbacks( link, packet_received )
    add_peer( link )


def packet_received( message, packet ):

    # handle message here...
    # TODO: decide if message is a public key or not ( based on size ).
    # can we make it more flexible somehow?
    if len(message) < 2048:
        add_public_key( message )
        print(Fore.LIGHTMAGENTA_EX + "[+] received a public key" + Fore.RESET, "(", len(message), ") bytes long")
        
    else:
        add_message( message )
        print( Fore.LIGHTBLUE_EX + "[*] Received a packet:", len(message), "bytes long", Fore.RESET )
