from django.contrib.auth.models import Group, User
from rest_framework import serializers

from nexus.core.models import Signal

class UserSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = User
        fields = ['url', 'username', 'email', 'groups', 'signals']

class GroupSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = Group
        fields = ['url', 'name']

class SignalSerializer(serializers.ModelSerializer):
    class Meta:
        model = Signal
        fields = ['url', 'title', 'content', 'priority', 'author', 'created_at', 'updated_at']