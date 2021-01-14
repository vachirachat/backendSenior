from concurrent import futures
import time
import math
import logging

from pymongo import MongoClient
import grpc


import backup_pb2
import backup_pb2_grpc


class BackupServicer(backup_pb2_grpc.BackupServicer):
    """Provides methods that implement functionality of route guide server."""
    def __init__(self):
        self.tricker = "triker"
        self.conn = MongoClient('mongodb://localhost:27017')

    def OnMessageIn(self, request, context):
        print(request)
        # data = backup_pb2.
        print("Access OnMessageIn", self.tricker,"OnMessageIn")
        db = self.conn['backup']
        dbCollection = db.message
        result = dbCollection.insert_one(request)
        return backup_pb2.Empty()
    def IsReady(self, request, context):
        print(request)
        # data = backup_pb2.
        print("Access IsReady", self.tricker,"IsReady")
        return backup_pb2.Status()


if __name__ == '__main__':
    print("start-server")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=5))
    server.add_insecure_port('[::]:5005')
    backup_pb2_grpc.add_BackupServicer_to_server(BackupServicer(), server)

    try:
        server.start()
        print('Running Discount service on %s' %"5005")
        while True:
            time.sleep(1)
    except Exception as e:
        print('[error] %s' % e)
        server.stop(0)