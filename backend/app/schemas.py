from pydantic import BaseModel, EmailStr
from typing import Optional
from datetime import datetime
from enum import Enum


class UserRole(str, Enum):
    USER = "user"
    ADMIN = "admin"


class TokenType(str, Enum):
    ACCESS = "access"
    REFRESH = "refresh"


class ProjectStatus(str, Enum):
    ACTIVE = "active"
    ARCHIVED = "archived"


class Token(BaseModel):
    access_token: str
    refresh_token: str
    token_type: str = "bearer"


class TokenData(BaseModel):
    user_id: int | None = None
    email: str | None = None


class UserBase(BaseModel):
    email: EmailStr
    full_name: str | None = None


class UserCreate(UserBase):
    password: str


class UserUpdate(BaseModel):
    email: EmailStr | None = None
    full_name: str | None = None
    password: str | None = None
    avatar_url: str | None = None
    bio: str | None = None


class UserResponse(UserBase):
    id: int
    avatar_url: str | None = None
    bio: str | None = None
    role: UserRole = UserRole.USER
    is_active: bool = True
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ProjectBase(BaseModel):
    name: str
    description: str | None = None


class ProjectCreate(ProjectBase):
    webhook_url: str | None = None


class ProjectUpdate(BaseModel):
    name: str | None = None
    description: str | None = None
    webhook_url: str | None = None
    status: ProjectStatus | None = None


class ProjectResponse(ProjectBase):
    id: int
    owner_id: int
    webhook_url: str | None = None
    status: ProjectStatus = ProjectStatus.ACTIVE
    created_at: datetime
    updated_at: datetime
    api_keys: list["APIKeyResponse"] = []

    class Config:
        from_attributes = True


class APIKeyBase(BaseModel):
    name: str | None = None


class APIKeyCreate(APIKeyBase):
    pass


class APIKeyResponse(BaseModel):
    id: int
    key: str
    name: str | None = None
    is_active: bool = True
    created_at: datetime

    class Config:
        from_attributes = True


class ReviewBase(BaseModel):
    rating: int
    content: str


class ReviewCreate(ReviewBase):
    project_id: int


class ReviewUpdate(BaseModel):
    rating: int | None = None
    content: str | None = None


class ReviewResponse(ReviewBase):
    id: int
    user_id: int
    project_id: int
    created_at: datetime
    updated_at: datetime
    user: UserResponse

    class Config:
        from_attributes = True


class AdminStats(BaseModel):
    total_users: int
    total_projects: int
    total_reviews: int
    users_by_day: list[dict]
    projects_by_day: list[dict]