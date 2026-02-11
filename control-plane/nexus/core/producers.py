"""
Kafka producer singleton for publishing events to Redpanda.

This module maintains a single confluent-kafka Producer instance
that is reused across the entire Django process. The producer is
thread-safe and manages its own internal buffer and background threads.
"""

from confluent_kafka import Producer
from django.conf import settings

import atexit

_producer = None


def get_producer():
    """
    Returns a singleton Kafka producer instance.
    
    The producer is configured with:
    - acks=all: Wait for all in-sync replicas to acknowledge
    - enable.idempotence=True: Exactly-once semantics at broker level
    
    Returns:
        Producer: Thread-safe confluent-kafka Producer instance
    """
    global _producer
    if _producer is None:
        _producer = Producer({
            "bootstrap.servers": settings.REDPANDA_BROKERS,
            "acks": "all",
            "enable.idempotence": True,
        })
        atexit.register(flush_producer)
    return _producer


def flush_producer():
    """
    Flushes any pending messages in the producer buffer.
    
    This should be called during application shutdown to ensure
    all buffered messages are delivered before the process exits.
    """
    global _producer
    if _producer is not None:
        _producer.flush(timeout=5)
