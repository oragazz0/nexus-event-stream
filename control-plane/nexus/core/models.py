import uuid
from django.db import models
from django.conf import settings

PRIORITY_CHOICES = [
    (1, 'Low'),
    (2, 'Medium'),
    (3, 'High'),
]

class Signal(models.Model):
    """
    Represents a directive or message to be broadcasted to the system.
    """

    class Priority(models.IntegerChoices):
        LOW = 1, 'Low'
        MEDIUM = 2, 'Medium'
        HIGH = 3, 'High'

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False) # As data might originate from different places, I'll stick with UUIDs to prevent collision.

    author = models.ForeignKey(
        settings.AUTH_USER_MODEL, # Django best practice for custom user models (even if using the default one for now, this code is "future-proof").
        on_delete=models.CASCADE,
        related_name='signals',
    )

    title = models.CharField(max_length=255)
    content = models.TextField()

    priority = models.IntegerField(
        choices=Priority.choices,
        default=Priority.LOW,
    )

    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = 'Signal'
        verbose_name_plural = 'Signals'
    
    def __str__(self):
        return f"[{self.get_priority_display()}] {self.title}"