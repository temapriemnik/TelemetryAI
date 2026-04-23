from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from sqlalchemy import func
from datetime import datetime, timedelta
from typing import List
from . import models, schemas
from .database import get_db
from .auth import get_current_active_admin


router = APIRouter(prefix="/admin", tags=["admin"])


@router.get("/stats", response_model=schemas.AdminStats)
def get_stats(
    db: Session = Depends(get_db),
    admin: models.User = Depends(get_current_active_admin)
):
    total_users = db.query(models.User).count()
    total_projects = db.query(models.Project).count()
    total_reviews = db.query(models.Review).count()
    
    today = datetime.utcnow().date()
    days_back = 30
    
    users_by_day = []
    projects_by_day = []
    
    for i in range(days_back):
        day = today - timedelta(days=days_back - i - 1)
        day_start = datetime.combine(day, datetime.min.time())
        day_end = datetime.combine(day, datetime.max.time())
        
        users_count = db.query(models.User).filter(
            models.User.created_at >= day_start,
            models.User.created_at <= day_end
        ).count()
        
        projects_count = db.query(models.Project).filter(
            models.Project.created_at >= day_start,
            models.Project.created_at <= day_end
        ).count()
        
        users_by_day.append({"date": str(day), "count": users_count})
        projects_by_day.append({"date": str(day), "count": projects_count})
    
    return schemas.AdminStats(
        total_users=total_users,
        total_projects=total_projects,
        total_reviews=total_reviews,
        users_by_day=users_by_day,
        projects_by_day=projects_by_day
    )


@router.get("/users", response_model=List[schemas.UserResponse])
def get_all_users(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
    admin: models.User = Depends(get_current_active_admin)
):
    users = db.query(models.User).offset(skip).limit(limit).all()
    return users


@router.get("/users/{user_id}", response_model=schemas.UserResponse)
def get_user(
    user_id: int,
    db: Session = Depends(get_db),
    admin: models.User = Depends(get_current_active_admin)
):
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user


@router.delete("/users/{user_id}")
def delete_user(
    user_id: int,
    db: Session = Depends(get_db),
    admin: models.User = Depends(get_current_active_admin)
):
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    db.delete(user)
    db.commit()
    return {"message": "User deleted"}


@router.post("/users/{user_id}/toggle-active")
def toggle_user_active(
    user_id: int,
    db: Session = Depends(get_db),
    admin: models.User = Depends(get_current_active_admin)
):
    user = db.query(models.User).filter(models.User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    
    user.is_active = not user.is_active
    db.commit()
    return {"message": f"User is now {'active' if user.is_active else 'inactive'}"}


from fastapi import HTTPException