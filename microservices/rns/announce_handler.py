import RNS

class AnnounceHandler:
    # get all the announces by default...?
    def __init__(self, callback=None, aspect_filter = None):
        self.aspect_filter = aspect_filter
        self.callback = callback

    def received_announce(self, destination_hash, announced_identity, app_data):
        #RNS.log("[*] received an announce from " + RNS.prettyhexrep(destination_hash))
        #if app_data:
        #    RNS.log('The announce contained the following app data: ' + b64encode(app_data).decode() )
        
        if self.callback:
            #print("[+] found callback function, calling it", self.callback, "(", RNS.prettyhexrep(destination_hash)[1:-1], ")" )
            self.callback._add_known_server( RNS.prettyhexrep(destination_hash)[1:-1] )
