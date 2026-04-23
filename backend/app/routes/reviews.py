from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from . import models, schemas
from .database import get_db
from .auth import get_current_user


router = APIRouter(prefix="/reviews", tags=["reviews"])


@router.post("", response_model=schemas.ReviewResponse)
def create_review(
    review: schemas.ReviewCreate,
    current_user: models.User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    project = db.query(models.Project).filter(models.Project.id == review.project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="Project not found")
    
    db_review = models.Review(
        rating=review.rating,
        content=review.content,
        user_id=current_user.id,
        project_id=review.project_id
    )
    db.add(db_review)
    db.commit()
    db.refresh(db_review)
    return db_review


@router.get("/project/{project_id}", response_model=List[schemas.ReviewResponse])
def get_project_reviews(
    project_id: int,
    db: Session = Depends(get_db)
):
    reviews = db.query(models.Review).filter(models.Review.project_id == project_id).all()
    for review in reviews:
        review.user  # Load user
    return reviews


@router.put("/{review_id}", response_model=schemas.ReviewResponse)
def update_review(
    review_id: int,
    review_update: schemas.ReviewUpdate,
    current_user: models.User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    review = db.query(models.Review).filter(
        models.Review.id == review_id,
        models.Review.user_id == current_user.id
    ).first()
    if not review:
        raise HTTPException(status_code=404, detail="Review not found")
    
    if review_update.rating is not None:
        review.rating = review_update.rating
    if review_update.content is not None:
        review.content = review_update.content
    
    db.commit()
    db.refresh(review)
    return review


@router.delete("/{review_id}")
def delete_review(
    review_id: int,
    current_user: models.User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    review = db.query(models.Review).filter(
        models.Review.id == review_id,
        models.Review.user_id == current_user.id
    ).first()
    if not review:
        raise HTTPException(status_code=404, detail="Review not found")
    
    db.delete(review)
    db.commit()
    return {"message": "Review deleted"}