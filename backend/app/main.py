from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from .database import engine, Base
from .routes import auth, users, projects, reviews, admin


def create_tables():
    Base.metadata.create_all(bind=engine)


app = FastAPI(
    title="TelemetryAI API",
    description="TelemetryAI - AI-powered telemetry platform",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

create_tables()

app.include_router(auth.router, prefix="/api/v1")
app.include_router(users.router, prefix="/api/v1")
app.include_router(projects.router, prefix="/api/v1")
app.include_router(reviews.router, prefix="/api/v1")
app.include_router(admin.router, prefix="/api/v1")


@app.get("/")
def root():
    return {"message": "TelemetryAI API", "version": "1.0.0"}


@app.get("/health")
def health():
    return {"status": "healthy"}