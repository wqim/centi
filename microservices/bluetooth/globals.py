import os
import asyncio
from threading import Lock
from base64 import b64encode, b64decode

public_keys = []
pk_mtx = Lock()
messages = []
msgs_mtx = Lock()
peers = []
peers_mtx = Lock()

# default sender of all the messages we receiving
UnknownSender = ""

publicKey = None
publicKeyMtx = Lock()

def have_public_key():
    global publicKeyMtx
    global publicKey

    publicKeyMtx.acquire()
    no_pk = (publicKey is None)
    publicKeyMtx.release()
    return not(no_pk)

def set_public_key( pubkey ):
    global publicKeyMtx
    global publicKey

    publicKeyMtx.acquire()
    publicKey = b64decode( pubkey )
    publicKeyMtx.release()

def get_public_keys():
    global pk_mtx
    global public_keys
    pk_mtx.acquire()
    pks = [pk for pk in public_keys]
    pk_mtx.release()
    return pks

async def send_public_key( client ):
    global publicKeyMtx
    global publicKey

    publicKeyMtx.acquire()
    await client.write_gatt_char( publicKey )
    publicKeyMtx.release()

def get_peers():
    global peers_mtx
    global peers
    peers_mtx.acquire()
    tmp = [i for i in peers]
    peers_mtx.release()
    return tmp

def add_peer( device ):
    global peers_mtx
    global peers
    peers_mtx.acquire()
    peers.append( device )
    peers_mtx.release()

def remove_peer( device ):
    global peers_mtx
    global peers
    peers_mtx.acquire()
    for i, dev in enumerate(peers):
        if dev == device:
            try:
                dev.close()
            except:
                pass
            peers = peers[:i] + peers[i+1:]
            break
    peers_mtx.release()

def clear_peers():
    global peers_mtx
    global peers
    peers_mtx.acquire()
    for i, dev in enumerate(peers):
        try:
            peer.close()
        except:
            pass
    peers_mtx.release()

def add_public_key( pk_raw ):
    global public_keys
    global pk_mtx

    alreadyHave = False
    content = b64encode( pk_raw ).decode()
    
    pk_mtx.acquire()

    for pk in public_keys:
        if pk.get('content', '') == content:
            alreadyHave = True
            break

    if alreadyHave == False:
        public_keys.append({ # generate random identifier for a public key.
                "alias": b64encode( os.urandom(16) ).decode(),
                "platform": "reticulum",
                "content": content
        })
    pk_mtx.release()


def add_message( bytes_ ):
    global messages
    global msgs_mtx

    msgs_mtx.acquire()
    messages.append( {
        "platform": "reticulum",
        "data": b64encode(bytes_).decode(),
        "sender": UnknownSender,
        "sent_by_us": False,
        "args": {}
    })
    msgs_mtx.release()

def get_messages():
    global messages
    global msgs_mtx

    msgs_mtx.acquire()
    msgs = [i for i in messages]
    messages = []
    msgs_mtx.release()

    return msgs
