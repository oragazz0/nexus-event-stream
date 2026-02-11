"""
Django signal handlers for publishing Signal model events to Redpanda.

These handlers listen to post_save and post_delete signals from the Signal model
and publish corresponding events to the 'nexus.signals' topic. Each event includes
an 'action' field indicating the CRUD operation type. Events are only published
after the database transaction commits successfully.
"""

import json
from django.db import transaction
from django.db.models.signals import post_save, post_delete
from django.dispatch import receiver

from nexus.core.models import Signal
from nexus.core.producers import get_producer


def _publish(topic, key, payload):
    """
    Publishes a message to Redpanda.
    
    Args:
        topic: Kafka topic name
        key: Message key (used for partitioning)
        payload: Dictionary to serialize as JSON
    """
    producer = get_producer()
    producer.produce(
        topic=topic,
        key=str(key).encode("utf-8"),
        value=json.dumps(payload).encode("utf-8"),
    )
    producer.poll(0)


def _signal_payload(instance):
    """
    Serializes a Signal instance to a dictionary for event publishing.
    
    Args:
        instance: Signal model instance
        
    Returns:
        Dictionary containing signal data
    """
    return {
        "id": str(instance.id),
        "title": instance.title,
        "content": instance.content,
        "priority": instance.get_priority_display(),
        "author": instance.author.username,
        "created_at": instance.created_at.isoformat(),
        "updated_at": instance.updated_at.isoformat(),
    }


@receiver(post_save, sender=Signal)
def publish_signal_save(sender, instance, created, **kwargs):
    """
    Publishes an event when a Signal is created or updated.
    
    Publishes to 'nexus.signals' topic with action field set to
    'created' for new records or 'updated' for modified records.
    """
    action = "created" if created else "updated"
    payload = {
        "action": action,
        **_signal_payload(instance)
    }
    
    transaction.on_commit(lambda: _publish("nexus.signals", instance.id, payload))


@receiver(post_delete, sender=Signal)
def publish_signal_delete(sender, instance, **kwargs):
    """
    Publishes an event when a Signal is deleted.
    
    Publishes to 'nexus.signals' topic with action field set to 'deleted'
    and minimal payload containing only the ID.
    """
    payload = {
        "action": "deleted",
        "id": str(instance.id)
    }
    
    transaction.on_commit(lambda: _publish("nexus.signals", instance.id, payload))
