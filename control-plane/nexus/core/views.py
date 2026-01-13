from django.contrib.auth.models import Group, User
from rest_framework import permissions, viewsets

from nexus.core.models import Signal
from nexus.core.serializers import GroupSerializer, SignalSerializer, UserSerializer

class UserViewSet(viewsets.ModelViewSet):
    queryset = User.objects.all().order_by('-date_joined')
    serializer_class = UserSerializer
    permission_classes = [permissions.IsAuthenticated]

class GroupViewSet(viewsets.ModelViewSet):
    queryset = Group.objects.all().order_by('-name')
    serializer_class = GroupSerializer
    permission_classes = [permissions.IsAuthenticated]

class SignalViewSet(viewsets.ModelViewSet):
    queryset = Signal.objects.all().order_by('-created_at')
    serializer_class = SignalSerializer
    permission_classes = [permissions.IsAuthenticated]