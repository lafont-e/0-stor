import grpc

from .generated import daemon_pb2_grpc as stubs
from .generated import daemon_pb2 as model

from . import namespace
from . import file as file


class Client:
    def __init__(self, address='172.0.0.1:8080'):
        channel = grpc.insecure_channel(address)

        # initialize stubs
        self._namespace = namespace.Namespace(channel)
        self._file = file.File(channel)

    @property
    def namespace(self):
        return self._namespace

    @property
    def file(self):
        return self._file
