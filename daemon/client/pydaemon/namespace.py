from .generated import daemon_pb2_grpc as stubs
from .generated import daemon_pb2 as model


class Namespace:
    def __init__(self, channel):
        self._stub = stubs.NamespaceServiceStub(channel)

    def create(self, namespace):
        return self._stub.CreateNamespace(
            model.CreateNamespaceRequest(namespace=namespace)
        )

    def delete(self, namespace):
        return self._stub.DeleteNamespace(
            model.DeleteNamespaceRequest(namespace=namespace)
        )

    def get_permission(self, namespace, user):
        return self._stub.GetPermission(
            model.GetPermissionRequest(namespace=namespace, userID=user)
        )

    def set_permission(self, namespace, user, admin=False, read=False, write=False, delete=False):
        perm = model.Permission(admin=admin, read=read, write=write, delete=delete)
        return self._stub.SetPermission(
            model.SetPermissionRequest(namespace=namespace, userID=user, permission=perm)
        )
