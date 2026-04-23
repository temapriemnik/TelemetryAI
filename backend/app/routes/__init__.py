from .auth import router as auth_router
from .users import router as users_router
from .projects import router as projects_router
from .reviews import router as reviews_router
from .admin import router as admin_router

__all__ = ["auth_router", "users_router", "projects_router", "reviews_router", "admin_router"]