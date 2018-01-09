import grpc

from .schema import daemon_pb2_grpc as stubs
from .schema import daemon_pb2 as model


class Client:
    def __init__(self, address='172.0.0.1:8080'):
        channel = grpc.insecure_channel(address)

        # initialize stubs
        self._namespace = stubs.NamespaceServiceStub(channel)

    @property
    def namespace(self):
        return self._namespace
