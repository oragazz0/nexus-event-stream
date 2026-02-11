from django.apps import AppConfig


class CoreConfig(AppConfig):
    default_auto_field = "django.db.models.BigAutoField"
    name = "nexus.core"
    verbose_name = "Core"

    def ready(self):
        import nexus.core.signals  # noqa: F401
