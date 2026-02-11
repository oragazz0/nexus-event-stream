"""
Unit tests for Signal model event publishing to Redpanda.

Tests cover signal creation, update, and deletion events to ensure
proper payload structure and message publishing behavior.
"""

import json
from unittest.mock import Mock, patch, call
from django.test import TransactionTestCase
from django.contrib.auth import get_user_model
from nexus.core.models import Signal


User = get_user_model()


class SignalEventPublishingTestCase(TransactionTestCase):
    """Tests for Signal model event publishing to Redpanda."""

    def setUp(self):
        """Create test user and mock producer for all tests."""
        self.user = User.objects.create_user(
            username="testuser",
            email="test@example.com",
            password="testpass123"
        )
        
        # Mock the producer to prevent actual Redpanda calls
        self.mock_producer = Mock()
        self.patcher = patch('nexus.core.signals.get_producer', return_value=self.mock_producer)
        self.patcher.start()

    def tearDown(self):
        """Clean up mocks."""
        self.patcher.stop()

    def test_signal_creation_publishes_created_event(self):
        """Test that creating a Signal publishes an event with action='created'."""
        signal = Signal.objects.create(
            title="Test Signal",
            content="This is a test signal",
            priority=Signal.Priority.HIGH,
            author=self.user
        )

        # Verify producer.produce was called
        self.mock_producer.produce.assert_called_once()
        
        # Extract call arguments
        call_kwargs = self.mock_producer.produce.call_args[1]
        
        # Verify topic
        self.assertEqual(call_kwargs['topic'], 'nexus.signals')
        
        # Verify key (should be the signal ID)
        self.assertEqual(call_kwargs['key'], str(signal.id).encode('utf-8'))
        
        # Verify payload structure
        payload = json.loads(call_kwargs['value'].decode('utf-8'))
        self.assertEqual(payload['action'], 'created')
        self.assertEqual(payload['id'], str(signal.id))
        self.assertEqual(payload['title'], 'Test Signal')
        self.assertEqual(payload['content'], 'This is a test signal')
        self.assertEqual(payload['priority'], 'High')
        self.assertEqual(payload['author'], 'testuser')
        self.assertIn('created_at', payload)
        self.assertIn('updated_at', payload)
        
        # Verify poll was called
        self.mock_producer.poll.assert_called_once_with(0)

    def test_signal_update_publishes_updated_event(self):
        """Test that updating a Signal publishes an event with action='updated'."""
        signal = Signal.objects.create(
            title="Original Title",
            content="Original content",
            priority=Signal.Priority.LOW,
            author=self.user
        )
        
        # Reset mock to ignore creation event
        self.mock_producer.reset_mock()
        
        # Update the signal
        signal.title = "Updated Title"
        signal.priority = Signal.Priority.MEDIUM
        signal.save()

        # Verify producer.produce was called for update
        self.mock_producer.produce.assert_called_once()
        
        # Extract call arguments
        call_kwargs = self.mock_producer.produce.call_args[1]
        
        # Verify topic
        self.assertEqual(call_kwargs['topic'], 'nexus.signals')
        
        # Verify payload action is 'updated'
        payload = json.loads(call_kwargs['value'].decode('utf-8'))
        self.assertEqual(payload['action'], 'updated')
        self.assertEqual(payload['title'], 'Updated Title')
        self.assertEqual(payload['priority'], 'Medium')

    def test_signal_deletion_publishes_deleted_event(self):
        """Test that deleting a Signal publishes an event with action='deleted'."""
        signal = Signal.objects.create(
            title="To Be Deleted",
            content="This signal will be deleted",
            priority=Signal.Priority.LOW,
            author=self.user
        )
        
        signal_id = signal.id
        
        # Reset mock to ignore creation event
        self.mock_producer.reset_mock()
        
        # Delete the signal
        signal.delete()

        # Verify producer.produce was called for deletion
        self.mock_producer.produce.assert_called_once()
        
        # Extract call arguments
        call_kwargs = self.mock_producer.produce.call_args[1]
        
        # Verify topic
        self.assertEqual(call_kwargs['topic'], 'nexus.signals')
        
        # Verify key
        self.assertEqual(call_kwargs['key'], str(signal_id).encode('utf-8'))
        
        # Verify payload structure (minimal for delete)
        payload = json.loads(call_kwargs['value'].decode('utf-8'))
        self.assertEqual(payload['action'], 'deleted')
        self.assertEqual(payload['id'], str(signal_id))
        # Deleted events should only contain action and id
        self.assertEqual(len(payload), 2)

    def test_signal_priority_display_values(self):
        """Test that priority values are human-readable strings."""
        test_cases = [
            (Signal.Priority.LOW, 'Low'),
            (Signal.Priority.MEDIUM, 'Medium'),
            (Signal.Priority.HIGH, 'High'),
        ]
        
        for priority_value, expected_display in test_cases:
            with self.subTest(priority=priority_value):
                signal = Signal.objects.create(
                    title=f"Test {expected_display}",
                    content="Testing priority display",
                    priority=priority_value,
                    author=self.user
                )
                
                # Get the payload from the mock call
                call_kwargs = self.mock_producer.produce.call_args[1]
                payload = json.loads(call_kwargs['value'].decode('utf-8'))
                
                # Verify priority is the display string, not the integer
                self.assertEqual(payload['priority'], expected_display)
                self.assertNotEqual(payload['priority'], priority_value)
                
                # Clean up
                signal.delete()
                self.mock_producer.reset_mock()

    def test_multiple_signals_publish_separate_events(self):
        """Test that creating multiple signals publishes separate events."""
        Signal.objects.create(
            title="First Signal",
            content="Content 1",
            priority=Signal.Priority.LOW,
            author=self.user
        )
        
        Signal.objects.create(
            title="Second Signal",
            content="Content 2",
            priority=Signal.Priority.HIGH,
            author=self.user
        )

        # Should have two produce calls
        self.assertEqual(self.mock_producer.produce.call_count, 2)
        
        # Verify both have action='created'
        for call_item in self.mock_producer.produce.call_args_list:
            payload = json.loads(call_item[1]['value'].decode('utf-8'))
            self.assertEqual(payload['action'], 'created')

    def test_author_username_in_payload(self):
        """Test that author appears as username, not ID."""
        signal = Signal.objects.create(
            title="Test Signal",
            content="Testing author field",
            priority=Signal.Priority.MEDIUM,
            author=self.user
        )

        call_kwargs = self.mock_producer.produce.call_args[1]
        payload = json.loads(call_kwargs['value'].decode('utf-8'))
        
        # Verify author is username string
        self.assertEqual(payload['author'], 'testuser')
        self.assertIsInstance(payload['author'], str)
        # Ensure it's not the ID
        self.assertNotEqual(payload['author'], self.user.id)
        self.assertNotIn('author_id', payload)

    def test_transaction_commit_behavior(self):
        """Test that events are only published after transaction commits."""
        with patch('nexus.core.signals.transaction.on_commit') as mock_on_commit:
            signal = Signal.objects.create(
                title="Transaction Test",
                content="Testing transaction behavior",
                priority=Signal.Priority.LOW,
                author=self.user
            )
            
            # Verify on_commit was called with a callable
            mock_on_commit.assert_called_once()
            callback = mock_on_commit.call_args[0][0]
            self.assertTrue(callable(callback))
