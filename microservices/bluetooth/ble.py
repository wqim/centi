import asyncio
from bleak import BleakScanner, BleakClient
from bleak.backends.characteristic import BleakGATTCharacteristic

async def scan_devices( characteristic_uuid ):
    devices = await BleakScanner.discover()
    for device in devices:
        print(f'Device: {device.address} - {device.name}')
        try:
            async with BleakClient( device.address ) as client:
                await client.connect()
                print(f'[*] Connected to {device.address}')
                for service in client.services:
                    for char in service.characteristics:
                        if char.uuid == characteristic_uuid:
                            print(f'[+] Found device in network: {device.address}')
        except Exception as e:
            print(f'[-] scan_devices: {e}')

async def find_device_by_addr( address, macos_use_bdaddr = False ):
    device = await BleakScanner.find_device_by_address(
            address, cb = {'use_bdaddr': macos_use_bdaddr}
    )
    return device

def sample_notification_handler( characteristic: BleakGATTCharacteristic, data: bytearray ):
    print(characteristic.description, ':', data)

async def enable_notifications( bleak_client, characteristic, notification_handler ):
    await bleak_client.start_notify( characteristic, notification_handler )

"""

async def send_and_receive_data( device_address, characteristic_uuid, pair = False, timeout = 10 ):
    if (timeout == 10) and pair:
        timeout = 90
    
    # connect to device
    async with BleakClient( device_address ) as client:
        await client.connect()
        print(f'Connected to {device_address}')

        for service in client.services:
            print(f'[Service] {service}')
            for char in service.characteristics:
                if 'read' in char.properties:
                    try:
                        value = await client.read_gatt_char( char )
                        extra = f', Value: {value}'
                    except Exception as e:
                        extra = f', Error: {e}'
                else:
                    extra = ''

                if 'write-without-response' in char.properties:
                    extra += f', Max write w/o rsp size: {char.max_write_without_response_size}'
                print('Characteristic:', char, '(', ','.join(char.properties), ')', extra )
                for descriptor in char.descriptors:
                    try:
                        value = await client.read_gatt_descriptor( descriptor )
                        print('[Descriptor]:', descriptor, ', Value:', value)
                    except Exception as e:
                        print('[Descriptor]:', descriptor, ', Error:', e)


            
        # write data to a characteristic
        data_to_send = get_next_message()
        if data_to_send:
            await client.write_gatt_char( characteristic_uuid, data_to_send )
            print('[+] sent data')

        received_data = await client.read_gatt_char( characteristic_uuid )
        print('[+] received data: {received_data}')
"""
'''
loop = asyncio.get_event_loop()
loop.run_until_complete( scan_devices() )
'''
