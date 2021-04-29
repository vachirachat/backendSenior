from concurrent import futures
import time
import math
import logging

from pymongo import MongoClient
import grpc
from bson import ObjectId

import backup_pb2
import backup_pb2_grpc


class BackupServicer(backup_pb2_grpc.BackupServicer):
    """Provides methods that implement functionality of route guide server."""
    def __init__(self):
        self.conn = MongoClient('mongodb://172.17.0.2:27017')

    def OnMessageIn(self, request, context):
        print("Access OnMessageIn","OnMessageIn")
        print(request)

        temp_dict = {
            "_id": ObjectId(request.messageId),
            "timestamp": request.timestamp,
            "roomId":    ObjectId(request.roomId),
            "userId":    ObjectId(request.userId),
            "clientUID": request.clientUid,
            "data":      request.data,
            "type":      request.type,
        }

        db = self.conn['backup']
        dbCollection = db.message
        result = dbCollection.insert_one(temp_dict)
        # Return must look like interface class in [name]_pb2_GRPC
        return backup_pb2.Empty()

    def IsReady(self, request, context):
        print("Access IsReady","IsReady")
        print(request)
        # Look like class in Python, you have to manual add attribute follow backup_pb2
        # is this case Status have 1 attribute is Status{ok:bool}, you have to assign
        response = backup_pb2.Status()
        response.ok = True
        # Return must look like interface class in [name]_pb2_GRPC
        return response


if __name__ == '__main__':
    print("start-server")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=5))
    server.add_insecure_port('172.17.0.3:5005')
    backup_pb2_grpc.add_BackupServicer_to_server(BackupServicer(), server)

    try:
        server.start()
        print('Running Discount service on %s' %":5005")
        while True:
            time.sleep(1)
    except Exception as e:
        print('[error] %s' % e)
        server.stop(0)